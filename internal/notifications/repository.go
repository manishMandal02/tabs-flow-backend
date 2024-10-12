package notifications

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type notificationRepository interface {
	createNotification(userId string, notification *notification) error
	getUserNotifications(userId string) ([]notification, error)
	getNotification(userId, notificationId string) (notification, error)
	deleteNotification(userId, notificationId string) error
}

type noteRepo struct {
	db *database.DDB
}

func newNotificationRepository(db *database.DDB) notificationRepository {
	return &noteRepo{
		db: db,
	}
}

func (nr *noteRepo) createNotification(userId string, notification *notification) error {

	item, err := attributevalue.MarshalMap(notification)

	if err != nil {
		logger.Error("error marshalling notification", err)
		return err
	}

	item[database.PK_NAME] = &types.AttributeValueMemberS{
		Value: userId,
	}
	item[database.SK_NAME] = &types.AttributeValueMemberS{
		Value: database.SORT_KEY.Notifications(notification.Id),
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

func (nr *noteRepo) getNotification(userId, notificationId string) (notification, error) {

	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{
			Value: userId,
		},
		database.SK_NAME: &types.AttributeValueMemberS{
			Value: database.SORT_KEY.Notifications(notificationId),
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
		return notification{}, fmt.Errorf(errMsg.notificationsEmpty)
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

	key := expression.KeyAnd(expression.Key(database.PK_NAME).Equal(expression.Value(userId)), expression.Key(database.SK_NAME).BeginsWith(database.SORT_KEY.Notifications("")))
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
		return nil, fmt.Errorf(errMsg.notificationsEmpty)
	}

	var notifications []notification

	err = attributevalue.UnmarshalListOfMaps(result.Items, &notifications)

	if err != nil {
		logger.Error("error unmarshalling notifications", err)
		return notifications, err
	}

	return notifications, nil

}

func (nr *noteRepo) deleteNotification(userId, notificationId string) error {

	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{
			Value: userId,
		},
		database.SK_NAME: &types.AttributeValueMemberS{
			Value: database.SORT_KEY.Notifications(notificationId),
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
