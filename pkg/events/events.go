package eventTypes

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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

type EventType int

// Events
const (
	SEND_OTP EventType = iota
	USER_REGISTERED
	SCHEDULE_TASK
)

type IEvent interface {
	GetEventType() EventType
}

// String method to get the string representation of the Event
func (e EventType) String() string {
	return [...]string{"SEND_OTP", "USER_REGISTERED", "PASSWORD_RESET"}[e]
}

type SEND_OTP_EVENT struct {
	EventType EventType `json:"eventType"`
	Payload   struct {
		Email string `json:"email"`
		OTP   string `json:"otp"`
	}
}

func (e SEND_OTP_EVENT) GetEventType() EventType {
	return e.EventType
}

// sqs helper fn to send messages
func (q *Queue) AddMsgToQueue(ev IEvent) error {
	res, err := q.client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		DelaySeconds: *aws.Int32(1),
		QueueUrl:     &q.url,
		MessageBody:  aws.String("Send OTP to user"),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"Email": {
				DataType:    aws.String("String"),
				StringValue: aws.String("user@example.com"),
			},
			"EventType": {
				DataType:    aws.String("String"),
				StringValue: aws.String("OTP"),
			},
			"Payload": {
				DataType:    aws.String("String"),
				StringValue: aws.String(ev.GetEventType().String()),
			},
		},
	})

	if err != nil || res.MessageId == nil {
		logger.Error("Error sending message to SQS queue: %v", err)
		return err
	}

	return nil
}
