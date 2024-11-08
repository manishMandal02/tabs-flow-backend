package integration_test

import (
	"context"

	"github.com/PaddleHQ/paddle-go-sdk"
	"github.com/stretchr/testify/mock"
)

type PaddleClientMock struct {
	mock.Mock
}

func NewPaddleClientMock() *PaddleClientMock {
	return new(PaddleClientMock)
}

func (p *PaddleClientMock) GetSubscription(ctx context.Context, req *paddle.GetSubscriptionRequest) (res *paddle.Subscription, err error) {

	args := p.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*paddle.Subscription), args.Error(1)
}
