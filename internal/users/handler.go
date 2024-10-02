package users

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

type userHandler struct {
	r userRepository
}

func newUserHandler(r userRepository) *userHandler {
	return &userHandler{
		r: r,
	}
}

// profile handlers
func (h *userHandler) userById(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	if id == "" {
		http.Error(w, errMsg.invalidUserId, http.StatusBadRequest)
		return
	}

	user, err := h.r.getUserByID(id)

	if err != nil {
		if err.Error() == errMsg.userNotFound {
			http.Error(w, errMsg.userNotFound, http.StatusBadRequest)
		} else {
			http.Error(w, errMsg.getUser, http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Data: user})
}

func (h *userHandler) createUser(w http.ResponseWriter, r *http.Request) {

	user, err := userFromJSON(r.Body)

	if err != nil {
		logger.Error("decoding user from body at createUser()", err)
		http.Error(w, errMsg.createUser, http.StatusBadRequest)
		return
	}

	logger.Dev("user: %v", user)

	err = user.validate()

	if err != nil {
		logger.Error("error validating user at createUser()", err)
		http.Error(w, errMsg.createUser, http.StatusBadRequest)
		return
	}

	//  check if the user with this id
	userExits := checkUserExits(user.Id, h.r, w)

	if !userExits {
		return
	}

	// check user id from auth service (api call)
	body := struct {
		Email string `json:"email"`
	}{
		Email: user.Email,
	}

	bodyJson, err := json.Marshal(body)

	if err != nil {
		logger.Error("Error marshaling json body", err)
		http.Error(w, errMsg.createUser, http.StatusBadRequest)
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	s := "https"

	if config.LOCAL_DEV_ENV {
		s = "http"
	}

	authServiceAPI := fmt.Sprintf("%s://%s/auth/user/", s, r.Host)

	res, respBody, err := utils.MakeHTTPRequest(http.MethodGet, authServiceAPI, headers, bodyJson)

	if err != nil {
		logger.Error(fmt.Sprintf("Error fetching user id from Auth Service for email: %v", body.Email), err)
		http.Error(w, errMsg.createUser, http.StatusInternalServerError)
		return
	}

	if res.StatusCode != 200 {
		logger.Error(fmt.Sprintf("User does not have a valid session profile for email: %v", body.Email), err)
		//  Logout
		http.Redirect(w, r, "/auth/logout", http.StatusTemporaryRedirect)
		// http.Error(w, errMsg.createUser, http.StatusInternalServerError)
		return
	}

	// check user id

	var userIdData struct {
		Data struct {
			UserId string `json:"userId"`
		} `json:"data"`
	}

	err = json.Unmarshal([]byte(respBody), &userIdData)

	if err != nil {
		logger.Error(fmt.Sprintf("Error un_marshaling user id data for email: %v", body.Email), err)
		http.Error(w, errMsg.createUser, http.StatusInternalServerError)
		return
	}
	if userIdData.Data.UserId != user.Id {
		logger.Error(fmt.Sprintf("User Id mismatch for email: %v", body.Email), err)
		http.Redirect(w, r, "/auth/logout", http.StatusTemporaryRedirect)
		return
	}

	err = h.r.insertUser(user)

	if err != nil {
		http.Error(w, errMsg.createUser, http.StatusBadRequest)
		return
	}

	// TODO - create subscription

	// TODO - create default subscription

	// send USER_REGISTERED event to email service (queue)
	event := &events.UserRegisteredPayload{
		Email:        user.Email,
		Name:         user.FullName,
		TrailEndDate: "22/12/24",
	}

	sqs := events.NewQueue()

	err = sqs.AddMessage(event)

	if err != nil {
		http.Error(w, errMsg.createUser, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Message: "user created"})
}

func (h *userHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if id == "" {
		http.Error(w, errMsg.invalidUserId, http.StatusBadRequest)
		return
	}

	//  check if the user with this id
	userExits := checkUserExits(id, h.r, w)

	if !userExits {
		return
	}

	var n struct {
		Name string `json:"fullName"`
	}

	err := json.NewDecoder(r.Body).Decode(&n)

	if err != nil {
		logger.Error("error un_marshaling name from JSON at updateUser()", err)
		http.Error(w, errMsg.updateUser, http.StatusBadRequest)
		return
	}

	err = h.r.updateUser(id, n.Name)

	if err != nil {
		http.Error(w, errMsg.updateUser, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Message: "user updated"})

}

func (h *userHandler) deleteUser(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	if id == "" {
		http.Error(w, errMsg.invalidUserId, http.StatusBadRequest)
		return
	}

	//  check if the user with this id
	userExits := checkUserExits(id, h.r, w)

	if !userExits {
		return
	}

	err := h.r.deleteAccount(id)

	if err != nil {
		http.Error(w, errMsg.deleteUser, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Message: "user deleted"})

}

// profile handlers
func (h *userHandler) getPreferences(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	if id == "" {
		http.Error(w, errMsg.invalidUserId, http.StatusBadRequest)
		return
	}

	//  check if the user with this id
	userExits := checkUserExits(id, h.r, w)

	if !userExits {
		return
	}

	http.Error(w, "Not Implemented", http.StatusNotImplemented)

}

// * helpers
func checkUserExits(id string, r userRepository, w http.ResponseWriter) bool {
	//  check if the user with this id
	userExists, err := r.getUserByID(id)

	if err != nil {
		if err.Error() == errMsg.userNotFound {
			http.Error(w, errMsg.userNotFound, http.StatusBadRequest)
		} else {
			http.Error(w, errMsg.getUser, http.StatusInternalServerError)
		}
		return false
	}

	if userExists == nil {
		http.Error(w, errMsg.userNotFound, http.StatusNotFound)
		return false
	}
	return true
}
