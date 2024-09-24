package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/user"
)

func main() {

	// Initialize
	config.Init()

	// Start Lambda
	lambda.Start(user.Routes)
}
