package users

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type userRepository interface {
	getUserByID(id string) (*User, error)
	insertUser(user *User) error
	updateUser(id, name string) error
	deleteAccount(id string) error
	getAllPreferences(id string) (*preferences, error)
	updatePreferences(id string, pData map[string]interface{}) error
}

type userRepo struct {
	db *database.DDB
}

func newUserRepository(db *database.DDB) userRepository {
	return &userRepo{
		db: db,
	}
}

// profile
func (r *userRepo) getUserByID(id string) (*User, error) {

	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: id},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.Profile},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't query for userId: %v", id), err)
		return nil, err
	}

	user := &User{}

	err = attributevalue.UnmarshalMap(response.Item, &user)

	if response.Item == nil || response.Item["PK"] == nil {
		return nil, fmt.Errorf(errMsg.userNotFound)
	}

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't unmarshal query result for user_id: %v", id), err)
		return nil, err
	}

	if user.Id == "" {
		return nil, fmt.Errorf(errMsg.userNotFound)
	}

	return user, nil
}

func (r *userRepo) insertUser(user *User) error {
	profile := userWithSK{
		User: user,
		SK:   database.SORT_KEY.Profile,
	}

	item, err := attributevalue.MarshalMap(profile)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't marshal user: %#v", user), err)
		return err
	}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      item,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't put item for user: %v", user.Id), err)
		return err
	}

	return nil
}

func (r *userRepo) updateUser(id, name string) error {

	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: id},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.Profile},
	}

	// build update expression
	updateExpr := expression.Set(expression.Name("FullName"), expression.Value(name))
	expr, err := expression.NewBuilder().WithUpdate(updateExpr).Build()

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't build expression for updateUser query for the user_id: %v", id), err)
		return err
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
		logger.Error(fmt.Sprintf("Couldn't updateUser, user_id: %v", id), err)
		return err
	}

	return nil
}

// delete user account with all their data
func (r *userRepo) deleteAccount(id string) error {

	// DynamoDB allows a maximum batch size of 25 items.
	batchSize := 25

	allSKs, err := database.Helpers.GetAllSKs(r.db, id)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't get all dynamic sort keys for user_id: %v", id), err)
		return errors.New(errMsg.deleteUser)
	}

	//  prepare delete requests for all SKs
	var deleteRequests []types.WriteRequest
	for _, sk := range allSKs {
		req := types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					database.PK_NAME: &types.AttributeValueMemberS{Value: id},
					database.SK_NAME: &types.AttributeValueMemberS{Value: sk},
				},
			},
		}
		deleteRequests = append(deleteRequests, req)
	}

	//  perform batch request in batches of allowed limit
	for i := 0; i < len(deleteRequests); i += batchSize {
		end := i + batchSize
		if end > len(deleteRequests) {
			end = len(deleteRequests)
		}

		_, err := r.db.Client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
			RequestItems: map[string][]types.WriteRequest{
				r.db.TableName: deleteRequests[i:end],
			},
		})

		if err != nil {
			logger.Error(fmt.Sprintf("Couldn't batch delete items for user_id: %v", id), err)
			return errors.New(errMsg.deleteUser)
		}
	}

	return nil
}

// preferences
func (r *userRepo) getAllPreferences(id string) (*preferences, error) {
	// primary key - partition+sort key
	keyCondition := expression.KeyAnd(expression.Key("PK").Equal(expression.Value(id)), expression.Key("SK").BeginsWith(database.SORT_KEY.PreferencesBase))

	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't build getPreferences expression for userId: %#v", id), err)
		return nil, err
	}

	response, err := r.db.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 &r.db.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't get preferences for userId : %#v", id), err)
		return nil, err
	}

	if len(response.Items) < 1 {
		return nil, nil
	}

	var p preferences

	err = attributevalue.UnmarshalMap(response.Items[0], &p)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't unmarshal preferences for userId: %#v", id), err)
		return nil, err
	}

	logger.Dev("preferences: %v", p)

	return &p, nil
}

func (r *userRepo) updatePreferences(id string, pData map[string]interface{}) error {

	d := map[string]interface{}{
		"PK": id,
	}

	for k, v := range pData {
		d[k] = v
	}

	logger.Dev("data: %v", d)

	item, err := attributevalue.MarshalMap(d)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't marshal preferences for userId: %#v", id), err)
		return err
	}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      item,
	})
	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't put preferences item for userId: %#v", id), err)
		return err
	}

	return nil
}

// TODO - subscription
