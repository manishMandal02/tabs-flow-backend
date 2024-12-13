package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
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
	r          authRepository
	emailQueue *events.Queue
}

func newAuthHandler(repo authRepository, q *events.Queue) *authHandler {
	return &authHandler{
		r:          repo,
		emailQueue: q,
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
		http_api.ErrorRes(w, errMsg.sendOTP, http.StatusBadRequest)
		return
	}

	otp := utils.GenerateOTP()

	err = h.r.saveOTP(&emailOTP{
		OTP:   otp,
		Email: b.Email,
		TTL:   time.Now().Add(time.Minute * time.Duration(config.OTP_EXPIRY_TIME_IN_MIN)).Unix(),
	})

	if err != nil {
		http_api.ErrorRes(w, errMsg.sendOTP, http.StatusBadGateway)
		return
	}

	// send email message to SQS queue
	// send USER_REGISTERED event to email service (queue)
	event := events.New(events.EventTypeSendOTP, &events.SendOTPPayload{
		Email: b.Email,
		OTP:   otp,
	})

	err = h.emailQueue.AddMessage(event)

	if err != nil {
		http_api.ErrorRes(w, errMsg.sendOTP, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "OTP sent successfully")

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
		http_api.ErrorRes(w, errMsg.validateOTP, http.StatusBadRequest)
		return
	}

	valid, err := h.r.validateOTP(b.Email, b.OTP)

	if err != nil {
		if err.Error() == errMsg.expiredOTP {
			http_api.ErrorRes(w, errMsg.expiredOTP, http.StatusBadRequest)
			return
		}
		http_api.ErrorRes(w, errMsg.validateOTP, http.StatusBadGateway)
		return
	}

	if !valid {
		http_api.ErrorRes(w, errMsg.inValidOTP, http.StatusBadRequest)
		return
	}

	// check if user exists
	resData, err := checkIfNewUser(b.Email, h.r)

	if err != nil {
		http_api.ErrorRes(w, errMsg.googleAuth, http.StatusBadGateway)
		return
	}

	// create new session
	cookie, err := createNewSession(resData.UserId, userAgent, h.r)

	if err != nil {
		http_api.ErrorRes(w, errMsg.createSession, http.StatusInternalServerError)
		return
	}

	// set session cookie
	http.SetCookie(w, cookie)

	http_api.SuccessResMsgWithBody(w, "OTP verified successfully", resData)
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
		http_api.ErrorRes(w, errMsg.googleAuth, http.StatusBadRequest)
		return
	}

	// check if user exists
	resData, err := checkIfNewUser(b.Email, h.r)

	if err != nil {
		http_api.ErrorRes(w, errMsg.googleAuth, http.StatusBadGateway)
		return
	}

	// create new session
	cookie, err := createNewSession(resData.UserId, userAgent, h.r)

	if err != nil {
		http_api.ErrorRes(w, errMsg.createSession, http.StatusInternalServerError)
		return
	}

	// set session cookie
	http.SetCookie(w, cookie)

	http_api.SuccessResMsgWithBody(w, "OTP verified successfully", resData)
}

func (h *authHandler) getUserId(w http.ResponseWriter, r *http.Request) {

	email := r.PathValue("email")

	userId, err := h.r.userIdByEmail(email)

	if err != nil {
		http_api.ErrorRes(w, errMsg.getUserId, http.StatusBadRequest)
		return
	}

	if userId == "" {
		http_api.ErrorRes(w, "No user with given email", http.StatusBadRequest)
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
			Name:     "session",
			Value:    "",
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			MaxAge:   -1,
			Domain:   config.APP_DOMAIN_NAME,
		}

		http.SetCookie(w, cookie)

		http_api.SuccessResMsg(w, "logged out successfully")
	}

	c, err := r.Cookie("session")

	if err != nil {
		logoutResponse()
		return
	}

	sId, userId, err := GetSessionValues(c.Value)

	if err != nil {
		logger.Error(errMsg.ValidateSession, err)
		logoutResponse()
		return
	}

	if sId == "" || userId == "" {
		logger.Error(errMsg.ValidateSession, errors.New(errMsg.invalidSessionValue))
		logoutResponse()
		return
	}

	err = h.r.deleteSession(userId, sId)

	if err != nil {
		logger.Error(errMsg.deleteSession, err)
		logoutResponse()
		return
	}

	logoutResponse()
}

func (h *authHandler) lambdaAuthorizer(ev *lambda_events.APIGatewayCustomAuthorizerRequestTypeRequest) (*lambda_events.APIGatewayCustomAuthorizerResponse, error) {

	// allow paddle webhook url, without auth tokens
	if strings.Contains(ev.Path, "/users/subscription/webhook") {
		//  allow access
		return generatePolicy("paddle-webhook", "Allow", ev.MethodArn, "", nil), nil
	}

	cookieHeader := ev.Headers["Cookie"]

	if ev.Headers["Cookie"] == "" {
		cookieHeader = ev.Headers["cookie"]
	}

	cookies := parseCookiesStrToMap(cookieHeader)

	if len(cookies) == 0 {
		logger.Error("No cookies found in header", errors.New(errMsg.invalidSessionValue))
		return nil, errors.New("Unauthorized")
	}

	sId, userId, err := GetSessionValues(cookies["session"])

	if err != nil {
		logger.Error("Error getting session values from session cookie", errors.New(errMsg.invalidSessionValue))
		return nil, errors.New("Unauthorized")
	}

	// validate session
	isValid, err := h.r.ValidateSession(userId, sId)

	if err != nil {
		logger.Error("Error validating session", errors.New(errMsg.ValidateSession))
		return nil, errors.New("Unauthorized")
	}

	// if session, valid then refresh token and allow access
	if !isValid {
		logger.Error("Error validating session", errors.New(errMsg.ValidateSession))
		return nil, errors.New("Unauthorized")
	}

	cookie, err := createNewSession(userId, ev.Headers["User-Agent"], h.r)

	if err != nil {
		return nil, errors.New("Unauthorized")
	}

	newCookies := map[string]string{
		"session": cookie.Value,
	}

	return generatePolicy(userId, "Allow", ev.MethodArn, userId, newCookies), nil
}
