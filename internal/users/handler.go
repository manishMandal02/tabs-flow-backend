package users

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	paddle "github.com/PaddleHQ/paddle-go-sdk"
	"github.com/PaddleHQ/paddle-go-sdk/pkg/paddlenotification"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
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

	http_api.SuccessResData(w, user)
}

func (h *userHandler) createUser(w http.ResponseWriter, r *http.Request) {

	user, err := userFromJSON(r.Body)

	if err != nil {
		logger.Error("decoding user from body at createUser()", err)
		http.Error(w, errMsg.createUser, http.StatusBadRequest)
		return
	}

	err = user.validate()

	if err != nil {
		logger.Error("error validating user at createUser()", err)
		http.Error(w, errMsg.createUser, http.StatusBadRequest)
		return
	}

	//  check if the user with this id
	userExists, err := h.r.getUserByID(user.Id)

	logger.Dev("userExists: %v\n, err: %v", userExists, err)

	if err != nil {
		if err.Error() != errMsg.userNotFound {
			http.Error(w, errMsg.getUser, http.StatusInternalServerError)
			return
		}
	}

	if userExists != nil {
		http.Error(w, errMsg.userExists, http.StatusBadRequest)
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

	p := "https"

	if config.LOCAL_DEV_ENV {
		p = "http"
	}

	authServiceURL := fmt.Sprintf("%s://%s/auth/user/", p, r.Host)

	res, respBody, err := utils.MakeHTTPRequest(http.MethodGet, authServiceURL, headers, bodyJson)

	if err != nil {
		logger.Error(fmt.Sprintf("Error fetching user id from Auth Service for email: %v", body.Email), err)
		http.Error(w, errMsg.createUser, http.StatusInternalServerError)
		return
	}

	if res.StatusCode != http.StatusOK {
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

	err = setDefaultUserData(user, h.r)

	if err != nil {
		logger.Error("Error setting user default data", err)
		http.Error(w, errMsg.createUser, http.StatusBadGateway)
		return
	}

	http_api.SuccessResMsg(w, "user created")
}

func (h *userHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

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

	http_api.SuccessResMsg(w, "user updated")
}

func (h *userHandler) deleteUser(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	err := h.r.deleteAccount(id)

	if err != nil {
		http.Error(w, errMsg.deleteUser, http.StatusBadRequest)
		return
	}

	http_api.SuccessResMsg(w, "user deleted")
}

// preferences handlers
func (h *userHandler) getPreferences(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	preferences, err := h.r.getAllPreferences(id)

	if err != nil {
		http.Error(w, errMsg.preferencesGet, http.StatusBadRequest)
		return
	}

	http_api.SuccessResData(w, preferences)

}

type updatePerfBody struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (h *userHandler) updatePreferences(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	var updateB updatePerfBody

	err := json.NewDecoder(r.Body).Decode(&updateB)

	if err != nil {
		logger.Error("error un_marshaling preferences from req body at updatePreferences()", err)
		http.Error(w, errMsg.preferencesUpdate, http.StatusBadRequest)
		return
	}

	sk, subPref, err := parseSubPreferencesData(id, updateB)

	logger.Dev("subPref: %v ", *subPref)

	if err != nil {
		http.Error(w, errMsg.preferencesUpdate, http.StatusBadRequest)
		return
	}

	err = h.r.updatePreferences(id, sk, *subPref)

	if err != nil {
		http.Error(w, errMsg.preferencesUpdate, http.StatusBadRequest)
		return
	}
	http_api.SuccessResMsg(w, "preferences updated")
}

// subscription handlers
func (h *userHandler) getSubscription(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	subscription, err := h.r.getSubscription(id)

	if err != nil {
		http.Error(w, errMsg.subscriptionGet, http.StatusBadRequest)
		return
	}

	http_api.SuccessResData(w, subscription)
}

func (h *userHandler) checkSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	s, err := h.r.getSubscription(id)

	if err != nil {
		http.Error(w, errMsg.subscriptionGet, http.StatusBadRequest)
		return
	}

	active := false

	if s != nil {
		active = s.Status == SubscriptionStatusActive
	}

	if active {
		// if subscription is active, check the end date
		endDate, err := time.Parse(time.DateOnly, s.End)

		if err != nil {
			logger.Error("error parsing subscription end date", err)
			http.Error(w, errMsg.subscriptionGet, http.StatusBadRequest)
			return
		}

		if endDate.Unix() < time.Now().Unix() {
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
func (h *userHandler) getPaddleURL(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	//  check if the user with this id

	userExits := checkUserExits(id, h.r, w)
	if !userExits {
		return
	}

	p, err := newPaddleClient()

	if err != nil {
		http.Error(w, errMsg.subscriptionPaddleURL, http.StatusInternalServerError)
		return
	}

	s, err := h.r.getSubscription(id)

	if err != nil {
		http.Error(w, errMsg.subscriptionGet, http.StatusBadGateway)
		return
	}

	client := p.SubscriptionsClient
	res, err := client.GetSubscription(context.TODO(), &paddle.GetSubscriptionRequest{
		SubscriptionID: s.Id,
	})

	if err != nil {
		logger.Error("error getting paddle subscription", err)
		http.Error(w, errMsg.subscriptionPaddleURL, http.StatusBadRequest)
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
		http.Error(w, errMsg.subscriptionPaddleURL, http.StatusBadRequest)
		return
	}

	if !shouldSendCancelURL && resBody.UpdateURL == "" {
		http.Error(w, errMsg.subscriptionPaddleURL, http.StatusBadRequest)
		return
	}

	http_api.SuccessResData(w, resBody)

}

// paddle webhook handler
func (h *userHandler) subscriptionWebhook(w http.ResponseWriter, r *http.Request) {

	v := paddle.NewWebhookVerifier(config.PADDLE_WEBHOOK_SECRET_KEY)

	ok, err := v.Verify(r)

	if err != nil {
		logger.Error("error verifying paddle webhook", err)
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	if !ok {
		logger.Dev("paddle webhook verification failed")
		http.Error(w, "Error bad_request", http.StatusBadRequest)
		return
	}

	http_api.SuccessResMsg(w, "event acknowledged")

	// process the event ><> * <>< <>< <>< <><

	// get the event type
	var ev *paddle.GenericEvent

	err = json.NewDecoder(r.Body).Decode(&ev)

	if err != nil {
		logger.Error("error decoding paddle webhook event", err)
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}

	switch ev.EventType {
	case paddle.EventTypeNameSubscriptionCreated:
		// get data from the event
		d := ev.Data.(paddlenotification.SubscriptionNotification)

		err := subscriptionEventHandler(h.r, d, false)

		if err != nil {
			logger.Error("Error processing SubscriptionCreated event as subscriptionWebhook()", err)
		}

	case paddle.EventTypeNameSubscriptionUpdated:

		// get data from the event
		d := ev.Data.(paddlenotification.SubscriptionNotification)

		err := subscriptionEventHandler(h.r, d, true)

		if err != nil {
			logger.Error("Error processing SubscriptionUpdated event as subscriptionWebhook()", err)
		}

		//  update subscription status

	case paddle.EventTypeNameTransactionPaymentFailed:
		// TODO -  handle payment failed
	}

}

// * helpers
func setDefaultUserPreferences(userId string, r userRepository) error {

	pref := make(map[string]interface{})

	pref[database.SORT_KEY.P_General] = &defaultUserPref.General
	pref[database.SORT_KEY.P_CmdPalette] = &defaultUserPref.CmdPalette
	pref[database.SORT_KEY.P_Note] = &defaultUserPref.Notes
	pref[database.SORT_KEY.P_LinkPreview] = &defaultUserPref.LinkPreview
	pref[database.SORT_KEY.P_AutoDiscard] = &defaultUserPref.AutoDiscard

	var wg sync.WaitGroup

	for k, v := range pref {
		wg.Add(1)
		go func(userId, k string, v interface{}) {
			defer wg.Done()
			err := r.setPreferences(userId, k, v)
			if err != nil {
				logger.Error(fmt.Sprintf("Error setting default preferences for userId: %v\n, data: %v ", userId, v), err)
			}
		}(userId, k, v)
	}

	wg.Wait()

	return nil

}
func setDefaultUserData(user *User, r userRepository) error {
	// set default preferences for user
	err := setDefaultUserPreferences(user.Id, r)

	if err != nil {
		return err
	}

	today := time.Now()

	trialEndDate := today.AddDate(0, 0, config.TRAIL_DAYS)

	//  start trail subscription
	s := &subscription{
		Plan:   SubscriptionPlanTrial,
		Status: SubscriptionStatusActive,
		Start:  today.Format(time.DateOnly),
		End:    trialEndDate.Format(time.DateOnly),
	}

	err = r.setSubscription(user.Id, s)

	if err != nil {
		return err
	}

	// send USER_REGISTERED event to email service (queue)
	event := &events.UserRegisteredPayload{
		Email:        user.Email,
		Name:         user.FullName,
		TrailEndDate: trialEndDate.Format(time.DateOnly),
	}

	sqs := events.NewQueue()

	err = sqs.AddMessage(event)

	if err != nil {
		return err
	}

	return nil

}

func checkUserExits(id string, r userRepository, w http.ResponseWriter) bool {

	if id == "" {
		http.Error(w, errMsg.invalidUserId, http.StatusBadRequest)
		return false
	}

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

// unmarshal json to sub preference struct
func unmarshalSubPref[T any](data json.RawMessage) (*T, error) {
	var pref T
	if err := json.Unmarshal(data, &pref); err != nil {
		return &pref, err
	}

	return &pref, nil
}

// associate req body data to a sub preference struct of a specific type
func parseSubPreferencesData(userId string, perfBody updatePerfBody) (string, *interface{}, error) {
	var subP interface{}
	var err error

	sk := fmt.Sprintf("P#%s", perfBody.Type)
	switch sk {
	case database.SORT_KEY.P_General:
		subP, err = unmarshalSubPref[generalP](perfBody.Data)
	case database.SORT_KEY.P_CmdPalette:
		subP, err = unmarshalSubPref[cmdPaletteP](perfBody.Data)
	case database.SORT_KEY.P_AutoDiscard:
		subP, err = unmarshalSubPref[autoDiscardP](perfBody.Data)
	case database.SORT_KEY.P_Note:
		subP, err = unmarshalSubPref[notesP](perfBody.Data)
	case database.SORT_KEY.P_LinkPreview:
		subP, err = unmarshalSubPref[linkPreviewP](perfBody.Data)
	default:
		err = fmt.Errorf("invalid preference sub type: %s", sk)
	}

	if err != nil {
		logger.Error(fmt.Sprintf("Error  un_marshaling sub preferences for userId: %v", userId), err)
		return "", nil, err
	}

	return sk, &subP, nil
}

// create a new instance of paddle sdk with configs
func newPaddleClient() (*paddle.SDK, error) {
	client, err := paddle.New(config.PADDLE_API_KEY)

	if err != nil {
		logger.Error("error creating paddle client", err)
		return nil, err
	}

	return client, nil
}

// get subscription plan type from paddle pice id
func parsePaddlePlan(priceId string) *SubscriptionPlan {

	plan := SubscriptionPlanTrial

	if priceId == paddlePlanId.Yearly {
		plan = SubscriptionPlanYearly
	}

	if priceId == paddlePlanId.LifeTime {
		plan = SubscriptionPlanLifetime
	}

	return &plan
}

// process paddle subscription (create/update) event in webhook
func subscriptionEventHandler(r userRepository, data paddlenotification.SubscriptionNotification, isUpdatedEvent bool) error {
	userId, ok := data.CustomData["userId"].(string)

	if !ok {
		logger.Error("Error getting userId from event custom data subscriptionWebhook()", errors.New("error paddle event"))
	}

	plan := *parsePaddlePlan(data.Items[0].Price.ID)

	s := &subscription{
		Id:     data.ID,
		Plan:   plan,
		Status: SubscriptionStatus(data.Status),
		Start:  *data.StartedAt,
		End:    data.CurrentBillingPeriod.EndsAt,
	}

	if plan == SubscriptionPlanYearly {
		// save next bill date if, subscription plan is yearly
		s.NextBillDate = *data.NextBilledAt
	}

	var err error

	if isUpdatedEvent {
		err = r.updateSubscription(userId, s)

	} else {

		err = r.setSubscription(userId, s)
	}

	if err != nil {
		return err
	}

	return nil
}
