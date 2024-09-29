package auth

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type authRepository interface {
	saveOTP(data *emailOTP) error
	attachUserId(data *emailWithUserId) error
	userIdByEmail(email string) (string, error)
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

	ttl := strconv.FormatInt(data.TTL, 10)

	saveItem := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: data.Email},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY_SESSIONS.OTP(data.OTP)},
		"TTL":            &types.AttributeValueMemberS{Value: ttl},
	}

	_, err := r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      saveItem,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't save OTP to db for email: %v", data.Email), err)
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

	if response.Item == nil || response.Item["TTL"] == nil {
		return false, nil
	}

	// check if OTP has expired
	var ttl struct {
		TTL string
	}

	err = attributevalue.UnmarshalMap(response.Item, &ttl)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't unmarshal OTP ttl from db for email: %#v", email), err)
		return false, errors.New(errMsg.inValidOTP)
	}

	ttlInt, err := strconv.ParseInt(ttl.TTL, 10, 64)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't convert TTL to int for email: %#v", email), err)
		return false, errors.New(errMsg.inValidOTP)
	}

	if ttlInt < int64(time.Now().Unix()) {
		return false, nil
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

func (r *authRepo) userIdByEmail(email string) (string, error) {
	// primary key - partition+sort key

	keyCondition := expression.KeyAnd(expression.Key("PK").Equal(expression.Value(email)), expression.Key("SK").BeginsWith(database.SORT_KEY_SESSIONS.UserId("")))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()

	logger.Dev(fmt.Sprintf("Querying expr: %#v", expr.Values()))

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't build getUserID expression for email: %#v", email), err)
		return "", errors.New(errMsg.createSession)
	}

	response, err := r.db.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 &r.db.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't get user id from db for email: %#v", email), err)
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
		logger.Error(fmt.Sprintf("Couldn't unmarshal user id from db for email: %#v", email), err)
		return "", errors.New(errMsg.getUserId)
	}

	// get user id from sort key
	userId := strings.Split(s.UserId, "#")[1]

	if userId == "" {
		logger.Error(fmt.Sprintf("Couldn't get user id from sort_key: %#v", s.UserId), err)
		return "", nil
	}

	return userId, nil

}

func (r *authRepo) createSession(s *session) error {

	item := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: s.Email},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY_SESSIONS.Session(s.Id)},
		"TTL":            &types.AttributeValueMemberS{Value: strconv.FormatInt(s.TTL, 10)},
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

	if userSession.TTL < time.Now().Unix() {
		return false, errors.New(errMsg.validateSession)
	}

	return true, nil
}
