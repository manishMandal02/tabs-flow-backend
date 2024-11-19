package e2e_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

type NotesSuite struct {
	E2ETestSuite
}

func (s *NotesSuite) SetupSuite() {
	s.initSuite()
}

var note1Id = time.Now().Unix()
var note1TextLexicalFormat = `{\"root\":{\"children\":[{\"children\":[{\"detail\":0,\"format\":1,\"mode\":\"normal\",\"style\":\"\",\"text\":\"Know the Features\",\"type\":\"text\",\"version\":1},{\"detail\":0,\"format\":0,\"mode\":\"normal\",\"style\":\"\",\"text\":\"UnsubscribeWith a single click, bid farewell to newsletters and promotional emails, streamlining your inbox and ensuring that only the content that truly matters remains.\",\"type\":\"text\",\"version\":1}],\"direction\":\"ltr\",\"format\":\"\",\"indent\":0,\"type\":\"paragraph\",\"version\":1},{\"children\":[],\"direction\":\"ltr\",\"format\":\"\",\"indent\":0,\"type\":\"paragraph\",\"version\":1},{\"children\":[{\"detail\":0,\"format\":0,\"mode\":\"normal\",\"style\":\"\",\"text\":\"Bulk DeleteSay goodbye to tedious, manual deletions, and effortlessly remove hundreds or thousands of emails from selected senders.\",\"type\":\"text\",\"version\":1}],\"direction\":\"ltr\",\"format\":\"\",\"indent\":0,\"type\":\"paragraph\",\"version\":1},{\"type\":\"horizontalrule\",\"version\":1},{\"children\":[{\"detail\":0,\"format\":0,\"mode\":\"normal\",\"style\":\"\",\"text\":\"Advanced SearchEasily locate specific emails using various filters. Once identified the bulk delete feature can clear out emails in 100s or 1000s.Privacy & SecurityFreshInbox is developed with a commitment to keeping your data secure. No data ever leaves your browser, ensuring that your sensitive information remains confidential and protected.\",\"type\":\"text\",\"version\":1}],\"direction\":\"ltr\",\"format\":\"\",\"indent\":0,\"type\":\"paragraph\",\"version\":1},{\"children\":[],\"direction\":\"ltr\",\"format\":\"\",\"indent\":0,\"type\":\"paragraph\",\"version\":1},{\"children\":[{\"detail\":0,\"format\":1,\"mode\":\"normal\",\"style\":\"\",\"text\":\"This is bold\",\"type\":\"text\",\"version\":1},{\"detail\":0,\"format\":0,\"mode\":\"normal\",\"style\":\"\",\"text\":\" \",\"type\":\"text\",\"version\":1},{\"detail\":0,\"format\":2,\"mode\":\"normal\",\"style\":\"\",\"text\":\"this is italic\",\"type\":\"text\",\"version\":1},{\"detail\":0,\"format\":0,\"mode\":\"normal\",\"style\":\"\",\"text\":\" \",\"type\":\"text\",\"version\":1}],\"direction\":\"ltr\",\"format\":\"\",\"indent\":0,\"type\":\"paragraph\",\"version\":1},{\"children\":[{\"children\":[{\"detail\":0,\"format\":0,\"mode\":\"normal\",\"style\":\"\",\"text\":\" This a a numbered list\",\"type\":\"text\",\"version\":1}],\"direction\":\"ltr\",\"format\":\"\",\"indent\":0,\"type\":\"listitem\",\"version\":1,\"value\":1}],\"direction\":\"ltr\",\"format\":\"\",\"indent\":0,\"type\":\"list\",\"version\":1,\"listType\":\"number\",\"start\":1,\"tag\":\"ol\"},{\"children\":[{\"children\":[{\"detail\":0,\"format\":0,\"mode\":\"normal\",\"style\":\"\",\"text\":\" This is a bulletin  list \",\"type\":\"text\",\"version\":1}],\"direction\":\"ltr\",\"format\":\"\",\"indent\":0,\"type\":\"listitem\",\"version\":1,\"value\":1}],\"direction\":\"ltr\",\"format\":\"\",\"indent\":0,\"type\":\"list\",\"version\":1,\"listType\":\"bullet\",\"start\":1,\"tag\":\"ul\"},{\"children\":[{\"detail\":0,\"format\":0,\"mode\":\"normal\",\"style\":\"\",\"text\":\" This is a quote\",\"type\":\"text\",\"version\":1}],\"direction\":\"ltr\",\"format\":\"\",\"indent\":0,\"type\":\"quote\",\"version\":1}],\"direction\":\"ltr\",\"format\":\"\",\"indent\":0,\"type\":\"root\",\"version\":1}}`

