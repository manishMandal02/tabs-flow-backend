package e2e_test

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/joho/godotenv"
	"github.com/manishMandal02/tabsflow-backend/internal/users"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/stretchr/testify/suite"
)

var TestUser = users.User{
	Email:      "mmjdd67@gmail.com",
	FullName:   "Manish Mandal",
	ProfilePic: "https://avatars.githubusercontent.com/u/123456789?v=4",
}

type ENV struct {
	ApiDomainName         string
	AWS_ACCOUNT_PROFILE   string
	MainTable             string
	SessionTable          string
	SearchIndexTable      string
	EmailQueueURL         string
	NotificationsQueueURL string
}

type SQSEventSourceMappingUUID struct {
	EmailService        string
	NotificationService string
}

type UserFlowTestSuite struct {
	suite.Suite
	ENV              ENV
	AppURL           *url.URL
	SessionCookie    *http.Cookie
	AWSConfig        *aws.Config
	SQSClient        *sqs.Client
	DDBClient        *dynamodb.Client
	LambdaClient     *lambda.Client
	SQSLambdaMapping SQSEventSourceMappingUUID
}

func (s *UserFlowTestSuite) SetupSuite() {
	env := getENVs()

	s.ENV = env

	appURL, err := url.Parse("https://tabsflow.com")

	if err != nil {
		panic(err)
	}

	s.AppURL = appURL

	awsConfig := configureAWS(env.AWS_ACCOUNT_PROFILE)

	s.AWSConfig = awsConfig

	s.SQSClient = sqs.NewFromConfig(*awsConfig)
	s.DDBClient = dynamodb.NewFromConfig(*awsConfig)
	s.LambdaClient = lambda.NewFromConfig(*s.AWSConfig)

	e, err1 := s.LambdaClient.ListEventSourceMappings(context.TODO(), &lambda.ListEventSourceMappingsInput{
		FunctionName: aws.String("EmailService_test"),
	})

	n, err2 := s.LambdaClient.ListEventSourceMappings(context.TODO(), &lambda.ListEventSourceMappingsInput{
		FunctionName: aws.String("NotificationsService_test"),
	})

	if err1 != nil || err2 != nil {
		panic("Failed to list event source mappings")
	}

	s.SQSLambdaMapping = SQSEventSourceMappingUUID{
		EmailService:        *e.EventSourceMappings[0].UUID,
		NotificationService: *n.EventSourceMappings[0].UUID,
	}

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

func getSQSQueueMessage[T any](client *sqs.Client, queueURL string) (*events.Event[T], error) {
	// retry 3 times
	for i := 0; i < 3; i++ {
		output, err := client.ReceiveMessage(context.TODO(), &sqs.ReceiveMessageInput{
			QueueUrl:            aws.String(queueURL),
			MaxNumberOfMessages: 2,
			WaitTimeSeconds:     2,
		})

		if err != nil {
			return nil, err
		}

		logger.Dev("Received %v messages.", len(output.Messages))

		if len(output.Messages) == 0 {
			continue
		}

		if len(output.Messages) > 1 {
			return nil, errors.New("more than one message found")
		}

		for _, msg := range output.Messages {
			ev, err := events.NewFromJSON[T](*msg.Body)

			logger.Dev("Received event: %v", ev)

			if err != nil {
				logger.Dev("Invalid event body: %v", ev)
				continue
			}

			return ev, nil
		}

		time.Sleep(time.Second * 500)
	}

	return nil, errors.New("no message found")

}

func updateLambdaEventSourceMappingState(lambdaClient *lambda.Client, uuid string, disable bool) error {

	_, err := lambdaClient.UpdateEventSourceMapping(context.TODO(), &lambda.UpdateEventSourceMappingInput{
		UUID:    aws.String(uuid),
		Enabled: aws.Bool(!disable),
	})

	if err != nil {
		return err

	}

	return nil
}
