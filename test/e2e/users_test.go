package e2e_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
	"github.com/stretchr/testify/suite"
)

func (s *UserFlowTestSuite) TestUserRegisterFlow() {
	// otp email register

	// send otp to email
	reqBody := fmt.Sprintf(`{
		"email": "%s"
		}`, TestUser.Email)

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/auth/send-otp", nil, []byte(reqBody), http.DefaultClient)

	s.NoError(err)

	if err != nil {
		s.FailNow(err.Error())
	}

	s.Equal(http.StatusOK, res.StatusCode)

	logger.Info("OTP sent to email")

	// get otp from dynamodb
	OTPs, err := s.getOTPs()

	if err != nil {
		s.FailNow(err.Error())
	}

	s.NotEmpty(OTPs, "OTPs should not be empty")

	logger.Dev("OTPs: %v", OTPs)

	otpVerifiedRes := &http.Response{}
	otpVerifiedResBody := ""

	for _, otp := range OTPs {
		// verify otp
		reqBody = fmt.Sprintf(`{
			"email": "%s",
			"otp": "%s"
			}`, TestUser.Email, otp)

		logger.Dev("req body: %v", reqBody)

		res, respBody, err := utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/auth/verify-otp", nil, []byte(reqBody), http.DefaultClient)

		if err == nil && res.StatusCode == 200 {
			otpVerifiedResBody = respBody
			otpVerifiedRes = res
			break
		}
	}

	if otpVerifiedResBody == "" {
		s.FailNow("failed to verify otp")
	}

	logger.Info("OTP verified successfully")

	// res body
	var resData struct {
		Data struct {
			UserId  string `json:"userId"`
			NewUser bool   `json:"isNewUser"`
		} `json:"data"`
	}

	err = json.Unmarshal([]byte(otpVerifiedResBody), &resData)

	s.NoError(err, "failed to unmarshal response body")

	s.True(resData.Data.NewUser, "isNewUser should be true")

	s.NotEmpty(resData.Data.UserId, "user id should not be empty")

	for _, c := range otpVerifiedRes.Cookies() {
		logger.Dev("cookie name: %v", c.Name)
		if c.Name != "session" {
			continue
		}
		s.SessionCookie = c
	}

	// create user

	reqBody = fmt.Sprintf(`{
		"id": "%s",
		"fullName": "%s",
		"email": "%s",
		"profilePic": "%s"
		}`,
		resData.Data.UserId, TestUser.FullName, TestUser.Email, TestUser.ProfilePic)

	reqHeader := map[string]string{
		"Cookie": s.SessionCookie.String(),
	}

	res, respData, err := utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/users", reqHeader, []byte(reqBody), http.DefaultClient)

	s.NoError(err)
	s.Equal(200, res.StatusCode, "POST /users")

	logger.Dev("POST /users > resp data: %v", respData)

	logger.Info("User created")

}

// run e2e tests
func TestUserFlowTestSuite(t *testing.T) {
	suite.Run(t, new(UserFlowTestSuite))
}
