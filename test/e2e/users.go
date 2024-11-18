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

	// register user
	s.RegisterOrLoginUser()
}

func (s *UserAuthSuite) TestUsersProfile() {
	// get user profile

	res, profileBody, err := utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/users/me", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /users/me")
	profileJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(profileBody), &profileJson)

	s.Require().NoError(err)
	s.Require().NotEmpty(profileJson.Data, "profile should not be empty")

	logger.Info("GET /users/me > success")

	newName := "Manish Mandal Updated"

	// update user profile
	reqBody := fmt.Sprintf(`{
		"fullName": "%s"
	}`, newName)

	res, _, err = utils.MakeHTTPRequest(http.MethodPatch, s.ENV.ApiDomainName+"/users/", s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "PATCH /users/me")

	// verify updated profile
	res, profileBody, err = utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/users/me", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /users/me")
	profileJson = struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(profileBody), &profileJson)
	s.Require().NoError(err)
	s.Require().NotEmpty(profileJson.Data, "profile should not be empty")
	s.Require().Equal(profileJson.Data["fullName"], newName, "fullName should be updated")

	logger.Info("PATCH /users/ > success")
}

func (s *UserAuthSuite) TestUsersPreferences() {
	// get user preferences

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

	for _, field := range fields {
		key := strings.Split(field.Tag.Get("json"), ",")[0]

		if _, ok := prefJson.Data[key]; !ok {
			s.FailNow(fmt.Sprintf("%s not found in preferences", field.Name))
		}
	}

	logger.Info("GET /users/preferences > success")

	// update user preferences
	reqBody := `
	{
		"type": "General",
		"data": {
			"openSpace": "sameWindow"
		}
	}
	`

	res, _, err = utils.MakeHTTPRequest(http.MethodPatch, s.ENV.ApiDomainName+"/users/preferences/", s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "Patch /users/preferences")

	// verify updated preferences
	res, preferencesBody, err = utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/users/preferences/", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /users/preferences")
	prefJson = struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(preferencesBody), &prefJson)

	s.Require().NoError(err)

	hasUpdated := prefJson.Data["general"].(map[string]interface{})["openSpace"] == "sameWindow"

	s.Require().True(hasUpdated, "preferences should be updated")

	logger.Info("PATCH /users/preferences > success")
}

func (s *UserAuthSuite) TestUsersSubscription() {
	// get user subscriptions
	res, subscriptionsBody, err := utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/users/subscription/", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /users/subscription")

	subJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(subscriptionsBody), &subJson)

	s.Require().NoError(err)
	s.Require().NotEmpty(subJson.Data, "subscriptions should not be empty")

	logger.Info("GET /users/subscription > success")

	// get user subscription status

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

	logger.Info("GET /users/subscription/status > success")

}
