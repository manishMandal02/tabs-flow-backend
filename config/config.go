package config

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	AWS_REGION              string
	SQS_QUEUE_URL           string
	DDB_MAIN_TABLE_NAME     string
	DDB_SESSIONS_TABLE_NAME string
	OTP_EXPIRY_TIME_IN_MIN  int32
)

func Init() {
	_, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	AWS_REGION = os.Getenv("AWS_REGION")
	SQS_QUEUE_URL = os.Getenv("SQS_QUEUE_URL")
	DDB_MAIN_TABLE_NAME = os.Getenv("DDB_TABLE_NAME")
	DDB_SESSIONS_TABLE_NAME = os.Getenv("DDB_SESSIONS_TABLE_NAME")
	OTP_EXPIRY_TIME_IN_MIN = 5
}
