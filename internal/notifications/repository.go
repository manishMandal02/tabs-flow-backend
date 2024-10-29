package notifications

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type notificationRepository interface {
	create(userId string, notification *notification) error
	get(userId, notificationId string) (notification, error)
	delete(userId, notificationId string) error
	subscribe(userId string, s *PushSubscription) error
	getNotificationSubscription(userId string) (*PushSubscription, error)
	getUserNotifications(userId string) ([]notification, error)
	deleteNotificationSubscription(userId string) error
}

type noteRepo struct {
	db *db.DDB
}

func newRepository(db *db.DDB) notificationRepository {
	return &noteRepo{
		db: db,
	}
}

func (nr *noteRepo) create(userId string, notification *notification) error {

	item, err := attributevalue.MarshalMap(notification)

	if err != nil {
		logger.Error("error marshalling notification", err)
		return err
	}

	item[db.PK_NAME] = &types.AttributeValueMemberS{
		Value: userId,
	}
	item[db.SK_NAME] = &types.AttributeValueMemberS{
		Value: db.SORT_KEY.Notifications(notification.Id),
	}

	_, err = nr.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &nr.db.TableName,
		Item:      item,
	})

	if err != nil {
		logger.Error("error putting notification to dynamodb", err)
		return err
	}

	return nil
}

func (nr *noteRepo) get(userId, notificationId string) (notification, error) {

	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{
			Value: userId,
		},
		db.SK_NAME: &types.AttributeValueMemberS{
			Value: db.SORT_KEY.Notifications(notificationId),
		},
	}

	result, err := nr.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &nr.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Error("error getting notification from dynamodb", err)
		return notification{}, err
	}

	if result.Item == nil {
		return notification{}, errors.New(errMsg.notificationsEmpty)
	}

	var n notification
	err = attributevalue.UnmarshalMap(result.Item, &n)

	if err != nil {
		logger.Error("error unmarshalling notification", err)
		return notification{}, err
	}

	return n, nil
}

func (nr *noteRepo) getUserNotifications(userId string) ([]notification, error) {

	key := expression.KeyAnd(expression.Key(db.PK_NAME).Equal(expression.Value(userId)), expression.Key(db.SK_NAME).BeginsWith(db.SORT_KEY.Notifications("")))
	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()

	if err != nil {
		logger.Error("error building expression", err)
		return nil, err
	}

	result, err := nr.db.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 &nr.db.TableName,
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	})

	if err != nil {
		logger.Error("error querying dynamodb", err)
		return nil, err
	}

	if result.Count < 1 {
		return nil, errors.New(errMsg.notificationsEmpty)
	}

	var notifications []notification

	err = attributevalue.UnmarshalListOfMaps(result.Items, &notifications)

	if err != nil {
		logger.Error("error unmarshalling notifications", err)
		return notifications, err
	}

	return notifications, nil

}

func (nr *noteRepo) delete(userId, notificationId string) error {

	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{
			Value: userId,
		},
		db.SK_NAME: &types.AttributeValueMemberS{
			Value: db.SORT_KEY.Notifications(notificationId),
		},
	}

	_, err := nr.db.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: &nr.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Error("error deleting notification", err)
		return err
	}

	return nil
}

// notification subscription
func (nr *noteRepo) subscribe(userId string, s *PushSubscription) error {
	item, err := attributevalue.MarshalMap(s)

	if err != nil {
		logger.Error("error marshalling notification", err)
		return err
	}

	item[db.PK_NAME] = &types.AttributeValueMemberS{
		Value: userId,
	}

	item[db.SK_NAME] = &types.AttributeValueMemberS{
		Value: db.SORT_KEY.NotificationSubscription,
	}

	_, err = nr.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &nr.db.TableName,
		Item:      item,
	})

	if err != nil {
		logger.Error("error putting notification subscription to db", err)
		return err
	}

	return nil
}

func (nr *noteRepo) getNotificationSubscription(userId string) (*PushSubscription, error) {

	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{
			Value: userId,
		},
		db.SK_NAME: &types.AttributeValueMemberS{
			Value: db.SORT_KEY.NotificationSubscription,
		},
	}

	result, err := nr.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &nr.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Error("error getting notification subscription from dynamodb", err)
		return nil, err
	}

	if result.Item == nil {
		return nil, errors.New(errMsg.notificationsSubscribeEmpty)
	}

	var s PushSubscription

	err = attributevalue.UnmarshalMap(result.Item, &s)

	if err != nil {
		logger.Error("error un_marshalling notification subscription", err)
		return nil, err
	}

	return &s, nil

}

func (nr *noteRepo) deleteNotificationSubscription(userId string) error {

	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{
			Value: userId,
		},
		db.SK_NAME: &types.AttributeValueMemberS{
			Value: db.SORT_KEY.NotificationSubscription,
		},
	}

	_, err := nr.db.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: &nr.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Error("error deleting notification subscription", err)
		return err
	}

	return nil
}
