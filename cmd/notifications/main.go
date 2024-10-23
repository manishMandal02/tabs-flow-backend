package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/notifications"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func main() {

	// load config
	config.Init()

	handler := http_api.NewAPIGatewayHandlerWithSQSHandler("/notifications/", notifications.Router(), notifications.EventsHandler)

	lambda.Start(handler.Handle)

}
