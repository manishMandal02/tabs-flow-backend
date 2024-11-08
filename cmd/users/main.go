package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/users"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func main() {
	// load config
	config.Init()

	ddb := db.New()

	queue := events.NewEmailQueue()

	httpClient := &http.Client{}

	paddle, err := users.NewPaddleSubscriptionClient()

	if err != nil {
		panic(err)
	}

	handler := http_api.NewAPIGatewayHandler("/users/", users.Router(ddb, queue, httpClient, paddle))

	lambda.Start(handler.Handle)

}
