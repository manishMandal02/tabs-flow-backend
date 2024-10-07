package email

import (
	"context"
	"fmt"

	lambda_events "github.com/aws/aws-lambda-go/events"

	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

func SendEmail(_ context.Context, ev lambda_events.SQSEvent) error {
	// TODO:handle multiple events process

	eventType, err := events.ParseEventType(*ev.Records[0].MessageAttributes["event_type"].StringValue)

	if err != nil {
		logger.Error("Error paring SQS event", err)
		return err
	}

	switch eventType {
	case events.SEND_OTP:
		to := &NameAddr{
			Name:    *ev.Records[0].MessageAttributes["email"].StringValue,
			Address: *ev.Records[0].MessageAttributes["email"].StringValue,
		}

		z := NewZeptoMail()

		otp := *ev.Records[0].MessageAttributes["otp"].StringValue

		err := z.SendOTPMail(otp, to)

		if err != nil {
			return err
		}
		// remove message from sqs
		q := events.NewQueue()

		err = q.DeleteMessage(ev.Records[0].ReceiptHandle)

		if err != nil {
			return err
		}

		return nil

	case events.USER_REGISTERED:
		z := NewZeptoMail()

		to := &NameAddr{
			Name:    *ev.Records[0].MessageAttributes["name"].StringValue,
			Address: *ev.Records[0].MessageAttributes["email"].StringValue,
		}

		trailEndDate := *ev.Records[0].MessageAttributes["trail_end_date"].StringValue

		err := z.sendWelcomeMail(to, trailEndDate)

		if err != nil {
			return err
		}

		// remove message from sqs
		q := events.NewQueue()

		err = q.DeleteMessage(ev.Records[0].ReceiptHandle)

		if err != nil {
			return err
		}
		return nil

	default:
		logger.Error(fmt.Sprintf("Unknown sqs event: %v", eventType), fmt.Errorf("unknown event"))
	}
	return nil
}
