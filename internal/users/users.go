package users

import (
	"encoding/json"
	"io"

	"github.com/go-playground/validator/v10"
)

type User struct {
	Id         string `json:"id" dynamodbav:"PK" validate:"required"`
	FirstName  string `json:"firstName" dynamodbav:"FirstName" validate:"required"`
	LastName   string `json:"lastName" dynamodbav:"LastName" validate:"required"`
	Email      string `json:"email" dynamodbav:"Email" validate:"required,email"`
	ProfilePic string `json:"profilePic,omitempty" dynamodbav:"ProfilePic" validate:"omitempty,url"`
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
	yearly   string
	lifeTime string
}{
	yearly:   "pri_01j9gharmwn4ev55kyzywy447w",
	lifeTime: "pri_01j9gh59zz1cs1zafbn95375h1",
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
	Start        int64              `json:"start" dynamodbav:"Start"`
	End          int64              `json:"end" dynamodbav:"End"`
	NextBillDate int64              `json:"nextBillDate,omitempty" dynamodbav:"NextBillDate,omitempty"`
}

type generalP struct {
	OpenSpace           string `json:"openSpace" dynamodbav:"OpenSpace"`
	DeleteUnsavedSpaces string `json:"deleteUnsavedSpaces" dynamodbav:"DeleteUnsavedSpaces"`
}

type searchP struct {
	Bookmarks bool `json:"bookmarks" dynamodbav:"Bookmarks"`
	Notes     bool `json:"notes" dynamodbav:"Notes"`
}

type cmdPaletteP struct {
	IsDisabled       bool     `json:"isDisabled" dynamodbav:"IsDisabled"`
	Search           searchP  `json:"search" dynamodbav:"Search"`
	DisabledCommands []string `json:"disabledCommands" dynamodbav:"DisabledCommands"`
}

type notesP struct {
	IsDisabled     bool   `json:"isDisabled" dynamodbav:"IsDisabled"`
	BubblePos      string `json:"bubblePos" dynamodbav:"BubblePos"`
	ShowOnAllSites bool   `json:"showOnAllSites" dynamodbav:"ShowOnAllSites"`
}
type autoDiscardP struct {
	IsDisabled              bool     `json:"isDisabled" dynamodbav:"IsDisabled"`
	DiscardTabsAfterIdleMin int32    `json:"discardTabsAfterIdleMin" dynamodbav:"DiscardTabsAfterIdleMin"`
	WhitelistedDomains      []string `json:"whitelistedDomains" dynamodbav:"WhitelistedDomains"`
}

type linkPreviewP struct {
	IsDisabled  bool   `json:"isDisabled" dynamodbav:"IsDisabled"`
	OpenTrigger string `json:"openTrigger" dynamodbav:"openTrigger"`
	Size        string `json:"size" dynamodbav:"Size"`
}

type Preferences struct {
	General     generalP     `json:"general,omitempty" dynamodbav:"P#General"`
	CmdPalette  cmdPaletteP  `json:"cmdPalette,omitempty" dynamodbav:"P#CmdPalette"`
	Notes       notesP       `json:"notes,omitempty" dynamodbav:"P#Notes"`
	AutoDiscard autoDiscardP `json:"autoDiscard,omitempty" dynamodbav:"P#AutoDiscard"`
	LinkPreview linkPreviewP `json:"linkPreview,omitempty" dynamodbav:"P#LinkPreview"`
}

var defaultUserPref = Preferences{
	General: generalP{
		DeleteUnsavedSpaces: "week",
		OpenSpace:           "newWindow",
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

var ErrMsg = struct {
	GetUser               string
	UserNotFound          string
	UserExists            string
	CreateUser            string
	UpdateUser            string
	DeleteUser            string
	InvalidUserId         string
	PreferencesGet        string
	PreferencesUpdate     string
	SubscriptionGet       string
	SubscriptionUpdate    string
	SubscriptionCheck     string
	SubscriptionPaddleURL string
}{
	GetUser:               "Error getting user",
	UserNotFound:          "User not found",
	UserExists:            "User already exits",
	CreateUser:            "Error creating user",
	UpdateUser:            "Error updating user",
	DeleteUser:            "Error deleting user",
	InvalidUserId:         "Invalid user id",
	PreferencesGet:        "Error getting preferences",
	PreferencesUpdate:     "Error updating preferences",
	SubscriptionGet:       "Error getting subscription",
	SubscriptionUpdate:    "Error updating subscription",
	SubscriptionCheck:     "Error checking subscription status",
	SubscriptionPaddleURL: "Error getting paddle url",
}
