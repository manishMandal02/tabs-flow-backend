package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/users"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func main() {
	// load config
	config.Init()

	handler := http_api.NewAPIGatewayHandler("/users/", users.Router())

	lambda.Start(handler.Handle)

}
