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
	Email      string `json:"email"`
	OTP        string `json:"otp"`
	TTL_Expiry int32  `json:"ttlExpiry"`
}

type session struct {
	Id         string `json:"id"`
	Email      string `json:"email"`
	TTL_Expiry int32  `json:"ttlExpiry"`
	DeviceInfo *DeviceInfo
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
	verifyToken     string
	createSession   string
	validateSession string
}{
	sendOTP:         "Error sending OTP",
	validateOTP:     "Error validating OTP",
	inValidOTP:      "Invalid OTP",
	createSession:   "Error creating session",
	validateSession: "Error validating session",
}

type authRepository interface {
	saveOTP(data *emailOTP) error
	validateOTP(email, otp string) (bool, error)
	validateSession(email, id string) (bool, error)
	createSession(s *session) error
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

	if response.Item == nil || response.Item["TTL_Expiry"] == nil {
		return false, errors.New(errMsg.validateSession)
	}

	// check if OTP has expired
	var ttl struct {
		TTL_Expiry int32
	}

	err = attributevalue.UnmarshalMap(response.Item, &ttl)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't unmarshal session expiry from db for email: %#v", email), err)
		return false, errors.New(errMsg.validateSession)
	}

	if ttl.TTL_Expiry < int32(time.Now().Unix()) {
		return false, errors.New(errMsg.validateSession)
	}

	return true, nil
}
