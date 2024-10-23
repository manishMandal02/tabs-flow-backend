package http_api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/awslabs/aws-lambda-go-api-proxy/httpadapter"
)

type sqsHandler func(context.Context, events.SQSEvent) (interface{}, error)

// handles API Gateway proxy events
type APIGatewayHandler struct {
	baseMux    *http.ServeMux // Original multiplexer with routes
	customMux  *http.ServeMux // Wrapped multiplexer with auth injection
	sqsHandler sqsHandler
}

func NewAPIGatewayHandler(baseMux *http.ServeMux) *APIGatewayHandler {
	return &APIGatewayHandler{
		baseMux:   baseMux,
		customMux: http.NewServeMux(),
	}
}
func NewAPIGatewayHandlerWithSQSHandler(baseMux *http.ServeMux, sH sqsHandler) *APIGatewayHandler {
	return &APIGatewayHandler{
		baseMux:    baseMux,
		customMux:  http.NewServeMux(),
		sqsHandler: sH,
	}
}

// inject UserId header
func (h *APIGatewayHandler) injectUserId(userId string) {
	// Wrap all routes with auth injection
	h.customMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("UserId", userId)
		h.baseMux.ServeHTTP(w, r)
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
				return h.sqsHandler(ctx, events.SQSEvent{})
			}
		}
		return nil, err
	}

	// Extract userId from authorizer context
	userId, ok := apiEvent.RequestContext.Authorizer["UserId"].(string)
	if ok {
		h.injectUserId(userId)
	}

	// serve the request
	adapter := httpadapter.New(h.customMux)
	return adapter.Proxy(apiEvent)
}
