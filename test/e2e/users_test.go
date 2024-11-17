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

	defaultReqHeaders := map[string]string{
		"Content-Type": "application/json",
	}

	// get otp from dynamodb
	OTPs, err := s.getOTPs()

	if err != nil {
		s.FailNow(err.Error())
	}

	s.NotEmpty(OTPs, "OTPs should not be empty")

	otpVerifiedResBody := ""

	for _, otp := range OTPs {
		// verify otp
		reqBody = fmt.Sprintf(`{
			"email": "%s",
			"otp": "%s"
			}`, TestUser.Email, otp)

		res, respBody, err := utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/auth/verify-otp", defaultReqHeaders, []byte(reqBody), s.HttpClient)

		if err == nil && res.StatusCode == 200 {
			cookies := s.CookieJar.Cookies(res.Request.URL)
			s.Require().NotEmpty(cookies, "cookies should not be empty")

			otpVerifiedResBody = respBody
			break
		}
	}

	s.Require().NotEmpty(otpVerifiedResBody, "otp verified res body should not be empty")

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

	// create user
	reqBody = fmt.Sprintf(`{
		"id": "%s",
		"fullName": "%s",
		"email": "%s",
		"profilePic": "%s"
		}`,
		resData.Data.UserId, TestUser.FullName, TestUser.Email, TestUser.ProfilePic)

	res, _, err = utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/users/", defaultReqHeaders, []byte(reqBody), s.HttpClient)

	s.NoError(err)
	s.Equal(200, res.StatusCode, "POST /users")

	logger.Info("User created")

	// get user preferences

	res, _, err = utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/users/preferences/", defaultReqHeaders, nil, s.HttpClient)

	s.NoError(err)
	s.Equal(200, res.StatusCode, "GET /users/preferences")

	logger.Info("User preferences fetched")

}

// run e2e tests
func TestUserFlowTestSuite(t *testing.T) {
	suite.Run(t, new(UserFlowTestSuite))
}
