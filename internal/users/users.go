package users

import (
	"encoding/json"
	"io"

	"github.com/go-playground/validator/v10"
)

type User struct {
	Id         string `json:"id" dynamodbav:"PK" validate:"required"`
	FullName   string `json:"fullName" dynamodbav:"FullName" validate:"required"`
	Email      string `json:"email" dynamodbav:"Email" validate:"required,email"`
	ProfilePic string `json:"profilePic" dynamodbav:"ProfilePic" validate:"url"`
}

type userWithSK struct {
	*User
	SK string `json:"sk" dynamodbav:"SK"`
}

func userFromJSON(body io.Reader) (*User, error) {

	var u *User

	err := json.NewDecoder(body).Decode(&u)

	if err != nil {
		return nil, err
	}

	return u, nil

}

func (u *User) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(u)

	if err != nil {
		return err
	}

	return nil

}

type SubscriptionPlan string

const (
	SubscriptionPlanTrial    SubscriptionPlan = "TRAIL"
	SubscriptionPlanYearly   SubscriptionPlan = "YEARLY"
	SubscriptionPlanLifetime SubscriptionPlan = "LIFE_TIME"
)

// paddle plan/price id
var paddlePlanId = struct {
	Yearly   string
	LifeTime string
}{
	Yearly:   "pri_01j9gharmwn4ev55kyzywy447w",
	LifeTime: "pri_01j9gh59zz1cs1zafbn95375h1",
}

type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusPaused   SubscriptionStatus = "paused"
	SubscriptionStatusTrialing SubscriptionStatus = "trialing"
)

type subscription struct {
	Id           string             `json:"id" dynamodbav:"Id"`
	Plan         SubscriptionPlan   `json:"plan" dynamodbav:"Plan"`
	Status       SubscriptionStatus `json:"status" dynamodbav:"Status"`
	End          string             `json:"end" dynamodbav:"End"`
	Start        string             `json:"start" dynamodbav:"Start"`
	NextBillDate string             `json:"nextBillDate,omitempty" dynamodbav:"NextBillDate,omitempty"`
}

type generalP struct {
	OpenSpace           string `json:"openSpace,omitempty" dynamodbav:"OpenSpace"`
	DeleteUnsavedSpaces string `json:"deleteUnsavedSpaces,omitempty" dynamodbav:"DeleteUnsavedSpaces"`
}

type searchP struct {
	Bookmarks bool `json:"bookmarks,omitempty" dynamodbav:"Bookmarks"`
	Notes     bool `json:"notes,omitempty" dynamodbav:"Notes"`
}

type cmdPaletteP struct {
	IsDisabled       bool     `json:"isDisabled,omitempty" dynamodbav:"IsDisabled"`
	Search           searchP  `json:"search,omitempty" dynamodbav:"Search"`
	DisabledCommands []string `json:"disabledCommands,omitempty" dynamodbav:"DisabledCommands"`
}

type notesP struct {
	IsDisabled     bool   `json:"isDisabled,omitempty" dynamodbav:"IsDisabled"`
	BubblePos      string `json:"bubblePos,omitempty" dynamodbav:"BubblePos"`
	ShowOnAllSites bool   `json:"showOnAllSites,omitempty" dynamodbav:"ShowOnAllSites"`
}
type autoDiscardP struct {
	IsDisabled              bool     `json:"isDisabled,omitempty" dynamodbav:"IsDisabled"`
	DiscardTabsAfterIdleMin int32    `json:"DiscardTabsAfterIdleMin,omitempty" dynamodbav:"DiscardTabsAfterIdleMin"`
	WhitelistedDomains      []string `json:"whitelistedDomains,omitempty" dynamodbav:"WhitelistedDomains"`
}

type linkPreviewP struct {
	IsDisabled  bool   `json:"isDisabled,omitempty" dynamodbav:"IsDisabled"`
	OpenTrigger string `json:"openTrigger,omitempty" dynamodbav:"openTrigger"`
	Size        string `json:"size,omitempty" dynamodbav:"Size"`
}

type preferences struct {
	General     generalP     `json:"general,omitempty" dynamodbav:"P#General"`
	CmdPalette  cmdPaletteP  `json:"cmdPalette,omitempty" dynamodbav:"P#CmdPalette"`
	Notes       notesP       `json:"notes,omitempty" dynamodbav:"P#Notes"`
	AutoDiscard autoDiscardP `json:"autoDiscard,omitempty" dynamodbav:"P#AutoDiscard"`
	LinkPreview linkPreviewP `json:"linkPreview,omitempty" dynamodbav:"P#LinkPreview"`
}

var defaultUserPref = preferences{
	General: generalP{
		OpenSpace:           "newWindow",
		DeleteUnsavedSpaces: "week",
	},
	CmdPalette: cmdPaletteP{
		IsDisabled:       false,
		Search:           searchP{Bookmarks: true, Notes: true},
		DisabledCommands: []string{},
	},
	Notes: notesP{
		IsDisabled:     false,
		BubblePos:      "bottom-right",
		ShowOnAllSites: true,
	},
	AutoDiscard: autoDiscardP{
		IsDisabled:              false,
		DiscardTabsAfterIdleMin: 10,
		WhitelistedDomains:      []string{},
	},
	LinkPreview: linkPreviewP{
		IsDisabled:  false,
		OpenTrigger: "shift-click",
		Size:        "tablet",
	},
}

// var defaultPreferences = map[string]interface{}{
// 	"General": map[string]interface{}{
// 		"OpenSpace":           "newWindow",
// 		"DeleteUnsavedSpaces": "week",
// 	},
// 	"CmdPalette": map[string]interface{}{
// 		"IsDisabled": false,
// 		"Search": map[string]interface{}{
// 			"Bookmarks": true,
// 			"Notes":     true,
// 		},
// 		"DisabledCommands": []string{},
// 	},
// 	"Notes": map[string]interface{}{
// 		"IsDisabled":     false,
// 		"BubblePos":      "bottom-right",
// 		"ShowOnAllSites": true,
// 	},
// 	"AutoDiscard": map[string]interface{}{
// 		"IsDisabled":              false,
// 		"DiscardTabsAfterIdleMin": 10,
// 		"WhitelistedDomains":      []string{},
// 	},
// 	"LinkPreview": map[string]interface{}{
// 		"IsDisabled":  false,
// 		"OpenTrigger": "shift-click",
// 		"Size":        "tablet",
// 	},
// }

var errMsg = struct {
	getUser               string
	userNotFound          string
	userExists            string
	createUser            string
	updateUser            string
	deleteUser            string
	invalidUserId         string
	preferencesGet        string
	preferencesUpdate     string
	subscriptionGet       string
	subscriptionUpdate    string
	subscriptionCheck     string
	subscriptionPaddleURL string
}{
	getUser:               "Error getting user",
	userNotFound:          "User not found",
	userExists:            "User already exits",
	createUser:            "Error creating user",
	updateUser:            "Error updating user",
	deleteUser:            "Error deleting user",
	invalidUserId:         "Invalid user id",
	preferencesGet:        "Error getting preferences",
	preferencesUpdate:     "Error updating preferences",
	subscriptionGet:       "Error getting subscription",
	subscriptionUpdate:    "Error updating subscription",
	subscriptionCheck:     "Error checking subscription status",
	subscriptionPaddleURL: "Error getting paddle url",
}
