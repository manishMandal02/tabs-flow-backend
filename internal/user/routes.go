package user

import (
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Routes(req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {

	db := database.New()

	ur := newUserRepository(db)

	handler := newUserHandler(ur)

	if !strings.Contains(req.RawPath, "/users/") {
		return http_api.APIResponse(404, http_api.RespBody{Message: http_api.ErrorInvalidRequest, Success: false})
	}

	reqMethod := req.RequestContext.HTTP.Method

	if reqMethod == "GET" {
		id := req.PathParameters["id"]

		if id == "" {
			return http_api.APIResponse(400, http_api.RespBody{Message: http_api.ErrorInvalidRequest, Success: false})
		}

		// get user by id
		return handler.userById("s")
	}

	if reqMethod == "POST" {
		// create user
		return handler.createUser(req.Body)
	}

	if reqMethod == "Patch" {
		id := req.PathParameters["id"]
		if id == "" {
			return http_api.APIResponse(400, http_api.RespBody{Message: http_api.ErrorInvalidRequest, Success: false})
		}
		// update user
		return handler.updateUser(id, req.Body)
	}

	if reqMethod == "DELETE" {
		id := req.PathParameters["id"]

		if id == "" {
			return http_api.APIResponse(400, http_api.RespBody{Message: http_api.ErrorInvalidRequest, Success: false})
		}

		// delete user
		return handler.deleteUser(id)

	}

	return http_api.APIResponse(404, http_api.RespBody{Message: http_api.ErrorMethodNotAllowed, Success: false})
}
