package user

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

type userHandler struct {
	ur userRepository
}

var errors = map[string]string{
	"error_creating_user": "error creating user",
}

func newUserHandler(ur userRepository) *userHandler {
	return &userHandler{
		ur: ur,
	}
}

func (h *userHandler) userById(id string) *events.APIGatewayProxyResponse {
	user, err := h.ur.getUserByID(id)
	if err != nil {
		return http_api.APIResponse(500, http_api.ErrorCouldNotMarshalItem)
	}

	return http_api.APIResponse(200, user)
}

func (h *userHandler) userByEmail(email string) *events.APIGatewayProxyResponse {
	user, err := h.ur.getUserByEmail(email)
	if err != nil {
		return http_api.APIResponse(500, http_api.ErrorCouldNotMarshalItem)
	}
	return http_api.APIResponse(200, user)
}

func (h *userHandler) createUser(user *User) *events.APIGatewayProxyResponse {
	err := h.ur.createUser(user)

	if err != nil {
		return http_api.APIResponse(500, "{'message':  'Error creating user' }")
	}
	return http_api.APIResponse(200, "user created")
}

func (h *userHandler) updateUser(user *User) *events.APIGatewayProxyResponse {
	err := h.ur.updateUser(user)

	if err != nil {
		return http_api.APIResponse(500, "{'message':  'Error creating user' }")
	}
	return http_api.APIResponse(200, "user created")
}

// not found handler
func (h *userHandler) notFound() *events.APIGatewayProxyResponse {

	return http_api.APIResponse(404, http_api.ErrorMethodNotAllowed)
}
