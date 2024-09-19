package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/auth"
)

func main() {

	// load config
	config.Init()

	lambda.Start(auth.Routes)
}
