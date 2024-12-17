package spaces

import (
	"github.com/go-playground/validator/v10"
	"github.com/manishMandal02/tabsflow-backend/config"
)

type space struct {
	Id             string `json:"id" validate:"required"`
	Title          string `json:"title" validate:"required"`
	Theme          string `json:"theme" validate:"required"`
	IsSaved        bool   `json:"isSaved" validate:"required"`
	Emoji          string `json:"emoji" validate:"required"`
	WindowId       int    `json:"windowId" validate:"required,number"`
	ActiveTabIndex int    `json:"activeTabIndex" validate:"number"`
	UpdatedAt      int64  `json:"updatedAt" validate:"number"`
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

var defaultSpaceId = "default2024"

var defaultUserSpace = space{
	Id:             defaultSpaceId,
	Title:          config.DEFAULT_SPACE_TITLE,
	Theme:          "#38bdf8",
	IsSaved:        true,
	Emoji:          "🗂️",
	WindowId:       0,
	ActiveTabIndex: 1,
}

var defaultUserGroup = []group{
	{
		Id:        8499388491,
		Name:      "Gmail Cleaner",
		Theme:     "blue",
		Collapsed: false,
	},
}

var defaultUserTabs = []tab{
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

var errMsg = struct {
	userDefaultSpace    string
	spaceNotFound       string
	spaceGet            string
	spaceId             string
	spaceCreate         string
	spaceUpdate         string
	spaceDelete         string
	spaceGetAllByUser   string
	tabsGet             string
	tabsSet             string
	groupsGet           string
	groupsSet           string
	snoozedTabsCreate   string
	snoozedTabsGet      string
	snoozedTabsNotFound string
	snoozedTabsDelete   string
}{
	userDefaultSpace:    "Error setting default space",
	spaceNotFound:       "Space not found",
	spaceGet:            "Error getting space",
	spaceId:             "Invalid space id",
	spaceCreate:         "Error creating space",
	spaceUpdate:         "Error updating space",
	spaceDelete:         "Error deleting space",
	spaceGetAllByUser:   "Error getting spaces for user",
	tabsGet:             "Error getting tabs",
	tabsSet:             "Error setting tabs",
	groupsGet:           "Error getting groups",
	groupsSet:           "Error setting groups",
	snoozedTabsNotFound: "Snoozed not found",
	snoozedTabsCreate:   "Error creating snoozed tab",
	snoozedTabsGet:      "Error getting snoozed tabs",
	snoozedTabsDelete:   "Error deleting snoozed tab",
}
