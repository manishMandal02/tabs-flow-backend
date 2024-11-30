package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/notifications"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func main() {

	// load config
	config.Init()

	// func EventsHandler(_ context.Context, event lambda_events.SQSEvent) (interface{}, error) {

	queue := events.NewNotificationQueue()

	ddb := db.New()

	sqsHandler := notifications.SQSMessagesHandler(queue)

	handler := http_api.NewAPIGatewayHandlerWithSQSHandler("/notifications/", notifications.Router(ddb), sqsHandler)

	lambda.Start(handler.Handle)

}
