package auth

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
	"github.com/mssola/useragent"
)

//* Handlers
// TODO - Send OTP: send SQS message with email & otp
// TODO - Verify OTP:
// TODO - Generate JWT Token
// TODO - verify JWT Token:
// if userId not found in Session table, add user profile (U#Profile) to main table

type AuthHandler struct {
	r authRepository
}

func newAuthHandler(repo authRepository) *AuthHandler {
	return &AuthHandler{
		r: repo,
	}
}

func (h *AuthHandler) sendOTP(body string) *events.APIGatewayProxyResponse {
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

	// TODO -

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Success: false, Message: "Error sending OTP"})
	}

	return http_api.APIResponse(200, http_api.RespBody{Success: true, Message: "OTP sent successfully"})
}

func (h *AuthHandler) VerifyOTP(opt int32) error {
	return nil
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

// remove jwt token
func (h *AuthHandler) logout() error {
	return nil
}
