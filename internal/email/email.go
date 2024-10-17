package email

import (
	"context"
	"fmt"

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

		logger.Info("processing record: %v", record.Body)

		eventType := *record.MessageAttributes["event_type"].StringValue

		logger.Info("processing event_type: %v", event.EventType)

		err := processEvent(eventType, record.Body)

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

func processEvent(eventType string, body string) error {

	switch events.EventType(eventType) {
	case events.EventTypeSendOTP:

		ev, err := events.NewFromJSON[events.SendOTPPayload](body)

		if err != nil {
			logger.Errorf("error un_marshalling event: %v", err)
			return err
		}

		return handleSendOTPMail(*ev.Payload)

	case events.EventTypeUserRegistered:
		ev, err := events.NewFromJSON[events.UserRegisteredPayload](body)

		if err != nil {
			logger.Errorf("error un_marshalling event: %v", err)
			return err
		}

		return handleUserRegistered(*ev.Payload)

	default:
		logger.Errorf("Unknown sqs event: %v", eventType)
	}

	return nil

}

func handleSendOTPMail(payload events.SendOTPPayload) error {

	to := &NameAddr{
		Name:    payload.Email,
		Address: payload.Email,
	}

	z := NewZeptoMail()

	otp := payload.OTP

	err := z.SendOTPMail(otp, to)

	if err != nil {
		return err
	}

	return nil
}

func handleUserRegistered(payload events.UserRegisteredPayload) error {
	z := NewZeptoMail()

	to := &NameAddr{
		Name:    payload.Email,
		Address: payload.Email,
	}

	err := z.sendWelcomeMail(to, payload.TrailEndDate)

	if err != nil {
		return err
	}

	return nil
}
