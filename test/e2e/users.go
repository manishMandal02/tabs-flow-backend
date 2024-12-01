package e2e_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/manishMandal02/tabsflow-backend/internal/users"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

type UserAuthSuite struct {
	E2ETestSuite
}

func (s *UserAuthSuite) SetupSuite() {
	s.initSuite()
}

func (s *UserAuthSuite) TestUsers1_RegisterLogin() {
	// otp email register

	// send otp to email
	reqBody := fmt.Sprintf(`{
		"email": "%s"
		}`, TestUser.Email)

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/auth/send-otp", nil, []byte(reqBody), http.DefaultClient)

	s.Require().NoError(err, "err sending otp - POST /auth/send-otp")

	s.Require().Equal(http.StatusOK, res.StatusCode, "POST /auth/send-otp")

	// get otp from dynamodb
	OTPs, err := getOTPs(s.DDBClient, s.ENV.SessionTable)

	if err != nil {
		s.FailNow(err.Error())
	}

	s.Require().NotEmpty(OTPs, "OTPs should not be empty")

	otpVerifiedResBody := ""

	// verify otp and start new session
	for _, otp := range OTPs {
		reqBody = fmt.Sprintf(`{
			"email": "%s",
			"otp": "%s"
			}`, TestUser.Email, otp)

		res, respBody, err := utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/auth/verify-otp", s.Headers, []byte(reqBody), s.HttpClient)

		s.Require().NoError(err, "failed to make verify otp [POST /auth/verify-otp]")

		if res.StatusCode == 200 {
			cookies := s.HttpClient.Jar.Cookies(res.Request.URL)
			s.Require().NotEmpty(cookies, "cookies should not be empty")

			otpVerifiedResBody = respBody
			break
		}
	}

	s.Require().NotEmpty(otpVerifiedResBody, "failed to verify otp")

	// res body
	var resData struct {
		Data struct {
			UserId  string `json:"userId"`
			NewUser bool   `json:"isNewUser"`
		} `json:"data"`
	}

	err = json.Unmarshal([]byte(otpVerifiedResBody), &resData)

	s.Require().NoError(err, "failed to unmarshal response body")

	s.Require().NotEmpty(resData.Data.UserId, "user id should not be empty")

	if !resData.Data.NewUser {
		logger.Info("User LoggedIn")
		return
	}

	// create new  user in db, if NewUser flag is true
	reqBody = fmt.Sprintf(`{
		"id": "%s",
		"firstName": "%s",
		"lastName": "%s",
		"email": "%s",
		"profilePic": "%s"
		}`,
		resData.Data.UserId, TestUser.FirstName, TestUser.LastName, TestUser.Email, TestUser.ProfilePic)

	res, _, err = utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/users/", s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /users")

	logger.Info("User Registered")

}

func (s *UserAuthSuite) TestUsers2_GetProfile() {
	res, profileBody, err := utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/users/me", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /users/me")
	profileJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(profileBody), &profileJson)

	s.Require().NoError(err)
	s.Require().NotEmpty(profileJson.Data, "profile should not be empty")

}

func (s *UserAuthSuite) TestUsers3_UpdateProfile() {

	newName := "Manish Mandal Updated"

	reqBody := fmt.Sprintf(`{
		"fullName": "%s"
	}`, newName)

	res, _, err := utils.MakeHTTPRequest(http.MethodPatch, s.ENV.ApiDomainName+"/users/", s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "PATCH /users/me")

	// verify updated profile
	res, profileBody, err := utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/users/me", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /users/me")
	profileJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(profileBody), &profileJson)
	s.Require().NoError(err)
	s.Require().NotEmpty(profileJson.Data, "profile should not be empty")
	s.Require().Equal(profileJson.Data["fullName"], newName, "fullName should be updated")
}

func (s *UserAuthSuite) TestUsers4_GetPreferences() {

	res, preferencesBody, err := utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/users/preferences/", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /users/preferences")

	prefJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(preferencesBody), &prefJson)

	s.NoError(err)
	s.NotEmpty(prefJson, "preferences should not be empty")

	fields := reflect.VisibleFields(reflect.TypeOf(users.Preferences{}))

	// check if all fields are present in preferences
	for _, field := range fields {
		key := strings.Split(field.Tag.Get("json"), ",")[0]

		if _, ok := prefJson.Data[key]; !ok {
			s.FailNow(fmt.Sprintf("%s not found in preferences", field.Name))
		}
	}
}

func (s *UserAuthSuite) TestUsers5_UpdatePreferences() {
	reqBody := `
	{
		"type": "General",
		"data": {
			"openSpace": "sameWindow"
		}
	}
	`

	res, _, err := utils.MakeHTTPRequest(http.MethodPatch, s.ENV.ApiDomainName+"/users/preferences/", s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "Patch /users/preferences")

	// verify updated preferences
	res, preferencesBody, err := utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/users/preferences/", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /users/preferences")
	prefJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(preferencesBody), &prefJson)

	s.Require().NoError(err)

	s.Require().Equal("sameWindow", prefJson.Data["general"].(map[string]interface{})["openSpace"])
}

func (s *UserAuthSuite) TestUsers6_GetSubscription() {
	res, subscriptionsBody, err := utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/users/subscription/", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /users/subscription")

	subJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(subscriptionsBody), &subJson)

	s.Require().NoError(err)
	s.Require().NotEmpty(subJson.Data, "subscriptions should not be empty")
}

func (s *UserAuthSuite) TestUsers7_GetSubscriptionStatus() {
	res, subscriptionStatusBody, err := utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/users/subscription/status/", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /users/subscription/status")
	subStatusJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(subscriptionStatusBody), &subStatusJson)

	s.Require().NoError(err)

	s.Require().NotEmpty(subStatusJson.Data, "subscription status data should not be empty")

	s.Require().NotEmpty(subStatusJson.Data["active"], "active field should not be empty in subscription status data")
}
