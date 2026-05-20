package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	questions_service "github.com/myselfBZ/satjade-backend/internal/services/questions"
	"github.com/myselfBZ/satjade-backend/internal/ws/challenge"
	"github.com/myselfBZ/satjade-backend/internal/ws/clients"
	"github.com/myselfBZ/satjade-backend/internal/ws/events"
)

var (
	ErrInvalidEventType = fmt.Errorf("error: invalid event type")
	ErrMalformedPayload = fmt.Errorf("error: malformed payload")
	ErrNoAuthToken      = fmt.Errorf("error: timed out, no auth token after 10 seconds")
	ErrUserOffline      = fmt.Errorf("error: user is not online")
	ErrConnClosed       = fmt.Errorf("error: connection closed")
)

type eventWrapper struct {
	event  events.ClientSentEvent
	client *clients.Client
}

func (a *api) handleWSConn(c echo.Context) error {
	conn, err := websocket.Accept(c.Response().Writer, c.Request(), &websocket.AcceptOptions{
		OriginPatterns: []string{
			a.config.frontEndUrl,
		},
	})

	if err != nil {
		return nil
	}

	client, err := a.authenticateWsConn(conn)

	if err != nil {
		switch err {
		case ErrNoAuthToken:
			return nil
		default:
			a.logger.Errorw("authentication failed", "err", err)
			if err := wsjson.Write(c.Request().Context(), conn, events.ServerSentEventPayload{
				Type: events.ErrEventType,
				Body: &events.ErrEvent{
					Message: err.Error(),
				},
			}); err != nil {
				a.logger.Warnw("could not write to ws connection", "err", err)
			}

			conn.Close(websocket.StatusPolicyViolation, "authentication failed")
			return nil
		}

	}

	a.wsClients.Set(client)

	client.WriteEvent(events.ServerSentEventPayload{
		Type: events.AckAuthEventType,
		Body: nil,
	})

	a.broadcastStatusChange(c.Request().Context(), client.User.ID, "online")

	a.readLoop(client)
	return nil
}

func (a *api) readLoop(client *clients.Client) {

	for {
		event, err := client.ReadEvent()

		if err != nil {

			if err == clients.ErrConnClosed {
				a.wsClientExitCh <- client.User.ID.String()
				break
			}
			a.logger.Errorw("ReadEvent() returned an error", "e", err)
		}

		a.eventCh <- eventWrapper{
			client: client,
			event:  event,
		}
	}
}

// TODO for this function, errors need to be changed.
// Should provide concrete distinguishable errors
// so the parent function can distinguish them and return user-friendly errors at that level
func (a *api) authenticateWsConn(conn *websocket.Conn) (*clients.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var payload events.ClientSentEventPayload
	if err := wsjson.Read(ctx, conn, &payload); err != nil {
		// TODO figure out a better way to do it, errors are nested af
		if strings.Contains(err.Error(), "context deadline exceeded") {
			return nil, ErrNoAuthToken
		}

		return nil, fmt.Errorf("error: malformed json input from client")
	}

	if payload.Type != events.AuthEventType {
		return nil, fmt.Errorf("error: client sent a non-auth packet")
	}

	var body events.AuthEvent

	if err := json.Unmarshal(payload.Body, &body); err != nil {
		return nil, fmt.Errorf("error: malformed json input from client")
	}

	validToken, err := a.auth.ValidateToken(body.Token)

	if err != nil {
		return nil, err
	}

	claims, ok := validToken.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}

	userID, ok := claims["sub"].(string)

	if !ok {
		return nil, fmt.Errorf("invalid subject claim")
	}

	validId, err := uuid.Parse(userID)

	if err != nil {
		return nil, fmt.Errorf("invalid user identity in token")
	}

	user, err := a.services.Users.GetById(ctx, validId)
	if err != nil {
		switch err {
		case domain.ErrRecordNotFound:
			return nil, fmt.Errorf("user not found")
		default:
			return nil, fmt.Errorf("server error")
		}
	}

	return clients.NewClient(clients.ClientParams{
		User:   user,
		Conn:   conn,
		ExitCh: a.wsClientExitCh,
	}), nil
}

func (a *api) onDissconnect() {
	for id := range a.wsClientExitCh {
		a.wsClients.Del(id)
		// TODO, handle the damn error
		validId, _ := uuid.Parse(id)
		a.broadcastStatusChange(context.TODO(), validId, "offline")
	}
}

func (a *api) handleEventLoop() {
	for e := range a.eventCh {
		switch eT := e.event.(type) {
		case *events.ChallengeRequest:
			go a.handleChallengeRequest(e.client, *eT)
		case *events.AcceptChallenge:
			go a.handleChallengeAccepted(e.client, *eT)
		case *events.QuitDuel:
			go a.handleQuitDuel(e.client, *eT)
		case *events.CheckResponse:
			go a.handleCheckResponse(e.client, *eT)
		case *events.DuelUserDone:
			go a.handleDuelDone(e.client, *eT)
		}
	}
}

