package config

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	AWS_REGION              string
	JWT_SECRET_KEY          string
	SQS_QUEUE_URL           string
	DDB_MAIN_TABLE_NAME     string
	DDB_SESSIONS_TABLE_NAME string

	OTP_EXPIRY_TIME_IN_MIN   = 5
	JWT_TOKEN_EXPIRY_IN_DAYS = 10
	USER_SESSION_EXPIRY_DAYS = 30
)

func Init() {
	_, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	AWS_REGION = os.Getenv("AWS_REGION")
	JWT_SECRET_KEY = os.Getenv("JWT_SECRET_KEY")
	SQS_QUEUE_URL = os.Getenv("SQS_QUEUE_URL")
	DDB_MAIN_TABLE_NAME = os.Getenv("DDB_TABLE_NAME")
	DDB_SESSIONS_TABLE_NAME = os.Getenv("DDB_SESSIONS_TABLE_NAME")
}
