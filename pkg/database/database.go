package database

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/manishMandal02/tabsflow-backend/config"
)

type DDB struct {
	Client    *dynamodb.Client
	TableName string
}

func New() *DDB {

	return &DDB{
		Client:    newDBB(),
		TableName: config.DDBTableName,
	}
}

// func NewWithTableName(tableName string) *DDB {
// 	return &DDB{
// 		Client:    newDBB(),
// 		TableName: tableName,
// 	}
// }

func newDBB() *dynamodb.Client {
	return dynamodb.New(dynamodb.Options{
		Region: config.AWSRegion,
	})
}
