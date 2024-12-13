package integration_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/PaddleHQ/paddle-go-sdk"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/internal/users"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/events"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type testSetup struct {
	router           http.Handler
	mockDB           *db.DDB
	mockClient       *mockClient
	mockQueue        *events.Queue
	mockPaddleClient *PaddleClientMock
}

func newTestSetup() *testSetup {
	db := NewDDBMock()
	q := NewQueueMock()
	p := NewPaddleClientMock()

	httpClient := new(mockClient)
	return &testSetup{
		mockDB:           db,
		router:           users.Router(db, q, httpClient, p),
		mockQueue:        q,
		mockClient:       httpClient,
		mockPaddleClient: p,
	}
}

// helper
func mockClientAuthAPISuccessRes(userId string) func(*mockClient) {
	return func(mockClient *mockClient) {
		mockClient.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(bytes.NewBufferString(fmt.Sprintf(`{
				"data":{
					"userId": "%v",
					"firstName": "New",
					"lastName": "User",
					"email": "test@test.com"
				}
				}`, userId))),
		}, nil)
	}
}

func mockClientSpacesAPISuccessRes() func(*mockClient) {
	return func(mockClient *mockClient) {
		mockClient.On("Do", mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
		}, nil)
	}
}

func mockDBGetUser(mockDB *DynamoDBClientMock) {

	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: testUser.Id},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.Profile},
	}
	mockDB.On("GetItem", mock.Anything, &dynamodb.GetItemInput{
		TableName: aws.String("MainTable_test"),
		Key:       key,
	}, mock.Anything).Return(&dynamodb.GetItemOutput{
		Item: map[string]types.AttributeValue{
			"PK":         &types.AttributeValueMemberS{Value: testUser.Id},
			"SK":         &types.AttributeValueMemberS{Value: db.SORT_KEY.Profile},
			"FirstName":  &types.AttributeValueMemberS{Value: testUser.FirstName},
			"LastName":   &types.AttributeValueMemberS{Value: testUser.LastName},
			"email":      &types.AttributeValueMemberS{Value: testUser.Email},
			"profilePic": &types.AttributeValueMemberS{Value: testUser.ProfilePic},
		},
	}, nil).Once()

}

func mockDBGetUserSubscription(mockDB *DynamoDBClientMock) {
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: testUser.Id},
		"SK": &types.AttributeValueMemberS{Value: db.SORT_KEY.Subscription},
	}
	mockDB.On("GetItem", mock.Anything, &dynamodb.GetItemInput{
		TableName: aws.String("MainTable_test"),
		Key:       key,
	}, mock.Anything).Return(&dynamodb.GetItemOutput{
		Item: map[string]types.AttributeValue{
			"PK":     &types.AttributeValueMemberS{Value: testUser.Id},
			"SK":     &types.AttributeValueMemberS{Value: db.SORT_KEY.Subscription},
			"Id":     &types.AttributeValueMemberS{Value: "1"},
			"Plan":   &types.AttributeValueMemberS{Value: "TRAIL"},
			"Status": &types.AttributeValueMemberS{Value: "active"},
			"Start":  &types.AttributeValueMemberN{Value: strconv.FormatInt((time.Now().Unix()), 10)},
			"End":    &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Add(time.Duration(time.Now().Day())+4).Unix(), 10)},
		},
	}, nil).Once()

}

var testUser = &users.User{
	Id:         "123",
	FirstName:  "Test",
	LastName:   "Name",
	Email:      "test@test.com",
	ProfilePic: "https://test.com/test.png",
}

type TestCase struct {
	name                      string
	method                    string
	path                      string
	body                      interface{}
	mockAuthHeader            func(r *http.Request) // mock authorizer's success res, add user id to header
	setupMockClient           func(*mockClient)
	setupMockQueue            func(*testing.T, *SQSClientMock)
	setupMockDB               func(*DynamoDBClientMock)
	setupMockPaddleClient     func(*PaddleClientMock)
	setupMockDBWithAssertions func(*testing.T, *DynamoDBClientMock)
	expectedStatus            int
	expectedBody              interface{}
}