func (a *api) handleQuitDuel(selfClient *clients.Client, event events.QuitDuel) {
	duel, ok := a.duels.Get(event.DuelId)

	if !ok {
		a.logger.Errorw("quit request to a non-existing duel match", "id", event.DuelId)
		selfClient.WriteErrEvent("duel not found")
		return
	}

	// ownership check
	if duel.user1 != selfClient.User.ID.String() && duel.user2 != selfClient.User.ID.String() {
		a.logger.Errorw("unauthorized duel access attempt",
			"duel_id", duel.Id,
			"attempted_by", selfClient.User.ID,
			"owner_1", duel.user1,
			"owner_2", duel.user2,
		)
		selfClient.WriteErrEvent("unauthorized duel access")
		return
	}

	duel.exitCh <- struct{}{}

	peerId := duel.user1
	if peerId == selfClient.User.ID.String() {
		peerId = duel.user2
	}

	peer, ok := a.wsClients.Get(peerId)

	// fine
	if !ok {
		return
	}

	peer.WriteEvent(events.ServerSentEventPayload{
		Type: events.DuelPeerQuitType,
		Body: &events.PeerQuitDuel{},
	})
}

func (a *api) handleChallengeRequest(selfClient *clients.Client, event events.ChallengeRequest) {
	client, ok := a.wsClients.Get(event.ToId)

	if !ok {
		selfClient.WriteErrEvent("challenge sent to an offline user")
		a.logger.Warnw("challenge sent to an offline user", "to", event.ToId, "from", selfClient.User.ID)
		return
	}

	state := client.GetState()

	if state.Type != clients.Idle {
		selfClient.WriteErrEvent("peer is already in a match")
		a.logger.Warnw("challenge sent to a user in a duel match", "to", event.ToId, "from", selfClient.User.ID)
		return
	}

	id, err := a.challenges.Create(event.ToId, selfClient.User.ID.String())

	if err != nil {
		selfClient.WriteErrEvent("internal server error")
		a.logger.Errorw("could not create a challenge", "err", err)
		return
	}

	client.WriteEvent(
		events.ServerSentEventPayload{
			Type: events.IncomingChallengeType,
			Body: &events.IncomingChallenge{
				ChallengeId:  id,
				FromId:       selfClient.User.ID.String(),
				ExpiresIsSec: 5,
			},
		},
	)
}

func (a *api) handleChallengeAccepted(selfClient *clients.Client, event events.AcceptChallenge) {
	ch, err := a.challenges.Accept(event.ChallengeId, selfClient.User.ID.String())

	if err != nil {
		switch err {
		case challenge.ErrChallengeNotFound:
			a.logger.Errorw("accept event to a non-existing challenge", "id", event.ChallengeId)
			selfClient.WriteErrEvent("challenge not found")
			return
		case challenge.ErrReciepientIdMismatch:
			a.logger.Errorw("unauthorized access to challenge", "user_id", selfClient.User.ID.String())
			selfClient.WriteErrEvent("you cannot accept this challenge")
			return
		}
	}

	// let's make sure the other guy is online
	peer, ok := a.wsClients.Get(ch.FromId)


	if !ok {
		selfClient.WriteErrEvent("peer went offlien")
		return
	}

	duelId, err := uuid.NewUUID()

	if err != nil {
		selfClient.WriteErrEvent("internal server error")
		a.logger.Errorw("could not generate a uuid", "err", err)
		return
	}

	// let's create the duel match
	d := duel{
		Id:    duelId,
		user1: ch.ToId,
		user2: ch.FromId,
		mu:    sync.Mutex{},
		correctAnswers: map[string]int{
			ch.FromId: 0,
			ch.ToId:   0,
		},
		doneUsers: make(map[string]bool),
		exitCh:    make(chan struct{}),
	}

	go func() {
		select {
		case <-time.After(time.Minute * 15):
			a.duels.Del(duelId.String())
		case <-d.exitCh:
			a.duels.Del(duelId.String())
		}
	}()

	ids, err := a.services.Questions.GetRandomIds(context.TODO())

	if err != nil {
		selfClient.WriteErrEvent("internal server error")
		a.logger.Errorw("could'nt fetch question ids", "err", err)
		return
	}

	a.duels.Set(duelId.String(), &d)

	go peer.WriteEvent(events.ServerSentEventPayload{
		Type: events.DuelCreatedType,
		Body: &events.DuelCreated{
			QuestionIds: ids,
			Id:          duelId.String(),
			PeerId:      selfClient.User.ID,
		},
	})

	go selfClient.WriteEvent(events.ServerSentEventPayload{
		Type: events.DuelCreatedType,
		Body: &events.DuelCreated{
			QuestionIds: ids,
			Id:          duelId.String(),
			PeerId:      peer.User.ID,
		},
	})
}

