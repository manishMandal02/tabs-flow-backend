package auth

import (
	"net/http"

	lambda_events "github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

// custom API_GW lambda authorizer
func LambdaAuthorizer(ev *lambda_events.APIGatewayCustomAuthorizerRequestTypeRequest) (*lambda_events.APIGatewayCustomAuthorizerResponse, error) {

	db := database.NewSessionTable()
	ar := newAuthRepository(db)

	handler := newAuthHandler(ar)

	return handler.lambdaAuthorizer(ev)
}

func Router(w http.ResponseWriter, r *http.Request) {

	db := database.NewSessionTable()

	ar := newAuthRepository(db)

	handler := newAuthHandler(ar)

	authRouter := http_api.NewRouter("/auth")

	// authRouter("/", handler.getUserId)

	authRouter.POST("/verify-otp", handler.verifyOTP)

	authRouter.POST("/send-otp", handler.sendOTP)

	authRouter.POST("/google", handler.googleAuth)

	authRouter.GET("/logout", handler.logout)

	authRouter.GET("/user", handler.getUserId)

	// serve API routes
	authRouter.ServeHTTP(w, r)
}
