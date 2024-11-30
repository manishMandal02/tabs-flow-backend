package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/auth"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func main() {

	// load config
	config.Init()

	ddb := db.NewSessionTable()

	queue := events.NewEmailQueue()

	handler := http_api.NewAPIGatewayHandler("/auth/", auth.Router(ddb, queue))

	lambda.Start(handler.Handle)
}
