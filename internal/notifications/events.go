package notifications

import (
	lambda_events "github.com/aws/aws-lambda-go/events"
)

func EventsHandler(event lambda_events.SQSEvent) error {
	// TODO: handle multiple events process
	return nil
}
