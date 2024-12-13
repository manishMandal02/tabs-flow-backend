package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	lambda_events "github.com/aws/aws-lambda-go/events"
	"github.com/mssola/useragent"

	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

// * helpers
// validate jwt token
func GetSessionValues(cookieValue string) (string, string, error) {

	//cookieValue = id={sessionId}//uid={userId}

	sessionValues := strings.Split(cookieValue, "//")

	if len(sessionValues) != 2 {
		return "", "", errors.New("invalid session value")
	}

	sessionId := strings.Split(sessionValues[0], "=")[1]

	userId := strings.Split(sessionValues[1], "=")[1]

	if sessionId == "" || userId == "" {
		return "", "", errors.New("invalid session value")
	}

	return sessionId, userId, nil
}

// parse cookie
func parseCookiesStrToMap(cookieHeader string) map[string]string {

	cookies := make(map[string]string)
	if cookieHeader == "" {
		return cookies
	}

	pairs := strings.Split(cookieHeader, ";")

	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			cookies[parts[0]] = parts[1]
		}
	}

	return cookies
}

type checkNewUserRes = struct {
	UserId  string `json:"userId"`
	NewUser bool   `json:"isNewUser"`
}

func checkIfNewUser(email string, aR authRepository) (*checkNewUserRes, error) {
	userId, err := aR.userIdByEmail(email)

	var res checkNewUserRes

	if err != nil || userId == "" {
		// new user
		newUserId := utils.GenerateID()

		err = aR.attachUserId(&emailWithUserId{
			Email:  email,
			UserId: newUserId,
		})

		if err != nil {
			return nil, err
		}

		res.NewUser = true
		res.UserId = newUserId
	} else {
		// old user
		res.NewUser = false
		res.UserId = userId
	}

	return &res, nil
}

func createNewSession(userId, userAgent string, aR authRepository) (*http.Cookie, error) {

	ua := useragent.New(userAgent)

	browser, _ := ua.Browser()

	sId := utils.GenerateID()

	session := session{
		UserId: userId,
		Id:     sId,
		TTL:    time.Now().AddDate(0, 0, config.USER_SESSION_EXPIRY_DAYS).Unix(),
		DeviceInfo: &deviceInfo{
			Browser:  browser,
			OS:       ua.OS(),
			Platform: ua.Platform(),
			IsMobile: ua.Mobile(),
		},
	}

	err := aR.createSession(&session)

	if err != nil {
		logger.Error(errMsg.createSession, err)
		return nil, err
	}

	sValue := fmt.Sprintf("id=%s//uid=%s", sId, userId)

	cookie := &http.Cookie{
		Name:     "session",
		Value:    sValue,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		Domain:   config.APP_DOMAIN_NAME,
		SameSite: http.SameSiteNoneMode,
	}

	return cookie, nil
}

// generate policy for lambda authorizer
func generatePolicy(principalId, effect, methodArn, userId string, cookies map[string]string) *lambda_events.APIGatewayCustomAuthorizerResponse {

	// remove the path and method from the arn, so it allows all the path and method even with cached data
	arnParts := strings.Split(methodArn, ":")
	apiGatewayArnTmp := strings.Split(arnParts[5], "/")
	apiID := apiGatewayArnTmp[0] // This gets just the API ID
	wildcardArn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/*",
		arnParts[3],         // region
		arnParts[4],         // account ID
		apiID,               // API ID
		apiGatewayArnTmp[1], // stage
	)

	authResponse := lambda_events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalId}

	if effect != "" && methodArn != "" {
		authResponse.PolicyDocument = lambda_events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []lambda_events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{wildcardArn},
				},
			},
		}
	}

	if cookies != nil {
		cookieStrings := make([]string, 0, len(cookies))
		for key, value := range cookies {
			cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s; HttpOnly; Secure; SameSite=Strict; Expires:%v", key, value, time.Now().AddDate(0, 0, config.JWT_TOKEN_EXPIRY_IN_DAYS*3).Format(time.RFC1123)))
		}
		authResponse.Context = map[string]interface{}{
			"Set-Cookie": strings.Join(cookieStrings, ", "),
		}
	}

	if userId != "" {
		if authResponse.Context == nil {
			authResponse.Context = map[string]interface{}{}
		}
		authResponse.Context["UserId"] = userId
	}

	if effect != "Allow" {

		if authResponse.Context == nil {
			authResponse.Context = map[string]interface{}{}
		}
		authResponse.Context["code"] = "401"
		authResponse.Context["message"] = "Unauthorized"
	}

	logger.Info("authorizer response: %v", authResponse)

	return &authResponse
}
