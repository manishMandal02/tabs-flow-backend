package spaces

type space struct {
	Id             string `json:"id"`
	Title          string `json:"title"`
	Theme          string `json:"theme"`
	IsSaved        bool   `json:"isSaved"`
	Emoji          string `json:"emoji"`
	WindowId       int    `json:"windowId"`
	ActiveTabIndex int    `json:"activeTabIndex"`
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
