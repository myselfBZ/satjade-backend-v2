package events

import "encoding/json"

type EventType string

// general events
const (
	AuthEventType        EventType = "event_auth"
	AckAuthEventType     EventType = "event_ack_auth"
	ErrEventType         EventType = "event_error"
	HeartBeatEventType   EventType = "event_heartbeat"
	UserStatusChangeType EventType = "event_user_status_change"
)

type ClientSentEventPayload struct {
	Type EventType       `json:"type"`
	Body json.RawMessage `json:"body"`
}

type ServerSentEventPayload struct {
	Type EventType       `json:"type"`
	Body ServerSentEvent `json:"body,omitempty"`
}

type ServerSentEvent interface {
	isServerSentEvent()
}

type ClientSentEvent interface {
	isClientSentEvent()
}

var _ ServerSentEvent = (*ErrEvent)(nil)

type ErrEvent struct {
	Message string `json:"message"`
}

func (e *ErrEvent) isServerSentEvent() {}
