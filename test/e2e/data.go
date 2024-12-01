package e2e_tests

import (
	"fmt"
	"time"

	"github.com/manishMandal02/tabsflow-backend/internal/users"
)

var TestUser = users.User{
	Email:      "mmjdd67@gmail.com",
	FirstName:  "Manish",
	LastName:   "Mandal",
	ProfilePic: "https://avatars.githubusercontent.com/u/123456789?v=4",
}

// spaces

var space = map[string]interface{}{
	"id":             "E34Y321",
	"title":          "Work",
	"theme":          "Green",
	"emoji":          "💼",
	"isSaved":        true,
	"windowId":       7890678432,
	"activeTabIndex": 1,
}

var tabs = []map[string]interface{}{
	{
		"id":    123456789,
		"url":   "https://github.com",
		"title": "GitHub",
		"icon":  "https://github.githubassets.com/favicons/favicon.svg",
	},
	{
		"id":    987654321,
		"url":   "https://twitter.com",
		"title": "Twitter",
		"icon":  "https://abs.twimg.com/favicons/twitter.2.ico",
	},
}

var groups = []map[string]interface{}{
	{
		"id":        49254834,
		"name":      "Backend",
		"collapsed": false,
		"theme":     "gray",
	},
	{
		"id":        49254835,
		"name":      "Extension",
		"collapsed": true,
		"theme":     "blue",
	},
}

var snoozedTabs = []map[string]interface{}{
	{
		"snoozedAt":    time.Now().Unix(),
		"url":          "https://freshinbox.xyz",
		"title":        "FreshInbox | Gmail Inbox Cleaner",
		"icon":         "https://freshinbox.xyz/favicon",
		"snoozedUntil": time.Now().Add(time.Hour * 6).Unix(),
	},
	{
		"snoozedAt":    time.Now().Add(time.Second + 4).Unix(),
		"url":          "https://manishmandal.com",
		"title":        "Manish Mandal | Fullstack Web Developer",
		"icon":         "https://manishmandal.com/favicon",
		"snoozedUntil": time.Now().Add(time.Second * 3).Unix(),
	},
}

// notes

var note1Text = `{"root":{"children":[{"children":[{"detail":0,"format":1,"mode":"normal","style":"","text":"Know the Features","type":"text","version":1},{"detail":0,"format":0,"mode":"normal","style":"","text":"UnsubscribeWith a single click, bid farewell to newsletters and promotional emails, streamlining your inbox and ensuring that only the content that truly matters remains.","type":"text","version":1}],"direction":"ltr","format":"","indent":0,"type":"paragraph","version":1},{"children":[],"direction":"ltr","format":"","indent":0,"type":"paragraph","version":1},{"children":[{"detail":0,"format":0,"mode":"normal","style":"","text":"Bulk DeleteSay goodbye to tedious, manual deletions, and effortlessly remove hundreds or thousands of emails from selected senders.","type":"text","version":1}],"direction":"ltr","format":"","indent":0,"type":"paragraph","version":1},{"type":"horizontalrule","version":1},{"children":[{"detail":0,"format":0,"mode":"normal","style":"","text":"Advanced SearchEasily locate specific emails using various filters. Once identified the bulk delete feature can clear out emails in 100s or 1000s.Privacy & SecurityFreshInbox is developed with a commitment to keeping your data secure. No data ever leaves your browser, ensuring that your sensitive information remains confidential and protected.","type":"text","version":1}],"direction":"ltr","format":"","indent":0,"type":"paragraph","version":1},{"children":[],"direction":"ltr","format":"","indent":0,"type":"paragraph","version":1},{"children":[{"detail":0,"format":1,"mode":"normal","style":"","text":"This is bold","type":"text","version":1},{"detail":0,"format":0,"mode":"normal","style":"","text":" ","type":"text","version":1},{"detail":0,"format":2,"mode":"normal","style":"","text":"this is italic","type":"text","version":1},{"detail":0,"format":0,"mode":"normal","style":"","text":" ","type":"text","version":1}],"direction":"ltr","format":"","indent":0,"type":"paragraph","version":1},{"children":[{"children":[{"detail":0,"format":0,"mode":"normal","style":"","text":" This a a numbered list","type":"text","version":1}],"direction":"ltr","format":"","indent":0,"type":"listitem","version":1,"value":1}],"direction":"ltr","format":"","indent":0,"type":"list","version":1,"listType":"number","start":1,"tag":"ol"},{"children":[{"children":[{"detail":0,"format":0,"mode":"normal","style":"","text":" This is a bulletin  list ","type":"text","version":1}],"direction":"ltr","format":"","indent":0,"type":"listitem","version":1,"value":1}],"direction":"ltr","format":"","indent":0,"type":"list","version":1,"listType":"bullet","start":1,"tag":"ul"},{"children":[{"detail":0,"format":0,"mode":"normal","style":"","text":" This is a quote","type":"text","version":1}],"direction":"ltr","format":"","indent":0,"type":"quote","version":1}],"direction":"ltr","format":"","indent":0,"type":"root","version":1}}`

var note2Text = `{"root": {"children": [ {"children":[ {"type": "text", "text": "Simple Note to test note remainder"}]}]}}`

var notes = []map[string]interface{}{
	{
		"id":        fmt.Sprintf("%d", time.Now().Unix()),
		"spaceId":   space["id"],
		"title":     "FreshTabs Launch",
		"text":      note1Text,
		"domain":    "freshinbox.xyz",
		"updatedAt": time.Now().Unix(),
	},
	{
		"id":          fmt.Sprintf("%d", time.Now().Add(time.Second+1).Unix()),
		"spaceId":     space["id"],
		"title":       "TabsFlow Launch",
		"text":        note2Text,
		"domain":      "tabsflow.com",
		"updatedAt":   time.Now().Unix(),
		"remainderAt": time.Now().Add(time.Second + 4).Unix(),
	},
}
