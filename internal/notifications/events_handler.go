package notifications

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	web_push "github.com/SherClockHolmes/webpush-go"
	lambda_events "github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/notes"
	"github.com/manishMandal02/tabsflow-backend/internal/spaces"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

func EventsHandler(_ context.Context, event lambda_events.SQSEvent) (interface{}, error) {
	if len(event.Records) < 1 {
		errMsg := "no events to process"
		logger.Errorf("%v", errMsg)

		return nil, errors.New(errMsg)
	}

	//  process batch of events
	for _, record := range event.Records {

		logger.Info("processing record: %v", record.Body)

		eventType := ""

		if _, ok := record.MessageAttributes["event_type"]; ok {
			eventType = *record.MessageAttributes["event_type"].StringValue
		} else {

			e, err := events.NewFromJSON[any](record.Body)

			if err != nil {
				logger.Errorf("error un_marshalling event from json: %v", err)
			}

			eventType = string(e.EventType)

		}

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

// set a schedule to trigger a note remainder notification
func scheduleNoteRemainder(p *events.ScheduleNoteRemainderPayload) error {
	var err error

	scheduler := events.NewScheduler()

	sId := fmt.Sprintf("note_%v", p.NoteId)

	switch p.SubEvent {
	case events.SubEventCreate:
		triggerEvent := events.New(events.EventTypeTriggerNoteRemainder, &events.ScheduleNoteRemainderPayload{
			UserId: p.UserId,
			NoteId: p.NoteId,
		})

		evStr := triggerEvent.ToJSON()

		t := time.Unix(p.TriggerAt, 0).UTC().Format(config.DATE_TIME_FORMAT)

		err = scheduler.CreateSchedule(sId, t, &evStr)
	case events.SubEventUpdate:
		t := time.Unix(p.TriggerAt, 0).UTC().Format(config.DATE_TIME_FORMAT)
		err = scheduler.UpdateSchedule(sId, t)
	case events.SubEventDelete:
		err = scheduler.DeleteSchedule(sId)
	}

	return err
}

// set a schedule to trigger a snoozed tab notification
func scheduleSnoozedTab(p *events.ScheduleSnoozedTabPayload) error {

	scheduler := events.NewScheduler()
	var err error

	sId := fmt.Sprintf("snoozedTab_%v", p.SnoozedTabId)

	switch p.SubEvent {
	case events.SubEventCreate:
		triggerEvent := events.New(events.EventTypeTriggerSnoozedTab, &events.ScheduleSnoozedTabPayload{
			UserId:       p.UserId,
			SpaceId:      p.SpaceId,
			SnoozedTabId: p.SnoozedTabId,
		})

		evStr := triggerEvent.ToJSON()

		t := time.Unix(p.TriggerAt, 0).UTC().Format(config.DATE_TIME_FORMAT)

		err = scheduler.CreateSchedule(sId, t, &evStr)
	case events.SubEventUpdate:
		t := time.Unix(p.TriggerAt, 0).UTC().Format(config.DATE_TIME_FORMAT)

		err = scheduler.UpdateSchedule(sId, t)
	case events.SubEventDelete:
		err = scheduler.DeleteSchedule(sId)
	}

	return err
}

// send note notification to user
func triggerNoteRemainder(p *events.ScheduleNoteRemainderPayload) error {
	db := db.New()
	r := newRepository(db)
	s, err := r.getNotificationSubscription(p.UserId)

	if err != nil && err.Error() != errMsg.notificationsSubscribeEmpty {

		return err
	}

	note, err := getNote(db, p.UserId, p.NoteId)

	if err != nil {
		return err
	}

	// create notification
	notification := notification{
		Id:        strconv.FormatInt(time.Now().UTC().Unix(), 10),
		Type:      NotificationTypeNoteRemainder,
		IsRead:    false,
		Timestamp: time.Now().UTC().Unix(),
		Note: &noteRemainderNotification{
			Id:     note.Id,
			Title:  note.Title,
			Domain: note.Domain,
		},
	}

	err = r.create(p.UserId, &notification)

	if err != nil {
		return err
	}

	// user has not subscribed for notifications
	if s == nil {
		return nil
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

	// remove remainder at
	err = removeNoteRemainder(db, p.UserId, p.NoteId)

	if err != nil {
		logger.Error("error removing note remainder", err)
		return err
	}

	return nil

}

// send snoozed tab notification to user
func triggerSnoozedTab(p *events.ScheduleSnoozedTabPayload) error {
	db := db.New()
	r := newRepository(db)
	s, err := r.getNotificationSubscription(p.UserId)

	if err != nil && err.Error() != errMsg.notificationsSubscribeEmpty {
		return err
	}

	snoozedTab, err := getSnoozedTab(db, p.UserId, p.SpaceId, p.SnoozedTabId)

	if err != nil {
		return err
	}

	// create notification
	notification := notification{
		Id:        strconv.FormatInt(time.Now().UTC().Unix(), 10),
		Type:      NotificationTypeUnSnoozedType,
		IsRead:    false,
		Timestamp: time.Now().UTC().Unix(),
		SnoozedTab: &snoozedTabNotification{
			Id:    p.SnoozedTabId,
			Title: snoozedTab.Title,
			Icon:  snoozedTab.Icon,
		},
	}

	err = r.create(p.UserId, &notification)

	if err != nil {
		return err
	}

	// user has not subscribed for notifications
	if s == nil {
		return nil
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

	// delete snoozed tab
	err = deleteSnoozedTab(db, p.UserId, p.SpaceId, p.SnoozedTabId)

	if err != nil {
		logger.Error("error deleting snoozed tab", err)
		return err
	}
	return nil

}

// * helpers
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

	var bEnc = make([]byte, base64.StdEncoding.EncodedLen(len(body)))

	base64.StdEncoding.Encode(bEnc, body)

	_, err := web_push.SendNotification(bEnc, ws, o)

	if err != nil {
		return err
	}

	return nil

}

func getNote(db *db.DDB, userId, noteId string) (*notes.Note, error) {

	r := notes.NewNoteRepository(db, nil)

	note, err := r.GetNote(userId, noteId)

	if err != nil {
		return nil, err
	}

	return note, nil
}

func getSnoozedTab(db *db.DDB, userId, spaceId, snoozedTabId string) (*spaces.SnoozedTab, error) {

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

func removeNoteRemainder(db *db.DDB, userId, noteId string) error {
	r := notes.NewNoteRepository(db, nil)

	err := r.RemoveNoteRemainder(userId, noteId)

	if err != nil {
		return err
	}

	return nil
}

func deleteSnoozedTab(db *db.DDB, userId, spaceId, snoozedTabId string) error {
	r := spaces.NewSpaceRepository(db)

	snoozedTabIdInt, err := strconv.ParseInt(snoozedTabId, 10, 64)

	if err != nil {
		logger.Error("error parsing snoozed tab id to int", err)
		return err

	}

	err = r.DeleteSnoozedTab(userId, spaceId, snoozedTabIdInt)

	if err != nil {
		return err
	}

	return nil
}
