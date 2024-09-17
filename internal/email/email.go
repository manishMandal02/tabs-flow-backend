package email

import (
	"context"
	"encoding/json"
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
		// TODO handle send OTP email
		return nil
	case events.USER_REGISTERED:
		// TODO handle send welcome email
		return nil
	default:
		logger.Error(fmt.Sprintf("Unknown sqs event: %v", eT), fmt.Errorf("unknown event"))
	}

	var emailEvent events.SendOTP_Payload

	jsonStr, err := json.Marshal(ev.Attributes)

	if err != nil {
		logger.Error("Error un_marshaling email event", err)
		return err
	}

	if err = json.Unmarshal(jsonStr, &emailEvent); err != nil {
		logger.Error("Error marshaling email event", err)
		return err
	}

	// TODO - send email

	// TODO - load config from env
	// cfg, err := config.LoadDefaultConfig(context.TODO())
	// if err != nil {
	// 	log.Fatalf("failed to load configuration, %v", err)
	// }
	// https://aws.github.io/aws-sdk-go-v2/docs/configuring-sdk/

	// TODO - read sqs message
	// TODO - send email through ses

	fmt.Println("sqs event: ", emailEvent)

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
