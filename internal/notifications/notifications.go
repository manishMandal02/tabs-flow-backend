package notifications

type NotificationType string

const (
	NotificationTypeAccount       NotificationType = "account"
	NotificationTypeNoteRemainder NotificationType = "note_remainder"
	NotificationTypeUnSnoozedType NotificationType = "un_snoozed_tab"
)

type noteNotification struct {
	Id     string `json:"id"`
	Title  string `json:"title"`
	Domain string `json:"domain"`
}

type snoozedTabNotification struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Icon  string `json:"icon"`
}

type notification struct {
	Id         string                 `json:"id"`
	Type       NotificationType       `json:"type"`
	IsRead     bool                   `json:"isRead"`
	Timestamp  int64                  `json:"timestamp"`
	Title      string                 `json:"title,omitempty"`
	Message    string                 `json:"message,omitempty"`
	Note       noteNotification       `json:"note,omitempty"`
	SnoozedTab snoozedTabNotification `json:"snoozedTab,omitempty"`
}

var errMsg = struct {
	notificationCreate string
	notificationUpdate string
	notificationGet    string
	notificationDelete string
	notificationsGet   string
	notificationsEmpty string
}{
	notificationCreate: "error creating notification",
	notificationUpdate: "error updating notification",
	notificationDelete: "error deleting notification",
	notificationGet:    "error getting notifications",
	notificationsEmpty: "no notifications found",
}