func (a *api) handleCheckResponse(selfClient *clients.Client, event events.CheckResponse) {
	duel, ok := a.duels.Get(event.DuelId)

	if !ok {
		a.logger.Errorw("check response event to a non-existing duel", "id", event.DuelId)
		selfClient.WriteErrEvent("duel not found")
		return
	}

	// ownership check
	if duel.user1 != selfClient.User.ID.String() && duel.user2 != selfClient.User.ID.String() {
		a.logger.Errorw("unauthorized duel access attempt",
			"duel_id", duel.Id,
			"attempted_by", selfClient.User.ID,
			"owner_1", duel.user1,
			"owner_2", duel.user2,
		)
		selfClient.WriteErrEvent("unauthorized duel access")
		return
	}

	questionId, err := uuid.Parse(event.QuestionId)

	if err != nil {
		a.logger.Errorw("handleCheckResponse(): invalid question id", "id", event.QuestionId)
		selfClient.WriteErrEvent("invalid question id")
		return
	}

	isCorrect, err := a.services.Questions.Check(context.TODO(), questions_service.CheckParams{
		Response:   event.Reponse,
		QuestionId: questionId,
	})

	if err != nil {

		if err == domain.ErrQuestionNotFound {
			a.logger.Errorw("handleCheckResponse(): question not found", "id", event.QuestionId)
			selfClient.WriteErrEvent("question not found")
			return
		}

		selfClient.WriteErrEvent("internal server error")
		a.logger.Errorw("handleCheckResponse(): server encountered a problem", "err", err)
		return
	}

	if isCorrect {
		duel.incrementScore(selfClient.User.ID.String())
		peerId := getPeerId(selfClient.User.ID.String(), duel.user1, duel.user2)

		peer, ok := a.wsClients.Get(peerId)
		if ok {
			peer.WriteEvent(events.ServerSentEventPayload{
				Type: events.ResponseCheckedType,
				Body: &events.ResponseChecked{
					IsCorrect:  isCorrect,
					QuestionId: questionId.String(),
					IsPeer:     true,
				},
			})
		}
	}

	selfClient.WriteEvent(events.ServerSentEventPayload{
		Type: events.ResponseCheckedType,
		Body: &events.ResponseChecked{
			IsCorrect:  isCorrect,
			QuestionId: questionId.String(),
			IsPeer:     false,
		},
	})

}

func (a *api) handleDuelDone(selfClient *clients.Client, event events.DuelUserDone) {
	userId := selfClient.User.ID.String()
	duel, ok := a.duels.Get(event.DuelId)
	if !ok {
		a.logger.Errorw("done request to a non-existing duel", "id", event.DuelId)
		selfClient.WriteErrEvent("duel not found")
		return
	}

	if ok := duel.isUserDone(userId); ok {
		a.logger.Errorw("user sent redundant done event to duel", "id", userId)
		selfClient.WriteErrEvent("redundant done event sent")
		return
	}

	if count := duel.doneUsersLen(); count < 1 {
		duel.setDoneUser(userId)
		peerId := duel.getPeer(userId)
		peer, ok := a.wsClients.Get(peerId)

		if !ok {
			duel.exitCh <- struct{}{}
			selfClient.WriteEvent(events.ServerSentEventPayload{
				Type: events.DuelPeerQuitType,
				Body: &events.PeerQuitDuel{},
			})
			return
		}

		peer.WriteEvent(events.ServerSentEventPayload{
			Type: events.DuelPeerDoneType,
			Body: &events.DuelPeerDone{},
		})
		return
	}

	peerId := duel.getPeer(userId)
	peer, ok := a.wsClients.Get(peerId)

	if !ok {
		duel.exitCh <- struct{}{}
		selfClient.WriteEvent(events.ServerSentEventPayload{
			Type: events.DuelPeerQuitType,
			Body: &events.PeerQuitDuel{},
		})
		return
	}

	winnerId := duel.getWinnerId()

	go peer.WriteEvent(events.ServerSentEventPayload{
		Type: events.DuelFinishedType,
		Body: &events.DuelFinished{
			WinnerId: winnerId,
			IsDraw:   winnerId == "",
		},
	})

	selfClient.WriteEvent(events.ServerSentEventPayload{
		Type: events.DuelFinishedType,
		Body: &events.DuelFinished{
			WinnerId: winnerId,
			IsDraw:   winnerId == "",
		},
	})
}

func getPeerId(selfId string, user1, user2 string) string {
	if user1 == selfId {
		return user2
	}

	return user1
}
