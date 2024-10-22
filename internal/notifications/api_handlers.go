package notifications

import (
	"encoding/json"
	"net/http"

	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type notificationHandler struct {
	r notificationRepository
}

func newHandler(r notificationRepository) *notificationHandler {
	return &notificationHandler{
		r: r,
	}
}

func (h *notificationHandler) get(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")
	notificationId := r.PathValue("id")

	if notificationId == "" {
		http.Error(w, errMsg.notificationGet, http.StatusBadRequest)
		return
	}

	n, err := h.r.get(userId, notificationId)
	if err != nil {
		if err.Error() == errMsg.notificationsEmpty {
			http_api.SuccessResData(w, []notification{})
			return
		}
		http.Error(w, errMsg.notificationGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, n)
}

func (h *notificationHandler) getUserNotifications(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")

	notifications, err := h.r.getUserNotifications(userId)

	if err != nil {
		if err.Error() == errMsg.notificationsEmpty {
			http_api.SuccessResData(w, []notification{})
			return
		}
		http.Error(w, errMsg.notificationGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, notifications)
}

func (h *notificationHandler) subscribe(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")

	var subscription PushSubscription

	err := json.NewDecoder(r.Body).Decode(&subscription)

	if err != nil {
		logger.Errorf("error decoding notification subscription for user_id: %v. \n[Error]: %v", userId, err)
		http.Error(w, errMsg.notificationsSubscribe, http.StatusBadRequest)
		return
	}

	err = subscription.validate()
	if err != nil {
		logger.Errorf("error validating notification subscription for user_id: %v. \n[Error]: %v", userId, err)
		http.Error(w, errMsg.notificationsSubscribe, http.StatusBadRequest)
		return
	}

	err = h.r.subscribe(userId, &subscription)

	if err != nil {
		http.Error(w, errMsg.notificationsSubscribe, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "Subscribed to  notifications")
}

func (h *notificationHandler) getNotificationSubscription(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")

	subscription, err := h.r.getNotificationSubscription(userId)

	if err != nil {
		if err.Error() == errMsg.notificationsSubscribeEmpty {
			http_api.SuccessResData(w, PushSubscription{})
			return
		}
		http.Error(w, errMsg.notificationsSubscribeEmpty, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, subscription)

}

func (h *notificationHandler) unsubscribe(w http.ResponseWriter, r *http.Request) {
	userId := r.PathValue("userId")

	err := h.r.deleteNotificationSubscription(userId)

	if err != nil {
		http.Error(w, errMsg.notificationsUnsubscribe, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "Unsubscribed from notifications")

}

func (h *notificationHandler) delete(w http.ResponseWriter, r *http.Request) {

	userId := r.PathValue("userId")
	notificationId := r.PathValue("id")

	if notificationId == "" {
		http.Error(w, errMsg.notificationDelete, http.StatusBadRequest)
		return
	}

	err := h.r.delete(userId, notificationId)

	if err != nil {
		logger.Error("error deleting space", err)
		http.Error(w, errMsg.notificationDelete, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "space deleted successfully")
}

// func (h *notificationHandler) create(w http.ResponseWriter, r *http.Request) {
// 	userId := r.PathValue("userId")
// 	n := notification{}
// 	err := json.NewDecoder(r.Body).Decode(&n)
// 	if err != nil {
// 		logger.Error("error decoding notification", err)
// 		http.Error(w, errMsg.notificationCreate, http.StatusBadRequest)
// 		return
// 	}
// 	err = h.r.createNotification(userId, &n)
// 	if err != nil {
// 		logger.Error("error creating notification", err)
// 		http.Error(w, errMsg.notificationCreate, http.StatusBadGateway)
// 		return
// 	}
// 	http_api.SuccessResMsg(w, "notification created successfully")
// }

//* helpers

// middleware to get userId from jwt token present in req cookies
func newUserIdMiddleware() http_api.Handler {
	return func(w http.ResponseWriter, r *http.Request) {

		// get userId from jwt token

		userId := r.Header.Get("UserId")

		if userId == "" {
			http.Redirect(w, r, "/logout", http.StatusTemporaryRedirect)
			return
		}

		r.SetPathValue("userId", userId)
	}
}
