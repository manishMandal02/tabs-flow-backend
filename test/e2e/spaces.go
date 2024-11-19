package e2e_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

type SpaceSuite struct {
	E2ETestSuite
}

func (s *SpaceSuite) SetupSuite() {
	s.initSuite()
}

var spaceId = "E34Y321"

func (s *SpaceSuite) TestSpaces1_CreateSpace() {
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
	apiURL := fmt.Sprintf("%s/spaces/%s/tabs/", s.ENV.ApiDomainName, spaceId)
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

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, apiURL, s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /spaces/tabs")
}

func (s *SpaceSuite) TestSpaces5_GetTabs() {
	apiURL := fmt.Sprintf("%s/spaces/%s/tabs/", s.ENV.ApiDomainName, spaceId)

	res, tabsBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/tabs")

	tabsJson := struct {
		Data []map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(tabsBody), &tabsJson)
	s.Require().NoError(err)
	s.Require().NotEmpty(tabsJson.Data, "tabs should not be empty")

	s.Require().Equal("FreshInbox | Gmail Inbox Cleaner", tabsJson.Data[0]["title"], "incorrect 1st tab's title")
	s.Require().Equal("Manish Mandal | Fullstack Web Developer", tabsJson.Data[1]["title"], "incorrect 2nd tab's title")
}

func (s *SpaceSuite) TestSpaces6_AddGroups() {
	apiURL := fmt.Sprintf("%s/spaces/%s/groups/", s.ENV.ApiDomainName, spaceId)

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

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, apiURL, s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /spaces/groups")
}

func (s *SpaceSuite) TestSpaces7_GetGroups() {
	apiURL := fmt.Sprintf("%s/spaces/%s/groups/", s.ENV.ApiDomainName, spaceId)

	res, groupsBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/groups")
	groupsJson := struct {
		Data []map[string]interface{} `json:"data"`
	}{}
	err = json.Unmarshal([]byte(groupsBody), &groupsJson)
	s.Require().NoError(err)

	s.Require().NotEmpty(groupsJson.Data, "groups should not be empty")

	s.Require().Equal("Backend", groupsJson.Data[0]["name"], "incorrect 1st group's name")
	s.Require().Equal("Extension", groupsJson.Data[1]["name"], "incorrect 2nd group's name")
}

var snoozedTab1Id = time.Now().Unix()
var snoozedTab2IdSaved = time.Now().Add(time.Second + 4).Unix()

var snoozedTab1SnoozedUntil = time.Now().Add(time.Hour * 6).Unix()
var snoozedTab2SnoozedUntil = time.Now().Add(time.Second * 3).Unix()

func (s *SpaceSuite) TestSpaces8_CreateSnoozedTabs() {

	apiURLSnoozedTabs := fmt.Sprintf("%s/spaces/%s/snoozed-tabs/", s.ENV.ApiDomainName, spaceId)

	reqBody1 := fmt.Sprintf(`{
		"snoozedAt": %v,
		"url": "https://freshinbox.xyz",
		"title": "FreshInbox | Gmail Inbox Cleaner",
		"icon": "https://freshinbox.xyz/favicon",
		"groupId": 49254834,
		"snoozedUntil": %v
	}`, snoozedTab1Id, snoozedTab1SnoozedUntil)

	reqBody2 := fmt.Sprintf(`{
		"snoozedAt": %v,
		"url": "https://manishmandal.com",
		"title": "Manish Mandal | Fullstack Web Developer",
		"icon": "https://manishmandal.com/favicon",
		"groupId": 49254834,
		"snoozedUntil": %v
	}`, snoozedTab2IdSaved, snoozedTab2SnoozedUntil)

	for _, reqBody := range []string{reqBody1, reqBody2} {
		res, _, err := utils.MakeHTTPRequest(http.MethodPost, apiURLSnoozedTabs, s.Headers, []byte(reqBody), s.HttpClient)

		s.Require().NoError(err)
		s.Require().Equal(200, res.StatusCode, "POST /spaces/snoozed-tabs")
	}
}

func (s *SpaceSuite) TestSpaces91_GetSnoozedTabById() {
	apiURL := fmt.Sprintf("%s/spaces/%s/snoozed-tabs/%v", s.ENV.ApiDomainName, spaceId, snoozedTab1Id)

	res, snoozedTabsBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /spaces/snoozed-tabs/:id")

	snoozedTabJson := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(snoozedTabsBody), &snoozedTabJson)

	s.Require().NoError(err)

	s.Require().NotEmpty(snoozedTabJson.Data, "snoozed tabs should not be empty")

	s.Require().Equal("https://freshinbox.xyz", snoozedTabJson.Data["url"])
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

	validTitle := "FreshInbox | Gmail Inbox Cleaner" == snoozedTabsJson.Data[0]["title"] || "FreshInbox | Gmail Inbox Cleaner" == snoozedTabsJson.Data[1]["title"]

	s.Require().True(validTitle, "one of the snoozed tab should contain title = 'FreshInbox | Gmail Inbox Cleaner'")
}

func (s *SpaceSuite) TestSpaces93_GetSpaceSnoozedTabs() {

	apiURL := fmt.Sprintf("%s/spaces/%s/snoozed-tabs/", s.ENV.ApiDomainName, spaceId)

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

	validTitle := "FreshInbox | Gmail Inbox Cleaner" == snoozedTabsJson.Data[0]["title"] || "FreshInbox | Gmail Inbox Cleaner" == snoozedTabsJson.Data[1]["title"]

	s.Require().True(validTitle, "one of the snoozed tab should contain title = 'FreshInbox | Gmail Inbox Cleaner'")
}

func (s *SpaceSuite) TestSpaces94_DeleteSnoozedTab() {
	apiURL := fmt.Sprintf("%s/spaces/%s/snoozed-tabs/%v", s.ENV.ApiDomainName, spaceId, snoozedTab1Id)
	res, _, err := utils.MakeHTTPRequest(http.MethodDelete, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "DELETE /spaces/snoozed-tabs/:id")
}
