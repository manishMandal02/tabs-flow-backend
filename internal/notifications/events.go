package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	web_push "github.com/SherClockHolmes/webpush-go"
	lambda_events "github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/notes"
	"github.com/manishMandal02/tabsflow-backend/internal/spaces"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

func EventsHandler(_ context.Context, event lambda_events.SQSEvent) (interface{}, error) {
	if len(event.Records) < 1 {
		errMsg := "no events to process"
		logger.Errorf(errMsg)

		return nil, errors.New(errMsg)
	}

	//  process batch of events
	for _, record := range event.Records {
		event := events.Event[any]{}

		logger.Info("processing record: %v", record.Body)

		eventType := *record.MessageAttributes["event_type"].StringValue

		logger.Info("processing event_type: %v", event.EventType)

		err := processEvent(eventType, record.Body)

		if err != nil {
			logger.Errorf("error processing event: %v", err)
			continue
		}

		// remove message from sqs
		q := events.NewNotificationQueue()

		err = q.DeleteMessage(record.ReceiptHandle)

		if err != nil {
			return nil, err
		}

	}

	return nil, nil
}

func processEvent(eventType string, body string) error {
	switch events.EventType(eventType) {
	case events.EventTypeScheduleNoteRemainder:

		ev, err := events.NewFromJSON[events.ScheduleNoteRemainderPayload](body)

		if err != nil {
			logger.Errorf("error un_marshalling event: %v", err)
			return err
		}

		return scheduleNoteRemainder(ev.Payload)

	case events.EventTypeScheduleSnoozedTab:

		ev, err := events.NewFromJSON[events.ScheduleSnoozedTabPayload](body)

		if err != nil {
			logger.Errorf("error un_marshalling event: %v", err)
			return err
		}

		return scheduleSnoozedTab(ev.Payload)

	case events.EventTypeTriggerNoteRemainder:

		ev, err := events.NewFromJSON[events.ScheduleNoteRemainderPayload](body)

		if err != nil {
			logger.Errorf("error un_marshalling event: %v", err)
			return err
		}

		return triggerNoteRemainder(ev.Payload)

	case events.EventTypeTriggerSnoozedTab:
		ev, err := events.NewFromJSON[events.ScheduleSnoozedTabPayload](body)

		if err != nil {
			logger.Errorf("error un_marshalling event: %v", err)
			return err
		}

		return triggerSnoozedTab(ev.Payload)
	}

	return nil
}

func scheduleNoteRemainder(p *events.ScheduleNoteRemainderPayload) error {
	scheduler := events.NewScheduler()
	var err error

	switch p.SubEvent {
	case events.SubEventCreate:
		triggerEvent := events.New(events.EventTypeTriggerNoteRemainder, &events.ScheduleNoteRemainderPayload{
			UserId: p.UserId,
			NoteId: p.NoteId,
		})

		evStr := triggerEvent.ToJSON()

		t := time.Unix(p.TriggerAt, 0).Format(config.DATE_TIME_FORMAT)

		err = scheduler.CreateSchedule(p.NoteId, t, &evStr)
	case events.SubEventUpdate:
		t := time.Unix(p.TriggerAt, 0).Format(config.DATE_TIME_FORMAT)
		err = scheduler.UpdateSchedule(p.NoteId, t)
	case events.SubEventDelete:
		err = scheduler.DeleteSchedule(p.NoteId)
	}

	return err
}

func scheduleSnoozedTab(p *events.ScheduleSnoozedTabPayload) error {

	scheduler := events.NewScheduler()
	var err error

	switch p.SubEvent {
	case events.SubEventCreate:
		triggerEvent := events.New(events.EventTypeTriggerNoteRemainder, &events.ScheduleSnoozedTabPayload{
			UserId:       p.UserId,
			SpaceId:      p.SpaceId,
			SnoozedTabId: p.SnoozedTabId,
		})

		evStr := triggerEvent.ToJSON()

		t := time.Unix(p.TriggerAt, 0).Format(config.DATE_TIME_FORMAT)

		err = scheduler.CreateSchedule(p.SnoozedTabId, t, &evStr)
	case events.SubEventUpdate:

		t := time.Unix(p.TriggerAt, 0).Format(config.DATE_TIME_FORMAT)

		err = scheduler.UpdateSchedule(p.SnoozedTabId, t)
	case events.SubEventDelete:
		err = scheduler.DeleteSchedule(p.SnoozedTabId)
	}

	return err
}

// send notification to user
func triggerNoteRemainder(p *events.ScheduleNoteRemainderPayload) error {
	db := database.New()
	r := newRepository(db)
	s, err := r.getSubscriptionInfo(p.UserId)

	if err != nil {
		return err
	}

	note, err := getNote(db, p.UserId, p.NoteId)

	if err != nil {
		return err
	}

	n, err := json.Marshal(note)

	if err != nil {
		logger.Error("error marshalling note", err)
		return err
	}

	err = sendWebPushNotification(p.UserId, s, n)

	if err != nil {
		logger.Error("error sending web push notification", err)
		return err

	}

	return nil

}

// send notification to user
func triggerSnoozedTab(p *events.ScheduleSnoozedTabPayload) error {
	db := database.New()
	r := newRepository(db)
	s, err := r.getSubscriptionInfo(p.UserId)

	if err != nil {
		return err
	}

	snoozedTab, err := getSnoozedTab(db, p.UserId, p.SpaceId, p.SnoozedTabId)

	if err != nil {
		return err
	}

	t, err := json.Marshal(snoozedTab)

	if err != nil {
		logger.Error("error marshalling snoozedTab", err)
		return err
	}

	err = sendWebPushNotification(p.UserId, s, t)

	if err != nil {
		logger.Error("error sending web push notification", err)
		return err

	}

	return nil

}

func sendWebPushNotification(userId string, s *PushSubscription, body []byte) error {

	ws := &web_push.Subscription{
		Endpoint: s.Endpoint,
		Keys: web_push.Keys{
			Auth:   s.AuthKey,
			P256dh: s.P256dhKey,
		},
	}
	o := &web_push.Options{
		TTL:             300,
		Subscriber:      userId,
		VAPIDPrivateKey: config.VAPID_PRIVATE_KEY,
		VAPIDPublicKey:  config.VAPID_PUBLIC_KEY,
	}

	_, err := web_push.SendNotification(body, ws, o)

	if err != nil {
		return err
	}

	return nil

}

func getNote(db *database.DDB, userId, noteId string) (*notes.Note, error) {

	r := notes.NewNoteRepository(db, nil)

	note, err := r.GetNote(userId, noteId)

	if err != nil {
		return nil, err
	}

	return note, nil
}

func getSnoozedTab(db *database.DDB, userId, spaceId, snoozedTabId string) (*spaces.SnoozedTab, error) {

	r := spaces.NewSpaceRepository(db)

	snoozedTabIdInt, err := strconv.ParseInt(snoozedTabId, 10, 64)

	if err != nil {
		logger.Error("error parsing snoozed tab id to int", err)
		return nil, err

	}

	snoozedTab, err := r.GetSnoozedTab(userId, spaceId, snoozedTabIdInt)

	if err != nil {
		return nil, err
	}

	return snoozedTab, nil
}
