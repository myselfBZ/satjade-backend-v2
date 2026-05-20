package events

import (
	"time"

	"github.com/google/uuid"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)

// client sent
const (
	// ChallengeRequest event is sent whe a user challenges another user.
	ChallengeRequestType EventType = "event_challenge_request"
	// AcceptChallenge event is sent when a user accepts an incoming challenge request.
	AcceptChallengeType EventType = "event_challenge_accepted"
	// QuitDuel event is sent when a user quits a duel
	QuitDuelType EventType = "event_quit_duel"
	// CheckResponse event is sent when a user checks their response for a question.
	CheckResponseType EventType = "event_duel_check_response"
	// DueluserDone event is sent when a user answers all the questions.
	DuelUserDoneType EventType = "event_user_done"
)

// server sent
const (
	// NewFriend event is sent when another user accepts their friendship request
	NewFriendType EventType = "event_new_friend"
	// FriendReqeust event is sent when a user makes a friendship request to another user.
	FriendRequestType EventType = "event_friend_request"
	// IncomingChallenge event is sent
	IncomingChallengeType EventType = "event_incoming_challenge"
	DuelCreatedType       EventType = "event_duel_created"
	DuelPeerQuitType      EventType = "event_duel_peer_quit"
	ResponseCheckedType   EventType = "event_response_checked"
	DuelPeerDoneType      EventType = "event_peer_done"
	DuelFinishedType      EventType = "event_duel_finished"
)

// ------------------------------------------- SERVER SENT EVENTS ------------------------------------------

var _ ServerSentEvent = (*UserStatusChange)(nil)
var _ ServerSentEvent = (*NewFriend)(nil)
var _ ServerSentEvent = (*FriendRequest)(nil)
var _ ServerSentEvent = (*IncomingChallenge)(nil)
var _ ServerSentEvent = (*DuelCreated)(nil)
var _ ServerSentEvent = (*PeerQuitDuel)(nil)
var _ ServerSentEvent = (*DuelFinished)(nil)

type UserStatusChange struct {
	UserId string `json:"user_id"`
	Status string `json:"status"`
}

func (u *UserStatusChange) isServerSentEvent() {}

type IncomingChallenge struct {
	ChallengeId  string `json:"challenge_id"`
	FromId       string `json:"from_id"`
	ExpiresIsSec int    `json:"expires_in_seconds"`
}

func (n *IncomingChallenge) isServerSentEvent() {}

type NewFriend struct {
	FullName     string    `json:"full_name"`
	IsOnline     bool      `json:"is_online"`
	FriendId     uuid.UUID `json:"id"`
	FriendshipId uuid.UUID `json:"frienship_id"`
	FriendsSince time.Time `json:"friends_since"`
}

func (n *NewFriend) isServerSentEvent() {}

type FriendRequest struct {
	domain.FriendshipRequest
}

func (n *FriendRequest) isServerSentEvent() {}

type DuelCreated struct {
	QuestionIds uuid.UUIDs `json:"question_ids"`
	Id          string     `json:"id"`
	PeerId      uuid.UUID  `json:"peer_id"`
}

func (d *DuelCreated) isServerSentEvent() {}

type PeerQuitDuel struct{}

func (d *PeerQuitDuel) isServerSentEvent() {}

type ResponseChecked struct {
	IsCorrect  bool   `json:"is_correct"`
	QuestionId string `json:"question_id"`
	IsPeer     bool   `json:"is_peer"`
}

func (r *ResponseChecked) isServerSentEvent() {}

type DuelPeerDone struct{}

func (d *DuelPeerDone) isServerSentEvent() {}

type DuelFinished struct {
	WinnerId string `json:"winner_id"`
	IsDraw   bool   `json:"is_draw"`
}

func (d *DuelFinished) isServerSentEvent() {}

// ------------------------------------------- CLIENT SENT EVENTS ---------------------------------
var _ ClientSentEvent = (*HeartBeatEvent)(nil)
var _ ClientSentEvent = (*AcceptChallenge)(nil)
var _ ClientSentEvent = (*ChallengeRequest)(nil)
var _ ClientSentEvent = (*QuitDuel)(nil)
var _ ClientSentEvent = (*CheckResponse)(nil)
var _ ClientSentEvent = (*DuelUserDone)(nil)

type HeartBeatEvent struct{}

func (h *HeartBeatEvent) isClientSentEvent() {}

type ChallengeRequest struct {
	ToId string `json:"to_id"`
}

func (n *ChallengeRequest) isClientSentEvent() {}

type AcceptChallenge struct {
	ChallengeId string `json:"challenge_id"`
}

func (n *AcceptChallenge) isClientSentEvent() {}

type QuitDuel struct {
	DuelId string `json:"duel_id"`
}

func (d *QuitDuel) isClientSentEvent() {}

type CheckResponse struct {
	DuelId     string `json:"duel_id"`
	Reponse    string `json:"response"`
	QuestionId string `json:"question_id"`
}

func (ch *CheckResponse) isClientSentEvent() {}

type DuelUserDone struct {
	DuelId string `json:"duel_id"`
}

func (d *DuelUserDone) isClientSentEvent() {}
