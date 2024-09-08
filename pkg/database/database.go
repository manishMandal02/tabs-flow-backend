package database

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/manishMandal02/tabsflow-backend/config"
)

type Database struct {
	Client    *dynamodb.Client
	TableName string
}

func New(tableName string) *Database {

	dbTableName := tableName

	if tableName == "" {
		dbTableName = config.DDBTableName
	}

	return &Database{
		Client: dynamodb.New(dynamodb.Options{
			Region: config.AWSRegion,
		}),
		TableName: dbTableName,
	}

}
