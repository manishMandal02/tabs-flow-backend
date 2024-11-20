package e2e_tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

type SpaceSuite struct {
	E2ETestSuite
}

func (s *SpaceSuite) SetupSuite() {
	s.initSuite()
}

func (s *SpaceSuite) TestSpaces1_CreateSpace() {
	reqBody, err := json.Marshal(space)

	s.Require().NoError(err, "err marshaling space")

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/spaces/", s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /spaces")
}

func (s *SpaceSuite) TestSpaces2_GetUserSpaces() {
	res, spacesBody, err := utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/spaces/my", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/my")

	spacesJson := struct {
		Data []map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(spacesBody), &spacesJson)
	s.Require().NoError(err)
	s.Require().NotEmpty(spacesJson.Data, "spaces should not be empty")

	s.Require().Equal("Work", spacesJson.Data[0]["title"], "space title not correct")
}

func (s *SpaceSuite) TestSpaces3_UpdateSpace() {

	reqBody := `{
		"id": "E34Y321",
		"title": "WorkSpace"
		}`

	res, _, err := utils.MakeHTTPRequest(http.MethodPatch, s.ENV.ApiDomainName+"/spaces/", s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "PATCH /spaces")

	res, spacesBody, err := utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/spaces/E34Y321", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/:spaceId")
	spaceJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(spacesBody), &spaceJson)
	s.Require().NoError(err)
	s.Require().NotEmpty(spaceJson.Data, "spaces should not be empty")

	s.Require().Equal("WorkSpace", spaceJson.Data["title"], "incorrect space title ")
}

func (s *SpaceSuite) TestSpaces4_AddTabs() {
	apiURL := fmt.Sprintf("%s/spaces/%s/tabs/", s.ENV.ApiDomainName, space["id"])
	tabsJson, err := json.Marshal(tabs)

	reqBody := fmt.Sprintf(`{
	"tabs": %s
	}`, tabsJson)

	s.Require().NoError(err, "Error marshalling tabs")

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, apiURL, s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /spaces/tabs")
}

func (s *SpaceSuite) TestSpaces5_GetTabs() {
	apiURL := fmt.Sprintf("%s/spaces/%s/tabs/", s.ENV.ApiDomainName, space["id"])

	res, tabsBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/tabs")

	tabsJson := struct {
		Data []map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(tabsBody), &tabsJson)
	s.Require().NoError(err)
	s.Require().NotEmpty(tabsJson.Data, "tabs should not be empty")

	s.Require().Equal(tabs[0]["title"].(string), tabsJson.Data[0]["title"], "incorrect 1st tab's title")
	s.Require().Equal(tabs[1]["title"], tabsJson.Data[1]["title"], "incorrect 2nd tab's title")
}

func (s *SpaceSuite) TestSpaces6_AddGroups() {
	apiURL := fmt.Sprintf("%s/spaces/%s/groups/", s.ENV.ApiDomainName, space["id"])

	groupsJson, err := json.Marshal(groups)

	reqBody := fmt.Sprintf(`{
	"groups": %s
	}`, groupsJson)

	s.Require().NoError(err, "Error marshalling groups")

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, apiURL, s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /spaces/groups")
}

func (s *SpaceSuite) TestSpaces7_GetGroups() {
	apiURL := fmt.Sprintf("%s/spaces/%s/groups/", s.ENV.ApiDomainName, space["id"])

	res, groupsBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/groups")
	groupsJson := struct {
		Data []map[string]interface{} `json:"data"`
	}{}
	err = json.Unmarshal([]byte(groupsBody), &groupsJson)
	s.Require().NoError(err)

	s.Require().NotEmpty(groupsJson.Data, "groups should not be empty")

	s.Require().Equal(groups[0]["name"], groupsJson.Data[0]["name"], "incorrect 1st group's name")
	s.Require().Equal(groups[1]["name"], groupsJson.Data[1]["name"], "incorrect 2nd group's name")
}

