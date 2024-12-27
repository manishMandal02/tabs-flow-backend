package notifications

import (
	"encoding/json"

	"github.com/go-playground/validator/v10"
	"github.com/manishMandal02/tabsflow-backend/pkg/db"
	"github.com/manishMandal02/tabsflow-backend/pkg/logger"
)

type NotificationType string

const (
	NotificationTypeAccount       NotificationType = "account"
	NotificationTypeNoteRemainder NotificationType = "note_remainder"
	NotificationTypeUnSnoozedType NotificationType = "un_snoozed_tab"
)

type snoozedTabNotification struct {
	SnoozedAt string `json:"snoozedAt"`
	URL       string `json:"url"`
	Title     string `json:"title"`
	Icon      string `json:"icon"`
	SpaceId   string `json:"spaceId"`
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
	Message    string                     `json:"message,omitempty"`
	Note       *noteRemainderNotification `json:"note,omitempty"`
	SnoozedTab *snoozedTabNotification    `json:"snoozedTab,omitempty"`
}

// notification subscription
type PushSubscription struct {
	Endpoint  string `json:"endpoint,omitempty" validate:"required"`
	AuthKey   string `json:"authKey,omitempty" validate:"required"`
	P256dhKey string `json:"p256dhKey,omitempty" validate:"required"`
}

func (p PushSubscription) validate() error {
	validate := validator.New(validator.WithRequiredStructEnabled())

	err := validate.Struct(p)
	if err != nil {
		return err
	}

	return nil
}

// push notification event
type PushNotificationEventType string

const (
	PushNotificationEventTypeNotification        PushNotificationEventType = "USER_NOTIFICATION"
	PushNotificationEventTypeSubscriptionUpdated PushNotificationEventType = "SUBSCRIPTION_UPDATED"
	PushNotificationEventTypeProfileUpdated      PushNotificationEventType = "PROFILE_UPDATED"
	PushNotificationEventTypePreferencesUpdated  PushNotificationEventType = "PREFERENCES_UPDATED"
	PushNotificationEventTypeSpacesUpdated       PushNotificationEventType = "SPACES_UPDATED"
	PushNotificationEventTypeTabsUpdated         PushNotificationEventType = "TABS_UPDATED"
	PushNotificationEventTypeGroupsUpdated       PushNotificationEventType = "GROUPS_UPDATED"
)

type WebPushEvent[T any] struct {
	Event   PushNotificationEventType
	Payload *T
}

func (n *WebPushEvent[T]) send(userId string, r notificationRepository) error {
	if r == nil {
		db := db.New()
		r = newRepository(db)
	}

	s, err := r.getNotificationSubscription(userId)

	if err != nil && err.Error() != errMsg.notificationsSubscribeEmpty {
		return err
	}

	// user has not subscribed for notifications
	if s == nil {
		logger.Errorf("No notification subscription found for userId: %s", userId)
		return nil
	}

	b, err := json.Marshal(n)

	if err != nil {
		logger.Error("error marshalling WebPushEvent", err)
		return err
	}

	err = sendWebPushNotification(userId, s, b)

	if err != nil {
		logger.Error("error sending web push notification", err)
		return err
	}
	return nil
}

var errMsg = struct {
	notificationGet              string
	notificationPublishEvent     string
	notificationDelete           string
	notificationsEmpty           string
	notificationsSubscribe       string
	notificationsSubscribeEmpty  string
	notificationsUnsubscribe     string
	notificationsSubscriptionGet string
}{
	notificationDelete:           "error deleting notification",
	notificationGet:              "error getting notifications",
	notificationPublishEvent:     "error sending notifications",
	notificationsEmpty:           "no notifications found",
	notificationsSubscribe:       "error subscribing to notifications",
	notificationsUnsubscribe:     "error unsubscribing from notifications",
	notificationsSubscribeEmpty:  "Not subscribed to notifications",
	notificationsSubscriptionGet: "error getting notification subscription",
}
