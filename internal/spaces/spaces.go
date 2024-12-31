package spaces

import (
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/go-playground/validator/v10"
	"github.com/manishMandal02/tabsflow-backend/config"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
)

type space struct {
	Id        string `json:"id" validate:"required"`
	Title     string `json:"title" validate:"required"`
	Theme     string `json:"theme" validate:"required"`
	IsSaved   bool   `json:"isSaved" validate:"required"`
	Emoji     string `json:"emoji" validate:"required"`
	WindowId  int    `json:"windowId" validate:"required,number"`
	UpdatedAt int64  `json:"updatedAt" validate:"number"`
}

func (s space) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(s)
	if err != nil {
		return err
	}

	return nil
}

type tab struct {
	Id      string `json:"id"`
	URL     string `json:"url"`
	Title   string `json:"title"`
	Index   int    `json:"index"`
	Icon    string `json:"icon"`
	GroupId int    `json:"groupId"`
}

type group struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Theme     string `json:"theme"`
	Collapsed bool   `json:"collapsed"`
}

type SnoozedTab struct {
	URL          string `json:"url,omitempty"`
	Title        string `json:"title,omitempty"`
	Icon         string `json:"icon,omitempty"`
	SnoozedAt    int64  `json:"snoozedAt,omitempty"`
	SnoozedUntil int64  `json:"snoozedUntil,omitempty"`
}

// initial space for new users
var defaultSpace = &space{
	Id:        "default2025",
	Title:     config.DEFAULT_SPACE_TITLE,
	Theme:     "#38bdf8",
	IsSaved:   true,
	Emoji:     "🗂️",
	WindowId:  0,
	UpdatedAt: time.Now().UnixMilli(),
}

// initial group for new users
var defaultGroups = []group{
	{
		Id:        8499388491,
		Name:      "Gmail Cleaner",
		Theme:     "blue",
		Collapsed: false,
	},
}

// initial tabs for new users
var defaultTabs = []tab{
	{
		URL:     "https://tabsflow.com",
		Title:   "TabsFlow",
		Index:   0,
		Icon:    "https://tabsflow.com/favicon.ico",
		GroupId: 0,
	},
	{
		URL:     "https://freshinbox.xyz",
		Title:   "Clean Inbox, Total Privacy | FreshInbox",
		Index:   1,
		Icon:    "https://freshinbox.xyz/favicon.ico",
		GroupId: 8499388491,
	},
	{
		URL:     "https://x.com/manishMandalJ",
		Title:   "Manish Mandal (manishMandalJ) / X",
		Index:   2,
		Icon:    "https://abs.twimg.com/favicons/twitter.2.ico",
		GroupId: 0,
	},
}

// It generates the initial space, groups, and tabs data for the provided userId
// and returns them as a Dynamodb Items.
func GetDefaultSpaceData(userId string) ([]map[string]types.AttributeValue, error) {
	items := []map[string]types.AttributeValue{}

	// set default space
	space, err := attributevalue.MarshalMap(defaultSpace)

	if err != nil {
		return nil, fmt.Errorf("Couldn't marshal space. [Error]: %v", err)
	}

	space[db.PK_NAME] = &types.AttributeValueMemberS{Value: userId}
	space[db.SK_NAME] = &types.AttributeValueMemberS{Value: db.SORT_KEY.Space(defaultSpace.Id)}
	space["UpdatedAt"] = &types.AttributeValueMemberN{Value: strconv.FormatInt(defaultSpace.UpdatedAt, 10)}

	items = append(items, space)

	// set default groups

	groups, err := attributevalue.MarshalList(defaultGroups)

	if err != nil {
		return nil, fmt.Errorf("Couldn't marshal groups. [Error]: %v", err)
	}

	groupsItem := map[string]types.AttributeValue{
		db.PK_NAME:  &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME:  &types.AttributeValueMemberS{Value: db.SORT_KEY.GroupsInSpace(defaultSpace.Id)},
		"Groups":    &types.AttributeValueMemberL{Value: groups},
		"UpdatedAt": &types.AttributeValueMemberN{Value: strconv.FormatInt(defaultSpace.UpdatedAt, 10)},
	}

	items = append(items, groupsItem)

	// set default tabs
	tabs, err := attributevalue.MarshalList(defaultTabs)

	if err != nil {
		return nil, fmt.Errorf("Couldn't marshal tabs. [Error]: %v", err)
	}

	tabsItem := map[string]types.AttributeValue{
		db.PK_NAME: &types.AttributeValueMemberS{Value: userId},
		db.SK_NAME: &types.AttributeValueMemberS{Value: db.SORT_KEY.TabsInSpace(defaultSpace.Id)},
		"Tabs":     &types.AttributeValueMemberL{Value: tabs},
	}

	items = append(items, tabsItem)

	return items, nil
}

var errMsg = struct {
	userDefaultSpace       string
	spaceNotFound          string
	dataConflict           string
	spaceGet               string
	spaceId                string
	spaceCreate            string
	spaceUpdate            string
	spaceDelete            string
	spaceActiveTabIndexGet string
	spaceActiveTabIndexSet string
	spaceGetAllByUser      string
	tabsGet                string
	tabsSet                string
	groupsGet              string
	groupsSet              string
	snoozedTabsCreate      string
	snoozedTabsGet         string
	snoozedTabsNotFound    string
	snoozedTabsSwitchSpace string
	snoozedTabsDelete      string
}{
	userDefaultSpace:       "Error setting default space",
	spaceNotFound:          "Space not found",
	dataConflict:           "data_conflict",
	spaceGet:               "Error getting space",
	spaceId:                "Invalid space id",
	spaceCreate:            "Error creating space",
	spaceUpdate:            "Error updating space",
	spaceDelete:            "Error deleting space",
	spaceActiveTabIndexGet: "Error getting active tab index",
	spaceActiveTabIndexSet: "Error setting active tab index",
	spaceGetAllByUser:      "Error getting spaces for user",
	tabsGet:                "Error getting tabs",
	tabsSet:                "Error setting tabs",
	groupsGet:              "Error getting groups",
	groupsSet:              "Error setting groups",
	snoozedTabsNotFound:    "Snoozed not found",
	snoozedTabsCreate:      "Error creating snoozed tab",
	snoozedTabsGet:         "Error getting snoozed tabs",
	snoozedTabsSwitchSpace: "Error switching snoozed tab space",
	snoozedTabsDelete:      "Error deleting snoozed tab",
}
