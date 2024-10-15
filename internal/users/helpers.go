package users

import (
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
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

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
				logger.Errorf("Error setting default preferences for userId: %v\n. data: %v.  \n[Error]: %v", userId, v, err)
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
	event := events.New(events.EventTypeUserRegistered, &events.UserRegisteredPayload{
		Email:        user.Email,
		Name:         user.FullName,
		TrailEndDate: trialEndDate.Format(time.DateOnly),
	})

	sqs := events.NewEmailQueue()

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
		logger.Errorf("Error  un_marshaling sub preferences for u \n[Error]: %vserId:. %v", userId, err)
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
