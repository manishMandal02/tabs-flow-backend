package http_api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

const (
	ErrorInvalidRequest        = "Invalid request"
	ErrorRouteNotFound         = "Route not found"
	ErrorMethodNotAllowed      = "Method not allowed"
	ErrorEmptyLambdaEvent      = "Empty lambda event"
	ErrorCouldNotMarshalItem   = "Could not marshal item"
	ErrorCouldNotUnMarshalItem = "Could not  unmarshal item"
)

type ErrorBody struct {
	ErrorMsg *string `json:"error,omitempty"`
}

type RespBody struct {
	Message string      `json:"message"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
}

func APIResponse(statusCode int, body interface{}) (*events.APIGatewayV2HTTPResponse, error) {
	resp := events.APIGatewayV2HTTPResponse{Headers: map[string]string{"Content-Type": "application/json"}}
	resp.StatusCode = statusCode

	stringBody, err := json.Marshal(body)
	if err != nil {
		resp.StatusCode = 500
		logger.Error("Error marshalling response body", err)
		resp.Body = `{"error": "Internal Server Error"}`
		return &resp, nil
	}
	resp.Body = string(stringBody)
	return &resp, nil
}
func APIResponseWithCookies(statusCode int, body interface{}, cookies map[string]string) (*events.APIGatewayV2HTTPResponse, error) {
	resp := events.APIGatewayV2HTTPResponse{Headers: map[string]string{"Content-Type": "application/json"}}
	resp.StatusCode = statusCode

	if cookies != nil {
		cookieStrings := make([]string, 0, len(cookies))
		for key, value := range cookies {
			cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s; HttpOnly; Secure; SameSite=Strict", key, value))
		}
		resp.MultiValueHeaders = map[string][]string{
			"Set-Cookie": {strings.Join(cookieStrings, ", ")},
		}
	}

	stringBody, _ := json.Marshal(body)
	resp.Body = string(stringBody)
	return &resp, nil
}
