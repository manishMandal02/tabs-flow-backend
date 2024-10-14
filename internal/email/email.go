package email

import (
	"context"
	"fmt"

	lambda_events "github.com/aws/aws-lambda-go/events"

	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

func SendEmail(_ context.Context, event lambda_events.SQSEvent) error {
	// TODO:handle multiple events process

	var err error

	if len(event.Records) == 0 {
		err = fmt.Errorf("no records found in event")
		logger.Error(err.Error(), err)
		return err
	}

	message := event.Records[0]

	eventType := *message.MessageAttributes["event_type"].StringValue

	if eventType == "" {
		err = fmt.Errorf("event_type is empty")
		logger.Error("event_type field missing in message", err)
		return err
	}

	logger.Info("event_type: %v", eventType)

	switch events.EventType(eventType) {
	case events.EventTypeSendOTP:
		ev := &events.Event[events.SendOTPPayload]{}

		err := ev.FromJSON(message.Body)

		if err != nil {
			logger.Error("error un_marshalling event", err)
			return err
		}

		to := &NameAddr{
			Name:    ev.Payload.Email,
			Address: ev.Payload.Email,
		}

		z := NewZeptoMail()

		otp := ev.Payload.OTP

		err = z.SendOTPMail(otp, to)

		if err != nil {
			return err
		}

		// remove message from sqs
		q := events.NewEmailQueue()

		err = q.DeleteMessage(message.ReceiptHandle)

		if err != nil {
			return err
		}

		return nil

	case events.EventTypeUserRegistered:
		z := NewZeptoMail()

		ev := &events.Event[events.UserRegisteredPayload]{}

		err = ev.FromJSON(message.Body)

		if err != nil {
			logger.Error("error un_marshalling event", err)
			return err
		}

		to := &NameAddr{
			Name:    ev.Payload.Email,
			Address: ev.Payload.Email,
		}

		err = z.sendWelcomeMail(to, ev.Payload.TrailEndDate)

		if err != nil {
			return err
		}

		// remove message from sqs
		q := events.NewEmailQueue()

		err = q.DeleteMessage(message.ReceiptHandle)

		if err != nil {
			return err
		}
		return nil

	default:
		logger.Error(fmt.Sprintf("Unknown sqs event: %v", eventType), fmt.Errorf("unknown event"))
	}
	return nil
}