func (s *SpaceSuite) TestSpaces8_CreateSnoozedTabs() {

	apiURLSnoozedTabs := fmt.Sprintf("%s/spaces/%s/snoozed-tabs/", s.ENV.ApiDomainName, space["id"])

	reqBody1, err1 := json.Marshal(snoozedTabs[0])
	reqBody2, err2 := json.Marshal(snoozedTabs[1])

	s.Require().NoError(err1, "json marshal failed tab 1")
	s.Require().NoError(err2, "json marshal failed tab 2")

	for _, reqBody := range [][]byte{reqBody1, reqBody2} {
		res, _, err := utils.MakeHTTPRequest(http.MethodPost, apiURLSnoozedTabs, s.Headers, reqBody, s.HttpClient)

		s.Require().NoError(err)
		s.Require().Equal(200, res.StatusCode, "POST /spaces/snoozed-tabs")
	}
}

func (s *SpaceSuite) TestSpaces91_GetSnoozedTabById() {
	apiURL := fmt.Sprintf("%s/spaces/%s/snoozed-tabs/%v", s.ENV.ApiDomainName, space["id"], snoozedTabs[0]["snoozedAt"])

	res, snoozedTabsBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/snoozed-tabs/:id")

	snoozedTabJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(snoozedTabsBody), &snoozedTabJson)

	s.Require().NoError(err)

	s.Require().NotEmpty(snoozedTabJson.Data, "snoozed tabs should not be empty")

	s.Require().Equal(snoozedTabs[0]["url"], snoozedTabJson.Data["url"])
}

func (s *SpaceSuite) TestSpaces92_GetUserSnoozedTabs() {
	apiURL := fmt.Sprintf("%s/spaces/snoozed-tabs/my", s.ENV.ApiDomainName)

	// get users snoozed tabs
	res, snoozedTabsBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/snoozed-tabs/my")
	snoozedTabsJson := struct {
		Data []map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(snoozedTabsBody), &snoozedTabsJson)
	s.Require().NoError(err)
	s.Require().NotEmpty(snoozedTabsJson.Data, "snoozed tabs should not be empty")

	s.Require().Len(snoozedTabsJson.Data, 2, "2 snoozed tabs should be returned")

	validTitle := snoozedTabs[0]["title"] == snoozedTabsJson.Data[0]["title"] || snoozedTabs[0]["title"] == snoozedTabsJson.Data[1]["title"]

	s.Require().True(validTitle, "one of the snoozed tab should contain title = 'FreshInbox | Gmail Inbox Cleaner'")
}

func (s *SpaceSuite) TestSpaces93_GetSpaceSnoozedTabs() {

	apiURL := fmt.Sprintf("%s/spaces/%s/snoozed-tabs/", s.ENV.ApiDomainName, space["id"])

	res, snoozedTabsBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/snoozed-tabs")
	snoozedTabsJson := struct {
		Data []map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(snoozedTabsBody), &snoozedTabsJson)

	s.Require().NoError(err)
	s.Require().NotEmpty(snoozedTabsJson.Data, "snoozed tabs should not be empty")

	s.Require().Len(snoozedTabsJson.Data, 2, "2 snoozed tabs should be returned")

	validTitle := snoozedTabs[0]["title"] == snoozedTabsJson.Data[0]["title"] || snoozedTabs[0]["title"] == snoozedTabsJson.Data[1]["title"]

	s.Require().True(validTitle, "one of the snoozed tab should contain title = 'FreshInbox | Gmail Inbox Cleaner'")
}

func (s *SpaceSuite) TestSpaces94_DeleteSnoozedTab() {
	apiURL := fmt.Sprintf("%s/spaces/%s/snoozed-tabs/%v", s.ENV.ApiDomainName, space["id"], snoozedTabs[0]["snoozedAt"])
	res, _, err := utils.MakeHTTPRequest(http.MethodDelete, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "DELETE /spaces/snoozed-tabs/:id")
}
