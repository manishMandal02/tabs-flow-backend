package events

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type Queue struct {
	client *sqs.Client
	url    string
}

func NewQueue() *Queue {
	client := sqs.NewFromConfig(config.AWS_CONFIG)

	return &Queue{
		client: client,
		url:    config.EMAIL_SQS_QUEUE_URL,
	}

}

// sqs helper fn to send messages
func (q Queue) AddMessage(ev Event) error {

	res, err := q.client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		DelaySeconds:      *aws.Int32(1),
		QueueUrl:          &q.url,
		MessageBody:       aws.String(ev.GetEventType().String()),
		MessageAttributes: ev.ToMsgAttributes(),
	})

	if err != nil || res.MessageId == nil {
		logger.Error(fmt.Sprintf("Error sending message to SQS queue for event_type: %v", ev.GetEventType().String()), err)
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
		logger.Error(fmt.Sprintf("Error deleting message from SQS queue for receipt_handle: %v", r), err)
		return err
	}

	return nil
}
