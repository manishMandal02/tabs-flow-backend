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
	client := sqs.New(sqs.Options{
		Region: config.AWS_REGION,
	})
	return &Queue{
		client: client,
		url:    config.SQS_QUEUE_URL,
	}

}

// sqs helper fn to send messages
func (q *Queue) AddMsgToQueue(ev Event) error {

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
