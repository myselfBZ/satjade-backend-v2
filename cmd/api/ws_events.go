package main

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)

type event interface {
	isEvent()
}

type eventType int

const (
	_ eventType = iota
	authEventType
	errEventType
	heartBeatEventType
	ackAuthEventType

	onlineStatusType
	offlineStatusType

	newFriendType
	friendRequestType

	challengeRequestType
	challengeReceivedType
	challengeAcceptedType

	duelCreatedType
	quitDuelType

	duelPeerQuitType
	checkResponseType
	responseCheckedType

	duelPeerDoneType
	duelUserDoneType
	duelFinishedType
)

type eventPayload struct {
	Type eventType       `json:"type"`
	Body json.RawMessage `json:"body"`

	user *domain.User
}

type eventWrapper struct {
	event  event
	client *wsClient
}

type serverSentEvent struct {
	Type eventType `json:"type"`
	Body event     `json:"body,omitempty"`
}

type authEvent struct {
	Token string `json:"token"`
}

func (a *authEvent) isEvent() {}

type errEvent struct {
	Message string `json:"message"`
}

func (e *errEvent) isEvent() {}

type heartBeatEvent struct{}

func (h *heartBeatEvent) isEvent() {}

type onlineStatus struct {
	UserId string `json:"user_id"`
}

func (o *onlineStatus) isEvent() {}

type offlineStatus struct {
	UserId string `json:"user_id"`
}

func (o *offlineStatus) isEvent() {}

type newFriend struct {
	FullName     string    `json:"full_name"`
	IsOnline     bool      `json:"is_online"`
	FriendId     uuid.UUID `json:"id"`
	FriendshipId uuid.UUID `json:"frienship_id"`
	FriendsSince time.Time `json:"friends_since"`
}

func (n *newFriend) isEvent() {}

type friendRequest struct {
	domain.FriendshipRequest
}

func (n *friendRequest) isEvent() {}

type challengeRequest struct {
	ToId string `json:"to_id"`
}

func (n *challengeRequest) isEvent() {}

type challengeReceived struct {
	ChallengeId  string `json:"challenge_id"`
	FromId       string `json:"from_id"`
	ExpiresIsSec int    `json:"expires_in_seconds"`
}

func (n *challengeReceived) isEvent() {}

type challengeAccepted struct {
	ChallengeId string `json:"challenge_id"`
}

func (n *challengeAccepted) isEvent() {}

type duelCreated struct {
	QuestionIds uuid.UUIDs `json:"question_ids"`
	Id          string     `json:"id"`
	PeerId      uuid.UUID  `json:"peer_id"`
}

func (d *duelCreated) isEvent() {}

type quitDuel struct {
	DuelId string `json:"duel_id"`
}

func (d *quitDuel) isEvent() {}

type peerQuitDuel struct{}

func (d *peerQuitDuel) isEvent() {}

type checkResponse struct {
	DuelId     string `json:"duel_id"`
	Reponse    string `json:"response"`
	QuestionId string `json:"question_id"`
}

func (ch *checkResponse) isEvent() {}

type responseChecked struct {
	IsCorrect  bool   `json:"is_correct"`
	QuestionId string `json:"question_id"`
	IsPeer     bool   `json:"is_peer"`
}

func (r *responseChecked) isEvent() {}


type duelUserDone struct{
	DuelId string `json:"duel_id"`
}
func (d *duelUserDone) isEvent() {}


type duelPeerDone struct {}
func (d *duelPeerDone) isEvent() {}


type duelFinished struct {
	WinnerId string `json:"winner_id"`
	IsDraw 	 bool 	`json:"is_draw"`
}
func (d *duelFinished) isEvent() {}
