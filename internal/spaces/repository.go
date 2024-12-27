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
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/http_api"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type spaceRepository interface {
	createSpace(userId string, s *space) error
	getSpaceById(userId, spaceId string) (*space, error)
	getSpacesByUser(userId string) (*[]space, error)
	updateSpace(userId string, s *space) error
	deleteSpace(userId, spaceId string) error
	setActiveTabIndex(userId, spaceId string, tabIndex int64) error
	getActiveTabIndex(userId, spaceId string) (int64, error)
	setTabsForSpace(userId, spaceId string, t *[]tab, m *http_api.Metadata) error
	setGroupsForSpace(userId, spaceId string, g *[]group, m *http_api.Metadata) error
	getTabsForSpace(userId, spaceId string) (*[]tab, *http_api.Metadata, error)
	getGroupsForSpace(userId, spaceId string) (*[]group, *http_api.Metadata, error)
	addSnoozedTab(userId, spaceId string, t *SnoozedTab) error
	getAllSnoozedTabsByUser(userId string, lastSnoozedTabID int64) (*[]SnoozedTab, error)
	geSnoozedTabsInSpace(userId, spaceId string, lastSnoozedTabId int64) (*[]SnoozedTab, error)
	DeleteSnoozedTab(userId, spaceId string, snoozedAt int64) error
	GetSnoozedTab(userId, spaceId string, snoozedAt int64) (*SnoozedTab, error)
}

type spaceRepo struct {
	db *db.DDB
}

func NewSpaceRepository(db *db.DDB) spaceRepository {
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

	av[db.PK_NAME] = &types.AttributeValueMemberS{Value: userId}
	av[db.SK_NAME] = &types.AttributeValueMemberS{Value: db.SORT_KEY.Space(s.Id)}
	av["UpdatedAt"] = &types.AttributeValueMemberN{Value: strconv.FormatInt(s.UpdatedAt, 10)}

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
		db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.Space(spaceId)},
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
		return nil, errors.New(errMsg.spaceNotFound)
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

	key := expression.KeyAnd(expression.Key(db.PK_NAME).Equal(expression.Value(userId)), expression.Key(db.SK_NAME).BeginsWith(db.SORT_KEY.Space("")))

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
		return nil, errors.New(errMsg.spaceNotFound)
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
		"SK": &types.AttributeValueMemberS{Value: db.SORT_KEY.Space(s.Id)},
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
		"SK": &types.AttributeValueMemberS{Value: db.SORT_KEY.Space(spaceId)},
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

// space active tab index
func (r spaceRepo) getActiveTabIndex(userId, spaceId string) (int64, error) {
	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.SpaceActiveTab(spaceId)},
	}
	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't get active tab index for spaceId: %v. \n[Error]: %v", spaceId, err)
		return 0, err
	}

	if len(response.Item) == 0 {
		return 0, errors.New(errMsg.spaceActiveTabIndexGet)
	}

	var activeTabIndex int64
	err = attributevalue.Unmarshal(response.Item["ActiveTabIndex"], &activeTabIndex)

	if err != nil {
		logger.Errorf("Couldn't unmarshal active tab index for spaceId: %v. \n[Error]: %v", spaceId, err)
		return 0, err
	}
	return activeTabIndex, nil
}

func (r spaceRepo) setActiveTabIndex(userId, spaceId string, activeTabIndex int64) error {
	item := map[string]types.AttributeValue{
		db.PK_NAME:       &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME:       &types.AttributeValueMemberS{Value: db.SORT_KEY.SpaceActiveTab(spaceId)},
		"ActiveTabIndex": &types.AttributeValueMemberN{Value: strconv.FormatInt(activeTabIndex, 10)},
	}

	_, err := r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      item,
	})

	if err != nil {
		logger.Errorf("Couldn't set active tab index for spaceId: %v. \n[Error]: %v", spaceId, err)
		return err
	}

	return nil
}

// tabs
func (r spaceRepo) setTabsForSpace(userId, spaceId string, t *[]tab, m *http_api.Metadata) error {

	tabs, err := attributevalue.MarshalListWithOptions(t)

	if err != nil {
		logger.Errorf("Couldn't marshal tabs: %v. \n[Error]: %v", t, err)
		return err
	}

	item := map[string]types.AttributeValue{
		db.PK_NAME:  &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME:  &types.AttributeValueMemberS{Value: db.SORT_KEY.TabsInSpace(spaceId)},
		"Tabs":      &types.AttributeValueMemberL{Value: tabs},
		"UpdatedAt": &types.AttributeValueMemberN{Value: strconv.FormatInt(m.UpdatedAt, 10)},
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

func (r spaceRepo) getTabsForSpace(userId, spaceId string) (*[]tab, *http_api.Metadata, error) {
	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.TabsInSpace(spaceId)},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't get tabs for space for userId: %v. \n[Error]: %v", userId, err)
		return nil, nil, err
	}
	if len(response.Item) == 0 {
		return nil, nil, errors.New(errMsg.tabsGet)
	}

	// tabs
	tabsAttr, ok := response.Item["Tabs"]

	if !ok {
		errStr := fmt.Sprintf("Tab attribute not found for spaceId: %v for userId: %v", spaceId, userId)
		logger.Error(errStr, err)
		return nil, nil, errors.New(errStr)
	}

	tabs := []tab{}

	err = attributevalue.Unmarshal(tabsAttr, &tabs)

	if err != nil {
		logger.Errorf("Couldn't unmarshal tabs for space for userId: %v. \n[Error]: %v", userId, err)
		return nil, nil, err
	}

	// get updatedAt time for tabs
	updatedAtAttr, err := strconv.ParseInt(response.Item["UpdatedAt"].(*types.AttributeValueMemberN).Value, 10, 64)

	if err != nil {
		logger.Errorf("Couldn't get updatedAt for tabs for userId: %v. \n[Error]: %v", userId, err)
		return nil, nil, err
	}

	m := &http_api.Metadata{
		UpdatedAt: updatedAtAttr,
	}

	return &tabs, m, nil
}

