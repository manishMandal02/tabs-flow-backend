package db

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
	Profile                  string
	Subscription             string
	UsageAnalytics           string
	PreferencesBase          string
	P_General                string
	P_Notes                  string
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
	P_Notes:                  "P#Notes",
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
