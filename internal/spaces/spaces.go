package spaces

import "github.com/go-playground/validator/v10"

type space struct {
	Id             string `json:"id" validate:"required"`
	Title          string `json:"title" validate:"required"`
	Theme          string `json:"theme" validate:"required"`
	IsSaved        bool   `json:"isSaved" validate:"required"`
	Emoji          string `json:"emoji" validate:"required"`
	WindowId       int    `json:"windowId" validate:"required,number"`
	ActiveTabIndex int    `json:"activeTabIndex" validate:"required,number"`
}

func (s *space) validate() error {
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
	Id        string `json:"id"`
	Name      string `json:"name"`
	Collapsed bool   `json:"collapsed"`
	Theme     string `json:"theme"`
}

type SnoozedTab struct {
	URL          string `json:"url"`
	Title        string `json:"title"`
	Icon         string `json:"icon"`
	SnoozedAt    int64  `json:"snoozedAt"`
	SnoozedUntil int64  `json:"snoozedUntil"`
}

var errMsg = struct {
	spaceNotFound     string
	spaceGet          string
	spaceCreate       string
	spaceUpdate       string
	spaceDelete       string
	spaceGetAllByUser string
	tabsGet           string
	tabsSet           string
	groupsGet         string
	groupsSet         string
	snoozedTabsCreate string
	snoozedTabsGet    string
	snoozedTabsDelete string
}{
	spaceNotFound:     "Space not found",
	spaceGet:          "Error getting space",
	spaceCreate:       "Error creating space",
	spaceUpdate:       "Error updating space",
	spaceDelete:       "Error deleting space",
	spaceGetAllByUser: "Error getting spaces for user",
	tabsGet:           "Error getting tabs for space",
	tabsSet:           "Error setting tabs for space",
	groupsGet:         "Error getting groups for space",
	groupsSet:         "Error setting groups for space",
	snoozedTabsCreate: "Error creating snoozed tab for space",
	snoozedTabsGet:    "Error getting snoozed tabs for space",
	snoozedTabsDelete: "Error deleting snoozed tab for space",
}
