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

		otp := ev.Attributes["otp"]

		err := sendOTPMail(otp, to)

		if err != nil {
			return err
		}
		// TODO - remove message/event from sqs
		return nil

	case events.USER_REGISTERED:
		// TODO handle send welcome email
		return nil
	default:
		logger.Error(fmt.Sprintf("Unknown sqs event: %v", eT), fmt.Errorf("unknown event"))
	}
	return nil
}

/*
curl "https://api.zeptomail.in/v1.1/email" \
        -X POST \
        -H "Accept: application/json" \
        -H "Content-Type: application/json" \
        -H "Authorization:Zoho-enczapikey KEY" \
        -d '{
        "from": {"address": "support@tabsflow.com"},
        "to": [{"email_address": {"address": "hello@manishmandal.com","name": "Manish"}}],
        "subject":"Test Email",
        "htmlbody":"<div><b> Test email sent successfully. </b></div>"}'
*/
