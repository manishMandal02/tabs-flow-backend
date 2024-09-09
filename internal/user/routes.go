package user

import (
	"encoding/json"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

func Routes(req events.APIGatewayProxyRequest) *events.APIGatewayProxyResponse {

	db := database.New()

	ur := newUserRepository(*db)

	handler := newUserHandler(ur)

	if req.Resource != "/users" {
		return handler.notFound()
	}

	if req.HTTPMethod == "GET" {
		if req.Body != "" && strings.Contains(req.Body, "email") {
			// get user by email
			var evBody struct {
				Email string `json:"email"`
			}

			err := json.Unmarshal([]byte(req.Body), &evBody)

			if err != nil {
				return http_api.APIResponse(500, `{"message":  "invalid email" }`)
			}

			return handler.userByEmail(evBody.Email)
		}
		// get user by id
		return handler.userById("s")
	}

	if req.HTTPMethod == "POST" {
		//TODO - validate req body
		var createUser User

		err := json.Unmarshal([]byte(req.Body), &createUser)

		if err != nil {
			return http_api.APIResponse(500, `{"message":  "error creating user" }`)
		}

		// create user
		return handler.createUser(&createUser)
	}

	if req.HTTPMethod == "Patch" {
		//TODO - validate req body
		var updateUser User

		err := json.Unmarshal([]byte(req.Body), &updateUser)

		if err != nil {
			return http_api.APIResponse(500, `{"message":  "error updating user" }`)
		}

		return handler.updateUser(&updateUser)

	}

	return handler.notFound()
}

// user api routes
// GET users/{id}
// GET users/find ~body: {email}
// POST users
// PUT users/{id}
// DELETE users/{id}
