package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/internal/email"
)

func main() {
	// TODO - load config from env
	// cfg, err := config.LoadDefaultConfig(context.TODO())
	// if err != nil {
	// 	log.Fatalf("failed to load configuration, %v", err)
	// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/
	// }
	lambda.Start(email.SendEmail)
}
