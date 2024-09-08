package config

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
)

var (
	AWSRegion    string
	DDBTableName string
)

func init() {
	_, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("failed to load configuration, %v", err)
	}

	AWSRegion = os.Getenv("AWS_REGION")
	DDBTableName = os.Getenv("DDB_TABLE_NAME")

}
