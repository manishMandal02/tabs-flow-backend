package notifications

import (
	web_push "github.com/SherClockHolmes/webpush-go"
	"github.com/manishMandal02/tabsflow-backend/config"
)

func sendWebPushNotification(userId string, s *PushSubscription, body []byte) error {
	ws := &web_push.Subscription{
		Endpoint: s.Endpoint,
		Keys: web_push.Keys{
			Auth:   s.AuthKey,
			P256dh: s.P256dhKey,
		},
	}
	o := &web_push.Options{
		TTL:             300,
		Subscriber:      userId,
		VAPIDPrivateKey: config.VAPID_PRIVATE_KEY,
		VAPIDPublicKey:  config.VAPID_PUBLIC_KEY,
	}
	_, err := web_push.SendNotification(body, ws, o)

	if err != nil {
		return err
	}

	return nil

}
