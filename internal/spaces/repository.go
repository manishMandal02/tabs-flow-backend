package spaces

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type spaceRepository interface {
	createSpace(userId string, s *space) error
	getSpaceById(userId, spaceId string) (*space, error)
	getSpacesByUser(userId string) (*[]space, error)
	updateSpace(userId string, s *space) error
	deleteSpace(userId, spaceId string) error
	setTabsForSpace(userId, spaceId string, t *[]tab) error
	getTabsForSpace(userId, spaceId string) (*[]tab, error)
	setGroupsForSpace(userId, spaceId string, g *[]group) error
	getGroupsForSpace(userId, spaceId string) (*[]group, error)
	addSnoozedTab(userId, spaceId string, t *SnoozedTab) error
	getAllSnoozedTabs(userId string, lastSnoozedTabID int64) (*[]SnoozedTab, error)
	geSnoozedTabsInSpace(userId, spaceId string, lastSnoozedTabId int64) (*[]SnoozedTab, error)
	deleteSnoozedTab(userId, spaceId string, snoozedAt int64) error
	GetSnoozedTab(userId, spaceId string, snoozedAt int64) (*SnoozedTab, error)
}

type spaceRepo struct {
	db *database.DDB
}

func NewSpaceRepository(db *database.DDB) spaceRepository {
	return &spaceRepo{
		db: db,
	}
}

func (r spaceRepo) createSpace(userId string, s *space) error {
	av, err := attributevalue.MarshalMap(s)

	if err != nil {
		logger.Errorf("Couldn't marshal space: %v. \n[Error]: %v", s, err)
		return err
	}

	av[database.PK_NAME] = &types.AttributeValueMemberS{Value: userId}
	av[database.SK_NAME] = &types.AttributeValueMemberS{Value: database.SORT_KEY.Space(s.Id)}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      av,
	})

	if err != nil {
		logger.Errorf("Couldn't create space: %v. \n[Error]: %v", s, err)
		return err
	}

	return nil
}

func (r spaceRepo) getSpaceById(userId, spaceId string) (*space, error) {

	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.Space(spaceId)},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't get space for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	if len(response.Item) == 0 {
		return nil, errors.New(errMsg.spaceGet)
	}

	s := &space{}

	err = attributevalue.UnmarshalMap(response.Item, s)

	if err != nil {
		logger.Errorf("Couldn't unmarshal space for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	if s.Id == "" {
		return nil, errors.New(errMsg.spaceGet)
	}

	return s, nil
}

func (r spaceRepo) getSpacesByUser(userId string) (*[]space, error) {

	key := expression.KeyAnd(expression.Key("PK").Equal(expression.Value(userId)), expression.Key("SK").BeginsWith(database.SORT_KEY.Space("")))

	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()

	if err != nil {
		logger.Errorf("Couldn't build getSpacesByUser expression for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}
	response, err := r.db.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 &r.db.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	if err != nil {
		logger.Errorf("Couldn't get spaces for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	if len(response.Items) < 1 {
		return nil, errors.New(errMsg.spaceGetAllByUser)
	}

	spaces := []space{}

	err = attributevalue.UnmarshalListOfMaps(response.Items, &spaces)

	if err != nil {
		logger.Errorf("Couldn't unmarshal spaces for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return &spaces, nil
}

func (r spaceRepo) updateSpace(userId string, s *space) error {
	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: userId},
		"SK": &types.AttributeValueMemberS{Value: database.SORT_KEY.Space(s.Id)},
	}

	var update expression.UpdateBuilder

	// iterate over the fields of the struct
	v := reflect.ValueOf(s)

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
		logger.Errorf("Couldn't update space for userId: %v. \n[Error]: %v", userId, err)
		return err
	}

	return nil
}

func (r spaceRepo) deleteSpace(userId, spaceId string) error {

	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: userId},
		"SK": &types.AttributeValueMemberS{Value: database.SORT_KEY.Space(spaceId)},
	}

	_, err := r.db.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't delete space for userId: %v. \n[Error]: %v", userId, err)
		return err
	}

	return nil
}

// tabs
func (r spaceRepo) setTabsForSpace(userId, spaceId string, t *[]tab) error {
	tabs, err := attributevalue.MarshalList(*t)

	if err != nil {
		logger.Errorf("Couldn't marshal tabs: %v. \n[Error]: %v", t, err)
		return err
	}

	item := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.TabsInSpace(spaceId)},
		"Tabs":           &types.AttributeValueMemberL{Value: tabs},
	}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      item,
	})

	if err != nil {
		logger.Errorf("Couldn't set tabs for space for userId: %v. \n[Error]: %v", userId, err)
		return err
	}

	return nil
}

func (r spaceRepo) getTabsForSpace(userId, spaceId string) (*[]tab, error) {
	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.TabsInSpace(spaceId)},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't get tabs for space for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}
	if len(response.Item) == 0 {
		return nil, errors.New(errMsg.tabsGet)
	}

	tabsAttr, ok := response.Item["Tabs"]

	if !ok {
		errStr := fmt.Sprintf("Tab attribute not found for spaceId: %v for userId: %v", spaceId, userId)
		logger.Error(errStr, err)
		return nil, errors.New(errStr)
	}

	tabs := []tab{}

	err = attributevalue.Unmarshal(tabsAttr, &tabs)

	if err != nil {
		logger.Errorf("Couldn't unmarshal tabs for space for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return &tabs, nil
}

// groups
func (r spaceRepo) setGroupsForSpace(userId, spaceId string, g *[]group) error {
	groups, err := attributevalue.MarshalList(*g)

	if err != nil {
		logger.Errorf("Couldn't marshal groups: %v. \n[Error]: %v", g, err)
		return err
	}

	item := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.GroupsInSpace(spaceId)},
		"Groups":         &types.AttributeValueMemberL{Value: groups},
	}
	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      item,
	})

	if err != nil {
		logger.Errorf("Couldn't set groups for space for userId: %v. \n[Error]: %v", userId, err)
		return err
	}

	return nil

}

