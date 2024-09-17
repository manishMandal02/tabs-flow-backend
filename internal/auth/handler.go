package auth

import (
	"encoding/json"
	"strings"

	lambda_events "github.com/aws/aws-lambda-go/events"
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
		TTL_Expiry: config.OTP_EXPIRY_TIME_IN_MIN,
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

	err = sqs.AddMsgToQueue(event)

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Success: false, Message: "Error sending OTP"})
	}

	return http_api.APIResponse(200, http_api.RespBody{Success: true, Message: "OTP sent successfully"})
}

func (h *AuthHandler) verifyOTP(body, userAgent string) *lambda_events.APIGatewayV2HTTPResponse {
	// TODO - handle verify otp

	ua := useragent.New(userAgent)

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

	// TODO - handle creating user session (session table) and user id in main table
	return http_api.APIResponse(200, http_api.RespBody{Success: true, Message: "OTP verified successfully"})
}

func (h *AuthHandler) googleAuth(body, userAgent string) *lambda_events.APIGatewayV2HTTPResponse {

	ua := useragent.New(userAgent)

	// TODO - handle google auth
	var b struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(strings.NewReader(body))
	err := decoder.Decode(&b)
	if err != nil {
		logger.Error("Error decoding request body for google auth", err)
		return http_api.APIResponse(400, http_api.RespBody{Success: false, Message: errMsg.googleAuth})
	}

	return http_api.APIResponse(200, http_api.RespBody{Success: true, Message: "Login successful"})
}

func (h *AuthHandler) generatedToken() error {
	return nil
}

func (h *AuthHandler) validateToken() error {
	return nil
}

func (h *AuthHandler) generateSession(email, userAgent string) error {
	sId := utils.GenerateRandomString(20)

	ua := useragent.New(userAgent)

	browser, _ := ua.Browser()

	deviceInfo := &DeviceInfo{
		Browser:  browser,
		OS:       ua.OS(),
		Platform: ua.Platform(),
		IsMobile: ua.Mobile(),
	}
	session := &session{
		Id:         sId,
		Email:      email,
		TTL_Expiry: 10,
		DeviceInfo: deviceInfo,
	}

	return nil
}

func (h *AuthHandler) logout() *lambda_events.APIGatewayV2HTTPResponse {
	// TODO - remove jwt token & delete session
	return nil
}

func (h *AuthHandler) lambdaAuthorizer(ev *lambda_events.APIGatewayCustomAuthorizerRequest) *lambda_events.APIGatewayCustomAuthorizerResponse {
	token := ev.AuthorizationToken

	return nil
}
