package users

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type repository interface {
	getUserByID(id string) (*User, *http_api.Metadata, error)
	insertUser(user *User) (*http_api.Metadata, error)
	updateUser(id, name string) (*http_api.Metadata, error)
	deleteAccount(id string) error
	getAllPreferences(id string) (*Preferences, *http_api.Metadata, error)
	setPreferences(userId, sk string, pData interface{}) error
	updatePreferences(userId, sk string, pData interface{}) (*http_api.Metadata, error)
	getSubscription(userId string) (*subscription, error)
	setSubscription(userId string, s *subscription) error
	updateSubscription(userId string, sData *subscription) error
}

type userRepo struct {
	db *db.DDB
}

func newRepository(db *db.DDB) repository {
	return &userRepo{
		db: db,
	}
}

// profile
func (r *userRepo) getUserByID(id string) (*User, *http_api.Metadata, error) {

	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: id},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.Profile},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't query for userId: %v. \n[Error]: %v", id, err)
		return nil, nil, err
	}

	if response == nil || response.Item == nil {
		return nil, nil, errors.New(ErrMsg.UserNotFound)
	}

	if _, ok := response.Item["PK"]; !ok {
		return nil, nil, errors.New(ErrMsg.UserNotFound)
	}

	user := &User{}

	updatedAt, err := db.GetUpdatedAtTime(response.Item)

	if err != nil {
		logger.Error("Error getting updatedAt time", err)
		return nil, nil, err
	}

	m := &http_api.Metadata{
		UpdatedAt: updatedAt,
	}

	err = attributevalue.UnmarshalMap(response.Item, &user)

	if err != nil {
		logger.Errorf("Couldn't unmarshal query result for user_id: %v. \n[Error]: %v", id, err)
		return nil, nil, err
	}

	if user.Id == "" {
		return nil, nil, errors.New(ErrMsg.UserNotFound)
	}

	return user, m, nil
}

func (r *userRepo) insertUser(user *User) (*http_api.Metadata, error) {
	profile := userWithSK{
		User: user,
		SK:   db.SORT_KEY.Profile,
	}

	item, err := attributevalue.MarshalMap(profile)

	m := &http_api.Metadata{
		UpdatedAt: time.Now().UnixMilli(),
	}

	item["UpdatedAt"] = &types.AttributeValueMemberN{Value: strconv.FormatInt(m.UpdatedAt, 10)}

	if err != nil {
		logger.Errorf("Couldn't marshal user: %#v. \n[Error]: %v", user, err)
		return nil, err
	}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      item,
	})

	if err != nil {
		logger.Errorf("Couldn't put item for userId: %v. \n[Error]: %v", user.Id, err)
		return nil, err
	}

	return m, nil
}

func (r userRepo) updateUser(id, name string) (*http_api.Metadata, error) {

	m := &http_api.Metadata{
		UpdatedAt: time.Now().UnixMilli(),
	}

	key := map[string]types.AttributeValue{
		db.PK_NAME:  &types.AttributeValueMemberS{Value: id},
		db.SK_NAME:  &types.AttributeValueMemberS{Value: db.SORT_KEY.Profile},
		"UpdatedAt": &types.AttributeValueMemberN{Value: strconv.FormatInt(m.UpdatedAt, 10)},
	}

	// build update expression
	updateExpr := expression.Set(expression.Name("FullName"), expression.Value(name))
	expr, err := expression.NewBuilder().WithUpdate(updateExpr).Build()

	if err != nil {
		logger.Errorf("Couldn't build expression for updateUser query for the user_id: %v. \n[Error]: %v", id, err)
		return nil, err
	}

	// execute the query
	_, err = r.db.Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName:                 &r.db.TableName,
		Key:                       key,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	})

	if err != nil {
		logger.Errorf("Couldn't updateUser, user_id: %v. \n[Error]: %v", id, err)
		return nil, err
	}

	return m, nil
}

// delete user account with all their data
func (r userRepo) deleteAccount(id string) error {

	allSKs, err := r.db.GetAllSKs(id)

	if err != nil {
		logger.Errorf("Couldn't get all SKs for userId: %v. \n[Error]: %v", id, err)
		return err
	}

	// channel to collect errors from goroutines
	errChan := make(chan error, len(allSKs)/db.DDB_MAX_BATCH_SIZE+1)

	var wg sync.WaitGroup

	// context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqs := []types.WriteRequest{}

	for _, sk := range allSKs {
		reqs = append(reqs, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					db.PK_NAME: &types.AttributeValueMemberS{Value: id},
					db.SK_NAME: &types.AttributeValueMemberS{Value: sk},
				},
			},
		})
	}

	r.db.BatchWriter(ctx, r.db.TableName, &wg, errChan, reqs)

	// Wait for all goroutines to complete
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	// Return combined errors if any
	if len(errs) > 0 {
		return fmt.Errorf("delete search index errors: %v", errs)
	}

	return nil
}

// preferences
func (r userRepo) getAllPreferences(id string) (*Preferences, *http_api.Metadata, error) {
	// primary key - partition+sort key
	keyCondition := expression.KeyAnd(expression.Key("PK").Equal(expression.Value(id)), expression.Key("SK").BeginsWith(db.SORT_KEY.PreferencesBase))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()

	if err != nil {
		logger.Errorf("Couldn't build getPreferences expression for userId: %#v. \n[Error]: %v", id, err)
		return nil, nil, err
	}

	response, err := r.db.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 &r.db.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	if err != nil {
		logger.Errorf("Couldn't get preferences for userId : %#v. \n[Error]: %v", id, err)
		return nil, nil, err
	}

	if len(response.Items) < 1 {
		return nil, nil, errors.New(ErrMsg.PreferencesGet)
	}
	p, m, err := unMarshalPreferences(response)

	if err != nil {
		return nil, nil, err
	}

	return p, m, nil
}

