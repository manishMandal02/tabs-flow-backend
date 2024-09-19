package email

import (
	"context"
	"fmt"

	lambda_events "github.com/aws/aws-lambda-go/events"

	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

func SendEmail(_ context.Context, ev lambda_events.SQSMessage) error {

	eT, err := events.ParseEventType(ev.Attributes["EventType"])

	if err != nil {
		logger.Error("Error paring SQS event", err)
	}

	switch eT {
	case events.SEND_OTP:
		to := &nameAddr{
			Name:    ev.Attributes["email"],
			Address: ev.Attributes["email"],
		}

		z := newZeptoMail()

		otp := ev.Attributes["otp"]

		err := z.sendOTPMail(otp, to)

		if err != nil {
			return err
		}
		// remove message from sqs
		q := events.NewQueue()

		err = q.DeleteMessage(ev.ReceiptHandle)

		if err != nil {
			return err
		}

		return nil

	case events.USER_REGISTERED:
		z := newZeptoMail()

		to := &nameAddr{
			Name:    ev.Attributes["name"],
			Address: ev.Attributes["email"],
		}

		trailEndDate := ev.Attributes["trail_end_date"]

		z.sendWelcomeMail(to, trailEndDate)

		if err != nil {
			return err
		}
		// remove message from sqs
		q := events.NewQueue()

		err = q.DeleteMessage(ev.ReceiptHandle)

		if err != nil {
			return err
		}
		return nil

	default:
		logger.Error(fmt.Sprintf("Unknown sqs event: %v", eT), fmt.Errorf("unknown event"))
	}
	return nil
}
