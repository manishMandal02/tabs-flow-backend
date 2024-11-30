package http_api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	lambda_events "github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

type SQSHandler func(messages []lambda_events.SQSMessage) (interface{}, error)

// API Gateway proxy events handler
type APIGatewayHandler struct {
	baseURL    string
	handler    http.Handler
	sqsHandler SQSHandler
}

func NewAPIGatewayHandler(baseURL string, handler http.Handler) *APIGatewayHandler {
	return &APIGatewayHandler{
		baseURL: baseURL,
		handler: handler,
	}
}

func NewAPIGatewayHandlerWithSQSHandler(baseURL string, handler http.Handler, sH SQSHandler) *APIGatewayHandler {
	return &APIGatewayHandler{
		baseURL:    baseURL,
		handler:    handler,
		sqsHandler: sH,
	}
}

// wrapper handler that injects the userId
func (h *APIGatewayHandler) withUserID(userId string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("UserId", userId)
		handler.ServeHTTP(w, r)
	})
}

// processes the Lambda api event
func (h *APIGatewayHandler) Handle(ctx context.Context, event json.RawMessage) (interface{}, error) {
	// Parse API GW event
	var apiEvent events.APIGatewayProxyRequest
	err := json.Unmarshal(event, &apiEvent)
	if err != nil || apiEvent.RequestContext.APIID == "" {
		// not a valid API Gateway event
		// check if it is an SQS event
		if h.sqsHandler != nil {
			// Try to parse the event as an SQS event
			var sqsEvent events.SQSEvent
			if err := json.Unmarshal(event, &sqsEvent); err == nil && len(sqsEvent.Records) > 0 {
				// This is an SQS event
				return h.sqsHandler(sqsEvent.Records)
			}
		}
		return nil, err
	}

	// Create mux for this request
	mux := http.NewServeMux()

	// Extract userId from authorizer context
	if userId, ok := apiEvent.RequestContext.Authorizer["UserId"].(string); ok {
		// Wrap the handler with userId injection
		mux.Handle(h.baseURL, h.withUserID(userId, h.handler))
	} else {
		// Use original handler without userId injection
		mux.Handle(h.baseURL, h.handler)
	}

	// serve the request
	adapter := httpadapter.New(mux)
	return adapter.Proxy(apiEvent)
}
