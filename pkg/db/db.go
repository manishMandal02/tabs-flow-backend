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
}

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

const (
	DDB_MAX_BATCH_SIZE int    = 25
	PK_NAME            string = "PK"
	SK_NAME            string = "SK"
)

// sort keys
type keySuffix func(s string) string

func generateKey(base string) keySuffix {
	return func(id string) string {
		return base + id
	}
}

var SORT_KEY = struct {
	Profile                  string
	Subscription             string
	UsageAnalytics           string
	PreferencesBase          string
	P_General                string
	P_Note                   string
	P_CmdPalette             string
	P_LinkPreview            string
	P_AutoDiscard            string
	NotificationSubscription string
	Notifications            keySuffix
	Space                    keySuffix
	TabsInSpace              keySuffix
	GroupsInSpace            keySuffix
	SnoozedTabs              keySuffix
	Notes                    keySuffix
}{
	Profile:                  "U#Profile",
	Subscription:             "U#Subscription",
	UsageAnalytics:           "U#UsageAnalytics",
	PreferencesBase:          "P#",
	P_General:                "P#General",
	P_Note:                   "P#Note",
	P_CmdPalette:             "P#CmdPalette",
	P_LinkPreview:            "P#LinkPreview",
	P_AutoDiscard:            "P#AutoDiscard",
	NotificationSubscription: "U#NotificationSubscription",
	Notifications:            generateKey("U#Notification#"),
	Space:                    generateKey("S#Info#"),
	TabsInSpace:              generateKey("S#Tabs#"),
	GroupsInSpace:            generateKey("S#Groups#"),
	SnoozedTabs:              generateKey("S#SnoozedTabs#"),
	Notes:                    generateKey("N#"),
}

var SORT_KEY_SESSIONS = struct {
	Session keySuffix
	OTP     keySuffix
	UserId  keySuffix
}{
	Session: generateKey("Session#"),
	OTP:     generateKey("OTP#"),
	UserId:  generateKey("UserId#"),
}

var SORT_KEY_SEARCH_INDEX = struct {
	Note keySuffix
}{
	Note: generateKey("N#"),
}
