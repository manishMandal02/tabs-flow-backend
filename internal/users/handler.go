package users

import (
	"encoding/json"
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

	// set default preferences for user
	err = setDefaultUserPreferences(user.Id, h.r)

	if err != nil {
		http.Error(w, errMsg.createUser, http.StatusBadRequest)
		return
	}

	today := time.Now()

	trialEndDate := today.AddDate(0, 0, config.TRAIL_DAYS)

	//  start trail subscription
	s := &subscription{
		Plan:   plans.Trail,
		Status: subscriptionStatus.Active,
		Start:  today.Format(time.DateOnly),
		End:    trialEndDate.Format(time.DateOnly),
	}

	err = h.r.setSubscription(user.Id, s)

	if err != nil {
		http.Error(w, errMsg.createUser, http.StatusBadRequest)
		return
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
		http.Error(w, errMsg.createUser, http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Message: "user created"})
}

func (h *userHandler) updateUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

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

// preferences handlers
func (h *userHandler) getPreferences(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	//  check if the user with this id
	userExits := checkUserExits(id, h.r, w)

	if !userExits {
		return
	}

	preferences, err := h.r.getAllPreferences(id)

	if err != nil {
		http.Error(w, errMsg.preferencesGet, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Data: preferences})
}

type updatePerfBody struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

func (h *userHandler) updatePreferences(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	//  check if the user with this id
	userExits := checkUserExits(id, h.r, w)
	if !userExits {
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Message: "preferences updated"})
}

// subscription handlers
func (h *userHandler) getSubscription(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	//  check if the user with this id
	userExits := checkUserExits(id, h.r, w)
	if !userExits {
		return
	}

	subscription, err := h.r.getSubscription(id)

	if err != nil {
		http.Error(w, errMsg.subscriptionGet, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Data: subscription})
}

func (h *userHandler) checkSubscriptionStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	//  check if the user with this id
	userExits := checkUserExits(id, h.r, w)

	if !userExits {
		return
	}

	s, err := h.r.getSubscription(id)

	if err != nil {
		http.Error(w, errMsg.subscriptionGet, http.StatusBadRequest)
		return
	}

	active := false

	if s != nil {
		active = s.Status == subscriptionStatus.Active
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

	res := struct {
		Active bool `json:"active"`
	}{
		Active: active,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Data: res})
}

func (h *userHandler) updateSubscription(w http.ResponseWriter, r *http.Request) {

	id := r.PathValue("id")

	//  check if the user with this id
	userExits := checkUserExits(id, h.r, w)
	if !userExits {
		return
	}

	var sub subscription

	err := json.NewDecoder(r.Body).Decode(&sub)

	if err != nil {
		logger.Error("error un_marshaling subscription from req body at updateSubscription()", err)
		http.Error(w, errMsg.subscriptionUpdate, http.StatusBadRequest)
		return
	}

	err = h.r.setSubscription(id, &sub)

	if err != nil {
		http.Error(w, errMsg.subscriptionUpdate, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Message: "subscription updated"})
}

func (h *userHandler) cancelSubscription(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	//  check if the user with this id

	userExits := checkUserExits(id, h.r, w)
	if !userExits {
		return
	}

	// TODO - cancel paddle subscription

	s := &subscription{
		Status:    subscriptionStatus.Cancelled,
		End:       time.Now().Format(time.DateOnly),
		Plan:      "",
		CancelUrl: "",
	}

	err := h.r.updateSubscription(id, s)

	if err != nil {
		http.Error(w, errMsg.subscriptionCancel, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(http_api.RespBody{Success: true, Message: "subscription cancelled"})

}

// paddle webhook handler
func (h *userHandler) subscriptionWebhook(w http.ResponseWriter, r *http.Request) {

	v := paddle.NewWebhookVerifier(config.PADDLE_WEBHOOK_SECRET_KEY)

	ok, err := v.Verify(r)

	if err != nil {
		logger.Error("error verifying paddle webhook", err)
		http.Error(w, "Error ", http.StatusInternalServerError)
		return
	}

	if !ok {
		logger.Dev("paddle webhook verification failed")
		http.Error(w, "Invalid webhook ", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(http_api.RespBody{Success: true})

	// process the event
	// p, err := newPaddleClient()

	if err != nil {
		logger.Error("error initializing paddle client", err)
		http.Error(w, "Error", http.StatusInternalServerError)
		return
	}

	// get the event type
	var ev *paddle.GenericEvent

	err = json.NewDecoder(r.Body).Decode(&ev)

	switch ev.EventType {
	case paddle.EventTypeNameSubscriptionCreated:
		//  create subscription
		d := ev.Data.(paddlenotification.SubscriptionCreatedNotification)
		// s := &subscription{
		// 	Plan:      plans.Yearly,
		// 	Status:    string(d.Status),
		// 	Start:     *d.StartedAt,
		// 	End:       d.CurrentBillingPeriod.EndsAt,
		// 	CancelUrl: "",
		// }

	case paddle.EventTypeNameSubscriptionUpdated:
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

// unmarshalPreference is a generic function to unmarshal preference data
func unmarshalSubPref[T any](data json.RawMessage) (*T, error) {
	var pref T
	if err := json.Unmarshal(data, &pref); err != nil {
		return &pref, err
	}

	return &pref, nil
}

func newPaddleClient() (*paddle.SDK, error) {
	client, err := paddle.New(config.PADDLE_API_KEY)

	if err != nil {
		logger.Error("error creating paddle client", err)
		return nil, err
	}

	return client, nil
}
