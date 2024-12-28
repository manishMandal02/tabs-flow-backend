package db

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/manishMandal02/tabsflow-backend/config"
	"golang.org/x/time/rate"
)

type DynamoDBClientInterface interface {
	GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	Query(ctx context.Context, params *dynamodb.QueryInput, optFns ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	BatchGetItem(ctx context.Context, params *dynamodb.BatchGetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
	TransactWriteItems(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
}

const DDB_MAX_BATCH_SIZE int = 25

type DDB struct {
	Client    DynamoDBClientInterface
	TableName string
	Limiter   *rate.Limiter
}

// new instance of main table
func New() *DDB {
	return &DDB{
		Client:    newDBB(),
		TableName: config.DDB_MAIN_TABLE_NAME,
		Limiter:   newLimiter(),
	}
}

// new instance of session table
func NewSessionTable() *DDB {
	return &DDB{
		Client:    newDBB(),
		TableName: config.DDB_SESSIONS_TABLE_NAME,
	}
}

// new instance od search index table
func NewSearchIndexTable() *DDB {
	return &DDB{
		Client:    newDBB(),
		TableName: config.DDB_SEARCH_INDEX_TABLE_NAME,
		Limiter:   newLimiter(),
	}
}

// new db client helper internal helper
func newDBB() *dynamodb.Client {
	return dynamodb.NewFromConfig(config.AWS_CONFIG)
}

func newLimiter() *rate.Limiter {
	return rate.NewLimiter(rate.Every(20*time.Millisecond), 10)
}
