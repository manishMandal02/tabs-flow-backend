package http_api

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

const (
	ErrorMethodNotAllowed      = "Method not allowed"
	ErrorEmptyLambdaEvent      = "Empty lambda event"
	ErrorCouldNotMarshalItem   = "Could not marshal item"
	ErrorCouldNotUnMarshalItem = "Could not  unmarshal item"
)

type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty"`
}

func APIResponse(statusCode int, body interface{}) *events.APIGatewayProxyResponse {
	resp := events.APIGatewayProxyResponse{Headers: map[string]string{"Content-Type": "application/json"}}
	resp.StatusCode = statusCode

	stringBody, _ := json.Marshal(body)
	resp.Body = string(stringBody)
	return &resp
}
