package main

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/notifications"
)

func lambdaHandler(ctx context.Context) {
	// context, _ := lambdacontext.FromContext(ctx)

}

func main() {

	// load config
	config.Init()

	// TODO: call handle SQS event handler

	http.HandleFunc("/notifications/", notifications.Router)

	lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)

}