func getUserByIDTestCases() []TestCase {
	return []TestCase{
		{
			name:           "GET-/users/me > empty user id",
			method:         "GET",
			path:           "/me",
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.InvalidUserId},
		},
		{
			name:           "GET-/users/me > dynamodb error",
			method:         "GET",
			path:           "/me",
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.GetUser},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{}, errors.New("error getting user by id"))
			},
		},
		{
			name:           "GET-/users/me > user not found",
			method:         "GET",
			path:           "/me",
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.UserNotFound},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{Item: nil}, nil)
			},
		},
		{
			name:           "GET-/users/me > success",
			method:         "GET",
			path:           "/me",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"id":         "123",
					"firstName":  "Test",
					"lastName":   "Name",
					"email":      "test@test.com",
					"profilePic": "https://test.com/test.png",
				},
			},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
			},
			setupMockClient: mockClientSpacesAPISuccessRes(),
		},
	}
}

func createUserTestCases() []TestCase {
	return []TestCase{
		{
			name:           "POST-/users/ > empty body error",
			method:         "POST",
			path:           "/",
			body:           nil,
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.CreateUser},
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
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.CreateUser},
		},
		{
			name:           "POST-/users/ > error checking if user exists",
			method:         "POST",
			path:           "/",
			body:           testUser,
			expectedStatus: http.StatusBadGateway,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.GetUser},
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(nil, errors.New("error checking if user exists"))
			},
		},
		{
			name:           "POST-/users/ > user exists error",
			method:         "POST",
			path:           "/",
			body:           testUser,
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.UserExists},
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
			},
		},
		{
			name:           "POST-/users/ > unable to reach auth service",
			method:         "POST",
			path:           "/",
			body:           testUser,
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.CreateUser},
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)
			},
			setupMockClient: func(mockClient *mockClient) {
				mockClient.On("Do", mock.Anything).Return(&http.Response{}, errors.New("failed to reach auth service"))
			},
		},
		{
			name:           "POST-/users/ > session not found for user from auth service, redirect to logout",
			method:         "POST",
			path:           "/",
			body:           testUser,
			expectedStatus: http.StatusTemporaryRedirect,
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)
			},
			setupMockClient: func(mockClient *mockClient) {
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
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)
			},
			setupMockClient: mockClientAuthAPISuccessRes("123-wrong-user-id"),
		},
		{
			name:           "POST-/users/ > error inserting data into dynamodb",
			method:         "POST",
			path:           "/",
			body:           testUser,
			expectedStatus: http.StatusBadGateway,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.CreateUser},
			setupMockClient: mockClientAuthAPISuccessRes(testUser.Id),
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)
				mockDB.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput"), mock.Anything).Return(nil, errors.New("error inserting data into dynamodb"))
			},
		},
		{
			name:           "POST-/users/ > error user_registered event to sqs queue",
			method:         "POST",
			path:           "/",
			body:           testUser,
			expectedStatus: http.StatusInternalServerError,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.CreateUser},
			setupMockClient: mockClientAuthAPISuccessRes(testUser.Id),
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Return(&dynamodb.GetItemOutput{}, nil)
				mockDB.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput"), mock.Anything).Return(&dynamodb.PutItemOutput{}, nil)

			},
			setupMockQueue: func(t *testing.T, mockQueue *SQSClientMock) {
				mockQueue.On("SendMessage", mock.AnythingOfType("*sqs.SendMessageInput"), mock.Anything).Return(nil, errors.New("sqs error"))
			},
		},

		{
			name:            "POST-/users/ > success",
			method:          "POST",
			path:            "/",
			body:            testUser,
			expectedStatus:  http.StatusOK,
			expectedBody:    map[string]interface{}{"success": true, "message": "user created"},
			setupMockClient: mockClientAuthAPISuccessRes(testUser.Id),
			setupMockDBWithAssertions: func(t *testing.T, mockDB *DynamoDBClientMock) {

				mockDB.On("GetItem", mock.Anything, mock.AnythingOfType("*dynamodb.GetItemInput"), mock.Anything).Run(
					(func(args mock.Arguments) {
						// verify ddb get item key
						input := args.Get(1).(*dynamodb.GetItemInput)
						assert.Equal(t, "MainTable_test", *input.TableName)
						assert.Equal(t, &types.AttributeValueMemberS{Value: testUser.Id}, input.Key["PK"])
						assert.Equal(t, &types.AttributeValueMemberS{Value: db.SORT_KEY.Profile}, input.Key["SK"])
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
			setupMockQueue: func(t *testing.T, mockQueue *SQSClientMock) {
				mockQueue.On("SendMessage", mock.AnythingOfType("*sqs.SendMessageInput"), mock.Anything).Run(
					(func(args mock.Arguments) {
						// verify message body and type
						input := args.Get(0).(*sqs.SendMessageInput)

						ev, err := events.NewFromJSON[events.UserRegisteredPayload](*input.MessageBody)
						require.NoError(t, err)

						assert.Equal(t, string(events.EventTypeUserRegistered), *input.MessageAttributes["event_type"].StringValue)
						assert.Equal(t, testUser.Email, ev.Payload.Email)
						assert.Equal(t, testUser.FirstName, ev.Payload.Name)
						assert.NotEmpty(t, ev.Payload.TrailEndDate)
					}),
				).Return(&sqs.SendMessageOutput{MessageId: aws.String("123")}, nil)

			},
		},
	}
}

func updateUserTestCases() []TestCase {
	return []TestCase{
		{
			name:           "PATCH-/users/ > invalid data",
			method:         "PATCH",
			path:           "/",
			body:           `{}`,
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.UpdateUser},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
			},
		},
		{
			name:   "PATCH-/users/ > dynamodb error updating user",
			method: "PATCH",
			path:   "/",
			body: map[string]string{
				"fullName": "Test Name 2",
			},
			expectedStatus: http.StatusBadGateway,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.UpdateUser},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
				mockDB.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput"), mock.Anything).Return(&dynamodb.UpdateItemOutput{}, errors.New("error updating user"))
			},
		},
		{
			name:   "PATCH-/users/ > update user success",
			method: "PATCH",
			path:   "/",
			body: map[string]string{
				"fullName": "Test Name 2",
			},
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"success": true, "message": "user updated"},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
				mockDB.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput"), mock.Anything).Return(&dynamodb.UpdateItemOutput{}, nil)
			},
		},
	}
}

