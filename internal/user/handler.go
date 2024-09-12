package user

import (
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

func (h *userHandler) userById(id string) *events.APIGatewayProxyResponse {
	user, err := h.r.getUserByID(id)
	if err != nil {
		if errors.Is(err, errors.New(errMsg.UserNotFound)) {
			return http_api.APIResponse(404, http_api.RespBody{Message: errMsg.UserNotFound, Success: false})
		}

		return http_api.APIResponse(500, http_api.RespBody{Message: errMsg.GetUser, Success: false})
	}

	return http_api.APIResponse(200, http_api.RespBody{Success: true, Data: user})

}

func (h *userHandler) upsertUser(body string, isCreateReq bool) *events.APIGatewayProxyResponse {

	var user *User

	errResponseMsg := errMsg.CreateUser
	successResponseMsg := "user created"

	if !isCreateReq {
		errResponseMsg = errMsg.UpdateUser
		successResponseMsg = "user updated"
	}

	err := user.fromJSON(body)

	if err != nil {
		logger.Error(fmt.Sprintf("error decoding user from JSON body: %v", body), err)
		return http_api.APIResponse(400, http_api.RespBody{Message: errResponseMsg, Success: false})
	}

	err = user.validate()

	if err != nil {
		logger.Error(fmt.Sprintf("error validating user: %v", body), err)
		return http_api.APIResponse(400, http_api.RespBody{Message: err.Error(), Success: false})
	}

	err = h.r.upsertUser(user)

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Message: errResponseMsg, Success: false})

	}

	return http_api.APIResponse(201, http_api.RespBody{Message: successResponseMsg, Success: true})

}

func (h *userHandler) deleteUser(id string) *events.APIGatewayProxyResponse {
	err := h.r.deleteAccount(id)

	if err != nil {
		return http_api.APIResponse(500, http_api.RespBody{Message: errMsg.DeleteUser, Success: false})
	}

	return http_api.APIResponse(200, http_api.RespBody{Message: "user deleted", Success: true})
}
