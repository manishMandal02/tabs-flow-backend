package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/manishMandal02/tabsflow-backend/internal/auth"
)

func main() {
	lambda.Start(auth.DefineAuthChallenge)
}
