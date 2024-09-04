package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Response structure to return a JSON response
type Response struct {
	StatusCode int               `json:"statusCode"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

func handler(_ context.Context, request events.APIGatewayProxyRequest) (Response, error) {
	fmt.Println("Hi %s!", request)

	// Create a simple response body
	responseBody := Response{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(request.Body),
	}

	return responseBody, nil
}

func main() {


	triggerHandler := lambda.NewHandler(handler)

	lambda.Start(triggerHandler)

}
