package e2e_tests

import (
	"context"
	"errors"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/joho/godotenv"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/stretchr/testify/suite"
)

var CookieJar, _ = cookiejar.New(nil)

type ENV struct {
	ApiDomainName         string
	AWS_ACCOUNT_PROFILE   string
	MainTable             string
	SessionTable          string
	SearchIndexTable      string
	EmailQueueURL         string
	NotificationsQueueURL string
}
type E2ETestSuite struct {
	suite.Suite
	ENV        ENV
	AppURL     *url.URL
	HttpClient *http.Client
	AWSConfig  *aws.Config
	DDBClient  *dynamodb.Client
	Headers    map[string]string
}

func (s *E2ETestSuite) initSuite() {
	s.T().Helper()

	env := getENVs()

	s.ENV = env

	s.Headers = map[string]string{
		"Content-Type": "application/json",
	}

	s.ENV.ApiDomainName = "https://" + s.ENV.ApiDomainName

	appURL, err := url.Parse("https://tabsflow.com")

	s.Require().NoError(err)

	s.AppURL = appURL

	s.Require().NoError(err)

	s.HttpClient = &http.Client{
		Jar: CookieJar,
	}

	awsConfig := configureAWS(env.AWS_ACCOUNT_PROFILE)

	s.AWSConfig = awsConfig

	s.DDBClient = dynamodb.NewFromConfig(*awsConfig)
}

func getENVs() ENV {

	err := godotenv.Load("./../../.env")

	if err != nil {
		logger.Errorf("Error loading .env file: [Error] %v", err)
		panic("Error loading .env file")
	}

	// Load environment variables
	awsProfile, awsProfileOK := os.LookupEnv("AWS_ACCOUNT_PROFILE")
	apiDomain, apiDomainOK := os.LookupEnv("API_DOMAIN_NAME")
	mainTableName, mainTableNameOK := os.LookupEnv("DDB_MAIN_TABLE_NAME")
	sessionTableName, sessionTableNameOK := os.LookupEnv("DDB_SESSIONS_TABLE_NAME")
	searchIndexTableName, searchIndexTableNameOK := os.LookupEnv("DDB_SEARCH_INDEX_TABLE_NAME")
	emailQueueURL, emailQueueUrlOK := os.LookupEnv("EMAIL_QUEUE_URL")
	notificationsQueueURL, notificationsQueueUrlOK := os.LookupEnv("NOTIFICATIONS_QUEUE_URL")

	if !apiDomainOK || !mainTableNameOK || !sessionTableNameOK || !awsProfileOK ||
		!searchIndexTableNameOK || !emailQueueUrlOK || !notificationsQueueUrlOK {
		panic("Missing environment variables")
	}

	return ENV{
		ApiDomainName:         apiDomain,
		AWS_ACCOUNT_PROFILE:   awsProfile,
		MainTable:             mainTableName,
		SessionTable:          sessionTableName,
		SearchIndexTable:      searchIndexTableName,
		EmailQueueURL:         emailQueueURL,
		NotificationsQueueURL: notificationsQueueURL,
	}
}

func configureAWS(profile string) *aws.Config {
	config, err := config.LoadDefaultConfig(context.Background(),
		config.WithSharedConfigProfile(profile),
		config.WithRetryer(func() aws.Retryer {
			return retry.AddWithMaxAttempts(retry.NewStandard(), 20)
		}),
	)

	if err != nil {
		logger.Errorf("Error loading AWS config: [Error] %v", err)
		panic("Error loading AWS config")
	}

	return &config
}

func getOTPs(client *dynamodb.Client, tableName string) ([]string, error) {
	keyCondition := expression.KeyAnd(expression.Key("PK").Equal(expression.Value(TestUser.Email)), expression.Key("SK").BeginsWith(db.SORT_KEY_SESSIONS.OTP("")))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()

	if err != nil {
		return nil, err
	}

	response, err := client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		Limit:                     aws.Int32(5),
		ScanIndexForward:          aws.Bool(false),
	})

	if err != nil {
		return nil, err
	}

	if len(response.Items) < 1 {
		return nil, errors.New("no otp found")
	}

	OTPs := []string{}

	otpMap := []struct {
		OTP string `dynamodbav:"SK"`
	}{}

	err = attributevalue.UnmarshalListOfMaps(response.Items, &otpMap)

	if err != nil {
		return nil, err
	}

	for _, o := range otpMap {
		OTPs = append(OTPs, strings.Trim(o.OTP, "OTP#"))
	}

	return OTPs, nil
}
