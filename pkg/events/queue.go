package events

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type Queue struct {
	client *sqs.Client
	url    string
}

func NewEmailQueue() *Queue {
	client := sqs.NewFromConfig(config.AWS_CONFIG)

	return &Queue{
		client: client,
		url:    config.EMAIL_QUEUE_URL,
	}
}

func NewNotificationQueue() *Queue {
	client := sqs.NewFromConfig(config.AWS_CONFIG)

	return &Queue{
		client: client,
		url:    config.NOTIFICATIONS_QUEUE_URL,
	}

}

// sqs helper fn to send messages
func (q Queue) AddMessage(ev IEvent) error {

	res, err := q.client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		DelaySeconds:      *aws.Int32(1),
		QueueUrl:          &q.url,
		MessageBody:       aws.String(ev.ToJSON()),
		MessageAttributes: ev.ToMsgAttributes(),
	})

	if err != nil || res.MessageId == nil {
		logger.Errorf("Error sending message to SQS queue for event_type: %v. \n [Error]: %v", ev.GetEventType(), err)
		return err
	}

	return nil
}

func (q Queue) DeleteMessage(r string) error {

	_, err := q.client.DeleteMessage(context.TODO(), &sqs.DeleteMessageInput{
		QueueUrl:      &q.url,
		ReceiptHandle: aws.String(r),
	})

	if err != nil {
		logger.Errorf("Error deleting message from SQS queue for receipt_handle: %v. \n [Error]: %v", r, err)
		return err
	}

	return nil
}
