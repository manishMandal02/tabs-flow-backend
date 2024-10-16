package notes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/manishMandal02/tabsflow-backend/pkg/database"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type noteRepository interface {
	createNote(userId string, n *Note) error
	GetNote(userId string, noteId string) (*Note, error)
	getNotesByIds(userId string, noteIds *[]string) (*[]Note, error)
	getNotesByUser(userId string, lastNoteId int64) (*[]Note, error)
	updateNote(userId string, n *Note) error
	deleteNote(userId string, noteId int64) (*Note, error)
	// search
	indexSearchTerms(userId, noteId string, terms []string) error
	noteIdsBySearchTerm(userId string, query string, limit int) ([]string, error)
	deleteSearchTerms(userId, noteId string, terms []string) error
}

type noteRepo struct {
	db               *database.DDB
	searchIndexTable *database.DDB
}

func NewNoteRepository(db *database.DDB, searchIndexTable *database.DDB) noteRepository {
	return &noteRepo{
		db:               db,
		searchIndexTable: searchIndexTable,
	}
}

func (r noteRepo) createNote(userId string, n *Note) error {
	av, err := attributevalue.MarshalMap(n)

	if err != nil {
		logger.Errorf("Couldn't marshal note: %v, \n[Error]: %v", n, err)
		return err
	}

	av[database.PK_NAME] = &types.AttributeValueMemberS{Value: userId}

	av[database.SK_NAME] = &types.AttributeValueMemberS{Value: database.SORT_KEY.Note(n.Id)}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      av,
	})

	if err != nil {
		logger.Errorf("Couldn't create note for userId: %v, \n[Error]: %v", userId, err)
		return err
	}

	return nil
}

func (r noteRepo) updateNote(userId string, n *Note) error {
	av, err := attributevalue.MarshalMap(n)

	if err != nil {
		logger.Errorf("Couldn't marshal note: %v. \n[Error]: %v", n, err)
		return err
	}

	av[database.PK_NAME] = &types.AttributeValueMemberS{Value: userId}
	av[database.SK_NAME] = &types.AttributeValueMemberS{Value: database.SORT_KEY.Note(n.Id)}

	_, err = r.db.Client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &r.db.TableName,
		Item:      av,
	})

	if err != nil {
		logger.Errorf("Couldn't update note for userId: %v. \n[Error]: %v", userId, err)
		return err
	}

	return nil
}

