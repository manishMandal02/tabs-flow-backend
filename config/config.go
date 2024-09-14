package config

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	AWS_REGION              string
	DDB_MAIN_TABLE_NAME     string
	DDB_SESSIONS_TABLE_NAME string
)

func Init() {
	_, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	AWS_REGION = os.Getenv("AWS_REGION")
	DDB_MAIN_TABLE_NAME = os.Getenv("DDB_TABLE_NAME")
	DDB_SESSIONS_TABLE_NAME = os.Getenv("DDB_SESSIONS_TABLE_NAME")

}
