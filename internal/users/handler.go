package users

import (
	"encoding/json"
	"net/http"

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

	err = h.r.insertUser(user)

	if err != nil {
		http.Error(w, errMsg.createUser, http.StatusBadRequest)
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

	err := h.r.deleteAccount(id)

	if err != nil {
		http.Error(w, errMsg.deleteUser, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Message: "user deleted"})

}
