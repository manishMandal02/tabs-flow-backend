package events

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type EventType int

// Events
const (
	SEND_OTP EventType = iota
	USER_REGISTERED
	SCHEDULE_TASK
)

type Event interface {
	GetEventType() EventType
	ToMsgAttributes() map[string]types.MessageAttributeValue
}

// String method to get the string representation of the Event
func (e EventType) String() string {
	return [...]string{"SEND_OTP", "USER_REGISTERED", "PASSWORD_RESET"}[e]
}

type SendOTP_Payload struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

func (e SendOTP_Payload) GetEventType() EventType {
	return SEND_OTP
}

func (e SendOTP_Payload) ToMsgAttributes() map[string]types.MessageAttributeValue {
	return map[string]types.MessageAttributeValue{
		"event_type": {
			DataType:    aws.String("String"),
			StringValue: aws.String(SEND_OTP.String()),
		},
		"email": {
			DataType:    aws.String("String"),
			StringValue: aws.String(e.Email),
		},
		"otp": {
			DataType:    aws.String("String"),
			StringValue: aws.String(e.OTP),
		},
	}
}

func (e UserRegisteredPayload) GetEventType() EventType {
	return USER_REGISTERED
}

type UserRegisteredPayload struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	TrailEndDate string `json:"trailEndDate"`
}

func (e UserRegisteredPayload) ToMsgAttributes() map[string]types.MessageAttributeValue {
	return map[string]types.MessageAttributeValue{
		"event_type": {
			DataType:    aws.String("String"),
			StringValue: aws.String(USER_REGISTERED.String()),
		},
		"email": {
			DataType:    aws.String("String"),
			StringValue: aws.String(e.Email),
		},
		"name": {
			DataType:    aws.String("String"),
			StringValue: aws.String(e.Name),
		},
		"trail_end_date": {
			DataType:    aws.String("String"),
			StringValue: aws.String(e.TrailEndDate),
		},
	}
}

// ParseEventType converts a string to EventType
func ParseEventType(s string) (EventType, error) {
	switch s {
	case "SEND_OTP":
		return SEND_OTP, nil
	case "USER_REGISTERED":
		return USER_REGISTERED, nil
	case "PASSWORD_RESET":
		return SCHEDULE_TASK, nil
	default:
		return EventType(-1), fmt.Errorf("unknown event type: %s", s)
	}
}
