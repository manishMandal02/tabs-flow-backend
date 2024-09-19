package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	lambda_events "github.com/aws/aws-lambda-go/events"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
	"github.com/mssola/useragent"
)

// if userId not found in Session table, add user profile (U#Profile) to main table

type AuthHandler struct {
	r authRepository
}

func newAuthHandler(repo authRepository) *AuthHandler {
	return &AuthHandler{
		r: repo,
	}
}

func (h *AuthHandler) sendOTP(body string) *lambda_events.APIGatewayV2HTTPResponse {
	var b struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(strings.NewReader(body))
	decoder.Decode(&b)

	otp := utils.GenerateOTP()

	err := h.r.saveOTP(&emailOTP{
		OTP:        otp,
		Email:      b.Email,
		TTL_Expiry: time.Now().Add(time.Minute * time.Duration(config.OTP_EXPIRY_TIME_IN_MIN)).Unix(),
	})

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Success: false, Message: "Error sending OTP"})
	}

	// send email message to SQS queue
	event := &events.SendOTP_Payload{
		Email: b.Email,
		OTP:   otp,
	}

	sqs := events.NewQueue()

	err = sqs.AddMessage(event)

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Success: false, Message: "Error sending OTP"})
	}

	return http_api.APIResponse(200, http_api.RespBody{Success: true, Message: "OTP sent successfully"})
}

func (h *AuthHandler) verifyOTP(body, userAgent string) *lambda_events.APIGatewayV2HTTPResponse {
	var b struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}
	decoder := json.NewDecoder(strings.NewReader(body))
	err := decoder.Decode(&b)
	if err != nil {
		logger.Error("Error decoding request body for verify otp", err)
		return http_api.APIResponse(400, http_api.RespBody{Success: false, Message: errMsg.validateOTP})
	}

	valid, err := h.r.validateOTP(b.Email, b.OTP)

	if err != nil {
		return http_api.APIResponse(400, http_api.RespBody{Success: false, Message: errMsg.validateOTP})
	}

	if !valid {
		return http_api.APIResponse(400, http_api.RespBody{Success: false, Message: errMsg.inValidOTP})
	}

	// create new session and set to cookie
	newToken, err := createNewSession(b.Email, userAgent, h.r)

	if err != nil {
		return http_api.APIResponse(400, http_api.RespBody{Success: false, Message: errMsg.createSession})
	}

	newCookies := map[string]string{
		"access_token": newToken,
	}

	//  check if user exits
	userId, err := h.r.userIdByEmail(b.Email)

	type respData struct {
		UserId  string `json:"userId"`
		NewUser bool   `json:"isNewUser"`
	}

	if err != nil {
		// new user
		newUserId := utils.GenerateID()

		err = h.r.attachUserId(&emailWithUserId{
			Email:  b.Email,
			UserId: newUserId,
		})

		if err != nil {
			return http_api.APIResponse(400, http_api.RespBody{Success: false, Message: errMsg.createSession})
		}

		resData := &respData{
			UserId:  newUserId,
			NewUser: true,
		}

		return http_api.APIResponseWithCookies(200, http_api.RespBody{Success: true, Message: "OTP verified successfully", Data: resData}, newCookies)
	}

	// old user
	resData := &respData{
		UserId:  userId,
		NewUser: false,
	}

	return http_api.APIResponseWithCookies(200, http_api.RespBody{Success: true, Message: "OTP verified successfully", Data: resData}, newCookies)
}

func (h *AuthHandler) googleAuth(body, userAgent string) *lambda_events.APIGatewayV2HTTPResponse {
	var b struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(strings.NewReader(body))

	err := decoder.Decode(&b)
	if err != nil {
		logger.Error("Error decoding request body for google auth", err)
		return http_api.APIResponse(400, http_api.RespBody{Success: false, Message: errMsg.googleAuth})
	}

	// create new session and set to cookie
	newToken, err := createNewSession(b.Email, userAgent, h.r)

	if err != nil {
		return http_api.APIResponse(400, http_api.RespBody{Success: false, Message: errMsg.createSession})
	}

	newCookies := map[string]string{
		"access_token": newToken,
	}

	//  check if user exits
	userId, err := h.r.userIdByEmail(b.Email)

	type respData struct {
		UserId  string `json:"userId"`
		NewUser bool   `json:"isNewUser"`
	}

	if err != nil {
		// new user
		newUserId := utils.GenerateID()

		err = h.r.attachUserId(&emailWithUserId{
			Email:  b.Email,
			UserId: newUserId,
		})

		if err != nil {
			return http_api.APIResponse(400, http_api.RespBody{Success: false, Message: errMsg.createSession})
		}

		resData := &respData{
			UserId:  newUserId,
			NewUser: true,
		}

		return http_api.APIResponseWithCookies(200, http_api.RespBody{Success: true, Message: "Google auth successful", Data: resData}, newCookies)
	}

	// old user
	resData := &respData{
		UserId:  userId,
		NewUser: false,
	}

	return http_api.APIResponseWithCookies(200, http_api.RespBody{Success: true, Message: "Google auth successful", Data: resData}, newCookies)

}

