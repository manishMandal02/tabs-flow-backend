package auth

import (
	"strings"

	lambda_events "github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

// user agent
// req.requestContext.http.userAgent

func Routes(req lambda_events.APIGatewayV2HTTPRequest) *lambda_events.APIGatewayV2HTTPResponse {

	db := database.New()

	ar := newAuthRepository(db)

	handler := newAuthHandler(ar)

	if !strings.Contains(req.RawPath, "/users/") {
		return http_api.APIResponse(404, http_api.RespBody{Message: http_api.ErrorInvalidRequest, Success: false})
	}

	reqMethod := req.RequestContext.HTTP.Method

	if reqMethod == "GET" {
		if req.RawPath == "/users/logout" {
			return handler.logout()
		}

		if req.RawPath == "/users/verify-otp" {
			ua := req.RequestContext.HTTP.UserAgent
			return handler.verifyOTP(req.Body, ua)
		}
	}

	if reqMethod == "POST" {
		if req.RawPath == "/users/login" {
			ua := req.RequestContext.HTTP.UserAgent
			return handler.googleAuth(req.Body, ua)
		}

		if req.RawPath == "/users/send-otp" {
			return handler.sendOTP(req.Body)
		}
	}

	return http_api.APIResponse(404, http_api.RespBody{Message: http_api.ErrorMethodNotAllowed, Success: false})

}
