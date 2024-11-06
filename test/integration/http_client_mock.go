package integration_test

import (
	"net/http"

	"github.com/stretchr/testify/mock"
)

type mockClient struct {
	mock.Mock
}

func (r *mockClient) Do(req *http.Request) (*http.Response, error) {
	args := r.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}
