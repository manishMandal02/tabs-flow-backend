package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	lambda_events "github.com/aws/aws-lambda-go/events"

	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

// if userId not found in Session table, add user profile (U#Profile) to main table

type authHandler struct {
	r authRepository
}

func newAuthHandler(repo authRepository) *authHandler {
	return &authHandler{
		r: repo,
	}
}

func (h *authHandler) sendOTP(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Email string `json:"email"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&b)

	if err != nil {
		logger.Error("Error decoding request body for sendOTP", err)
		http.Error(w, errMsg.sendOTP, http.StatusBadRequest)
		return
	}

	otp := utils.GenerateOTP()

	err = h.r.saveOTP(&emailOTP{
		OTP:   otp,
		Email: b.Email,
		TTL:   time.Now().Add(time.Minute * time.Duration(config.OTP_EXPIRY_TIME_IN_MIN)).Unix(),
	})

	if err != nil {
		http.Error(w, errMsg.sendOTP, http.StatusBadGateway)
		return
	}

	// send email message to SQS queue
	// send USER_REGISTERED event to email service (queue)
	event := events.New(events.EventTypeUserRegistered, &events.SendOTPPayload{
		Email: b.Email,
		OTP:   otp,
	})

	sqs := events.NewEmailQueue()

	err = sqs.AddMessage(event)

	if err != nil {
		http.Error(w, errMsg.sendOTP, http.StatusBadGateway)
		return
	}

	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Message: "OTP sent successfully"})
}

func (h *authHandler) verifyOTP(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	err := json.NewDecoder(r.Body).Decode(&b)

	userAgent := r.Header.Get("User-Agent")

	if err != nil {
		logger.Error("Error decoding request body for verify otp", err)
		http.Error(w, errMsg.validateOTP, http.StatusBadRequest)
		return
	}

	valid, err := h.r.validateOTP(b.Email, b.OTP)

	if err != nil {
		http.Error(w, errMsg.validateOTP, http.StatusBadGateway)
		return
	}

	logger.Dev(fmt.Sprintf("OTP valid: %v", valid))

	if !valid {
		http.Error(w, errMsg.inValidOTP, http.StatusBadRequest)
		return
	}

	// create new session and set to cookie
	res, err := createNewSession(b.Email, userAgent, h.r)

	if err != nil {
		http.Error(w, errMsg.createSession, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, res.cookie)
	http_api.SuccessResMsgWithBody(w, "OTP verified successfully", res.data)
}

func (h *authHandler) googleAuth(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Email string `json:"email"`
	}

	userAgent := r.Header.Get("User-Agent")

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&b)
	if err != nil {
		logger.Error("Error decoding request body for google auth", err)
		http.Error(w, errMsg.googleAuth, http.StatusBadRequest)
		return
	}

	// create new session and set to cookie
	res, err := createNewSession(b.Email, userAgent, h.r)

	if err != nil {
		logger.Error(errMsg.createSession, errors.New(errMsg.createSession))
		return
	}

	http.SetCookie(w, res.cookie)
	http.SetCookie(w, res.cookie)
	http_api.SuccessResMsgWithBody(w, "OTP verified successfully", res.data)
}

func (h *authHandler) getUserId(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&b)

	if err != nil {
		logger.Error("Error decoding request body for getUserId", err)
		http.Error(w, errMsg.getUserId, http.StatusBadRequest)
	}
	userId, err := h.r.userIdByEmail(b.Email)

	if err != nil {
		http.Error(w, errMsg.getUserId, http.StatusBadRequest)
		return
	}

	if userId == "" {
		http.Error(w, "No user with given email", http.StatusBadRequest)
		return
	}

	resData := &struct {
		UserId string `json:"userId"`
	}{
		UserId: userId,
	}

	http_api.SuccessResData(w, resData)
}

func (h *authHandler) logout(w http.ResponseWriter, r *http.Request) {

	logoutResponse := func() {
		cookie := &http.Cookie{
			Name:     "access_token",
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			MaxAge:   -1,
		}

		http.SetCookie(w, cookie)

		http_api.SuccessResMsg(w, "logged out successfully")
	}

	c, err := r.Cookie("access_token")

	if err != nil {
		logoutResponse()
		return
	}

	claims, err := ValidateToken(c.Value)

	if err != nil {
		logger.Error(errMsg.validateSession, err)
		logoutResponse()
		return
	}

	email, okEmail := claims["sub"].(string)
	sId, okSID := claims["session_id"].(string)

	if !okEmail || !okSID {
		logger.Error(errMsg.validateSession, errors.New(errMsg.invalidToken))
		logoutResponse()
		return
	}

	err = h.r.deleteSession(email, sId)

	if err != nil {
		logger.Error(errMsg.deleteSession, err)
		logoutResponse()
		return
	}

	logoutResponse()
}

func (h *authHandler) lambdaAuthorizer(ev *lambda_events.APIGatewayCustomAuthorizerRequestTypeRequest) (*lambda_events.APIGatewayCustomAuthorizerResponse, error) {
	cookies := parseCookiesStr(ev.Headers["Cookie"])

	claims, err := ValidateToken(cookies["access_token"])

	if err != nil {
		logger.Error("Error validating JWT token", errors.New(errMsg.invalidToken))

		return nil, errors.New("Unauthorized")
	}

	email, emailOK := claims["sub"].(string)
	userId, userIdOK := claims["user_id"].(string)
	sId, sIdOK := claims["session_id"].(string)
	expiryTime, expiryOK := claims["exp"].(float64)

	logger.Dev("emailOK: %v", emailOK)
	logger.Dev("sIdOK: %v", sIdOK)
	logger.Dev("expiryOK: %v", expiryOK)

	if !emailOK || !sIdOK || !expiryOK || !userIdOK {
		logger.Error("Error getting token claims", errors.New(errMsg.invalidToken))
		return nil, errors.New("Unauthorized")
	}

	if int64(expiryTime) > time.Now().Unix() {
		// token valid, allow access
		return generatePolicy(ev.MethodArn, "Allow", ev.MethodArn, userId, nil), nil
	}

	// validate session
	isValid, err := h.r.validateSession(email, sId)

	if err != nil {
		logger.Error("Error validating session", errors.New(errMsg.validateSession))
		return nil, errors.New("Unauthorized")
	}

	// if session, valid then refresh token and allow access
	if !isValid {
		logger.Error("Error validating session", errors.New(errMsg.validateSession))
		return nil, errors.New("Unauthorized")
	}

	res, err := createNewSession(email, ev.Headers["User-Agent"], h.r)

	if err != nil {
		return nil, errors.New("Unauthorized")
	}

	newCookies := map[string]string{
		"access_token": res.token,
	}

	return generatePolicy("user", "Allow", ev.MethodArn, userId, newCookies), nil
}