var note2Id = time.Now().Add(time.Second + 1).Unix()
var note2Text = `{\"root\": {\"children\": [ {\"children\":[ {\"type\": \"text\", \"text\": \"Simple Note to test note remainder\"}]}]}}`
var note2RemindAt = time.Now().Add(time.Second + 4).Unix()

func (s *NotesSuite) TestNotes1_Create() {
	note1 := fmt.Sprintf(`{
			"id": "%v",
			"title": "FreshTabs Launch",
			"spaceId": "%s",
			"domain": "freshinbox.xyz",
			"updatedAt": 1729256811,
			"text": "%v"
		}`, note1Id, spaceId, note1TextLexicalFormat)

	note2 := fmt.Sprintf(`{
			"id": "%v",
			"title": "TabsFlow Launch",
			"spaceId": "%s",
			"domain": "tabsflw.com",
			"updatedAt": 1729256811,
			"text": "%s",
			"remindAt": %v
		}`, note2Id, spaceId, note2Text, note2RemindAt)

	for _, reqBody := range []string{note1, note2} {

		res, _, err := utils.MakeHTTPRequest(http.MethodPost, s.ENV.ApiDomainName+"/notes/", s.Headers, []byte(reqBody), s.HttpClient)

		s.Require().NoError(err)
		s.Require().Equal(200, res.StatusCode, "POST /notes/")
	}
}

func (s *NotesSuite) TestNotes2_GetUserNotes() {

	res, body, err := utils.MakeHTTPRequest(http.MethodGet, s.ENV.ApiDomainName+"/notes/my", s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /notes/my")

	notesJson := struct {
		Data []map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(body), &notesJson)

	s.Require().NoError(err)

	s.Require().Len(notesJson.Data, 2, "number of notes should be 2")

	validTitle := "FreshTabs Launch" == notesJson.Data[0]["title"] || "FreshTabs Launch" == notesJson.Data[1]["title"]

	s.Require().True(validTitle, "one of the note should have a title = 'FreshTabs Launch'")

	validDomain := "freshinbox.xyz" == notesJson.Data[0]["domain"] || "freshinbox.xyz" == notesJson.Data[1]["domain"]

	s.Require().True(validDomain, "one of the note should have a domain = 'freshinbox.xyz'")
}

func (s *NotesSuite) TestNotes3_GetNote() {

	apiURL := fmt.Sprintf("%s/notes/%v", s.ENV.ApiDomainName, note1Id)

	res, body, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /notes/:id")

	noteBody := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(body), &noteBody)

	s.Require().NoError(err)

	s.Require().Equal("FreshTabs Launch", noteBody.Data["title"], "note title not correct")

	s.Require().Equal("freshinbox.xyz", noteBody.Data["domain"], "note domain not correct")

	s.Require().Equal(note1Id, noteBody.Data["id"], "note id should be valid")
}

func (s *NotesSuite) TestNotes4_Search() {

	apiURL := s.ENV.ApiDomainName + "/notes/search?query=inbox&limit=5"

	res, body, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /notes/search")

	searchRes := struct {
		Data []map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(body), &searchRes)
	s.Require().NoError(err)

	s.Require().Len(searchRes.Data, 1, "search result should have 1 note")

	s.Require().Equal("FreshTabs Launch", searchRes.Data[0]["title"], "note title not correct")

	s.Require().Equal("freshinbox.xyz", searchRes.Data[0]["domain"], "note domain not correct")
}

func (s *NotesSuite) TestNotes5_UpdateNote() {
	reqBody := fmt.Sprintf(`{
		"id": "%v",
		"title": "FreshTabs Launch updated",
		"spaceId": "%s"
	}`, note1Id, spaceId)

	res, _, err := utils.MakeHTTPRequest(http.MethodPatch, s.ENV.ApiDomainName+"/notes/", s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /notes/")
}

func (s *NotesSuite) TestNotes6_DeleteNote() {

	apiURL := fmt.Sprintf("%s/notes/%v", s.ENV.ApiDomainName, note1Id)

	res, _, err := utils.MakeHTTPRequest(http.MethodDelete, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "DELETE /notes/:id")
}