func (r noteRepo) deleteNote(userId string, noteId int64) (*Note, error) {
	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.Note(fmt.Sprintf("%d", noteId))},
	}

	res, err := r.db.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName:    &r.db.TableName,
		Key:          key,
		ReturnValues: types.ReturnValueAllOld,
	})

	if err != nil || res.Attributes == nil {
		logger.Errorf("Couldn't delete note for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	var n = Note{}

	err = attributevalue.UnmarshalMap(res.Attributes, &n)

	if err != nil {
		logger.Errorf("Couldn't unmarshal note for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return &n, nil
}

func (r noteRepo) GetNote(userId string, noteId string) (*Note, error) {

	key := map[string]types.AttributeValue{
		database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.Note(noteId)},
	}

	response, err := r.db.Client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &r.db.TableName,
		Key:       key,
	})

	if err != nil {
		logger.Errorf("Couldn't get note for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	if len(response.Item) == 0 {
		return nil, errors.New(errMsg.notesGetEmpty)
	}

	note := &Note{}

	err = attributevalue.UnmarshalMap(response.Item, note)
	if err != nil {
		logger.Errorf("Couldn't unmarshal note for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return note, nil
}

func (r noteRepo) getNotesByIds(userId string, noteIds *[]string) (*[]Note, error) {
	keys := []map[string]types.AttributeValue{}

	for _, noteId := range *noteIds {
		keys = append(keys, map[string]types.AttributeValue{
			database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
			database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.Note(noteId)},
		})
	}

	response, err := r.db.Client.BatchGetItem(context.TODO(), &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			r.db.TableName: {
				Keys: keys,
			},
		},
	})

	if err != nil {
		logger.Errorf("Couldn't get notes for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	if len(response.Responses[r.db.TableName]) < 1 {
		return nil, errors.New(errMsg.notesGet)
	}

	notes := []Note{}

	err = attributevalue.UnmarshalListOfMaps(response.Responses[r.db.TableName], &notes)

	if err != nil {
		logger.Errorf("Couldn't unmarshal notes for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return &notes, nil

}

func (r noteRepo) getNotesByUser(userId string, lastNoteId int64) (*[]Note, error) {

	key := expression.KeyAnd(expression.Key("PK").Equal(expression.Value(userId)), expression.Key("SK").BeginsWith(database.SORT_KEY.Note("")))

	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()

	if err != nil {
		logger.Errorf("Couldn't build getNotesByUser() expression for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	var startKey map[string]types.AttributeValue

	if lastNoteId != 0 {
		startKey = map[string]types.AttributeValue{
			database.PK_NAME: &types.AttributeValueMemberS{Value: userId},
			database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY.Note(fmt.Sprintf("%v", lastNoteId))},
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
		logger.Errorf("Couldn't get notes for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	if len(response.Items) < 1 {
		return nil, errors.New(errMsg.notesGet)
	}

	notes := []Note{}

	err = attributevalue.UnmarshalListOfMaps(response.Items, &notes)

	if err != nil {
		logger.Errorf("Couldn't unmarshal notes for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	return &notes, nil
}

// search index table

func (r noteRepo) indexSearchTerms(userId, noteId string, terms []string) error {
	// max batch size allowed
	batchSize := 25
	start := 0
	end := start + batchSize

	var err error

	wg := sync.WaitGroup{}

	// batch write to dynamodb search index table in batches
	for start < len(terms) {
		writeReqs := map[string][]types.WriteRequest{}
		if end > len(terms) {
			end = len(terms)
		}

		wg.Add(1)

		for _, term := range terms[start:end] {
			writeReqs[r.searchIndexTable.TableName] = append(writeReqs[r.searchIndexTable.TableName], types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						database.PK_NAME: &types.AttributeValueMemberS{Value: createSearchTermPK(userId, term)},
						database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY_SEARCH_INDEX.Note(noteId)},
					},
				},
			})
		}
		go func(reqs map[string][]types.WriteRequest) {
			defer wg.Done()
			_, err = r.searchIndexTable.Client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
				RequestItems: writeReqs,
			})
			if err != nil {
				logger.Errorf("error batch writing search terms for noteId: %v. \n[Error]: %v", noteId, err)
			}
		}(writeReqs)
	}
	return nil

}

func (r noteRepo) noteIdsBySearchTerm(userId string, query string, limit int) ([]string, error) {

	key := expression.KeyAnd(expression.Key("PK").Equal(expression.Value(createSearchTermPK(userId, query))), expression.Key("SK").BeginsWith(database.SORT_KEY_SEARCH_INDEX.Note("")))

	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()

	if err != nil {
		logger.Errorf("Couldn't build searchNotes expression for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	response, err := r.db.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 &r.db.TableName,
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		KeyConditionExpression:    expr.KeyCondition(),
		Limit:                     aws.Int32(int32(limit)),
	})

	if err != nil {
		logger.Errorf("Couldn't search notes for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	if len(response.Items) < 1 {
		return nil, errors.New(errMsg.notesSearchEmpty)
	}

	noteIdsSK := []struct {
		Id string `json:"id" dynamodbav:"SK"`
	}{}

	err = attributevalue.UnmarshalListOfMaps(response.Items, &noteIdsSK)

	if err != nil {
		logger.Errorf("Couldn't unmarshal notes for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	noteIds := []string{}

	for _, note := range noteIdsSK {
		id := strings.Split(note.Id, "#")[1]
		noteIds = append(noteIds, id)
	}

	return noteIds, nil
}

func (r noteRepo) deleteSearchTerms(userId, noteId string, terms []string) error {

	// max batch size allowed
	batchSize := 25
	start := 0
	end := start + batchSize
	var err error

	wg := sync.WaitGroup{}

	// batch write to dynamodb search index table in batches
	for start < len(terms) {
		deleteReqs := map[string][]types.WriteRequest{}
		if end > len(terms) {
			end = len(terms)
		}

		wg.Add(1)

		for _, term := range terms[start:end] {
			deleteReqs[r.searchIndexTable.TableName] = append(deleteReqs[r.searchIndexTable.TableName], types.WriteRequest{
				DeleteRequest: &types.DeleteRequest{
					Key: map[string]types.AttributeValue{
						database.PK_NAME: &types.AttributeValueMemberS{Value: createSearchTermPK(userId, term)},
						database.SK_NAME: &types.AttributeValueMemberS{Value: database.SORT_KEY_SEARCH_INDEX.Note(noteId)},
					},
				},
			})
		}

		go func(reqs map[string][]types.WriteRequest) {
			defer wg.Done()
			_, err = r.searchIndexTable.Client.BatchWriteItem(context.TODO(), &dynamodb.BatchWriteItemInput{
				RequestItems: reqs,
			})

			if err != nil {
				logger.Errorf("error batch deleting search terms for noteId: %v. \n[Error]: %v", noteId, err)
			}
		}(deleteReqs)

	}

	wg.Wait()

	return nil
}