func (r spaceRepo) getGroupsForSpace(userId, spaceId string) (*[]group, error) {
	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.GroupsInSpace(spaceId)},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't get groups for space for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	if len(response.Item) == 0 {
		return nil, errors.New(errMsg.groupsGet)
	}

	groupsAttr, ok := response.Item["Groups"]

	if !ok {
		errStr := fmt.Sprintf("Groups attribute not found for spaceId: %v for userId: %v", spaceId, userId)
		logger.Error(errStr, err)
		return nil, errors.New(errStr)
	}

	groups := []group{}

	err = attributevalue.Unmarshal(groupsAttr, &groups)

	if err != nil {
		logger.Errorf("Couldn't unmarshal groups for space for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return &groups, nil

}

// snoozed tabs
func (r spaceRepo) addSnoozedTab(userId, spaceId string, t *SnoozedTab) error {

	snoozedTabs, err := attributevalue.MarshalMap(*t)

	if err != nil {
		logger.Errorf("Couldn't marshal snoozed tabs: %v. \n[Error]: %v", t, err)
		return err
	}

	sk := fmt.Sprintf("%s#%v", database.SORT_KEY.SnoozedTabs(spaceId), t.SnoozedUntil)

	snoozedTabs[database.PK_NAME] = &types.AttributeValueMemberS{Value: userId}
	snoozedTabs[database.SK_NAME] = &types.AttributeValueMemberS{Value: sk}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      snoozedTabs,
	})
	if err != nil {
		logger.Errorf("Couldn't add snoozed tab for userId: %v. \n[Error]: %v", userId, err)
		return err
	}

	return nil
}

func (r spaceRepo) GetSnoozedTab(userId, spaceId string, snoozedAt int64) (*SnoozedTab, error) {

	sk := fmt.Sprintf("%s#%v", database.SORT_KEY.SnoozedTabs(spaceId), snoozedAt)

	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		database.SK_NAME: &types.AttributeValueMemberS{Value: sk},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't get snoozed tab for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	if len(response.Item) == 0 {
		return nil, errors.New(errMsg.snoozedTabsGet)
	}
	snoozedTab := &SnoozedTab{}

	err = attributevalue.UnmarshalMap(response.Item, snoozedTab)

	if err != nil {
		logger.Errorf("Couldn't unmarshal snoozed tab for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return snoozedTab, nil

}

func (r spaceRepo) getAllSnoozedTabs(userId string, lastSnoozedTabId int64) (*[]SnoozedTab, error) {

	key := expression.KeyAnd(expression.Key("PK").Equal(expression.Value(userId)), expression.Key("SK").BeginsWith(database.SORT_KEY.SnoozedTabs("")))

	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()

	if err != nil {
		logger.Errorf("Couldn't build getSnoozedTabs expression for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	var startKey map[string]types.AttributeValue

	if lastSnoozedTabId != 0 {
		startKey = map[string]types.AttributeValue{
			database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
			database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.SnoozedTabs(fmt.Sprintf("%v", lastSnoozedTabId))},
		}
	}

	response, err := r.db.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 &r.db.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		Limit:                     aws.Int32(10),
		ExclusiveStartKey:         startKey,
	})

	if err != nil {
		logger.Errorf("Couldn't get snoozed tabs for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	if len(response.Items) < 1 {
		return nil, errors.New(errMsg.snoozedTabsGet)
	}

	snoozedTabs := []SnoozedTab{}

	err = attributevalue.UnmarshalListOfMaps(response.Items, &snoozedTabs)

	if err != nil {
		logger.Errorf("Couldn't unmarshal snoozed tabs for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return &snoozedTabs, nil
}

func (r spaceRepo) geSnoozedTabsInSpace(userId, spaceId string, lastSnoozedTabId int64) (*[]SnoozedTab, error) {

	key := expression.KeyAnd(expression.Key("PK").Equal(expression.Value(userId)), expression.Key("SK").BeginsWith(database.SORT_KEY.SnoozedTabs(spaceId)))

	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()

	if err != nil {
		logger.Errorf("Couldn't build getSnoozedTabs expression for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	var startKey map[string]types.AttributeValue

	if lastSnoozedTabId != 0 {
		startKey = map[string]types.AttributeValue{
			database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
			database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.SnoozedTabs(fmt.Sprintf("%s#%v", spaceId, lastSnoozedTabId))},
		}
	}

	response, err := r.db.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 &r.db.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		Limit:                     aws.Int32(10),
		ExclusiveStartKey:         startKey,
	})

	if err != nil {
		logger.Errorf("Couldn't get snoozed tabs for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	if len(response.Items) < 1 {
		return nil, errors.New(errMsg.snoozedTabsGet)
	}
	snoozedTabs := []SnoozedTab{}

	err = attributevalue.UnmarshalListOfMaps(response.Items, &snoozedTabs)

	if err != nil {
		logger.Errorf("Couldn't unmarshal snoozed tabs for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return &snoozedTabs, nil
}

func (r spaceRepo) deleteSnoozedTab(userId, spaceId string, snoozedAt int64) error {
	sk := fmt.Sprintf("%s#%s", database.SORT_KEY.SnoozedTabs(spaceId), strconv.FormatInt(snoozedAt, 10))

	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: userId},
		"SK": &types.AttributeValueMemberS{Value: sk},
	}

	_, err := r.db.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't delete snoozed tab for userId: %v. \n[Error]: %v", userId, err)
		return err
	}

	return nil
}
