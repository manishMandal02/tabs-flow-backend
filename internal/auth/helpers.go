package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	lambda_events "github.com/aws/aws-lambda-go/events"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/mssola/useragent"

	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

// * helpers
// generate new token
func generateToken(email, sessionId string) (string, error) {
	// Create a new JWT token with claims
	claims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		// Subject (user identifier)
		"sub": email,
		// Audience (user role)
		"session_id": sessionId,
		// Issuer
		"iss": "tabsflow-app",
		// Expiration time
		"exp": time.Now().AddDate(0, 0, config.JWT_TOKEN_EXPIRY_IN_DAYS).Unix(),
		// Issued at
		"iat": time.Now().Unix(),
	})

	tokenStr, err := claims.SignedString([]byte(config.JWT_SECRET_KEY))

	if err != nil {
		logger.Error("Error generating JWT token", err)
		return "", err
	}

	return tokenStr, nil
}

// validate jwt token
func validateToken(tokenStr string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (interface{}, error) {
		return []byte(config.JWT_SECRET_KEY), nil
	})

	if err != nil {
		logger.Error("Error parsing JWT token", err)
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	logger.Dev("token claims: %v", token.Claims.(jwt.MapClaims))

	return token.Claims.(jwt.MapClaims), nil
}

// parse cookie
func parseCookiesStr(cookieHeader string) map[string]string {
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

type createSessionRes struct {
	token  string
	cookie *http.Cookie
	data   struct {
		UserId  string `json:"userId"`
		NewUser bool   `json:"isNewUser"`
	}
}

func createNewSession(email, userAgent string, aR authRepository, isAuthorizer bool) (*createSessionRes, error) {
	ua := useragent.New(userAgent)

	browser, _ := ua.Browser()

	session := session{
		Email: email,
		Id:    utils.GenerateRandomString(20),
		TTL:   time.Now().AddDate(0, 0, config.USER_SESSION_EXPIRY_DAYS*3).Unix(),
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

	newToken, err := generateToken(email, session.Id)

	if err != nil {
		logger.Error(errMsg.createToken, err)
		return nil, err
	}

	if isAuthorizer {
		// if authorizer, return token only
		return &createSessionRes{token: newToken}, nil
	}

	cookie := &http.Cookie{
		Name:     "access_token",
		Value:    newToken,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	}

	//  check if user exits
	userId, err := aR.userIdByEmail(email)

	type respData struct {
		UserId  string `json:"userId"`
		NewUser bool   `json:"isNewUser"`
	}

	resData := &respData{}

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

		resData = &respData{
			UserId:  newUserId,
			NewUser: true,
		}
	} else {
		// old user
		resData = &respData{
			UserId:  userId,
			NewUser: false,
		}

	}

	return &createSessionRes{
		cookie: cookie,
		data:   *resData,
	}, nil
}

// generate policy for lambda authorizer
func generatePolicy(principalId, effect, resource string, cookies map[string]string) *lambda_events.APIGatewayCustomAuthorizerResponse {
	authResponse := lambda_events.APIGatewayCustomAuthorizerResponse{PrincipalID: principalId}

	if effect != "" && resource != "" {
		authResponse.PolicyDocument = lambda_events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []lambda_events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{resource},
				},
			},
		}
	}

	if cookies != nil {
		cookieStrings := make([]string, 0, len(cookies))
		for key, value := range cookies {
			cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s; HttpOnly; Secure; SameSite=Strict", key, value))
		}
		authResponse.Context = map[string]interface{}{
			"Set-Cookie": strings.Join(cookieStrings, ", "),
		}
	}

	if effect != "Allow" {

		if authResponse.Context == nil {
			authResponse.Context = map[string]interface{}{}
		}
		authResponse.Context["code"] = "401"
		authResponse.Context["message"] = "Unauthorized"
	}

	logger.Dev("authorizer response: %v", authResponse)

	return &authResponse
}

func checkSubscriptionStatus(userId, urlHost string) (bool, error) {
	p := "https"

	if config.LOCAL_DEV_ENV {
		p = "http"
	}

	authServiceURL := fmt.Sprintf("%s://%s/users/%s/subscription/status", p, urlHost, userId)

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	res, respBody, err := utils.MakeHTTPRequest(http.MethodGet, authServiceURL, headers, nil)

	if err != nil {
		logger.Error("Error checking subscription status", err)
		return false, err
	}

	if res.StatusCode != http.StatusOK {
		logger.Error("Error checking subscription status", errors.New(respBody))
		return false, errors.New(respBody)
	}

	var sub struct {
		Data struct {
			Active bool `json:"active"`
		} `json:"data"`
	}

	err = json.Unmarshal([]byte(respBody), &sub)

	if err != nil {
		logger.Error("Error unmarshaling subscription status", err)
		return false, err
	}

	if !sub.Data.Active {
		return false, nil
	}

	return true, nil
}
