package events

import (
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type EventType string

// Events
const (
	EventTypeSendOTP        EventType = "send_otp"
	EventTypeUserRegistered EventType = "user_registered"

	EventTypeScheduleNoteRemainder EventType = "schedule_note_remainder"
	EventTypeScheduleSnoozedTab    EventType = "schedule_snoozed_tab"
	EventTypeTriggerNoteRemainder  EventType = "trigger_note_remainder"
	EventTypeTriggerSnoozedTab     EventType = "trigger_snoozed_tab"
)

type SubEvent string

const (
	SubEventCreate SubEvent = "create"
	SubEventUpdate SubEvent = "update"
	SubEventDelete SubEvent = "delete"
)

type IEvent interface {
	GetEventType() EventType
	ToMsgAttributes() map[string]types.MessageAttributeValue
	ToJSON() string
	FromJSON(jsonStr string) error
}

type Event[T any] struct {
	EventType EventType `json:"event_type"`
	Payload   *T        `json:"payload"`
}

func New[e any](eventType EventType, payload *e) IEvent {
	return &Event[e]{
		EventType: eventType,
		Payload:   payload,
	}
}

func NewFromJSON[T any](jsonStr string) (*Event[T], error) {
	var ev Event[T]

	err := json.Unmarshal([]byte(jsonStr), &ev)

	if err != nil {
		return nil, err
	}

	return &ev, nil
}

// event_type info as map for sqs message
func (e Event[any]) ToMsgAttributes() map[string]types.MessageAttributeValue {

	return map[string]types.MessageAttributeValue{
		"event_type": {
			DataType:    aws.String("String"),
			StringValue: aws.String(string(e.GetEventType())),
		},
	}
}

func (e Event[any]) ToJSON() string {
	jsonBytes, err := json.Marshal(e)

	if err != nil {
		return ""
	}

	return string(jsonBytes)
}

func (e *Event[T]) FromJSON(jsonStr string) error {

	return json.Unmarshal([]byte(jsonStr), &e)

}

func (e Event[any]) GetEventType() EventType {
	return e.EventType
}

//* Event Payloads

type SendOTPPayload struct {
	Email string `json:"email"`
	OTP   string `json:"otp"`
}

type UserRegisteredPayload struct {
	Email        string `json:"email"`
	Name         string `json:"name"`
	TrailEndDate string `json:"trailEndDate"`
}

type ScheduleNoteRemainderPayload struct {
	UserId    string   `json:"userId"`
	NoteId    string   `json:"noteId"`
	TriggerAt int64    `json:"triggerAt,omitempty"`
	SubEvent  SubEvent `json:"subEvent,omitempty"`
}

type ScheduleSnoozedTabPayload struct {
	UserId       string   `json:"userId"`
	SnoozedTabId string   `json:"snoozedTabId"`
	SpaceId      string   `json:"spaceId"`
	TriggerAt    int64    `json:"triggerAt,omitempty"`
	SubEvent     SubEvent `json:"subEvent,omitempty"`
}
