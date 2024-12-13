package auth

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type authRepository interface {
	saveOTP(data *emailOTP) error
	attachUserId(data *emailWithUserId) error
	userIdByEmail(email string) (string, error)
	validateOTP(email, otp string) (bool, error)
	ValidateSession(email, id string) (bool, error)
	createSession(s *session) error
	deleteSession(email, sessionId string) error
}

type authRepo struct {
	db *db.DDB
}

func newAuthRepository(db *db.DDB) authRepository {
	return &authRepo{
		db: db,
	}
}

// save OTP to DB
func (r *authRepo) saveOTP(data *emailOTP) error {

	ttl := strconv.FormatInt(data.TTL, 10)

	saveItem := map[string]types.AttributeValue{
		db.PK_NAME:      &types.AttributeValueMemberS{Value: data.Email},
		db.SK_NAME:      &types.AttributeValueMemberS{Value: db.SORT_KEY_SESSIONS.OTP(data.OTP)},
		db.TTL_KEY_NAME: &types.AttributeValueMemberN{Value: ttl},
	}

	_, err := r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      saveItem,
	})

	if err != nil {
		logger.Errorf("Couldn't save OTP to db for email: %v, \n[Error:] %v", data.Email, err)
		return errors.New(errMsg.sendOTP)
	}

	return nil
}

func (r *authRepo) validateOTP(email, otp string) (bool, error) {

	// primary key - partition+sort key
	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: email},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY_SESSIONS.OTP(otp)},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't get OTP from db for email: %#v: \n[Error]: %v", email, err)
		return false, errors.New(errMsg.validateOTP)
	}

	if response.Item == nil || response.Item[db.TTL_KEY_NAME] == nil {
		return false, nil
	}

	// check if OTP has expired
	var ttlAtr struct {
		TTL int64
	}

	err = attributevalue.UnmarshalMap(response.Item, &ttlAtr)

	if err != nil {
		logger.Errorf("Couldn't unmarshal OTP ttl from db for email: %#v: \n[Error]: %v", email, err)
		return false, errors.New(errMsg.inValidOTP)
	}

	if ttlAtr.TTL < time.Now().Unix() {
		return false, errors.New(errMsg.expiredOTP)
	}

	return true, nil
}

func (r *authRepo) attachUserId(data *emailWithUserId) error {
	// primary key - partition+sort key
	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: data.Email},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY_SESSIONS.UserId(data.UserId)},
	}

	_, err := r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      key,
	})
	if err != nil {
		logger.Errorf("Couldn't attach user id to email: %#v, \n[Error]: %v", data.Email, err)
		return errors.New(errMsg.createSession)
	}

	return nil
}

func (r *authRepo) userIdByEmail(email string) (string, error) {
	// primary key - partition+sort key
	keyCondition := expression.KeyAnd(expression.Key("PK").Equal(expression.Value(email)), expression.Key("SK").BeginsWith(db.SORT_KEY_SESSIONS.UserId("")))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()

	if err != nil {
		logger.Errorf("Couldn't build getUserID expression for email: %#v: \n[Error]: %v", email, err)
		return "", errors.New(errMsg.createSession)
	}

	response, err := r.db.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 &r.db.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	if err != nil {
		logger.Errorf("Couldn't get user id from db for email: %#v: \n[Error]: %v", email, err)
		return "", errors.New(errMsg.getUserId)
	}

	if len(response.Items) < 1 {
		return "", nil
	}

	var s struct {
		UserId string `json:"userId" dynamodbav:"SK"`
	}

	err = attributevalue.UnmarshalMap(response.Items[0], &s)

	if err != nil || s.UserId == "" {
		logger.Errorf("Couldn't unmarshal user id from db for email: %#v: \n[Error]: %v", email, err)
		return "", errors.New(errMsg.getUserId)
	}

	// get user id from sort key
	userId := strings.Split(s.UserId, "#")[1]

	if userId == "" {
		// logger.Errorf("Couldn't get user id from sort_key: %#v", s: \n[Error]: %v.UserId, err)
		logger.Errorf("Couldn't get user id from sort_key: %#v: \n[Error]: %v", s.UserId, err)

		return "", nil
	}

	return userId, nil

}

func (r *authRepo) createSession(s *session) error {

	item := map[string]types.AttributeValue{
		db.PK_NAME:      &types.AttributeValueMemberS{Value: s.UserId},
		db.SK_NAME:      &types.AttributeValueMemberS{Value: db.SORT_KEY_SESSIONS.Session(s.Id)},
		db.TTL_KEY_NAME: &types.AttributeValueMemberN{Value: strconv.FormatInt(s.TTL, 10)},
		"CreatedAt":     &types.AttributeValueMemberN{Value: strconv.FormatInt(time.Now().Unix(), 10)},
		"DeviceInfo": &types.AttributeValueMemberM{
			Value: map[string]types.AttributeValue{
				"os":       &types.AttributeValueMemberS{Value: s.DeviceInfo.OS},
				"browser":  &types.AttributeValueMemberS{Value: s.DeviceInfo.Browser},
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
		logger.Errorf("Couldn't create session for userId: %#v. \n[Error]: %v", s.UserId, err)
		return errors.New(errMsg.createSession)
	}

	return nil
}

func (r *authRepo) deleteSession(userId, sId string) error {
	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY_SESSIONS.Session(sId)},
	}

	_, err := r.db.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't delete session for user with userId: %#v: \n[Error]: %v", userId, err)

		return errors.New(errMsg.deleteSession)
	}

	return nil
}

func (r *authRepo) ValidateSession(userId, sId string) (bool, error) {
	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY_SESSIONS.Session(sId)},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't get session from db, for userId: %#v: \n[Error]: %v", userId, err)
		return false, errors.New(errMsg.ValidateSession)
	}

	var userSession session

	err = attributevalue.UnmarshalMap(response.Item, &userSession)

	if err != nil {
		logger.Errorf("Couldn't unmarshal session expiry from db for userId: %#v: \n[Error]: %v", userId, err)
		return false, errors.New(errMsg.ValidateSession)
	}

	if userSession.Id == "" {
		return false, errors.New(errMsg.ValidateSession)
	}

	if userSession.TTL < time.Now().Unix() {
		return false, errors.New(errMsg.ValidateSession)
	}

	return true, nil
}
