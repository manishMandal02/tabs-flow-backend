package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	paddle "github.com/PaddleHQ/paddle-go-sdk"
	"github.com/PaddleHQ/paddle-go-sdk/pkg/paddlenotification"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
	"github.com/manishMandal02/tabsflow-backend/pkg/utils"
)

type paddleClientInterface interface {
	GetSubscription(ctx context.Context, req *paddle.GetSubscriptionRequest) (res *paddle.Subscription, err error)
}

type handler struct {
	r          repository
	paddle     paddleClientInterface
	emailQueue *events.Queue
	httpClient http_api.Client
}

func newHandler(r repository, q *events.Queue, c http_api.Client, p paddleClientInterface) *handler {
	return &handler{
		r:          r,
		paddle:     p,
		emailQueue: q,
		httpClient: c,
	}
}

// profile handlers
func (h handler) userById(w http.ResponseWriter, r *http.Request) {

	id := r.Header.Get("UserId")

	if id == "" {
		http_api.ErrorRes(w, ErrMsg.InvalidUserId, http.StatusBadRequest)
		return
	}

	user, err := h.r.getUserByID(id)

	if err != nil {
		if err.Error() == ErrMsg.UserNotFound {
			http_api.ErrorRes(w, ErrMsg.UserNotFound, http.StatusBadRequest)
		} else {
			http_api.ErrorRes(w, ErrMsg.GetUser, http.StatusInternalServerError)
		}
		return
	}

	http_api.SuccessResData(w, user)
}

func (h handler) createUser(w http.ResponseWriter, r *http.Request) {

	user, err := userFromJSON(r.Body)

	if err != nil {
		logger.Error("decoding user from body at createUser()", err)
		http_api.ErrorRes(w, ErrMsg.CreateUser, http.StatusBadRequest)
		return
	}

	err = user.validate()

	if err != nil {
		logger.Error("error validating user at createUser()", err)
		http_api.ErrorRes(w, ErrMsg.CreateUser, http.StatusBadRequest)
		return
	}

	//  check if the user with this id
	userExists, err := h.r.getUserByID(user.Id)

	if err != nil && err.Error() != ErrMsg.UserNotFound {
		http_api.ErrorRes(w, ErrMsg.GetUser, http.StatusBadGateway)
		return
	}

	if userExists != nil {
		http_api.ErrorRes(w, ErrMsg.UserExists, http.StatusBadRequest)
		return
	}

	p := "https"

	if config.LOCAL_DEV_ENV {
		p = "http"
	}

	host := r.Host

	if strings.Contains(host, "amazonaws.com") {
		host += "/test"
	}

	authServiceURL := fmt.Sprintf("%s://%s/auth/user/%s", p, host, user.Email)

	origin := r.Header.Get("Origin")

	if origin == "" {
		referrer := r.Header.Get("Referer")
		if referrer == "" {
			origin = "http://tabsflow.com"
		} else {
			origin = referrer
		}
	}

	headers := map[string]string{
		"Origin": origin,
	}

	res, respBody, err := utils.MakeHTTPRequest(http.MethodGet, authServiceURL, headers, nil, h.httpClient)

	if err != nil {
		logger.Errorf("Error fetching user id from Auth Service for email: %v. \n [Error]: %v", user.Email, err)
		http_api.ErrorRes(w, ErrMsg.CreateUser, http.StatusInternalServerError)
		return
	}

	if res.StatusCode != http.StatusOK {
		logger.Errorf("User does not have a valid session profile for email: %v. \n [Error]: %v", user.Email, err)
		//  Logout
		http.Redirect(w, r, "/auth/logout", http.StatusTemporaryRedirect)
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
		logger.Errorf("Error un_marshaling user id data for email: %v. \n [Error]: %v", user.Email, err)
		http_api.ErrorRes(w, ErrMsg.CreateUser, http.StatusInternalServerError)
		return
	}

	if userIdData.Data.UserId != user.Id {
		logger.Errorf("User Id mismatch for email: %v. \n [Error]: %v", user.Email, err)
		http.Redirect(w, r, "/auth/logout", http.StatusTemporaryRedirect)
		return
	}

	err = h.r.insertUser(user)

	if err != nil {
		http_api.ErrorRes(w, ErrMsg.CreateUser, http.StatusBadGateway)
		return
	}

	logger.Info("user saved to db, with userId: %v", user.Id)

	err = setDefaultUserData(user, h.r, h.emailQueue)

	if err != nil {
		logger.Error("Error setting user default data", err)
		http_api.ErrorRes(w, ErrMsg.CreateUser, http.StatusInternalServerError)
		return
	}

	http_api.SuccessResMsg(w, "user created")
}

