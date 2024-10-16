package config

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/joho/godotenv"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

var (
	AWS_REGION                  string
	JWT_SECRET_KEY              string
	EMAIL_QUEUE_URL             string
	NOTIFICATIONS_QUEUE_URL     string
	NOTIFICATIONS_QUEUE_ARN     string
	SCHEDULER_ROLE_ARN          string
	DDB_MAIN_TABLE_NAME         string
	DDB_SEARCH_INDEX_TABLE_NAME string
	DDB_SESSIONS_TABLE_NAME     string
	ZEPTO_MAIL_API_KEY          string
	AWS_CONFIG                  aws.Config
	PADDLE_API_KEY              string
	PADDLE_WEBHOOK_SECRET_KEY   string
	VAPID_PRIVATE_KEY           string
	VAPID_PUBLIC_KEY            string

	ZEPTO_MAIL_API_URL       = "https://api.zeptomail.in/v1.1/email/template"
	DATE_TIME_FORMAT         = "2006-01-02T15:04:05"
	LOCAL_DEV_ENV            = false
	TRAIL_DAYS               = 14
	OTP_EXPIRY_TIME_IN_MIN   = 5
	JWT_TOKEN_EXPIRY_IN_DAYS = 10
	USER_SESSION_EXPIRY_DAYS = 10
)

func Init() {

	localDevFlag := flag.Bool("local_dev", false, "local development mode")

	flag.Parse()

	isLocalDev := *localDevFlag

	if isLocalDev {
		logger.Dev("Local development mode 🚧")
		LOCAL_DEV_ENV = true
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}

		// local development config
		profile := os.Getenv("AWS_ACCOUNT_PROFILE")

		cfg, err := config.LoadDefaultConfig(
			context.Background(),
			config.WithSharedConfigProfile(profile),
		)

		if err != nil {
			log.Fatalf("failed to load configuration, %v", err)
		}
		AWS_CONFIG = cfg

		DDB_MAIN_TABLE_NAME = "TabsFlow-Main_dev"
		DDB_SEARCH_INDEX_TABLE_NAME = "TabsFlow-SearchIndex_dev"
		DDB_SESSIONS_TABLE_NAME = "TabsFlow-Sessions_dev"
		EMAIL_QUEUE_URL = "TabsFlow-Emails_dev"
		NOTIFICATIONS_QUEUE_URL = "TabsFlow-Notifications_dev"
		// TODO - set role arn for scheduler
		SCHEDULER_ROLE_ARN = "arn:aws:iam::410381041692:role/TabsFlow-SchedulerRole_dev"
		NOTIFICATIONS_QUEUE_ARN = "arn:aws:sqs:ap-south-1:410381041692:TabsFlow-Notifications_dev"
	} else {
		// lambda config
		config, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(AWS_REGION))
		if err != nil {
			log.Fatalf("failed to load configuration, %v", err)
		}

		AWS_CONFIG = config

		DDB_MAIN_TABLE_NAME = os.Getenv("DDB_MAIN_TABLE_NAME")
		DDB_SESSIONS_TABLE_NAME = os.Getenv("DDB_SESSIONS_TABLE_NAME")
		EMAIL_QUEUE_URL = os.Getenv("EMAIL_QUEUE_URL")
		NOTIFICATIONS_QUEUE_URL = os.Getenv("NOTIFICATIONS_QUEUE_URL")
		SCHEDULER_ROLE_ARN = os.Getenv("SCHEDULER_ROLE_ARN")
		NOTIFICATIONS_QUEUE_ARN = os.Getenv("NOTIFICATIONS_QUEUE_ARN")
	}

	AWS_REGION = os.Getenv("AWS_REGION")
	JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")
	ZEPTO_MAIL_API_KEY = os.Getenv("ZEPTO_MAIL_API_KEY")
	PADDLE_API_KEY = os.Getenv("PADDLE_API_KEY")
	PADDLE_WEBHOOK_SECRET_KEY = os.Getenv("PADDLE_WEBHOOK_SECRET_KEY")
	VAPID_PRIVATE_KEY = os.Getenv("VAPID_PRIVATE_KEY")
	VAPID_PUBLIC_KEY = os.Getenv("VAPID_PUBLIC_KEY")
}
