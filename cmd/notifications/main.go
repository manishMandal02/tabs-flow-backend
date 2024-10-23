package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/notifications"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func main() {

	// load config
	config.Init()

	baseMux := http.NewServeMux()

	baseMux.HandleFunc("/notifications/", notifications.Router)

	handler := http_api.NewAPIGatewayHandlerWithSQSHandler(baseMux, notifications.EventsHandler)

	lambda.Start(handler.Handle)

}
