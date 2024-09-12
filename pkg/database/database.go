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

// sort keys

type keySuffix func(s string) string

func generateKey(base string) keySuffix {
	return func(id string) string {
		return base + id
	}
}

const (
	PK_NAME string = "PK"
	SK_NAME string = "SK"
)

var SORT_KEY = struct {
	Profile        string
	Subscription   string
	UsageAnalytics string
	P_General      string
	P_Note         string
	P_CmdPalette   string
	P_LinkPreview  string
	P_AutoDiscard  string
	Notifications  keySuffix
	Space          keySuffix
	TabsInSpace    keySuffix
	GroupsInSpace  keySuffix
	SnoozedTabs    keySuffix
	Note           keySuffix
}{
	Profile:        "P#Profile",
	Subscription:   "U#Subscription",
	UsageAnalytics: "U#UsageAnalytics",
	P_General:      "P#General",
	P_Note:         "P#Note",
	P_CmdPalette:   "P#CmdPalette",
	P_LinkPreview:  "P#LinkPreview",
	P_AutoDiscard:  "P#AutoDiscard",
	Notifications:  generateKey("U#Notification#"),
	Space:          generateKey("S#Info#"),
	TabsInSpace:    generateKey("S#Tabs#"),
	GroupsInSpace:  generateKey("S#Groups#"),
	SnoozedTabs:    generateKey("S#SnoozedTabs#"),
	Note:           generateKey("N#Note#"),
}
