package notifications

import "github.com/go-playground/validator/v10"

type NotificationType string

const (
	notificationTypeAccount       NotificationType = "account"
	notificationTypeNoteRemainder NotificationType = "note_remainder"
	notificationTypeUnSnoozedType NotificationType = "un_snoozed_tab"
)

type snoozedTabNotification struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Icon  string `json:"icon"`
}

type noteRemainderNotification struct {
	Id     string `json:"id"`
	Title  string `json:"title"`
	Domain string `json:"domain"`
}

type notification struct {
	Id         string                     `json:"id"`
	Type       NotificationType           `json:"type"`
	IsRead     bool                       `json:"isRead"`
	Timestamp  int64                      `json:"timestamp"`
	Title      string                     `json:"title,omitempty"`
	Message    string                     `json:"message,omitempty"`
	Note       *noteRemainderNotification `json:"note,omitempty"`
	SnoozedTab *snoozedTabNotification    `json:"snoozedTab,omitempty"`
}

var errMsg = struct {
	notificationGet              string
	notificationDelete           string
	notificationsEmpty           string
	notificationsSubscribe       string
	notificationsSubscriptionGet string
}{
	notificationDelete:           "error deleting notification",
	notificationGet:              "error getting notifications",
	notificationsEmpty:           "no notifications found",
	notificationsSubscribe:       "error subscribing to notifications",
	notificationsSubscriptionGet: "error getting notification subscription",
}

type PushSubscription struct {
	Endpoint  string `json:"endpoint" validate:"required"`
	AuthKey   string `json:"authKey" validate:"required"`
	P256dhKey string `json:"p256dhKey" validate:"required"`
}

func (p PushSubscription) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(p)
	if err != nil {
		return err
	}

	return nil
}
