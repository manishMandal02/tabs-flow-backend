package e2e_tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

type SpaceSuite struct {
	E2ETestSuite
}

func (s *SpaceSuite) SetupSuite() {
	s.initSuite()
}

var spaceId = "E34Y321"

func (s *SpaceSuite) TestSpaces1_Space() {
	// create space
	reqBody := fmt.Sprintf(`{
		"id": "%s",
		"title": "Work",
		"theme": "Green",
		"emoji": "💼",
		"isSaved": true,
		"windowId": 7890678432,
		"activeTabIndex":1
	}`, spaceId)

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/spaces/", s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /spaces")

	logger.Info("POST /spaces > success")

	// get spaces

	res, spacesBody, err := utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/spaces/my", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/my")

	spacesJson := struct {
		Data []map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(spacesBody), &spacesJson)
	s.Require().NoError(err)
	s.Require().NotEmpty(spacesJson.Data, "spaces should not be empty")

	validSpace := spacesJson.Data[0]["id"] == "E34Y321" && spacesJson.Data[0]["title"] == "Work"

	s.Require().True(validSpace, "space data should be valid")

	logger.Info("GET /spaces/my > success")

	// update space

	reqBody = `{
		"id": "E34Y321",
		"title": "WorkSpace"
		}`

	res, _, err = utils.MakeHTTPRequest(http.MethodPatch, s.ENV.ApiDomainName+"/spaces/", s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "PATCH /spaces")

	// verify updated space

	res, spacesBody, err = utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/spaces/E34Y321", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/:spaceId")
	spaceJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(spacesBody), &spaceJson)
	s.Require().NoError(err)
	s.Require().NotEmpty(spaceJson.Data, "spaces should not be empty")

	validSpace = spaceJson.Data["id"] == "E34Y321" && spaceJson.Data["title"] == "WorkSpace"

	s.Require().True(validSpace, "updated space data should be valid")

	logger.Info("PATCH /spaces > success")
}

func (s *SpaceSuite) TestSpaces2_Tabs() {

	apiURLTabs := fmt.Sprintf("%s/spaces/%s/tabs/", s.ENV.ApiDomainName, spaceId)

	// add tabs to spaces
	reqBody := `{
		"tabs": [
			{
				"id": 123456789,
				"url": "https://freshinbox.xyz",
				"title": "FreshInbox | Gmail Inbox Cleaner",
				"index": 0,
				"icon": "https://freshinbox.xyz/favicon",
				"groupId": 49254834
			},
			{
				"id": 879086463,
				"url": "https://manishmandal.com",
				"title": "Manish Mandal | Fullstack Web Developer",
				"index": 0,
				"icon": "https://manishmandal.com/favicon",
				"groupId": 49254834
			}
    	]
	}`

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, apiURLTabs, s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /spaces/tabs")

	logger.Info("POST /spaces/tabs > success")

	// get tabs

	res, tabsBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURLTabs, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/tabs")

	tabsJson := struct {
		Data []map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(tabsBody), &tabsJson)
	s.Require().NoError(err)
	s.Require().NotEmpty(tabsJson.Data, "tabs should not be empty")

	validTab1 := tabsJson.Data[0]["title"] == "FreshInbox | Gmail Inbox Cleaner"

	validTab2 := tabsJson.Data[1]["title"] == "Manish Mandal | Fullstack Web Developer"

	s.Require().True(validTab1 && validTab2, "tabs data should be valid")

}

