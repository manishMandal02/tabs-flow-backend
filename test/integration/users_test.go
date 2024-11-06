package integration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/manishMandal02/tabsflow-backend/internal/users"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/manishMandal02/tabsflow-backend/pkg/test_utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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

type testSetup struct {
	router     http.Handler
	mockDB     *db.DDB
	mockClient *mockClient
	mockQueue  *events.Queue
}

func newTestSetup() *testSetup {
	db := test_utils.NewDDBMock()
	q := test_utils.NewQueueMock()

	httpClient := new(mockClient)
	return &testSetup{
		mockDB:     db,
		router:     users.Router(db, q, httpClient),
		mockQueue:  q,
		mockClient: httpClient,
	}
}

var testUser = &users.User{
	Id:         "123",
	FullName:   "Test Name",
	Email:      "test@test.com",
	ProfilePic: "https://test.com/test.png",
}

type TestCase struct {
	name            string
	method          string
	path            string
	body            interface{}
	setupAuth       func(r *http.Request)
	setupMockClient func(*mockClient)
	setupMockQueue  func(*testing.T, *test_utils.SQSClientMock)
	setupMockDB     func(*testing.T, *test_utils.DynamoDBClientMock)
	expectedStatus  int
	expectedBody    interface{}
}

var tests = []TestCase{
	{
		name:           "POST-/users/ > empty body error",
		method:         "POST",
		path:           "/",
		body:           nil,
		expectedStatus: http.StatusBadRequest,
		expectedBody:   users.ErrMsg.CreateUser,
	},
	{
		name:   "POST-/users/ > invalid body error",
		method: "POST",
		path:   "/",
		body: map[string]string{
			"id":   "123",
			"name": "Test Name",
		},
		expectedStatus: http.StatusBadRequest,
		expectedBody:   users.ErrMsg.CreateUser,
	},
	{
		name:           "POST-/users/ > error checking if user exists",
		method:         "POST",
		path:           "/",
		body:           testUser,
		expectedStatus: http.StatusBadGateway,
		expectedBody:   users.ErrMsg.GetUser,
		setupMockDB: func(t *testing.T, mockDB *test_utils.DynamoDBClientMock) {
			// Mock DynamoDB PutItem response
			mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(nil, errors.New("error checking if user exists"))
		},
	},
	{
		name:           "POST-/users/ > user exists error",
		method:         "POST",
		path:           "/",
		body:           testUser,
		expectedStatus: http.StatusBadRequest,
		expectedBody:   users.ErrMsg.UserExists,
		setupMockDB: func(t *testing.T, mockDB *test_utils.DynamoDBClientMock) {
			// Mock DynamoDB PutItem response
			mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"PK":       &types.AttributeValueMemberS{Value: testUser.Id},
					"SK":       &types.AttributeValueMemberS{Value: "Profile"},
					"FullName": &types.AttributeValueMemberS{Value: testUser.FullName},
					"Email":    &types.AttributeValueMemberS{Value: testUser.Email},
				},
			}, nil)
		},
	},
	{
		name:           "POST-/users/ > unable to reach auth service",
		method:         "POST",
		path:           "/",
		body:           testUser,
		expectedStatus: http.StatusInternalServerError,
		expectedBody:   users.ErrMsg.CreateUser,
		setupMockDB: func(t *testing.T, mockDB *test_utils.DynamoDBClientMock) {
			// Mock DynamoDB PutItem response
			mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)
		},
		setupMockClient: func(mockClient *mockClient) {
			// Mock HTTP request
			mockClient.On("Do", mock.Anything).Return(&http.Response{}, errors.New("failed to reach auth service"))
		},
	},
	{
		name:           "POST-/users/ > session not found for user from auth service, redirect to logout",
		method:         "POST",
		path:           "/",
		body:           testUser,
		expectedStatus: http.StatusTemporaryRedirect,
		setupMockDB: func(t *testing.T, mockDB *test_utils.DynamoDBClientMock) {
			// Mock DynamoDB PutItem response
			mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)
		},
		setupMockClient: func(mockClient *mockClient) {
			// Mock HTTP request
			mockClient.On("Do", mock.Anything).Return(&http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader("{\"error\":\"session not found\"}")),
			}, nil)
		},
	},
	{
		name:           "POST-/users/ > invalid user session, redirect to logout",
		method:         "POST",
		path:           "/",
		body:           testUser,
		expectedStatus: http.StatusTemporaryRedirect,
		setupMockDB: func(t *testing.T, mockDB *test_utils.DynamoDBClientMock) {
			// Mock DynamoDB PutItem response
			mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)
		},
		setupMockClient: func(mockClient *mockClient) {
			// Mock HTTP request
			mockClient.On("Do", mock.Anything).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"data":{
						"userId": "1222-Wrong user id",
						"name": "New User",
						"email": "test@test.com"
					}
					}`)),
			}, nil)
		},
	},
	{
		name:           "POST-/users/ > error inserting data into dynamodb",
		method:         "POST",
		path:           "/",
		body:           testUser,
		expectedStatus: http.StatusBadGateway,
		expectedBody:   users.ErrMsg.CreateUser,
		setupMockDB: func(t *testing.T, mockDB *test_utils.DynamoDBClientMock) {
			// Mock DynamoDB PutItem response
			mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)

			mockDB.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput"), mock.Anything).Return(nil, errors.New("error inserting data into dynamodb"))
		},
		setupMockClient: func(mockClient *mockClient) {
			// Mock HTTP request
			mockClient.On("Do", mock.Anything).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"data":{
						"userId": "123",
						"name": "New User",
						"email": "test@test.com"
					}
					}`)),
			}, nil)
		},
	},
	{
		name:           "POST-/users/ > error user_registered event to sqs queue",
		method:         "POST",
		path:           "/",
		body:           testUser,
		expectedStatus: http.StatusInternalServerError,
		expectedBody:   users.ErrMsg.CreateUser,
		setupMockDB: func(t *testing.T, mockDB *test_utils.DynamoDBClientMock) {
			// Mock DynamoDB PutItem response
			mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)

			mockDB.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput"), mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)

		},
		setupMockQueue: func(t *testing.T, mockQueue *test_utils.SQSClientMock) {
			mockQueue.On("SendMessage", mock.AnythingOfType("*sqs.SendMessageInput"), mock.Anything).Return(nil, errors.New("sqs error"))
		},
		setupMockClient: func(mockClient *mockClient) {
			// Mock HTTP request
			mockClient.On("Do", mock.Anything).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
					"data":{
						"userId": "123",
						"name": "New User",
						"email": "test@test.com"
					}
					}`)),
			}, nil)
		},
	},

	{
		name:           "POST-/users/ > success",
		method:         "POST",
		path:           "/",
		body:           testUser,
		expectedStatus: http.StatusOK,
		expectedBody:   map[string]interface{}{"success": true, "message": "user created"},
		setupMockDB: func(t *testing.T, mockDB *test_utils.DynamoDBClientMock) {

			mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Run(
				(func(args mock.Arguments) {
					// verify ddb get item key
					input := args.Get(1).(*dynamodb.GetItemInput)
					assert.Equal(t, "MainTable_test", *input.TableName)
					assert.Equal(t, &types.AttributeValueMemberS{Value: testUser.Id}, input.Key["PK"])
					assert.Equal(t, &types.AttributeValueMemberS{Value: "P#Profile"}, input.Key["SK"])
				}),
			).Return(&dynamodb.GetItemOutput{}, nil)

			mockDB.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput"), mock.Anything).Run(
				(func(args mock.Arguments) {
					// verify ddb put item values
					input := args.Get(1).(*dynamodb.PutItemInput)
					assert.Equal(t, "MainTable_test", *input.TableName)
					assert.Equal(t, &types.AttributeValueMemberS{Value: testUser.Id}, input.Item["PK"])
				}),
			).Return(&dynamodb.PutItemOutput{}, nil)

		},
		setupMockQueue: func(t *testing.T, mockQueue *test_utils.SQSClientMock) {
			mockQueue.On("SendMessage", mock.AnythingOfType("*sqs.SendMessageInput"), mock.Anything).Run(
				(func(args mock.Arguments) {
					// verify message body and type
					input := args.Get(0).(*sqs.SendMessageInput)

					ev, err := events.NewFromJSON[events.UserRegisteredPayload](*input.MessageBody)
					require.NoError(t, err)

					assert.Equal(t, string(events.EventTypeUserRegistered), *input.MessageAttributes["event_type"].StringValue)
					assert.Equal(t, testUser.Email, ev.Payload.Email)
					assert.Equal(t, testUser.FullName, ev.Payload.Name)
					assert.NotEmpty(t, ev.Payload.TrailEndDate)
				}),
			).Return(&sqs.SendMessageOutput{MessageId: aws.String("123")}, nil)

		},
		setupMockClient: func(requestIntr *mockClient) {
			// Mock the outgoing request to the authentication service
			requestIntr.On("Do", mock.Anything).Return(&http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(bytes.NewBufferString(`{
				"data":{
					"userId": "123",
					"name": "New User",
					"email": "test@test.com"
				}
				}`)),
			}, nil)
		},
	}}

// * run test cases
func TestUsersService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create new test setup for each test case
			setup := newTestSetup()

			var mockedQueue *test_utils.SQSClientMock

			mockedDB, ok := setup.mockDB.Client.(*test_utils.DynamoDBClientMock)

			if !ok {
				t.Fatal("failed to get mock db client")
			}

			// setup mock interceptor
			if tc.setupMockClient != nil {
				tc.setupMockClient(setup.mockClient)
			}

			// setup mock db
			if tc.setupMockDB != nil {
				tc.setupMockDB(t, mockedDB)
			}

			// setup mock queue
			if tc.setupMockQueue != nil {
				mockedQueue = setup.mockQueue.Client.(*test_utils.SQSClientMock)
				tc.setupMockQueue(t, mockedQueue)
			}

			// create http request
			var reqBody []byte
			var err error
			if tc.body != nil {
				reqBody, err = json.Marshal(tc.body)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(tc.method, tc.path, bytes.NewBuffer(reqBody))
			if tc.setupAuth != nil {
				tc.setupAuth(req)
			}

			//  headers
			req.Header.Set("Content-Type", "application/json")

			// recorder
			w := httptest.NewRecorder()

			// serve request through the router
			setup.router.ServeHTTP(w, req)

			// assertions
			assert.Equal(t, tc.expectedStatus, w.Code)

			if tc.expectedBody != nil {
				// check if expected body is a string
				if s, ok := tc.expectedBody.(string); ok {
					assert.Equal(t, s, strings.TrimSpace(w.Body.String()))
				} else {
					var response map[string]interface{}
					err := json.NewDecoder(w.Body).Decode(&response)
					require.NoError(t, err)
					assert.Equal(t, tc.expectedBody, response)
				}
			}

			// verify all mock assertions
			mockedDB.AssertExpectations(t)

			if mockedQueue != nil {
				mockedQueue.AssertExpectations(t)
			}

		})
	}
}
