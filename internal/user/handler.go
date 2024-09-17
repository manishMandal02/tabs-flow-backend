package user

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type userHandler struct {
	r userRepository
}

func newUserHandler(r userRepository) *userHandler {
	return &userHandler{
		r: r,
	}
}

func (h *userHandler) userById(id string) *events.APIGatewayV2HTTPResponse {
	user, err := h.r.getUserByID(id)
	if err != nil {
		if errors.Is(err, errors.New(errMsg.UserNotFound)) {
			return http_api.APIResponse(404, http_api.RespBody{Success: false, Message: errMsg.UserNotFound})
		}

		return http_api.APIResponse(500, http_api.RespBody{Success: false, Message: errMsg.GetUser})
	}

	return http_api.APIResponse(200, http_api.RespBody{Success: true, Data: user})

}

func (h *userHandler) createUser(body string) *events.APIGatewayV2HTTPResponse {

	var user *User

	err := user.fromJSON(body)

	if err != nil {
		logger.Error(fmt.Sprintf("error decoding user from JSON body: %v", body), err)
		return http_api.APIResponse(400, http_api.RespBody{Success: false, Message: errMsg.CreateUser})
	}

	err = user.validate()

	if err != nil {
		logger.Error(fmt.Sprintf("error validating user: %v", body), err)
		return http_api.APIResponse(400, http_api.RespBody{Success: false, Message: err.Error()})
	}

	err = h.r.insertUser(user)

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Success: false, Message: errMsg.CreateUser})

	}

	return http_api.APIResponse(201, http_api.RespBody{Message: "user created", Success: true})

}

func (h *userHandler) updateUser(id, body string) *events.APIGatewayV2HTTPResponse {

	var err error
	var n struct {
		Name string `json:"name"`
	}

	err = json.Unmarshal([]byte(body), &n)

	if err != nil {
		logger.Error(fmt.Sprintf("error un_marshaling name from JSON body: %v", body), err)
		return http_api.APIResponse(400, http_api.RespBody{Success: false, Message: errMsg.UpdateUser})
	}

	err = h.r.updateUser(id, n.Name)

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Success: false, Message: errMsg.UpdateUser})
	}

	return http_api.APIResponse(200, http_api.RespBody{Success: true, Message: "user updated"})
}

func (h *userHandler) deleteUser(id string) *events.APIGatewayV2HTTPResponse {
	err := h.r.deleteAccount(id)

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Success: false, Message: errMsg.DeleteUser})
	}

	return http_api.APIResponse(200, http_api.RespBody{Message: "user deleted", Success: true})
}
