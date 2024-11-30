package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/spaces"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func main() {

	// load config
	config.Init()

	ddb := db.New()

	queue := events.NewNotificationQueue()

	handler := http_api.NewAPIGatewayHandler("/spaces/", spaces.Router(ddb, queue))

	lambda.Start(handler.Handle)

}
