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

	mux := http.NewServeMux()

	http.HandleFunc("/", users.Router)

	lambda.Start(httpadapter.New(mux).ProxyWithContext)

}
