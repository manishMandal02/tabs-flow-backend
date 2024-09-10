package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type userRepository interface {
	getUserByID(id string) (*User, error)
	upsertUser(user *User) error
	deleteUser(id string) error
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

	keyEx := expression.Key("sk").Equal(expression.Value(id))
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
		pk         string
		sk         string
		FullName   string
		Email      string
		ProfilePic string
	}{
		pk:         user.Id,
		sk:         database.SortKey.Profile,
		FullName:   user.FullName,
		Email:      user.Email,
		ProfilePic: user.ProfilePic,
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

func (r *userRepo) deleteUser(Id string) error {

	logger.Dev(fmt.Sprintf("Delete user_id:%v", Id))

	// TODO - delete user 👇
	// get all the possible sort keys including all the dynamic ones
	// batch delete them in batches of allowed limit

	// _, err := r.db.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
	// 	TableName: &r.db.TableName,
	// 	Key: map[string]types.AttributeValue{
	// 		"pk": fmt.Sprintf("%v"),
	// 		"sk": "sk",
	// 	},
	// })

	// if err != nil {
	// 	logger.Error(fmt.Sprintf("Couldn't delete item for user_id: %v", Id), err)
	// 	return errors.New(errMsg.DeleteUser)
	// }
	return nil
}
