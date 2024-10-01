package main

import (
	"net/http"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/users"
)

func main() {
	// load config
	config.Init()

	http.HandleFunc("/users/", users.Router)

	lambda.Start(httpadapter.New(http.DefaultServeMux).ProxyWithContext)

}
