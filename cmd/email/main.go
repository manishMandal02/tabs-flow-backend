package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/internal/email"
)

func main() {

	lambda.Start(email.SendEmail)
}
