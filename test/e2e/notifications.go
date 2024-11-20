package e2e_tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/manishMandal02/tabsflow-backend/internal/notifications"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

type NotificationSuite struct {
	E2ETestSuite
}

func (s *NotificationSuite) SetupSuite() {
	s.initSuite()

}

func (s *NotificationSuite) TestNotifications1_Subscribe() {
	apiURL := fmt.Sprintf("%s/notifications/subscription", s.ENV.ApiDomainName)

	reqBody := `{
		"endpoint": "endpoint",
		"authKey": "1212ertwerwewr1",
		"p256dhKey": "T34RGGH345634GERFDS"
	}`

	res, _, err := utils.MakeHTTPRequest(http.MethodPost, apiURL, s.Headers, []byte(reqBody), s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "POST /notifications/subscription")
}

func (s *NotificationSuite) TestNotifications2_GetSubscription() {
	apiURL := fmt.Sprintf("%s/notifications/subscription", s.ENV.ApiDomainName)
	res, resBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "GET /notifications/subscription")

	subscription := struct {
		Data map[string]interface{} `json:"data"`
	}{}

	err = json.Unmarshal([]byte(resBody), &subscription)

	s.Require().NoError(err)

	s.Require().Equal("endpoint", subscription.Data["endpoint"], "notification endpoint not correct")
	s.Require().Equal("1212ertwerwewr1", subscription.Data["authKey"], "notification authKey not correct")
	s.Require().Equal("T34RGGH345634GERFDS", subscription.Data["p256dhKey"], "notification p256dhKey not correct")
}

func (s *NotificationSuite) TestNotifications3_Unsubscribe() {
	apiURL := fmt.Sprintf("%s/notifications/subscription", s.ENV.ApiDomainName)

	res, _, err := utils.MakeHTTPRequest(http.MethodDelete, apiURL, s.Headers, nil, s.HttpClient)

	s.Require().NoError(err)
	s.Require().Equal(200, res.StatusCode, "DELETE /notifications/subscription")
}

var notificationIds = []string{}

func (s *NotificationSuite) TestNotifications3_UserNotifications() {
	// wait for 4 seconds for notifications to be processed
	logger.Info(" ⏳ Waiting for a few seconds, for notifications to be processed...")
	time.Sleep(5 * time.Second)

	apiURL := fmt.Sprintf("%s/notifications/my/", s.ENV.ApiDomainName)

	notificationRes := struct {
		Data []map[string]interface{} `json:"data"`
	}{}

	// retry 3 times
	// sometimes it takes a little more than usual to process notification events
	for i := 0; i <= 3; i++ {
		res, resBody, err := utils.MakeHTTPRequest(http.MethodGet, apiURL, s.Headers, nil, s.HttpClient)

		s.Require().NoError(err)
		s.Require().Equal(200, res.StatusCode, "GET /notificationRes/my")

		err = json.Unmarshal([]byte(resBody), &notificationRes)

		s.Require().NoError(err)

		if len(notificationRes.Data) >= 2 {
			// stop if more 2 notification found
			break
		}

		time.Sleep(2 * time.Second)
	}

	for _, notification := range notificationRes.Data {
		validType := notification["type"] == notifications.NotificationTypeNoteRemainder || notification["type"] == notifications.NotificationTypeUnSnoozedType
		s.Require().True(validType, "notification should have a valid type")

		notificationIds = append(notificationIds, notification["id"].(string))

		if notification["type"] == notifications.NotificationTypeNoteRemainder {
			note := notification["note"].(map[string]interface{})
			s.Require().Equal(note["domain"], "tabsflow.com", "note remainder notification should have a domain = 'tabsflow.com'")
			s.Require().Equal(note["title"], "TabsFlow Launch", "note remainder notification should have a title = 'FreshTabs Launch'")
		}

		if notification["type"] == notifications.NotificationTypeUnSnoozedType {
			snoozedTab := notification["snoozedTab"].(map[string]interface{})

			s.Require().Equal(snoozedTab["title"], snoozedTabs[1]["title"], "un_snoozed tab notification should have a title = 'Manish Mandal | Fullstack Web Developer'")
			s.Require().Equal(snoozedTab["url"], snoozedTabs[1]["url"], "un_snoozed tab notification should have a url = 'https://manishmandal.com'")
		}
	}

}

func (s *NotificationSuite) TestNotifications4_NotificationById() {
	apiURL := fmt.Sprintf("%s/notifications/", s.ENV.ApiDomainName)

	for _, id := range notificationIds {
		res, _, err := utils.MakeHTTPRequest(http.MethodGet, apiURL+id, s.Headers, nil, s.HttpClient)

		s.Require().NoError(err)
		s.Require().Equal(http.StatusOK, res.Status, "GET /notifications/:id")
	}

}

func (s *NotificationSuite) TestNotifications4_DeleteNotification() {
	apiURL := fmt.Sprintf("%s/notifications/", s.ENV.ApiDomainName)

	for _, id := range notificationIds {
		res, _, err := utils.MakeHTTPRequest(http.MethodDelete, apiURL+id, s.Headers, nil, s.HttpClient)

		s.Require().NoError(err)
		s.Require().Equal(200, res.StatusCode, "DELETE /notifications/:id")
	}
}
