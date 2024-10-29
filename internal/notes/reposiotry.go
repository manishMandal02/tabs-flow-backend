package notes

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type noteRepository interface {
	createNote(userId string, n *Note) error
	GetNote(userId string, noteId string) (*Note, error)
	getNotesByIds(userId string, noteIds *[]string) (*[]Note, error)
	getNotesByUser(userId string, lastNoteId int64) (*[]Note, error)
	updateNote(userId string, n *Note) error
	deleteNote(userId, noteId string) error
	// search
	indexSearchTerms(userId, noteId string, terms []string) error
	noteIdsBySearchTerm(userId string, query string, limit int) ([]string, error)
	deleteSearchTerms(userId, noteId string, terms []string) error
}

type noteRepo struct {
	db               *db.DDB
	searchIndexTable *db.DDB
}

func NewNoteRepository(db *db.DDB, searchIndexTable *db.DDB) noteRepository {
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

	av[db.PK_NAME] = &types.AttributeValueMemberS{Value: userId}

	av[db.SK_NAME] = &types.AttributeValueMemberS{Value: db.SORT_KEY.Notes(n.Id)}

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

	key := map[string]types.AttributeValue{
		"PK": &types.AttributeValueMemberS{Value: userId},
		"SK": &types.AttributeValueMemberS{Value: db.SORT_KEY.Notes(n.Id)},
	}

	var update expression.UpdateBuilder
	// iterate over the fields of the struct
	v := reflect.ValueOf(n)

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
		logger.Errorf("Couldn't update note for userId: %v. \n[Error]: %v", userId, err)
		return err
	}

	return nil
}

func (r noteRepo) deleteNote(userId string, noteId string) error {
	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.Notes(noteId)},
	}

	_, err := r.db.Client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName:    &r.db.TableName,
		Key:          key,
		ReturnValues: types.ReturnValueAllOld,
	})

	if err != nil {
		logger.Errorf("Couldn't delete note for userId: %v. \n[Error]: %v", userId, err)
		return err
	}

	return nil
}

func (r noteRepo) GetNote(userId string, noteId string) (*Note, error) {

	key := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.Notes(noteId)},
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
			db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
			db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.Notes(noteId)},
		})
	}

	response, err := r.db.Client.BatchGetItem(context.TODO(), &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			r.db.TableName: {
				Keys: keys,
				AttributesToGet: []string{
					db.PK_NAME,
					db.SK_NAME,
					"Title",
					"Domain",
					"UpdatedAt",
					"SpaceId",
					"Id",
					"RemainderAt",
				},
			},
		},
	})

	if err != nil {
		logger.Errorf("Couldn't get notes for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	if len(response.Responses[r.db.TableName]) < 1 {
		return nil, errors.New(errMsg.notesGetEmpty)
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

	key := expression.KeyAnd(expression.Key(db.PK_NAME).Equal(expression.Value(userId)), expression.Key(db.SK_NAME).BeginsWith(db.SORT_KEY.Notes("")))

	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()

	if err != nil {
		logger.Errorf("Couldn't build getNotesByUser() expression for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	var startKey map[string]types.AttributeValue

	if lastNoteId != 0 {
		startKey = map[string]types.AttributeValue{
			db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
			db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.Notes(fmt.Sprintf("%v", lastNoteId))},
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
		return nil, errors.New(errMsg.notesGetEmpty)
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

	// channel to collect errors from goroutines
	errChan := make(chan error, len(terms)/db.DDB_MAX_BATCH_SIZE+1)

	var wg sync.WaitGroup

	// context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqs := []types.WriteRequest{}

	for _, term := range terms {
		reqs = append(
			reqs,
			types.WriteRequest{
				PutRequest: &types.PutRequest{
					Item: map[string]types.AttributeValue{
						db.PK_NAME: &types.AttributeValueMemberS{
							Value: createSearchTermPK(userId, term),
						},
						db.SK_NAME: &types.AttributeValueMemberS{
							Value: db.SORT_KEY_SEARCH_INDEX.Note(noteId),
						},
					},
				},
			},
		)
	}

	r.db.BatchWriter(ctx, &wg, errChan, reqs)

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
		return fmt.Errorf("indexSearchTerms errors: %v", errs)
	}

	return nil
}

func (r noteRepo) noteIdsBySearchTerm(userId string, query string, limit int) ([]string, error) {

	key := expression.KeyAnd(expression.Key(db.PK_NAME).Equal(expression.Value(createSearchTermPK(userId, query))), expression.Key(db.SK_NAME).BeginsWith(db.SORT_KEY_SEARCH_INDEX.Note("")))

	expr, err := expression.NewBuilder().WithKeyCondition(key).Build()

	if err != nil {
		logger.Errorf("Couldn't build searchNotes expression for userId: %v. \n[Error]: %v", userId, err)
		return nil, err
	}

	response, err := r.searchIndexTable.Client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 &r.searchIndexTable.TableName,
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

	// channel to collect errors from goroutines
	errChan := make(chan error, len(terms)/db.DDB_MAX_BATCH_SIZE+1)

	var wg sync.WaitGroup

	// context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	reqs := []types.WriteRequest{}

	for _, term := range terms {
		reqs = append(reqs, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					db.PK_NAME: &types.AttributeValueMemberS{Value: createSearchTermPK(userId, term)},
					db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY_SEARCH_INDEX.Note(noteId)},
				},
			},
		})
	}

	r.db.BatchWriter(ctx, &wg, errChan, reqs)

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

// * helpers
func createSearchTermPK(userId string, term string) string {
	return fmt.Sprintf("%s#%s", userId, term)
}