func (r userRepo) setPreferences(userId, sk string, pData interface{}) error {
	av, err := attributevalue.MarshalMap(pData)

	if err != nil {
		return err
	}

	av["PK"] = &types.AttributeValueMemberS{Value: userId}
	av["SK"] = &types.AttributeValueMemberS{Value: sk}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      av,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r userRepo) updatePreferences(userId, sk string, pData interface{}) (*http_api.Metadata, error) {

	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: userId},
		"SK": &types.AttributeValueMemberS{Value: sk},
	}

	m := &http_api.Metadata{
		UpdatedAt: time.Now().UnixMilli(),
	}

	var update expression.UpdateBuilder

	// iterate over the fields of the struct
	v := reflect.ValueOf(pData)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	} else {
		logger.Error("unexpected type", errors.New(v.Kind().String()))
		return nil, errors.ErrUnsupported
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if fieldValue.IsZero() {
			continue
		}

		update = update.Set(expression.Name(field.Name), expression.Value(v.Field(i).Interface()))
	}

	update.Set(expression.Name("UpdatedAt"), expression.Value(m.UpdatedAt))

	expr, err := expression.NewBuilder().WithUpdate(update).Build()

	if err != nil {
		return nil, err
	}

	_, err = r.db.Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName:                 &r.db.TableName,
		Key:                       key,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	})

	if err != nil {
		logger.Errorf("Couldn't update preferences for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return m, nil
}

// subscription
func (r userRepo) getSubscription(userId string) (*subscription, error) {

	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: userId},
		"SK": &types.AttributeValueMemberS{Value: db.SORT_KEY.Subscription},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't get subscription for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}
	if _, ok := response.Item["PK"]; !ok {
		return nil, errors.New(ErrMsg.SubscriptionGet)
	}
	s := &subscription{}

	err = attributevalue.UnmarshalMap(response.Item, s)

	if err != nil {
		logger.Errorf("Couldn't unmarshal subscription for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return s, nil

}

func (r userRepo) setSubscription(userId string, s *subscription) error {

	av, err := attributevalue.MarshalMap(s)
	if err != nil {
		return err
	}
	av["PK"] = &types.AttributeValueMemberS{Value: userId}
	av["SK"] = &types.AttributeValueMemberS{Value: db.SORT_KEY.Subscription}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      av,
	})

	if err != nil {
		logger.Errorf("Couldn't set subscription for userId: %v. \n[Error]: %v", userId, err)
		return err
	}

	return nil
}

func (r userRepo) updateSubscription(userId string, sData *subscription) error {
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: userId},
		"SK": &types.AttributeValueMemberS{Value: db.SORT_KEY.Subscription},
	}

	var update expression.UpdateBuilder

	// iterate over the fields of the struct
	v := reflect.ValueOf(sData)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	} else {
		logger.Error("unexpected type", errors.New(v.Kind().String()))
		return errors.ErrUnsupported
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if fieldValue.IsZero() {
			continue
		}

		update = update.Set(expression.Name(field.Name), expression.Value(v.Field(i).Interface()))
	}

	expr, err := expression.NewBuilder().WithUpdate(update).Build()

	if err != nil {
		return err
	}

	_, err = r.db.Client.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName:                 &r.db.TableName,
		Key:                       key,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
	})

	if err != nil {
		logger.Errorf("Couldn't update subscription for userId: %v. \n[Error]: %v", userId, err)
		return err
	}

	return nil
}

// * helper
func unMarshalPreferences(res *dynamodb.QueryOutput) (*Preferences, *http_api.Metadata, error) {

	m := http_api.Metadata{
		UpdatedAtMap: map[string]int64{},
	}

	unmarshal := func(item map[string]types.AttributeValue, v interface{}) error {

		if err := attributevalue.UnmarshalMap(item, &v); err != nil {
			return err
		}

		return nil
	}

	getUpdatedAt := func(item map[string]types.AttributeValue, key string) error {
		u, err := db.GetUpdatedAtTime(item)

		if err != nil {
			logger.Error("Error getting updatedAt time", err)
			return err
		}
		m.UpdatedAtMap[key] = u
		return nil
	}

	var err, err1 error

	p := Preferences{}

	for _, item := range res.Items {
		if item["PK"] == nil || item["SK"] == nil {
			err = errors.New("invalid item")
			continue
		}
		sk := item["SK"].(*types.AttributeValueMemberS).Value

		switch sk {
		case "P#General":
			err = unmarshal(item, &p.General)
			err1 = getUpdatedAt(item, "general")

		case "P#CmdPalette":
			err = unmarshal(item, &p.CmdPalette)
			err1 = getUpdatedAt(item, "cmdPalette")

		case "P#Notes":
			err = unmarshal(item, &p.Notes)
			err1 = getUpdatedAt(item, "notes")

		case "P#AutoDiscard":
			err = unmarshal(item, &p.AutoDiscard)
			err1 = getUpdatedAt(item, "autoDiscard")

		case "P#LinkPreview":
			err = unmarshal(item, &p.LinkPreview)
			err1 = getUpdatedAt(item, "linkPreview")

		default:
			err = errors.New("invalid preferences SK")
		}
	}

	if err != nil || err1 != nil {
		logger.Error("Couldn't unmarshal preferences", err)
		return nil, nil, err
	}

	return &p, &m, nil
}