func deleteUserTestCases() []TestCase {
	return []TestCase{
		{
			name:           "DELETE-/users/ > success",
			method:         "DELETE",
			path:           "/",
			expectedStatus: http.StatusOK,
			expectedBody:   map[string]interface{}{"success": true, "message": "user deleted"},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
				mockDB.On("Query", mock.Anything, mock.AnythingOfType("*dynamodb.QueryInput"), mock.Anything).Return(&dynamodb.QueryOutput{}, nil)
				mockDB.On("BatchWriteItem", mock.Anything, mock.AnythingOfType("*dynamodb.BatchWriteItemInput"), mock.Anything).Return(&dynamodb.BatchWriteItemOutput{}, nil)
			},
		},
	}
}

func getUserPreferencesTestCases() []TestCase {
	return []TestCase{
		{
			name:           "GET-/users/preferences > success",
			method:         "GET",
			path:           "/preferences",
			expectedStatus: http.StatusOK,
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
				mockDB.On("Query", mock.Anything, mock.AnythingOfType("*dynamodb.QueryInput"), mock.Anything).Return(&dynamodb.QueryOutput{
					Items: []map[string]types.AttributeValue{
						{
							"PK": &types.AttributeValueMemberS{Value: testUser.Id},
							"SK": &types.AttributeValueMemberS{Value: db.SORT_KEY.P_General},
						},
						{
							"PK": &types.AttributeValueMemberS{Value: testUser.Id},
							"SK": &types.AttributeValueMemberS{Value: db.SORT_KEY.P_CmdPalette},
						},
					},
				}, nil)
			},
		},
	}
}

func updateUserPreferencesTestCases() []TestCase {
	return []TestCase{
		{
			name:   "PATCH-/users/preferences > invalid data",
			method: "PATCH",
			path:   "/preferences",
			body: map[string]interface{}{
				"Data": json.RawMessage(`{"theme": "dark"}`),
			},
			expectedStatus: http.StatusBadRequest,

			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.PreferencesUpdate},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
			},
		},
		{
			name:   "PATCH-/users/preferences > invalid preference sub type",
			method: "PATCH",
			path:   "/preferences",
			body: map[string]interface{}{
				"Type": "invalid",
				"Data": json.RawMessage(`{"theme": "dark"}`),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.PreferencesUpdate},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
			},
		},
		{

			name:   "PATCH-/users/preferences > invalid preference type",
			method: "PATCH",
			path:   "/preferences",
			body: map[string]interface{}{
				"Type": "General",
				"Data": json.RawMessage(`{"theme": "dark"}`),
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.PreferencesUpdate},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
			},
		},
		{
			name:   "PATCH-/users/preferences > success",
			method: "PATCH",
			path:   "/preferences",
			body: map[string]interface{}{
				"Type": "General",
				"Data": json.RawMessage(`{"openSpace": "sameWindow"}`),
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"message": "user preferences updated",
			},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
				mockDB.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput"), mock.Anything).Return(&dynamodb.UpdateItemOutput{}, nil)
			},
		},
	}

}

