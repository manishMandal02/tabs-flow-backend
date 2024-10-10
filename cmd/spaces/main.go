package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/spaces"
)

func main() {

	// load config
	config.Init()

	http.HandleFunc("/users/", spaces.Router)

	lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)

}
