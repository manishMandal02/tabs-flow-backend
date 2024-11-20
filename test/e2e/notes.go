package e2e_tests

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

type NotesSuite struct {
	E2ETestSuite
}

func (s *NotesSuite) SetupSuite() {
	s.initSuite()
}

func (s *NotesSuite) TestNotes1_Create() {
	reqBody1, err1 := json.Marshal(notes[0])
	reqBody2, err2 := json.Marshal(notes[1])
	s.Require().NoError(err1, "err marshaling note1")
	s.Require().NoError(err2, "err marshaling note2")

	for _, reqBody := range [][]byte{reqBody1, reqBody2} {
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

	validTitle := notes[0]["title"] == notesJson.Data[0]["title"] || notes[0]["title"] == notesJson.Data[1]["title"]

	s.Require().True(validTitle, "one of the note should have a title = 'FreshTabs Launch'")

	validDomain := notes[0]["domain"] == notesJson.Data[0]["domain"] || notes[0]["domain"] == notesJson.Data[1]["domain"]

	s.Require().True(validDomain, "one of the note should have a domain = 'freshinbox.xyz'")
}

func (s *NotesSuite) TestNotes3_GetNote() {

	apiURL := fmt.Sprintf("%s/notes/%v", s.ENV.ApiDomainName, notes[0]["id"])

	res, body, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /notes/:id")

	noteBody := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(body), &noteBody)

	s.Require().NoError(err)

	s.Require().Equal(notes[0]["title"], noteBody.Data["title"], "note title not correct")

	s.Require().Equal(notes[0]["domain"], noteBody.Data["domain"], "note domain not correct")

	s.Require().Equal(notes[0]["id"], noteBody.Data["id"], "note id should be valid")
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

	s.Require().Equal(notes[0]["title"], searchRes.Data[0]["title"], "note title not correct")

	s.Require().Equal(notes[0]["domain"], searchRes.Data[0]["domain"], "note domain not correct")
}

func (s *NotesSuite) TestNotes5_UpdateNote() {
	reqBody := fmt.Sprintf(`{
		"id": "%v",
		"title": "FreshTabs Launch updated",
		"spaceId": "%s"
	}`, notes[0]["id"], space["id"])

	res, _, err := utils.MakeHTTPRequest(http.MethodPatch, s.ENV.ApiDomainName+"/notes/", s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /notes/")
}

func (s *NotesSuite) TestNotes6_DeleteNote() {

	apiURL := fmt.Sprintf("%s/notes/%v", s.ENV.ApiDomainName, notes[0]["id"])

	res, _, err := utils.MakeHTTPRequest(http.MethodDelete, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "DELETE /notes/:id")
}
