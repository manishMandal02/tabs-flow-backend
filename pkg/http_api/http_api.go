package http_api

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
)

const (
	ErrorInvalidRequest        = "invalid request"
	ErrorMethodNotAllowed      = "method not allowed"
	ErrorEmptyLambdaEvent      = "empty lambda event"
	ErrorCouldNotMarshalItem   = "could not marshal item"
	ErrorCouldNotUnMarshalItem = "could not  unmarshal item"
)

type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty"`
}

type RespBody struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

func APIResponse(statusCode int, body interface{}) *events.APIGatewayV2HTTPResponse {
	resp := events.APIGatewayV2HTTPResponse{Headers: map[string]string{"Content-Type": "application/json"}}
	resp.StatusCode = statusCode

	stringBody, _ := json.Marshal(body)
	resp.Body = string(stringBody)
	return &resp
}