func (h *AuthHandler) logout(cookieStr []string) *lambda_events.APIGatewayV2HTTPResponse {

	newCookies := map[string]string{
		"access_token": "",
	}

	logoutResponse := http_api.APIResponseWithCookies(200, http_api.RespBody{Success: true, Message: "Logged out"}, newCookies)

	cookies := parseCookiesPair(cookieStr)

	token := cookies["access_token"]

	claims, err := validateToken(token)

	if err != nil {
		logger.Error(errMsg.validateSession, err)
		return logoutResponse
	}

	email, okEmail := claims["email"].(string)
	sId, okSID := claims["session_id"].(string)

	if !okEmail || !okSID {
		logger.Error(errMsg.validateSession, errors.New(errMsg.invalidToken))
		return logoutResponse
	}

	err = h.r.deleteSession(email, sId)

	if err != nil {
		logger.Error(errMsg.deleteSession, err)
		return logoutResponse
	}

	return logoutResponse
}

func (h *AuthHandler) lambdaAuthorizer(ev *lambda_events.APIGatewayCustomAuthorizerRequestTypeRequest) (lambda_events.APIGatewayCustomAuthorizerResponse, error) {
	cookies := parseCookiesStr(ev.Headers["Cookie"])

	claims, err := validateToken(cookies["access_token"])

	if err != nil {
		logger.Error("Error validating JWT token", errors.New(errMsg.invalidToken))
		return lambda_events.APIGatewayCustomAuthorizerResponse{}, errors.New(errMsg.invalidToken)
	}

	email, emailOK := claims["email"].(string)
	sId, sIdOK := claims["session_id"].(string)
	expiryTime, expiryOK := claims["exp"].(int64)

	if !emailOK || !sIdOK || !expiryOK {
		logger.Error("Error getting token claims", errors.New(errMsg.invalidToken))
		return lambda_events.APIGatewayCustomAuthorizerResponse{}, errors.New(errMsg.invalidToken)
	}

	if expiryTime > time.Now().Unix() {
		// token valid, allow access
		return generatePolicy(ev.MethodArn, "Allow", ev.MethodArn, nil), nil
	}

	// validate session
	isValid, err := h.r.validateSession(email, sId)

	if err != nil {
		logger.Error("Error validating session", errors.New(errMsg.validateSession))
		return lambda_events.APIGatewayCustomAuthorizerResponse{}, errors.New(errMsg.validateSession)
	}

	// if session, valid then refresh token and allow access
	if !isValid {
		logger.Error("Error validating session", errors.New(errMsg.validateSession))
		return lambda_events.APIGatewayCustomAuthorizerResponse{}, errors.New(errMsg.validateSession)
	}

	newToken, err := createNewSession(email, ev.Headers["User-Agent"], h.r)

	if err != nil {
		return lambda_events.APIGatewayCustomAuthorizerResponse{}, errors.New(errMsg.createToken)
	}

	newCookies := map[string]string{
		"access_token": newToken,
	}

	return generatePolicy("user", "Allow", ev.MethodArn, newCookies), nil
}

// helpers
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

	tokenStr, err := claims.SignedString(config.JWT_SECRET_KEY)

	if err != nil {
		logger.Error("Error generating JWT token", err)
		return "", err
	}

	return tokenStr, nil
}

// validate jwt token
func validateToken(tokenStr string) (jwt.MapClaims, error) {

	token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (interface{}, error) {
		return config.JWT_SECRET_KEY, nil
	})

	if err != nil {
		logger.Error("Error parsing JWT token", err)
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// get session id from token
	claims, _ := token.Claims.(jwt.MapClaims)

	return claims, nil
}

// generate authorizer policy
func generatePolicy(principalId, effect, resource string, cookies map[string]string) lambda_events.APIGatewayCustomAuthorizerResponse {
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

	return authResponse
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

func parseCookiesPair(cookiePairs []string) map[string]string {
	cookies := make(map[string]string)
	if len(cookiePairs) < 1 {
		return cookies
	}

	for _, pair := range cookiePairs {
		parts := strings.SplitN(strings.TrimSpace(pair), "=", 2)
		if len(parts) == 2 {
			cookies[parts[0]] = parts[1]
		}
	}
	return cookies
}

func createNewSession(email, userAgent string, aR authRepository) (string, error) {
	ua := useragent.New(userAgent)

	browser, _ := ua.Browser()

	session := session{
		Email:      email,
		Id:         utils.GenerateRandomString(20),
		TTL_Expiry: time.Now().AddDate(0, 0, config.USER_SESSION_EXPIRY_DAYS*3).Unix(),
		DeviceInfo: &DeviceInfo{
			Browser:  browser,
			OS:       ua.OS(),
			Platform: ua.Platform(),
			IsMobile: ua.Mobile(),
		},
	}
	err := aR.createSession(&session)

	if err != nil {
		logger.Error("Error creating session", errors.New(errMsg.createSession))
		return "", err
	}

	newToken, err := generateToken(email, session.Id)

	if err != nil {
		logger.Error("Error creating token", errors.New(errMsg.createToken))
		return "", err
	}

	return newToken, nil
}

// helper
