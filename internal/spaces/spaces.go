package spaces

import "github.com/go-playground/validator/v10"

type space struct {
	Id             string `json:"id" validate:"required"`
	Title          string `json:"title" validate:"required"`
	Theme          string `json:"theme" validate:"required"`
	IsSaved        bool   `json:"isSaved" validate:"required"`
	Emoji          string `json:"emoji" validate:"required"`
	WindowId       int    `json:"windowId" validate:"required,number"`
	ActiveTabIndex int    `json:"activeTabIndex" validate:"number"`
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
	Id      int    `json:"id"`
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
	URL          string `json:"url"`
	Title        string `json:"title"`
	Icon         string `json:"icon"`
	SnoozedAt    int64  `json:"snoozedAt"`
	SnoozedUntil int64  `json:"snoozedUntil"`
}

var errMsg = struct {
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
