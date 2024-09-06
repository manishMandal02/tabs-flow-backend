package email

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

func SendEmail(_ context.Context, event events.SQSMessage) error {

	// TODO - load config from env
	// cfg, err := config.LoadDefaultConfig(context.TODO())
	// if err != nil {
	// 	log.Fatalf("failed to load configuration, %v", err)
	// }
	// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/

	// TODO - read sqs message
	// TODO - send email through ses

	fmt.Println("sqs event: ", event)

	return nil
}
