package notifications

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	lambda_events "github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/notes"
	"github.com/manishMandal02/tabsflow-backend/internal/spaces"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

func SQSMessagesHandler(q *events.Queue) http_api.SQSHandler {
	return func(messages []lambda_events.SQSMessage) (interface{}, error) {
		if len(messages) < 1 {
			errMsg := "no events to process"
			logger.Errorf("%v", errMsg)

			return nil, errors.New(errMsg)
		}

		//  process batch of events
		for _, msg := range messages {

			logger.Info("processing msg: %v", msg.Body)

			eventType := ""

			if _, ok := msg.MessageAttributes["event_type"]; ok {
				eventType = *msg.MessageAttributes["event_type"].StringValue
			} else {

				e, err := events.NewFromJSON[any](msg.Body)

				if err != nil {
					logger.Errorf("error un_marshalling event from json: %v", err)
				}

				eventType = string(e.EventType)

			}

			err := processEvent(eventType, msg.Body)

			if err != nil {
				logger.Errorf("error processing event: %v", err)
				continue
			}

			// remove message from sqs
			err = q.DeleteMessage(msg.ReceiptHandle)

			if err != nil {
				return nil, err
			}

		}

		return nil, nil

	}
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

	note, err := getNote(db, p.UserId, p.NoteId)

	if err != nil {
		return err
	}

	// create notification
	n := &notification{
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

	err = r.create(p.UserId, n)

	if err != nil {
		return err
	}

	pushEvent := &WebPushEvent[notification]{
		Event:   PushNotificationEventTypeNotification,
		Payload: n,
	}

	err = pushEvent.send(p.UserId, r)

	if err != nil {
		return err
	}

	// remove remainder at
	err = removeNoteRemainder(db, p.UserId, p.NoteId)

	if err != nil {
		return err
	}

	return nil

}

// send snoozed tab notification to user
func triggerSnoozedTab(p *events.ScheduleSnoozedTabPayload) error {
	db := db.New()
	r := newRepository(db)

	snoozedTab, err := getSnoozedTab(db, p.UserId, p.SpaceId, p.SnoozedTabId)

	if err != nil {
		return err
	}

	// create notification
	n := &notification{
		Id:        strconv.FormatInt(time.Now().UTC().Unix(), 10),
		Type:      NotificationTypeUnSnoozedType,
		IsRead:    false,
		Timestamp: time.Now().UTC().Unix(),
		SnoozedTab: &snoozedTabNotification{
			URL:       snoozedTab.URL,
			SnoozedAt: p.SnoozedTabId,
			Title:     snoozedTab.Title,
			Icon:      snoozedTab.Icon,
		},
	}

	err = r.create(p.UserId, n)

	if err != nil {
		return err
	}

	pushEvent := &WebPushEvent[notification]{
		Event:   PushNotificationEventTypeNotification,
		Payload: n,
	}

	err = pushEvent.send(p.UserId, r)

	if err != nil {
		return err
	}

	// delete snoozed tab
	err = deleteSnoozedTab(db, p.UserId, p.SpaceId, p.SnoozedTabId)

	if err != nil {
		return err
	}
	return nil

}

// * helpers
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
		logger.Error("error removing note remainder", err)
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
		logger.Error("error deleting snoozed tab", err)
		return err
	}

	return nil
}
