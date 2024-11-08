package users

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	paddle "github.com/PaddleHQ/paddle-go-sdk"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

// middleware to get userId from header ( set by authorizer after validating jwt token claims)
// also check if user exits
func newUserMiddleware(ur repository) http_api.Handler {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get("UserId")

		if userId == "" || !checkUserExits(userId, ur, w) {
			http.Redirect(w, r, "/logout", http.StatusTemporaryRedirect)
			return
		}

		r.SetPathValue("id", userId)
	}
}

func setDefaultUserPreferences(userId string, r repository) error {

	pref := make(map[string]interface{})

	pref[db.SORT_KEY.P_General] = &defaultUserPref.General
	pref[db.SORT_KEY.P_CmdPalette] = &defaultUserPref.CmdPalette
	pref[db.SORT_KEY.P_Note] = &defaultUserPref.Notes
	pref[db.SORT_KEY.P_LinkPreview] = &defaultUserPref.LinkPreview
	pref[db.SORT_KEY.P_AutoDiscard] = &defaultUserPref.AutoDiscard

	var wg sync.WaitGroup

	for k, v := range pref {
		wg.Add(1)
		go func(userId, k string, v interface{}) {
			defer wg.Done()
			err := r.setPreferences(userId, k, v)
			if err != nil {
				logger.Errorf("Error setting default preferences for userId: %v\n. data: %v.  \n[Error]: %v", userId, v, err)
			}
		}(userId, k, v)
	}

	wg.Wait()

	return nil

}

func setDefaultUserData(user *User, r repository, emailQueue *events.Queue) error {
	// set default preferences for user
	err := setDefaultUserPreferences(user.Id, r)

	if err != nil {
		return err
	}

	today := time.Now().UTC()

	trialEndDate := time.Date(
		today.Year(),
		today.Month(),
		today.Day()+config.TRAIL_DAYS,
		23, // hour
		59, // min
		59, // sec
		0,  // nano sec
		time.UTC,
	)

	//  start trail subscription
	s := &subscription{
		Plan:   SubscriptionPlanTrial,
		Status: SubscriptionStatusActive,
		Start:  today.Unix(),
		End:    trialEndDate.Unix(),
	}

	err = r.setSubscription(user.Id, s)

	if err != nil {
		return err
	}

	// send USER_REGISTERED event to email service (queue)
	event := events.New(events.EventTypeUserRegistered, &events.UserRegisteredPayload{
		Email:        user.Email,
		Name:         user.FullName,
		TrailEndDate: trialEndDate.Format(time.DateOnly),
	})

	err = emailQueue.AddMessage(event)

	if err != nil {
		return err
	}

	return nil

}

func checkUserExits(id string, r repository, w http.ResponseWriter) bool {

	if id == "" {
		http.Error(w, ErrMsg.InvalidUserId, http.StatusBadRequest)
		return false
	}

	//  check if the user with this id
	userExists, err := r.getUserByID(id)

	if err != nil {
		if err.Error() == ErrMsg.UserNotFound {
			http.Error(w, ErrMsg.UserNotFound, http.StatusBadRequest)
		} else {
			http.Error(w, ErrMsg.GetUser, http.StatusInternalServerError)
		}
		return false
	}

	if userExists == nil {
		http.Error(w, ErrMsg.UserNotFound, http.StatusNotFound)
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
func parseSubPreferencesData(perfBody updatePerfBody) (string, *interface{}, error) {
	var subP interface{}
	var err error

	sk := fmt.Sprintf("P#%s", perfBody.Type)
	switch sk {
	case db.SORT_KEY.P_General:
		subP, err = unmarshalSubPref[generalP](perfBody.Data)
	case db.SORT_KEY.P_CmdPalette:
		subP, err = unmarshalSubPref[cmdPaletteP](perfBody.Data)
	case db.SORT_KEY.P_AutoDiscard:
		subP, err = unmarshalSubPref[autoDiscardP](perfBody.Data)
	case db.SORT_KEY.P_Note:
		subP, err = unmarshalSubPref[notesP](perfBody.Data)
	case db.SORT_KEY.P_LinkPreview:
		subP, err = unmarshalSubPref[linkPreviewP](perfBody.Data)
	default:
		err = fmt.Errorf("invalid preference sub type: %s", sk)
	}

	if err != nil {
		logger.Errorf("Error  un_marshaling sub preferences[Error]:  %v", err)
		return "", nil, err
	}

	return sk, &subP, nil
}

// create a new instance of paddle sdk with configs
func NewPaddleSubscriptionClient() (*paddle.SubscriptionsClient, error) {
	client, err := paddle.New(config.PADDLE_API_KEY)

	if err != nil {
		logger.Error("error creating paddle client", err)
		return nil, err
	}

	return client.SubscriptionsClient, nil
}

// get subscription plan type from paddle pice id
func parsePaddlePlan(priceId string) *SubscriptionPlan {

	plan := SubscriptionPlanTrial

	if priceId == paddlePlanId.yearly {
		plan = SubscriptionPlanYearly
	}

	if priceId == paddlePlanId.lifeTime {
		plan = SubscriptionPlanLifetime
	}

	return &plan
}

type subscriptionData struct {
	userId         string
	subscriptionId string
	status         string
	priceId        string
	startDate      string
	endDate        string
	nextBillDate   string
}

// process paddle subscription (create/update) event in webhook
func subscriptionEventHandler(r repository, data *subscriptionData, isUpdatedEvent bool) error {
	// parse date to convert it to unix timestamp for db
	startDate, err := time.Parse(time.RFC3339, data.startDate)
	endDate, err2 := time.Parse(time.RFC3339, data.endDate)
	nextBillDate, err3 := time.Parse(time.RFC3339, data.nextBillDate)

	if err != nil || err2 != nil || err3 != nil {
		logger.Errorf("subscriptionEventHandler(): error parsing subscription dates: %v", err)
		return err
	}

	if data.userId == "" {
		ErrMsg := "error getting userId from event custom data subscriptionWebhook()"
		logger.Errorf("%v", ErrMsg)
		return errors.New(ErrMsg)
	}

	plan := *parsePaddlePlan(data.priceId)

	s := &subscription{
		Id:     data.subscriptionId,
		Plan:   plan,
		Status: SubscriptionStatus(data.status),
		Start:  startDate.Unix(),
		End:    endDate.Unix(),
	}

	if plan == SubscriptionPlanYearly {
		// save next bill date if, subscription plan is yearly
		s.NextBillDate = nextBillDate.Unix()
	}

	if isUpdatedEvent {
		err = r.updateSubscription(data.userId, s)

	} else {
		err = r.setSubscription(data.userId, s)
	}

	if err != nil {
		return err
	}

	return nil
}
