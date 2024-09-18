package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type emailOTP struct {
	Email      string `json:"email" dynamodbav:"PK"`
	OTP        string `json:"otp"`
	TTL_Expiry int64  `json:"ttlExpiry"`
}

type emailWithUserId struct {
	Email  string `json:"email" dynamodbav:"PK"`
	UserId string `json:"userId" dynamodbav:"SK"`
}

type session struct {
	Email      string      `json:"email" dynamodbav:"PK"`
	Id         string      `json:"id" dynamodbav:"SK"`
	TTL_Expiry int64       `json:"ttlExpiry" dynamodbav:"TTL_Expiry"`
	DeviceInfo *DeviceInfo `json:"deviceInfo" dynamodbav:"DeviceInfo"`
}

type DeviceInfo struct {
	Browser  string `json:"browser" dynamodbav:"browser"`
	OS       string `json:"os" dynamodbav:"os"`
	Platform string `json:"platform" dynamodbav:"platform"`
	IsMobile bool   `json:"isMobile" dynamodbav:"isMobile"`
}

var errMsg = struct {
	sendOTP         string
	validateOTP     string
	inValidOTP      string
	createToken     string
	createSession   string
	deleteSession   string
	validateSession string
	googleAuth      string
	tokenExpired    string
	invalidToken    string
	invalidSession  string
	logout          string
}{
	sendOTP:         "Error sending OTP",
	validateOTP:     "Error validating OTP",
	inValidOTP:      "Invalid OTP",
	googleAuth:      "Error authenticating with google",
	createSession:   "Error creating session",
	deleteSession:   "Error deleting session",
	createToken:     "Error creating token",
	validateSession: "Error validating session",
	tokenExpired:    "Token expired",
	invalidToken:    "Invalid token",
	invalidSession:  "Invalid session",
	logout:          "Error logging out",
}

type authRepository interface {
	saveOTP(data *emailOTP) error
	attachUserId(data *emailWithUserId) error
	validateOTP(email, otp string) (bool, error)
	validateSession(email, id string) (bool, error)
	createSession(s *session) error
	deleteSession(email, sessionId string) error
}

type authRepo struct {
	db *database.DDB
}

func newAuthRepository(db *database.DDB) authRepository {
	return &authRepo{
		db: db,
	}
}

// save OTP to DB
func (r *authRepo) saveOTP(data *emailOTP) error {

	saveItem := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: data.Email},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY_SESSIONS.OTP(data.OTP)},
		"TTL_Expiry":     &types.AttributeValueMemberN{Value: string(data.TTL_Expiry)},
	}

	_, err := r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      saveItem,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't save OTP to db for email: %#v", data.Email), err)
		return errors.New(errMsg.sendOTP)
	}

	return nil
}

func (r *authRepo) validateOTP(email, otp string) (bool, error) {

	// primary key - partition+sort key
	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: email},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY_SESSIONS.OTP(otp)},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't get OTP from db for email: %#v", email), err)
		return false, errors.New(errMsg.validateOTP)
	}

	if response.Item == nil || response.Item["TTL_Expiry"] == nil {
		return false, errors.New(errMsg.inValidOTP)
	}

	// check if OTP has expired
	var ttl struct {
		TTL_Expiry int32
	}

	err = attributevalue.UnmarshalMap(response.Item, &ttl)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't unmarshal OTP ttl from db for email: %#v", email), err)
		return false, errors.New(errMsg.inValidOTP)
	}

	if ttl.TTL_Expiry < int32(time.Now().Unix()) {
		return false, errors.New(errMsg.inValidOTP)
	}

	return true, nil
}

func (r *authRepo) attachUserId(data *emailWithUserId) error {
	// primary key - partition+sort key
	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: data.Email},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY_SESSIONS.UserId(data.UserId)},
	}

	_, err := r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      key,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't attach user id to email: %#v", data.Email), err)
		return errors.New(errMsg.createSession)
	}

	return nil
}

func (r *authRepo) createSession(s *session) error {

	item := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: s.Email},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY_SESSIONS.Session(s.Id)},
		"DeviceInfo": &types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"browser":  &types.AttributeValueMemberS{Value: s.DeviceInfo.Browser},
				"os":       &types.AttributeValueMemberS{Value: s.DeviceInfo.OS},
				"platform": &types.AttributeValueMemberS{Value: s.DeviceInfo.Platform},
				"isMobile": &types.AttributeValueMemberBOOL{Value: s.DeviceInfo.IsMobile},
			},
		},
	}

	_, err := r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      item,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't create session for email: %#v", s.Email), err)
		return errors.New(errMsg.createSession)
	}

	return nil
}

func (r *authRepo) deleteSession(email, id string) error {
	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: email},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY_SESSIONS.Session(id)},
	}

	_, err := r.db.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't delete session for user with email: %#v", email), err)
		return errors.New(errMsg.deleteSession)
	}

	return nil
}

func (r *authRepo) validateSession(email, id string) (bool, error) {
	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: email},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY_SESSIONS.Session(id)},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't get session from db, for email: %#v", email), err)
		return false, errors.New(errMsg.validateSession)
	}

	var userSession session

	err = attributevalue.UnmarshalMap(response.Item, &userSession)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't unmarshal session expiry from db for email: %#v", email), err)
		return false, errors.New(errMsg.validateSession)
	}

	if userSession.Id == "" {
		return false, errors.New(errMsg.validateSession)
	}

	if userSession.TTL_Expiry < time.Now().Unix() {
		return false, errors.New(errMsg.validateSession)
	}

	return true, nil
}