func (s *SpaceSuite) TestSpaces3_Groups() {
	apiURLGroups := fmt.Sprintf("%s/spaces/%s/groups/", s.ENV.ApiDomainName, spaceId)

	// add groups to spaces

	reqBody := `{
    "groups": [
			{
				"id": 623678,
				"name": "Backend",
				"theme": "gray",
				"collapsed": true
	
			},
			{
				"id": 605489,
				"name": "Extension",
				"theme": "green",
				"collapsed": false
	
			}
   		]
	}`

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, apiURLGroups, s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /spaces/groups")

	logger.Info("POST /spaces/groups > success")

	// get groups
	res, groupsBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURLGroups, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/groups")
	groupsJson := struct {
		Data []map[string]interface{} `json:"data"`
	}{}
	err = json.Unmarshal([]byte(groupsBody), &groupsJson)
	s.Require().NoError(err)

	s.Require().NotEmpty(groupsJson.Data, "groups should not be empty")

	validGroup1 := groupsJson.Data[0]["name"] == "Backend"
	validGroup2 := groupsJson.Data[1]["name"] == "Extension"

	s.Require().True(validGroup1 && validGroup2, "groups data should be valid")

	logger.Info("GET /spaces/groups > success")

}

func (s *SpaceSuite) TestSpaces4_SnoozedTabs() {

	apiURLTabs := fmt.Sprintf("%s/spaces/%s/snoozed-tabs/", s.ENV.ApiDomainName, spaceId)
	snoozedTabId := "1731945819"
	// snooze tab
	reqBody := fmt.Sprintf(`{
		"snoozedAt": %v,
		"url": "https://freshinbox.xyz",
		"title": "FreshInbox | Gmail Inbox Cleaner",
		"icon": "https://freshinbox.xyz/favicon",
		"groupId": 49254834,
		"snoozedUntil": 1731946819
	}`, snoozedTabId)

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, apiURLTabs, s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /spaces/snoozed-tabs")
	logger.Info("POST /spaces/snoozed-tabs > success")

	// get snoozed tab by id
	res, snoozedTabsBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURLTabs+snoozedTabId, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/snoozed-tabs:id")

	snoozedTabJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}
	err = json.Unmarshal([]byte(snoozedTabsBody), &snoozedTabJson)

	s.Require().NoError(err)
	s.Require().NotEmpty(snoozedTabJson.Data, "snoozed tabs should not be empty")

	validSnoozedTab := snoozedTabJson.Data["url"] == "https://freshinbox.xyz"

	s.Require().True(validSnoozedTab, "snoozed tab data should be valid")

	logger.Info("GET /spaces/snoozed-tabs/:id > success")

	// get users snoozed tabs
	res, snoozedTabsBody, err = utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/spaces/snoozed-tabs/my", s.Headers, nil, s.HttpClient)
	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/snoozed-tabs/my")
	snoozedTabsJson := struct {
		Data []map[string]interface{} `json:"data"`
	}{}
	err = json.Unmarshal([]byte(snoozedTabsBody), &snoozedTabsJson)
	s.Require().NoError(err)
	s.Require().NotEmpty(snoozedTabsJson.Data, "snoozed tabs should not be empty")
	validSnoozedTab = snoozedTabsJson.Data[0]["url"] == "https://freshinbox.xyz"

	s.Require().True(validSnoozedTab, "snoozed tab data should be valid")

	logger.Info("GET /spaces/snoozed-tabs/my > success")

	// get snoozed tabs in space

	res, snoozedTabsBody, err = utils.MakeHTTPRequest(http.MethodGet, apiURLTabs, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/snoozed-tabs")
	snoozedTabsJson = struct {
		Data []map[string]interface{} `json:"data"`
	}{}
	err = json.Unmarshal([]byte(snoozedTabsBody), &snoozedTabsJson)

	s.Require().NoError(err)
	s.Require().NotEmpty(snoozedTabsJson.Data, "snoozed tabs should not be empty")
	validSnoozedTab = snoozedTabsJson.Data[0]["url"] == "https://freshinbox.xyz"
	s.Require().True(validSnoozedTab, "snoozed tab data should be valid")

	logger.Info("GET /spaces/snoozed-tabs > success")

	// delete snoozed tab

	res, _, err = utils.MakeHTTPRequest(http.MethodDelete, apiURLTabs+snoozedTabId, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "DELETE /spaces/snoozed-tabs/:id")

	logger.Info("DELETE /spaces/snoozed-tabs/:id > success")

}