func getUserSubscriptionsTestCases() []TestCase {
	return []TestCase{
		{
			name:           "GET-/users/subscription > ddb error",
			method:         "GET",
			path:           "/subscription",
			expectedStatus: http.StatusBadGateway,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.SubscriptionGet},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
				mockDB.On("GetItem", mock.Anything, &dynamodb.GetItemInput{
					TableName: aws.String("MainTable_test"),
					Key: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{Value: testUser.Id},
						"SK": &types.AttributeValueMemberS{Value: db.SORT_KEY.Subscription},
					},
				}, mock.Anything).Return(&dynamodb.GetItemOutput{}, errors.New("error getting user subscription")).Once()
			},
		},
		{
			name:           "GET-/users/subscription > success",
			method:         "GET",
			path:           "/subscription",
			expectedStatus: http.StatusOK,
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
				mockDBGetUserSubscription(mockDB)
			},
		},
	}
}

func checkSubscriptionStatusTestCases() []TestCase {
	return []TestCase{
		{
			name:           "GET-/users/subscription/status > success",
			method:         "GET",
			path:           "/subscription/status",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"active": true,
				},
			},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
				mockDBGetUserSubscription(mockDB)
			},
		},
	}
}

func getPaddleURLTestCases() []TestCase {

	paddleURL := map[string]string{
		"cancelURL": "https://paddle.com/subscription/cancel/123456789",
		"updateURL": "https://paddle.com/subscription/update/123456789",
	}

	return []TestCase{
		{
			name:           "GET-/users/subscription/paddle-url > paddle client error",
			method:         "GET",
			path:           "/subscription/paddle-url",
			expectedStatus: http.StatusBadGateway,
			expectedBody: map[string]interface{}{"success": false,
				"message": users.ErrMsg.SubscriptionPaddleURL},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
				mockDBGetUserSubscription(mockDB)
			},
			setupMockPaddleClient: func(mockPaddleClient *PaddleClientMock) {
				mockPaddleClient.On("GetSubscription", mock.Anything, mock.AnythingOfType("*paddle.GetSubscriptionRequest")).Return(nil, errors.New("paddle error")).Once()
			},
		},
		{
			name:           "GET-/users/subscription/paddle-url > success",
			method:         "GET",
			path:           "/subscription/paddle-url",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"updateURL": paddleURL["updateURL"],
				},
			},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
				mockDBGetUserSubscription(mockDB)
			},
			setupMockPaddleClient: func(mockPaddleClient *PaddleClientMock) {
				mockPaddleClient.On("GetSubscription", mock.Anything, mock.AnythingOfType("*paddle.GetSubscriptionRequest")).Return(&paddle.Subscription{

					ManagementURLs: paddle.SubscriptionManagementUrLs{
						Cancel:              paddleURL["cancelURL"],
						UpdatePaymentMethod: aws.String(paddleURL["updateURL"]),
					},
				}, nil).Once()
			},
		},
		{
			name:           "GET-/users/subscription/paddle-url?cancelURL=true > success",
			method:         "GET",
			path:           "/subscription/paddle-url?cancelURL=true",
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"data": map[string]interface{}{
					"cancelURL": paddleURL["cancelURL"],
				},
			},
			mockAuthHeader: func(r *http.Request) { r.Header.Set("UserId", testUser.Id) },
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDBGetUser(mockDB)
				mockDBGetUserSubscription(mockDB)
			},
			setupMockPaddleClient: func(mockPaddleClient *PaddleClientMock) {
				mockPaddleClient.On("GetSubscription", mock.Anything, mock.AnythingOfType("*paddle.GetSubscriptionRequest")).Return(&paddle.Subscription{

					ManagementURLs: paddle.SubscriptionManagementUrLs{
						Cancel:              paddleURL["cancelURL"],
						UpdatePaymentMethod: aws.String(paddleURL["updateURL"]),
					},
				}, nil).Once()
			},
		},
	}
}

