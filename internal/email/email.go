package email

import (
	"context"
	"fmt"
	"reflect"

	lambda_events "github.com/aws/aws-lambda-go/events"

	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

func SendEmail(_ context.Context, event lambda_events.SQSEvent) (interface{}, error) {
	if len(event.Records) == 0 {
		err := fmt.Errorf("no records found in event")
		logger.Error(err.Error(), err)
		return nil, err
	}
	//  process batch of events
	for _, record := range event.Records {
		event := events.Event[any]{}

		logger.Info("processing event_type: %v", event.EventType)

		err := event.FromJSON(record.Body)
		if err != nil {
			logger.Errorf("error un_marshalling event: %v", err)
			continue
		}

		err = processEvent(&event)

		if err != nil {
			logger.Errorf("error processing event: %v", err)
			continue
		}

		// remove message from sqs
		q := events.NewEmailQueue()

		err = q.DeleteMessage(record.ReceiptHandle)

		if err != nil {
			return nil, err
		}

	}

	return nil, nil
}

func processEvent(event *events.Event[any]) error {

	switch events.EventType(event.EventType) {
	case events.EventTypeSendOTP:
		return validateAndHandle(event, handleSendOTPMail)
	case events.EventTypeUserRegistered:
		return validateAndHandle(event, handleUserRegistered)

	default:
		logger.Errorf("Unknown sqs event: %v", event.EventType)
	}

	return nil

}

func handleSendOTPMail(event *events.Event[events.SendOTPPayload]) error {

	to := &NameAddr{
		Name:    event.Payload.Email,
		Address: event.Payload.Email,
	}

	z := NewZeptoMail()

	otp := event.Payload.OTP

	err := z.SendOTPMail(otp, to)

	if err != nil {
		return err
	}

	return nil
}

func handleUserRegistered(event *events.Event[events.UserRegisteredPayload]) error {
	z := NewZeptoMail()

	to := &NameAddr{
		Name:    event.Payload.Email,
		Address: event.Payload.Email,
	}

	err := z.sendWelcomeMail(to, event.Payload.TrailEndDate)

	if err != nil {
		return err
	}

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
