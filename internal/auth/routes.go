package auth

import (
	"fmt"
	"net/http"
	"strings"

	lambda_events "github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

// custom API_GW lambda authorizer
func LambdaAuthorizer(ev *lambda_events.APIGatewayCustomAuthorizerRequestTypeRequest) (lambda_events.APIGatewayCustomAuthorizerResponse, error) {
	logger.Dev(fmt.Sprintf("Lambda Authorizer event: %v", ev))

	db := database.NewSessionTable()
	ar := newAuthRepository(db)

	handler := newAuthHandler(ar)

	return handler.lambdaAuthorizer(ev)
}

func Router(w http.ResponseWriter, r *http.Request) {
	logger.Dev(fmt.Sprintf("Auth req: method: %v, path: %v", r.Method, r.URL.Path))

	db := database.NewSessionTable()

	ar := newAuthRepository(db)

	handler := newAuthHandler(ar)

	authRouter := http_api.NewRouter("/auth")

	authRouter.AddRoute("POST", "/verify-otp", handler.verifyOTP)

	authRouter.AddRoute("POST", "/send-otp", handler.sendOTP)

	authRouter.ServeHTTP(w, r)
}

// handle API routes
func Routes(req lambda_events.APIGatewayV2HTTPRequest) (*lambda_events.APIGatewayV2HTTPResponse, error) {
	logger.Dev(fmt.Sprintf("Auth lambda req: %v", req))

	db := database.NewSessionTable()

	ar := newAuthRepository(db)

	handler := newAuthHandler(ar)

	if !strings.Contains(req.RawPath, "/auth/") {
		return http_api.APIResponse(404, http_api.RespBody{Message: http_api.ErrorInvalidRequest, Success: false})
	}

	reqMethod := req.RequestContext.HTTP.Method

	if reqMethod == "GET" {

		if req.RawPath == "/auth/verify-otp" {
			// ua := req.RequestContext.HTTP.UserAgent
			// return handler.verifyOTP(req.Body, ua)
		}

		if req.RawPath == "/auth/user-id" {
			return handler.getUserId(req.Body)
		}
		if req.RawPath == "/auth/logout" {
			return handler.logout(req.Cookies)
		}

	}

	if reqMethod == "POST" {
		if req.RawPath == "/auth/google" {
			ua := req.RequestContext.HTTP.UserAgent
			return handler.googleAuth(req.Body, ua)
		}

		if req.RawPath == "/auth/send-otp" {
			// return handler.sendOTP(req.Body)s
		}
	}

	return http_api.APIResponse(404, http_api.RespBody{Message: http_api.ErrorMethodNotAllowed, Success: false})

}