// groups
func (r spaceRepo) setGroupsForSpace(userId, spaceId string, g *[]group, m *http_api.Metadata) error {
	groups, err := attributevalue.MarshalList(g)

	if err != nil {
		logger.Errorf("Couldn't marshal groups: %v. \n[Error]: %v", g, err)
		return err
	}

	item := map[string]types.AttributeValue{
		db.PK_NAME:  &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME:  &types.AttributeValueMemberS{Value: db.SORT_KEY.GroupsInSpace(spaceId)},
		"Groups":    &types.AttributeValueMemberL{Value: groups},
		"UpdatedAt": &types.AttributeValueMemberN{Value: strconv.FormatInt(m.UpdatedAt, 10)},
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

func (r spaceRepo) getGroupsForSpace(userId, spaceId string) (*[]group, *http_api.Metadata, error) {
	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.GroupsInSpace(spaceId)},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't get groups for space for the userId: %v. \n[Error]: %v", userId, err)
		return nil, nil, err
	}

	if len(response.Item) == 0 {
		return nil, nil, errors.New(errMsg.groupsGet)
	}

	groupsAttr, ok := response.Item["Groups"]

	if !ok {
		errStr := fmt.Sprintf("Groups attribute not found for spaceId: %v for userId: %v", spaceId, userId)
		logger.Error(errStr, err)
		return nil, nil, errors.New(errStr)
	}

	groups := []group{}

	err = attributevalue.Unmarshal(groupsAttr, &groups)

	if err != nil {
		logger.Errorf("Couldn't unmarshal groups for space for the userId: %v. \n[Error]: %v", userId, err)
		return nil, nil, err
	}

	// get updatedAt time for groups
	updatedAtAttr, err := strconv.ParseInt(response.Item["UpdatedAt"].(*types.AttributeValueMemberN).Value, 10, 64)

	if err != nil {
		logger.Errorf("Couldn't get updatedAt for groups for the userId: %v. \n[Error]: %v", userId, err)
		return nil, nil, err
	}

	m := &http_api.Metadata{
		UpdatedAt: updatedAtAttr,
	}

	return &groups, m, nil

}

// snoozed tabs
func (r spaceRepo) addSnoozedTab(userId, spaceId string, t *SnoozedTab) error {

	snoozedTabs, err := attributevalue.MarshalMap(*t)

	if err != nil {
		logger.Errorf("Couldn't marshal snoozed tabs: %v. \n[Error]: %v", t, err)
		return err
	}

	sk := fmt.Sprintf("%s#%v", db.SORT_KEY.SnoozedTabs(spaceId), t.SnoozedAt)

	snoozedTabs[db.PK_NAME] = &types.AttributeValueMemberS{Value: userId}
	snoozedTabs[db.SK_NAME] = &types.AttributeValueMemberS{Value: sk}

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

	skSuffix := fmt.Sprintf("%s#%v", spaceId, snoozedAt)

	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.SnoozedTabs(skSuffix)},
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
		return nil, errors.New(errMsg.snoozedTabsNotFound)
	}
	snoozedTab := &SnoozedTab{}

	err = attributevalue.UnmarshalMap(response.Item, snoozedTab)

	if err != nil {
		logger.Errorf("Couldn't unmarshal snoozed tab for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return snoozedTab, nil

}

func (r spaceRepo) getAllSnoozedTabsByUser(userId string, lastSnoozedTabId int64) (*[]SnoozedTab, error) {

	key := expression.KeyAnd(expression.Key(db.PK_NAME).Equal(expression.Value(userId)), expression.Key(db.SK_NAME).BeginsWith(db.SORT_KEY.SnoozedTabs("")))

	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()

	if err != nil {
		logger.Errorf("Couldn't build getSnoozedTabs expression for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	var startKey map[string]types.AttributeValue

	if lastSnoozedTabId != 0 {
		startKey = map[string]types.AttributeValue{
			db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
			db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.SnoozedTabs(fmt.Sprintf("%v", lastSnoozedTabId))},
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
		return nil, errors.New(errMsg.snoozedTabsNotFound)
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

	key := expression.KeyAnd(expression.Key("PK").Equal(expression.Value(userId)), expression.Key("SK").BeginsWith(db.SORT_KEY.SnoozedTabs(spaceId)))

	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()

	if err != nil {
		logger.Errorf("Couldn't build getSnoozedTabs expression for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	var startKey map[string]types.AttributeValue

	if lastSnoozedTabId != 0 {
		startKey = map[string]types.AttributeValue{
			db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
			db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.SnoozedTabs(fmt.Sprintf("%s#%v", spaceId, lastSnoozedTabId))},
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

func (r spaceRepo) DeleteSnoozedTab(userId, spaceId string, snoozedAt int64) error {
	sk := fmt.Sprintf("%s#%s", db.SORT_KEY.SnoozedTabs(spaceId), strconv.FormatInt(snoozedAt, 10))

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
