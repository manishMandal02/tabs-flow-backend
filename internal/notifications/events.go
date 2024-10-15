package notifications

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	lambda_events "github.com/aws/aws-lambda-go/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

func EventsHandler(_ context.Context, event lambda_events.SQSEvent) (interface{}, error) {
	// TODO: handle multiple events process
	if len(event.Records) < 1 {
		errMsg := "no events to process"
		logger.Errorf(errMsg)

		return nil, errors.New(errMsg)
	}

	for _, record := range event.Records {
		event := events.Event[any]{}

		err := event.FromJSON(record.Body)
		if err != nil {
			logger.Errorf("error unmarshalling event: %v", err)
			continue
		}
		eventType := event.EventType
		err = processEvent(&event)

		if err != nil {
			logger.Errorf("error processing event: %v", err)
			continue
		}
	}

	return nil, nil
}

func processEvent(event *events.Event[any]) error {

	switch event.EventType {
	case events.EventTypeScheduleNoteRemainder:
		return validateAndHandle(event, scheduleNoteRemainder)

	case events.EventTypeScheduleSnoozedTab:
		return validateAndHandle(event, scheduleSnoozedTab)

	case events.EventTypeTriggerNoteRemainder:
		return validateAndHandle(event, triggerNoteRemainder)

	case events.EventTypeTriggerSnoozedTab:
		return validateAndHandle(event, triggerSnoozedTab)
	}

	return nil
}

func scheduleNoteRemainder(p events.ScheduleNoteRemainderPayload) error {
	scheduler := events.NewScheduler()
	var err error

	switch p.SubEvent {
	case events.SubEventCreate:
		triggerEvent := events.New(events.EventTypeTriggerNoteRemainder, &events.ScheduleNoteRemainderPayload{
			NoteId:    p.NoteId,
			TriggerAt: p.TriggerAt,
		})

		evStr := triggerEvent.ToJSON()

		err = scheduler.CreateSchedule(p.NoteId, p.TriggerAt, &evStr)
	case events.SubEventUpdate:
		err = scheduler.UpdateSchedule(p.NoteId, p.TriggerAt)
	case events.SubEventDelete:
		err = scheduler.DeleteSchedule(p.NoteId)
	}

	return err
}

func scheduleSnoozedTab(p events.ScheduleSnoozedTabPayload) error {

	scheduler := events.NewScheduler()
	var err error

	switch p.SubEvent {
	case events.SubEventCreate:
		triggerEvent := events.New(events.EventTypeTriggerNoteRemainder, &events.ScheduleSnoozedTabPayload{
			SnoozedTabId: p.SnoozedTabId,
			TriggerAt:    p.TriggerAt,
		})

		evStr := triggerEvent.ToJSON()

		err = scheduler.CreateSchedule(p.SnoozedTabId, p.TriggerAt, &evStr)
	case events.SubEventUpdate:
		err = scheduler.UpdateSchedule(p.SnoozedTabId, p.TriggerAt)
	case events.SubEventDelete:
		err = scheduler.DeleteSchedule(p.SnoozedTabId)
	}

	return err
}

func triggerNoteRemainder(p events.ScheduleNoteRemainderPayload) error {

	// TODO: send notification to user

	return nil

}

func triggerSnoozedTab(p events.ScheduleSnoozedTabPayload) error {

	// TODO: send notification to user

	return nil
}

// * helpers
// assert payload and handle event
func validateAndHandle[T any](event *events.Event[any], handler func(T) error) error {
	payload, ok := (*event.Payload).(T)
	if !ok {
		err := fmt.Errorf("payload is not of type %s", reflect.TypeOf((*T)(nil)).Elem())
		logger.Errorf("Error asserting payload: %v", err)
		return err
	}
	return handler(payload)
}
