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
	DiscardTabsAfterIdleMin int32    `json:"discardTabsAfterIdleTime" dynamodbav:"DiscardTabsAfterIdleTime"`
	WhitelistedDomains      []string `json:"whitelistedDomains" dynamodbav:"WhitelistedDomains"`
}

type linkPreviewP struct {
	IsDisabled  bool   `json:"isDisabled" dynamodbav:"IsDisabled"`
	OpenTrigger string `json:"openTrigger" dynamodbav:"openTrigger"`
	Size        string `json:"size" dynamodbav:"Size"`
}

type preferences struct {
	General     generalP     `json:"general" dynamodbav:"P#General"`
	CmdPalette  cmdPaletteP  `json:"cmdPalette" dynamodbav:"P#CmdPalette"`
	Notes       notesP       `json:"notes" dynamodbav:"P#Notes"`
	AutoDiscard autoDiscardP `json:"autoDiscard" dynamodbav:"P#AutoDiscard"`
	LinkPreview linkPreviewP `json:"linkPreview" dynamodbav:"P#LinkPreview"`
}

var defaultPreferences = map[string]interface{}{
	"General": map[string]interface{}{
		"OpenSpace":           "newWindow",
		"DeleteUnsavedSpaces": "week",
	},
	"CmdPalette": map[string]interface{}{
		"IsDisabled": false,
		"Search": map[string]interface{}{
			"Bookmarks": true,
			"Notes":     true,
		},
		"DisabledCommands": []string{},
	},
	"Notes": map[string]interface{}{
		"IsDisabled":     false,
		"BubblePos":      "bottom-right",
		"ShowOnAllSites": true,
	},
	"AutoDiscard": map[string]interface{}{
		"IsDisabled":              false,
		"DiscardTabsAfterIdleMin": 10,
		"WhitelistedDomains":      []string{},
	},
	"LinkPreview": map[string]interface{}{
		"IsDisabled":  false,
		"OpenTrigger": "shift-click",
		"Size":        "tablet",
	},
}

var errMsg = struct {
	getUser           string
	userNotFound      string
	userExists        string
	createUser        string
	updateUser        string
	deleteUser        string
	invalidUserId     string
	preferencesGet    string
	preferencesUpdate string
}{
	getUser:           "Error getting user",
	userNotFound:      "User not found",
	userExists:        "User already exits",
	createUser:        "Error creating user",
	updateUser:        "Error updating user",
	deleteUser:        "Error deleting user",
	invalidUserId:     "Invalid user id",
	preferencesGet:    "Error getting user preferences",
	preferencesUpdate: "Error updating user preferences",
}
