package user

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

// TODO - handle get user by email in auth lambda, as it requires cognito access

// if req.Body != "" && strings.Contains(req.Body, "email") {
// // get user by email
// 	var evBody struct {
// 		Email string `json:"email"`
// 	}

// 	err := json.Unmarshal([]byte(req.Body), &evBody)

// 	if err != nil {
// 		return http_api.APIResponse(500, `{"message":  "invalid email" }`)
// 	}

// 	return handler.userByEmail(evBody.Email)
// }

func Routes(req events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {

	db := database.New()

	ur := newUserRepository(*db)

	handler := newUserHandler(ur)

	if req.Resource != "/users" {
		return http_api.APIResponse(404, http_api.RespBody{Message: http_api.ErrorInvalidRequest, Success: false})
	}

	if req.HTTPMethod == "GET" {
		id := req.PathParameters["id"]

		if id == "" {
			return http_api.APIResponse(400, http_api.RespBody{Message: http_api.ErrorInvalidRequest, Success: false})
		}

		// get user by id
		return handler.userById("s")
	}

	if req.HTTPMethod == "POST" {
		// create user
		return handler.upsertUser(req.Body, true)
	}

	if req.HTTPMethod == "Patch" {
		// update user
		return handler.upsertUser(req.Body, false)
	}

	if req.HTTPMethod == "DELETE" {
		id := req.PathParameters["id"]

		if id == "" {
			return http_api.APIResponse(400, http_api.RespBody{Message: http_api.ErrorInvalidRequest, Success: false})
		}

		// delete user
		return handler.deleteUser(id)

	}

	return http_api.APIResponse(404, http_api.RespBody{Message: http_api.ErrorMethodNotAllowed, Success: false})
}

// user api routes
// GET users/{id}
// POST users
// PUT users/{id}
// DELETE users/{id}