func subscriptionWebhookTestCases() []TestCase {

	generateWebhookBody := func(createdEvent bool) map[string]interface{} {

		eventBody := map[string]interface{}{
			"id":              "ntfsimevt_01jaqmyj1kg8gt3e02j25nfw25",
			"event_type":      "subscription.updated",
			"occurred_at":     "2024-10-21T13:41:01.619284Z",
			"notification_id": "ntfsimntf_01jaqmyj4n7d9e9hqkrkzmf9qt",
			"data": map[string]interface{}{
				"id": "sub_01hv8x29kz0t586xy6zn1a62ny",
				"items": []map[string]interface{}{
					{
						"price": map[string]interface{}{
							"id": "pri_01gsz8x8sawmvhz1pv30nge1ke",
						},
					},
				},
				"status":     "active",
				"started_at": "2024-04-12T10:37:59.556997Z",
				"custom_data": map[string]interface{}{
					"userId": "123",
				},
				"customer_id":     "ctm_01hv6y1jedq4p1n0yqn5ba3ky4",
				"currency_code":   "USD",
				"next_billed_at":  "2024-05-12T10:37:59.556997Z",
				"first_billed_at": "2024-04-12T10:37:59.556997Z",
				"current_billing_period": map[string]interface{}{
					"ends_at":   "2024-05-12T10:37:59.556997Z",
					"starts_at": "2024-04-12T10:37:59.556997Z",
				},
			},
		}

		if createdEvent {
			eventBody["event_type"] = "subscription.created"
		}

		return eventBody

	}

	return []TestCase{
		{
			name:           "POST-/users/subscription/webhook > subscription update event success",
			method:         "POST",
			path:           "/subscription/webhook",
			body:           generateWebhookBody(false),
			expectedStatus: http.StatusOK,
			mockAuthHeader: func(r *http.Request) {
				r.Header.Set("Paddle-Webhook-Test", "true")
			},
			expectedBody: map[string]interface{}{
				"success": true,
				"message": "event acknowledged",
			},
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDB.On("UpdateItem", mock.Anything, mock.AnythingOfType("*dynamodb.UpdateItemInput"), mock.Anything).Return(&dynamodb.UpdateItemOutput{}, nil).Once()
			},
		},
		{
			name:           "POST-/users/subscription/webhook > subscription create event success",
			method:         "POST",
			path:           "/subscription/webhook",
			body:           generateWebhookBody(true),
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"message": "event acknowledged",
			},
			mockAuthHeader: func(r *http.Request) {
				r.Header.Set("Paddle-Webhook-Test", "true")
			},
			setupMockDB: func(mockDB *DynamoDBClientMock) {
				mockDB.On("PutItem", mock.Anything, mock.AnythingOfType("*dynamodb.PutItemInput"), mock.Anything).Return(&dynamodb.PutItemOutput{}, nil).Once()
			},
		},
	}
}

// merge all test cases
func allTestCases() []TestCase {
	tests := slices.Concat(getUserByIDTestCases(), createUserTestCases(), updateUserTestCases(), deleteUserTestCases(),
		getUserPreferencesTestCases(), updateUserPreferencesTestCases(), getUserSubscriptionsTestCases(),
		checkSubscriptionStatusTestCases(), getPaddleURLTestCases(), subscriptionWebhookTestCases())

	return tests

}

// * run test cases
func TestUsersService(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	tests := allTestCases()

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create new test setup for each test case
			setup := newTestSetup()

			var mockedQueue *SQSClientMock

			mockedDB, ok := setup.mockDB.Client.(*DynamoDBClientMock)

			if !ok {
				t.Fatal("failed to get mock db client")
			}

			// setup mock interceptor
			if tc.setupMockClient != nil {
				tc.setupMockClient(setup.mockClient)
			}

			// setup mock paddle client
			if tc.setupMockPaddleClient != nil {
				tc.setupMockPaddleClient(setup.mockPaddleClient)
			}

			// setup mock db
			if tc.setupMockDB != nil {
				tc.setupMockDB(mockedDB)
			}

			if tc.setupMockDBWithAssertions != nil {
				tc.setupMockDBWithAssertions(t, mockedDB)
			}

			// setup mock queue
			if tc.setupMockQueue != nil {
				mockedQueue = setup.mockQueue.Client.(*SQSClientMock)
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
			if tc.mockAuthHeader != nil {
				tc.mockAuthHeader(req)
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Origin", config.AllowedOrigins[0])
			req.Header.Set("Cookie", "test=1")
			// recorder
			w := httptest.NewRecorder()

			// serve request through the router
			setup.router.ServeHTTP(w, req)

			// assert status code
			assert.Equal(t, tc.expectedStatus, w.Code)

			// assert body
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
