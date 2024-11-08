package e2e_test

import (
	"os"

	"github.com/manishMandal02/tabsflow-backend/internal/users"
	"github.com/stretchr/testify/suite"
)

var TestUser = users.User{
	Email:      "test@example.com",
	FullName:   "Manish Mandal",
	ProfilePic: "https://avatars.githubusercontent.com/u/123456789?v=4",
}

type TestConfig struct {
	ApiDomainName         string
	MainTable             string
	SessionTable          string
	SearchIndexTable      string
	EmailQueueURL         string
	NotificationsQueueURL string
}

type UserFlowTestSuite struct {
	suite.Suite
	Config TestConfig
	// HTTPClient *http.Client
	// DDBClient  *dynamodb.Client
	// SQSClient  *sqs.Client
}

func (s *UserFlowTestSuite) SetupSuite() {
	loadENV(s)
}

func loadENV(s *UserFlowTestSuite) {

	// Load environment variables
	apiDomain, apiDomainOK := os.LookupEnv("API_DOMAIN_NAME")
	mainTableName, mainTableNameOK := os.LookupEnv("DDB_MAIN_TABLE_NAME")
	sessionTableName, sessionTableNameOK := os.LookupEnv("DDB_SESSIONS_TABLE_NAME")
	searchIndexTableName, searchIndexTableNameOK := os.LookupEnv("DDB_SEARCH_INDEX_TABLE_NAME")
	emailQueueURL, emailQueueUrlOK := os.LookupEnv("EMAIL_QUEUE_URL")
	notificationsQueueURL, notificationsQueueUrlOK := os.LookupEnv("NOTIFICATIONS_QUEUE_URL")

	if !apiDomainOK || !mainTableNameOK || !sessionTableNameOK ||
		!searchIndexTableNameOK || !emailQueueUrlOK || !notificationsQueueUrlOK {
		panic("Missing environment variables")
	}

	s.Config = TestConfig{
		ApiDomainName:         apiDomain,
		MainTable:             mainTableName,
		SessionTable:          sessionTableName,
		SearchIndexTable:      searchIndexTableName,
		EmailQueueURL:         emailQueueURL,
		NotificationsQueueURL: notificationsQueueURL,
	}

}