func (h handler) updateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var n struct {
		Name string `json:"fullName"`
	}

	err := json.NewDecoder(r.Body).Decode(&n)

	if err != nil {
		logger.Error("error un_marshaling name from JSON at updateUser()", err)
		http_api.ErrorRes(w, ErrMsg.UpdateUser, http.StatusBadRequest)
		return
	}

	err = h.r.updateUser(id, n.Name)

	if err != nil {
		http_api.ErrorRes(w, ErrMsg.UpdateUser, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "user updated")
}

func (h handler) deleteUser(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	err := h.r.deleteAccount(id)

	if err != nil {
		http_api.ErrorRes(w, ErrMsg.DeleteUser, http.StatusBadRequest)
		return
	}

	http_api.SuccessResMsg(w, "user deleted")
}

// preferences handlers
func (h handler) getPreferences(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	preferences, err := h.r.getAllPreferences(id)

	if err != nil {
		http_api.ErrorRes(w, ErrMsg.PreferencesGet, http.StatusBadRequest)
		return
	}

	http_api.SuccessResData(w, preferences)

}

type updatePerfBody struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (h handler) updatePreferences(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var updateB updatePerfBody

	err := json.NewDecoder(r.Body).Decode(&updateB)

	if err != nil {
		logger.Error("error un_marshaling preferences from req body at updatePreferences()", err)
		http_api.ErrorRes(w, ErrMsg.PreferencesUpdate, http.StatusBadRequest)
		return
	}

	sk, subPref, err := parseSubPreferencesData(updateB)

	if err != nil {
		http_api.ErrorRes(w, ErrMsg.PreferencesUpdate, http.StatusBadRequest)
		return
	}

	err = h.r.updatePreferences(id, sk, *subPref)

	if err != nil {
		http_api.ErrorRes(w, ErrMsg.PreferencesUpdate, http.StatusBadRequest)
		return
	}
	http_api.SuccessResMsg(w, "user preferences updated")
}

// subscription handlers
func (h handler) getSubscription(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	subscription, err := h.r.getSubscription(id)

	if err != nil {
		http_api.ErrorRes(w, ErrMsg.SubscriptionGet, http.StatusBadGateway)
		return
	}

	http_api.SuccessResData(w, subscription)
}

func (h handler) checkSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s, err := h.r.getSubscription(id)

	if err != nil {
		http_api.ErrorRes(w, ErrMsg.SubscriptionGet, http.StatusBadRequest)
		return
	}

	active := false

	if s != nil {
		active = s.Status == SubscriptionStatusActive
	}

	if active {
		// if subscription is active, check the end date
		if err != nil {
			logger.Error("error parsing subscription end date", err)
			http_api.ErrorRes(w, ErrMsg.SubscriptionGet, http.StatusBadRequest)
			return
		}

		if s.End < time.Now().UTC().Unix() {
			active = false
		}
	}

	status := struct {
		Active bool `json:"active"`
	}{
		Active: active,
	}

	http_api.SuccessResData(w, status)

}

func (h handler) getPaddleURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s, err := h.r.getSubscription(id)

	if err != nil {
		http_api.ErrorRes(w, ErrMsg.SubscriptionGet, http.StatusBadGateway)
		return
	}

	res, err := h.paddle.GetSubscription(context.TODO(), &paddle.GetSubscriptionRequest{
		SubscriptionID: s.Id,
	})

	if err != nil {
		logger.Error("error getting paddle subscription", err)
		http_api.ErrorRes(w, ErrMsg.SubscriptionPaddleURL, http.StatusBadGateway)
		return
	}

	resBody := struct {
		CancelURL string `json:"cancelURL,omitempty"`
		UpdateURL string `json:"updateURL,omitempty"`
	}{}

	shouldSendCancelURL := r.URL.Query().Get("cancelURL") != ""

	if shouldSendCancelURL {
		resBody.CancelURL = res.ManagementURLs.Cancel
	} else {
		resBody.UpdateURL = *res.ManagementURLs.UpdatePaymentMethod
	}

	if shouldSendCancelURL && resBody.CancelURL == "" {
		http_api.ErrorRes(w, ErrMsg.SubscriptionPaddleURL, http.StatusBadRequest)
		return
	}

	if !shouldSendCancelURL && resBody.UpdateURL == "" {
		http_api.ErrorRes(w, ErrMsg.SubscriptionPaddleURL, http.StatusBadRequest)
		return
	}

	http_api.SuccessResData(w, resBody)

}

