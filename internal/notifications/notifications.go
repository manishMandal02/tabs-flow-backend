package notifications

type NotificationType string

const (
	NotificationTypeAccount       NotificationType = "account"
	NotificationTypeNoteRemainder NotificationType = "note_remainder"
	NotificationTypeUnSnoozedType NotificationType = "un_snoozed_tab"
)

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
	Note       string                 `json:"note,omitempty"`
	SnoozedTab snoozedTabNotification `json:"snoozedTab,omitempty"`
}

var errMsg = struct {
	notificationCreate           string
	notificationUpdate           string
	notificationGet              string
	notificationDelete           string
	notificationsEmpty           string
	notificationsSubscribe       string
	notificationsSubscriptionGet string
}{
	notificationCreate:           "error creating notification",
	notificationUpdate:           "error updating notification",
	notificationDelete:           "error deleting notification",
	notificationGet:              "error getting notifications",
	notificationsEmpty:           "no notifications found",
	notificationsSubscribe:       "error subscribing to notifications",
	notificationsSubscriptionGet: "error getting notification subscription",
}

type PushSubscription struct {
	Endpoint  string `json:"endpoint"`
	AuthKey   string `json:"authKey"`
	P256dhKey string `json:"p256dhKey"`
}
