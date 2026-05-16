package events

import "encoding/json"

type EventType string

// general events
const(
	AuthEventType      EventType = "event_auth"
	AckAuthEventType   EventType = "event_ack_auth"
	ErrEventType       EventType = "event_error"
	HeartBeatEventType EventType = "event_heartbeat"
	OnlineStatusType   EventType = "event_went_online"
	OfflineStatusType  EventType = "event_went_offline"
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


var _ ClientSentEvent = (*ErrEvent)(nil)

type ErrEvent struct {
	Message string `json:"message"`
}

func (e *ErrEvent) isClientSentEvent() {}


