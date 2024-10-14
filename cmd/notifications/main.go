package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/notifications"
)

func main() {

	// load config
	config.Init()

	// TODO: handle SQS events

	http.HandleFunc("/notifications/", notifications.Router)

	lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)

}
