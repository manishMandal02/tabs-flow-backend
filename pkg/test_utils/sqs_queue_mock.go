package test_utils

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/stretchr/testify/mock"
)

func NewQueueMock() *events.Queue {
	return &events.Queue{
		Client: new(SQSClientMock),
		URL:    "https://email-queue-url.com",
	}
}

type SQSClientMock struct {
	mock.Mock
}

func (m *SQSClientMock) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}
func (m *SQSClientMock) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}
