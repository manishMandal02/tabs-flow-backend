package user

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
	upsertUser(user *User) error
	deleteAccount(id string) error
}

type userRepo struct {
	db database.DDB
}

func newUserRepository(db database.DDB) userRepository {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) getUserByID(id string) (*User, error) {
	var err error
	var user User
	var response *dynamodb.QueryOutput

	// match primary key
	keyEx := expression.Key(database.PK_NAME).Equal(expression.Value(id))
	// match sort key
	keyEx.And(expression.Key(database.SK_NAME).Equal(expression.Value(database.SORT_KEY.Profile)))
	expr, err := expression.NewBuilder().WithKeyCondition(keyEx).Build()
	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't build expression for query for user_id: %v", id), err)
		return nil, errors.New(errMsg.GetUser)
	}

	response, err = r.db.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 &r.db.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.Condition(),
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't query for user_id: %v", id), err)
		return nil, errors.New(errMsg.GetUser)
	}

	if response.Count < 1 {
		return nil, errors.New(errMsg.UserNotFound)
	}

	err = attributevalue.UnmarshalMap(response.Items[0], &user)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't unmarshal query result for user_id: %v", id), err)
		return nil, errors.New(errMsg.GetUser)
	}

	logger.Dev(user)

	return &user, nil
}

func (r *userRepo) upsertUser(user *User) error {
	profile := struct {
		user User
		sk   string
	}{

		sk: database.SORT_KEY.Profile,
	}
	item, err := attributevalue.MarshalMap(profile)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't marshal user: %#v", user), err)
		return errors.New(errMsg.CreateUser)
	}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      item,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't put item for user: %v", user.Id), err)
		return errors.New(errMsg.CreateUser)
	}

	return nil
}

// deletes the user account with all associated data
func (r *userRepo) deleteAccount(Id string) error {

	// DynamoDB allows a maximum batch size of 25 items.
	batchSize := 25

	allSKs, err := database.Helpers.GetAllSKs(&r.db, Id)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't get all dynamic sort keys for user_id: %v", Id), err)
		return errors.New(errMsg.DeleteUser)
	}

	//  prepare delete requests for all SKs
	var deleteRequests []types.WriteRequest
	for _, sk := range allSKs {
		req := types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					database.PK_NAME: &types.AttributeValueMemberS{Value: Id},
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
			logger.Error(fmt.Sprintf("Couldn't batch delete items for user_id: %v", Id), err)
			return errors.New(errMsg.DeleteUser)
		}
	}

	return nil
}
