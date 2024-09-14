package database

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/manishMandal02/tabsflow-backend/config"
)

type DDB struct {
	Client    *dynamodb.Client
	TableName string
}

// new instance of main table
func New() *DDB {
	return &DDB{
		Client:    newDBB(),
		TableName: config.DDB_MAIN_TABLE_NAME,
	}
}

// new instance of session table
func NewSessionTable() *DDB {
	return &DDB{
		Client:    newDBB(),
		TableName: config.DDB_SESSIONS_TABLE_NAME,
	}
}

// new db client helper internal helper
func newDBB() *dynamodb.Client {
	return dynamodb.New(dynamodb.Options{
		Region: config.AWS_REGION,
	})
}

const (
	PK_NAME string = "PK"
	SK_NAME string = "SK"
)

// sort keys
type keySuffix func(s string) string

func generateKey(base string) keySuffix {
	return func(id string) string {
		return base + id
	}
}

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

var SORT_KEY_SESSIONS = struct {
	Session keySuffix
	OTP     keySuffix
	UserId  keySuffix
}{
	Session: generateKey("Session#"),
	OTP:     generateKey("OTP#"),
	UserId:  generateKey("UserId#"),
}