// paddle webhook handler
func (h handler) subscriptionWebhook(w http.ResponseWriter, r *http.Request) {
	var err error

	logger.Dev("paddle webhook test : %v ", r.Header.Get("Paddle-Webhook-Test"))

	if r.Header.Get("Paddle-Webhook-Test") != "true" {

		v := paddle.NewWebhookVerifier(config.PADDLE_WEBHOOK_SECRET_KEY)

		ok, err := v.Verify(r)

		if err != nil {
			logger.Error("error verifying paddle webhook", err)
			http_api.ErrorRes(w, "Error", http.StatusInternalServerError)
			return
		}

		if !ok {
			http_api.ErrorRes(w, "Error bad_request", http.StatusBadRequest)
			return
		}
	}

	body := r.Body

	http_api.SuccessResMsg(w, "event acknowledged")

	// process the event *******

	// get the event type
	var ev paddle.GenericEvent

	err = json.NewDecoder(body).Decode(&ev)

	if err != nil {
		logger.Error("error decoding paddle webhook event", err)
		http_api.ErrorRes(w, "Error", http.StatusBadRequest)
		return
	}

	switch ev.EventType {
	case paddle.EventTypeNameSubscriptionCreated:
		var c paddlenotification.SubscriptionCreatedNotification

		data, err := json.Marshal(ev.Data)

		if err != nil {
			logger.Error("error marshaling paddle webhook event data", err)
			return
		}

		if err := json.Unmarshal(data, &c); err != nil {
			logger.Error("error parsing subscription data", err)
			return
		}

		userId, ok := c.CustomData["userId"].(string)

		if !ok {
			logger.Errorf("userId not found in subscription created event data")
			userId = "01929a76-ce53-7e0d-b712-41a9fa1178d8"
		}

		subscriptionData := subscriptionData{
			userId:         userId,
			subscriptionId: c.ID,
			priceId:        c.Items[0].Price.ID,
			status:         string(c.Status),
			startDate:      *c.StartedAt,
			endDate:        c.CurrentBillingPeriod.EndsAt,
			nextBillDate:   *c.NextBilledAt,
		}

		err = subscriptionEventHandler(h.r, &subscriptionData, false)

		if err != nil {
			logger.Error("Error processing SubscriptionCreated event as subscriptionWebhook()", err)
		}

	case paddle.EventTypeNameSubscriptionUpdated:
		var u paddlenotification.SubscriptionNotification

		data, err := json.Marshal(ev.Data)

		if err != nil {
			logger.Error("error marshaling paddle webhook event data", err)
			return
		}

		if err := json.Unmarshal(data, &u); err != nil {
			logger.Error("error parsing subscription data", err)
			return
		}

		userId, ok := u.CustomData["userId"].(string)

		if !ok {
			logger.Errorf("userId not found in subscription updated event data")
			userId = "01929a76-ce53-7e0d-b712-41a9fa1178d8"
		}

		subscriptionData := subscriptionData{
			userId:         userId,
			subscriptionId: u.ID,
			priceId:        u.Items[0].Price.ID,
			status:         string(u.Status),
			startDate:      *u.StartedAt,
			endDate:        u.CurrentBillingPeriod.EndsAt,
			nextBillDate:   *u.NextBilledAt,
		}

		err = subscriptionEventHandler(h.r, &subscriptionData, true)

		if err != nil {
			logger.Error("Error processing SubscriptionUpdated event as subscriptionWebhook()", err)
		}

	case paddle.EventTypeNameTransactionPaymentFailed:
		var f paddlenotification.TransactionNotification

		data, err := json.Marshal(ev.Data)

		if err != nil {
			logger.Error("error marshaling paddle webhook event data", err)
			return
		}

		if err := json.Unmarshal(data, &f); err != nil {
			logger.Error("error parsing payment failed event data", err)
			return
		}

		userId, ok := f.CustomData["userId"].(string)

		if !ok {
			logger.Errorf("userId not found in subscription created event data")
		}

		logger.Dev("userId: %v", userId)

		// TODO: handle payment failed
		// show a notification to the user
		// send an email only if paddle doesn't send a payment failed email

	default:
		logger.Errorf("not handling paddle webhook event_type: %v", ev.EventType)
	}

}
