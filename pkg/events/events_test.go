package events_test

import (
	"testing"

	"github.com/manishMandal02/tabsflow-backend/pkg/events"
)

func TestNewEventFromJSON(t *testing.T) {
	eventStr := `{
		"event_type": "send_otp",
		"payload": {
			"email": "test@example.com",
			"otp": "123456"
		}
	}`

	event, err := events.NewFromJSON[events.SendOTPPayload](eventStr)

	if err != nil {
		t.Errorf("Error creating event from JSON: %v", err)
	}

	if event.EventType != events.EventTypeSendOTP {
		t.Errorf("Expected event type %s, got %s", events.EventTypeSendOTP, event.EventType)
	}

	if event.Payload == nil {
		t.Errorf("Expected payload to be non-nil")
	}
	if event.Payload.Email != "test@example.com" {
		t.Errorf("Expected email to be test@example.com, got %s", event.Payload.Email)
	}
	if event.Payload.OTP != "123456" {
		t.Errorf("Expected tp to be 123456, got %s", event.Payload.OTP)
	}
}
