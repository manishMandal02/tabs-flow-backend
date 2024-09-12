package user

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
)

type userHandler struct {
	r userRepository
}

func newUserHandler(r userRepository) *userHandler {
	return &userHandler{
		r: r,
	}
}

func (h *userHandler) userById(id string) *events.APIGatewayProxyResponse {
	user, err := h.r.getUserByID(id)
	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Message: http_api.ErrorCouldNotMarshalItem, Success: false})
	}

	return http_api.APIResponse(200, http_api.RespBody{Success: true, Data: user})

}

func (h *userHandler) createUser(body string) *events.APIGatewayProxyResponse {

	//TODO - validate req body

	var user *User

	err := json.Unmarshal([]byte(body), &user)

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Message: errMsg.CreateUser, Success: false})

	}

	err = h.r.upsertUser(user)

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Message: errMsg.CreateUser, Success: false})

	}

	return http_api.APIResponse(201, http_api.RespBody{Success: true, Message: "user created"})

}

func (h *userHandler) updateUser(body string) *events.APIGatewayProxyResponse {

	//TODO - validate req body
	var user *User

	err := json.Unmarshal([]byte(body), &user)

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Message: errMsg.UpdateUser, Success: false})
	}

	err = h.r.upsertUser(user)

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Message: errMsg.UpdateUser, Success: false})
	}

	return http_api.APIResponse(201, http_api.RespBody{Success: true, Message: "user updated"})

}

func (h *userHandler) deleteUser(id string) *events.APIGatewayProxyResponse {
	err := h.r.deleteAccount(id)

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Message: errMsg.DeleteUser, Success: false})

	}

	return http_api.APIResponse(200, http_api.RespBody{Success: true, Message: "user deleted"})

}
