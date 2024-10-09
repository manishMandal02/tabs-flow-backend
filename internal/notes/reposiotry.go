package notes

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

type noteRepository interface {
	createNote(userId string, n *note) error
	getNotes(userId string) (*[]note, error)
	updateNote(userId string, n *note) error
	deleteNote(userId, noteId string) error
}

type noteRepo struct {
	db *database.DDB
}

func newNoteRepository(db *database.DDB) noteRepository {
	return &noteRepo{
		db: db,
	}
}

func (r noteRepo) createNote(userId string, n *note) error {
	av, err := attributevalue.MarshalMap(n)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't marshal note: %v", n), err)
		return err
	}

	av[database.PK_NAME] = &types.AttributeValueMemberS{Value: userId}

	av[database.SK_NAME] = &types.AttributeValueMemberS{Value: database.SORT_KEY.Note(n.Id)}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      av,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't create note for userId: %v", userId), err)
		return err
	}

	return nil
}

func (r noteRepo) updateNote(userId string, n *note) error {
	av, err := attributevalue.MarshalMap(n)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't marshal note: %v", n), err)
		return err
	}

	av[database.PK_NAME] = &types.AttributeValueMemberS{Value: userId}
	av[database.SK_NAME] = &types.AttributeValueMemberS{Value: database.SORT_KEY.Note(n.Id)}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      av,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't update note for userId: %v", userId), err)
		return err
	}

	return nil
}

func (r noteRepo) deleteNote(userId string, noteId string) error {
	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.Note(noteId)},
	}

	_, err := r.db.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't delete note for userId: %v", userId), err)
		return err
	}
	return nil

}

func (r noteRepo) getNote(userId, noteId string) (*note, error) {

	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.Note(noteId)},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't get note for userId: %v", userId), err)
		return nil, err
	}

	if len(response.Item) == 0 {
		return nil, fmt.Errorf(errMsg.noteGet)
	}

	note := &note{}

	err = attributevalue.UnmarshalMap(response.Item, note)
	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't unmarshal note for userId: %v", userId), err)
		return nil, err
	}

	return note, nil
}

// TODO - get notes with  pagination
func (r noteRepo) getNotes(userId string) (*[]note, error) {

	key := expression.KeyAnd(expression.Key("PK").Equal(expression.Value(userId)), expression.Key("SK").BeginsWith(database.SORT_KEY.Note("")))

	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't build getNotes expression for userId: %v", userId), err)
		return nil, err
	}

	response, err := r.db.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 &r.db.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
	})

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't get notes for userId: %v", userId), err)
		return nil, err
	}

	if len(response.Items) < 1 {
		return nil, fmt.Errorf(errMsg.notesGet)
	}

	notes := []note{}

	err = attributevalue.UnmarshalListOfMaps(response.Items, &notes)

	if err != nil {
		logger.Error(fmt.Sprintf("Couldn't unmarshal notes for userId: %v", userId), err)
		return nil, err
	}

	return &notes, nil
}

// TODO - search notes service
// persistence storage for lambda
// store notes to s3 broken into searchable tokens, after they're created, updated or  deleted
// initialize lambda storage from s3
// handle search queries
