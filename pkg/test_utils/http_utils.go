package test_utils

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/stretchr/testify/suite"
)

// TestConfig holds environment-specific configuration
type TestConfig struct {
	BaseURL      string
	DDBTableName string
	SQSQueueURL  string
	TestEmail    string
}

// E2ETestSuite base suite with common utilities
type E2ETestSuite struct {
	suite.Suite
	Config     TestConfig
	HTTPClient *http.Client
	DDBClient  *dynamodb.Client
	SQSClient  *sqs.Client
}

// WaitForSQSMessage waits for and verifies an SQS message
func (s *E2ETestSuite) WaitForSQSMessage(ctx context.Context, expectedEmail string) (*events.SendOTPPayload, error) {
	for i := 0; i < 3; i++ { // retry 3 times
		output, err := s.SQSClient.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
			QueueUrl:            &s.Config.SQSQueueURL,
			MaxNumberOfMessages: 1,
			WaitTimeSeconds:     5,
		})
		if err != nil {
			return nil, err
		}

		for _, msg := range output.Messages {
			var payload events.SendOTPPayload
			if err := json.Unmarshal([]byte(*msg.Body), &payload); err != nil {
				continue
			}

			if payload.Email == expectedEmail {
				// Delete the message
				_, _ = s.SQSClient.DeleteMessage(ctx, &sqs.DeleteMessageInput{
					QueueUrl:      &s.Config.SQSQueueURL,
					ReceiptHandle: msg.ReceiptHandle,
				})
				return &payload, nil
			}
		}

		time.Sleep(time.Second * 2)
	}
	return nil, nil
}

// AssertSuccessResponse checks if the API response indicates success
func (s *E2ETestSuite) AssertSuccessResponse(resp *http.Response) interface{} {
	var apiResp http_api.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&apiResp)
	s.Require().NoError(err)
	s.Require().True(apiResp.Success)
	return apiResp.Data
}

// AssertErrorResponse checks if the API response indicates an error
func (s *E2ETestSuite) AssertErrorResponse(resp *http.Response, expectedStatus int) string {
	s.Require().Equal(expectedStatus, resp.StatusCode)
	var apiResp http_api.APIResponse
	err := json.NewDecoder(resp.Body).Decode(&apiResp)
	s.Require().NoError(err)
	s.Require().False(apiResp.Success)
	return apiResp.Message
}

// AssertHasCookie checks if a specific cookie is present
func (s *E2ETestSuite) AssertHasCookie(resp *http.Response, name string) *http.Cookie {
	for _, cookie := range resp.Cookies() {
		if cookie.Name == name {
			s.Assert().True(cookie.Secure)
			s.Assert().True(cookie.HttpOnly)
			return cookie
		}
	}
	s.Fail("Cookie not found: " + name)
	return nil
}
