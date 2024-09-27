package config

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	AWS_REGION              string
	JWT_SECRET_KEY          string
	EMAIL_SQS_QUEUE_URL     string
	DDB_MAIN_TABLE_NAME     string
	DDB_SESSIONS_TABLE_NAME string
	ZEPTO_MAIL_API_KEY      string
	ENVIRONMENT             string

	ZEPTO_MAIL_API_URL = "https://api.zeptomail.com/v1.1/email"

	OTP_EXPIRY_TIME_IN_MIN   = 5
	JWT_TOKEN_EXPIRY_IN_DAYS = 10
	USER_SESSION_EXPIRY_DAYS = 10
)

func Init() {
	_, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	env := flag.String("env", "lambda", "Environment")

	flag.Parse()

	ENVIRONMENT = *env
	AWS_REGION = os.Getenv("AWS_REGION")
	JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")
	EMAIL_SQS_QUEUE_URL = os.Getenv("EMAIL_SQS_QUEUE_URL")
	DDB_MAIN_TABLE_NAME = os.Getenv("DDB_TABLE_NAME")
	ZEPTO_MAIL_API_KEY = os.Getenv("ZEPTO_MAIL_API_KEY")
	DDB_SESSIONS_TABLE_NAME = os.Getenv("DDB_SESSIONS_TABLE_NAME")
}
