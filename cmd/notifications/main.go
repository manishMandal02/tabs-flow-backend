package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/notifications"
)

type sqsHandler func(context.Context, events.SQSEvent) (interface{}, error)

type lambdaHandler func(context.Context, json.RawMessage) (interface{}, error)

func newLambdaHandler(h sqsHandler) lambdaHandler {

	return func(ctx context.Context, event json.RawMessage) (interface{}, error) {

		// Try to parse the event as an API Gateway proxy request
		var apiGatewayEvent events.APIGatewayProxyRequest
		if err := json.Unmarshal(event, &apiGatewayEvent); err == nil && apiGatewayEvent.RequestContext.APIID != "" {
			// This is an API Gateway event
			adapter := httpadapter.New(http.DefaultServeMux)
			return adapter.ProxyWithContext(ctx, apiGatewayEvent)
		}

		// Try to parse the event as an SQS event
		var sqsEvent events.SQSEvent
		if err := json.Unmarshal(event, &sqsEvent); err == nil && len(sqsEvent.Records) > 0 {
			// This is an SQS event
			return h(ctx, sqsEvent)
		}

		// If we can't determine the event type, return an error
		return nil, errors.New("unknown event type")

	}
}

func main() {

	// load config
	config.Init()

	// add router to handle api gateway events
	http.HandleFunc("/notifications/", notifications.Router)

	// add event handler to handler sqs messages
	handler := newLambdaHandler(notifications.EventsHandler)

	lambda.Start(handler)

}
