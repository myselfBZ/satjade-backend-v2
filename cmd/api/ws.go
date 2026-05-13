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
)

var(
	ErrInvalidEventType = fmt.Errorf("error: invalid event type")
	ErrMalformedPayload = fmt.Errorf("error: malformed payload")
	ErrNoAuthToken 		= fmt.Errorf("error: timed out, no auth token after 10 seconds")
	ErrUserOffline 		= fmt.Errorf("error: user is not online")
	ErrConnClosed 	    = fmt.Errorf("error: connection closed")
)

type wsClient struct {
	user *domain.User
	conn *websocket.Conn

	disconnectCh chan struct {}
}

func (w *wsClient) readEvent() (event, error) {
	var payload eventPayload 
	var event event
	if err := wsjson.Read(context.TODO(), w.conn, &payload); err != nil {

		if isCloseError(err) {
			close(w.disconnectCh)
			return nil, ErrConnClosed
		}

		return nil, err
	}

	switch payload.Type {
	case authEventType:
		event = &authEvent{}
	case heartBeatEventType:
		event = &heartBeatEvent{}
	case challengeRequestType:
		event = &challengeRequest{}
	case challengeAcceptedType:
		event = &challengeAccepted{}
	case quitDuelType:
		event = &quitDuel{}
	case checkResponseType:
		event = &checkResponse{}
	case duelUserDoneType:
		event = &duelUserDone{}
	default:
		return nil, ErrInvalidEventType
	}


	if err := json.Unmarshal(payload.Body, event); err != nil {
		return nil, ErrMalformedPayload
	}

	return event, nil
}

func (w *wsClient) writeError(ctx context.Context, msg string) error {
	return w.writeEvent(ctx, serverSentEvent{
		Type: errEventType,
		Body: &errEvent{
			Message: msg,
		},
	})
}

func (w *wsClient) writeEvent(ctx context.Context, e serverSentEvent) error {

	if err := wsjson.Write(ctx, w.conn, e); err != nil {

		if isCloseError(err) {
			close(w.disconnectCh)
			return ErrConnClosed
		}
		return err
	}
	return nil
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
			if err := wsjson.Write(c.Request().Context(), conn, serverSentEvent{
				Type: errEventType,
				Body: &errEvent{
					Message: err.Error(),
				},
			}); err != nil {
				a.logger.Warnw("could not write to ws connection", "err", err)
			}

			conn.Close(websocket.StatusPolicyViolation, "authentication failed")
			return nil
		}

	}

	a.wsClients.Set(client.user.ID.String(), client)

	client.writeEvent(context.TODO(), serverSentEvent{
		Type: ackAuthEventType,
		Body: nil,
	})

	a.broadcastOnlineStatus(c.Request().Context(), client.user.ID)
	go client.disconnectCheck(a.wsConnCloseCh)
	a.readLoop(client)

	return nil
}



func (a *api) readLoop(client *wsClient) {

	for {
		event, err := client.readEvent()

		if err != nil {

			if err == ErrConnClosed {
				break
			}
			a.logger.Errorw("readEvent() returned an error", "e", err)
		}
		a.eventCh <- eventWrapper{
			client: client,
			event: event,
		}
	}

}

// TODO for this function, errors need to be changed.
// Should provide concrete distinguishable errors
// so the parent function can distinguish them and return user-friendly errors at that level
func (a *api) authenticateWsConn(conn *websocket.Conn) (*wsClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()


	var payload eventPayload
	if err := wsjson.Read(ctx, conn, &payload); err != nil {
		// TODO figure out a better way to do it, errors are nested af
		if strings.Contains(err.Error(), "context deadline exceeded"){
			return nil, ErrNoAuthToken
		}

		return nil, fmt.Errorf("error: malformed json input from client")
	}

	if payload.Type != authEventType {
		return nil, fmt.Errorf("error: client sent a non-auth packet")
	}

	var body authEvent

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

	return &wsClient{
		conn: conn,
		user: user,
		disconnectCh: make(chan struct{}),
	}, nil
}

func (a *api) onDissconnect()  {
	for id := range a.wsConnCloseCh {
		a.wsClients.Del(id)
		// TODO
		validId, _ := uuid.Parse(id)
		a.broadcastOfflineStatus(context.TODO(), validId)
	}
}

func (w *wsClient) disconnectCheck(write chan <- string ) {
	<- w.disconnectCh
	write <- w.user.ID.String()
}


func isCloseError(err error) bool {
	closeStatus := websocket.CloseStatus(err)

	if closeStatus != -1 {
		return true 
	}

	return strings.Contains(err.Error(), "EOF")
}

func (a *api) handleEventLoop() {
	for e := range a.eventCh {
		switch eT := e.event.(type) {
		case *challengeRequest:
			go a.handleChallengeRequest(e.client, *eT)
		case *challengeAccepted:
			go a.handleChallengeAccepted(e.client, *eT)
		case *quitDuel:
			go a.handleQuitDuel(e.client, *eT)
		case *checkResponse:
			go a.handleCheckResponse(e.client, *eT)
		case *duelUserDone:
			go a.handleDuelDone(e.client, *eT)
		}
	}
}

func (a *api) handleQuitDuel(selfClient *wsClient, event quitDuel) {
	duel, ok := a.duels.Get(event.DuelId)

	if !ok {
		a.logger.Errorw("quit request to a non-existing duel match", "id", event.DuelId)
		selfClient.writeError(context.TODO(), "duel not found")
		return
	}

	// ownership check
	if duel.user1 != selfClient.user.ID.String() && duel.user2 != selfClient.user.ID.String() {
		a.logger.Errorw("unauthorized duel access attempt",
		"duel_id", duel.Id,
		"attempted_by", selfClient.user.ID,
		"owner_1", duel.user1,
		"owner_2", duel.user2,
		)
		selfClient.writeError(context.TODO(), "unauthorized duel access")
		return
	}

	duel.exitCh <- struct{}{}

	peerId := duel.user1
	if peerId == selfClient.user.ID.String() {
		peerId = duel.user2
	} 

	peer, ok := a.wsClients.Get(peerId)

	// fine
	if !ok {
		return
	}

	peer.writeEvent(context.TODO(), serverSentEvent{
		Type: duelPeerQuitType,
		Body: nil,
	})
}

func (a *api) handleChallengeRequest(selfClient *wsClient, event challengeRequest) {
	client, ok := a.wsClients.Get(event.ToId)

	if !ok {
		selfClient.writeError(context.TODO(), "challenge sent to an offline user")
		a.logger.Warnw("challenge sent to an offline user", "to", event.ToId, "from", selfClient.user.ID)
		return 
	}

	challengeId, err := uuid.NewUUID()

	if err != nil {
		selfClient.writeError(context.TODO(), "internal server error")
		a.logger.Errorw("could not generate a uuid", "err", err)
		return 
	}

	challenge := challenge{
		ID: challengeId,
		User1: selfClient.user.ID,
		User2: client.user.ID,
	}

	a.challenges.Set(challenge.ID.String(), &challenge)

	if err := client.writeEvent(context.TODO(), serverSentEvent{
		Type: challengeReceivedType,
		Body: &challengeReceived{
			ChallengeId: challenge.ID.String(),
			FromId: selfClient.user.ID.String(),
			ExpiresIsSec: 5,
		},
	}); err != nil {
		a.logger.Errorw("error writing to ws connection", "type", "Challenge Recieved", "error", err)
	}


	// a clean-up for unaccepted challenges
	go func(id string) {
		<-time.After(5 * time.Second)

		_, ok := a.challenges.Get(challengeId.String())
		if ok {
			a.challenges.Del(id)
		}

	}(challengeId.String())
}


func (a *api) handleChallengeAccepted(selfClient *wsClient, event challengeAccepted) {
	ch, ok := a.challenges.Get(event.ChallengeId)

	if !ok {
		selfClient.writeError(context.TODO(), "challenge doesn't exist")
		a.logger.Warnw("challenge not found in the heap", "to", event.ChallengeId, "from", selfClient.user.ID)
		return 
	}

	other := ch.User1

	if other  == selfClient.user.ID {
		other = ch.User2
	}

	// let's make sure the other guy is online
	peer, ok := a.wsClients.Get(other.String())

	if !ok {
		selfClient.writeError(context.TODO(), "peer went offlien")
		return 
	}

	// no longer needed here baby
	a.challenges.Del(ch.ID.String())

	duelId, err := uuid.NewUUID()

	if err != nil {
		selfClient.writeError(context.TODO(), "internal server error")
		a.logger.Errorw("could not generate a uuid", "err", err)
		return 
	}

	// let's create the duel match
	d := duel{
		Id: duelId,
		user1: ch.User1.String(),
		user2: ch.User2.String(),
		mu: sync.Mutex{},
		correctAnswers: map[string]int{
			ch.User1.String():0,
			ch.User2.String():0,
		},
		doneUsers: make(map[string]bool),
		exitCh: make(chan struct{}),
	}

	go func() {
		select {
		case <- time.After(time.Minute * 15):
			a.duels.Del(duelId.String())
		case <- d.exitCh:
			a.duels.Del(duelId.String())
		}
	}()


	ids, err :=  a.services.Questions.GetRandomIds(context.TODO())

	if err != nil {
		selfClient.writeError(context.TODO(), "internal server error")
		a.logger.Errorw("could'nt fetch question ids", "err", err)
		return 
	}	

	a.duels.Set(duelId.String(), &d)

	go peer.writeEvent(context.TODO(), serverSentEvent{
		Type: duelCreatedType,
		Body: &duelCreated{
			QuestionIds: ids,
			Id: duelId.String(),
			PeerId: selfClient.user.ID,
		},
	})
	go selfClient.writeEvent(context.TODO(), serverSentEvent{
		Type: duelCreatedType,
		Body: &duelCreated{
			QuestionIds: ids,
			Id: duelId.String(),
			PeerId: peer.user.ID,
		},
	})
}


func (a *api) handleCheckResponse(selfClient *wsClient, event checkResponse) {
	duel, ok := a.duels.Get(event.DuelId)

	if !ok {
		a.logger.Errorw("check response event to a non-existing duel", "id", event.DuelId)
		selfClient.writeError(context.TODO(), "duel not found")
		return
	}

	// ownership check
	if duel.user1 != selfClient.user.ID.String() && duel.user2 != selfClient.user.ID.String() {
		a.logger.Errorw("unauthorized duel access attempt",
		"duel_id", duel.Id,
		"attempted_by", selfClient.user.ID,
		"owner_1", duel.user1,
		"owner_2", duel.user2,
		)
		selfClient.writeError(context.TODO(), "unauthorized duel access")
		return
	}
	

	questionId, err := uuid.Parse(event.QuestionId)

	if err != nil {
		a.logger.Errorw("handleCheckResponse(): invalid question id", "id", event.QuestionId)
		selfClient.writeError(context.TODO(), "invalid question id")
		return
	}

	isCorrect, err := a.services.Questions.Check(context.TODO(), questions_service.CheckParams{
		Response: event.Reponse,
		QuestionId: questionId,
	})

	if err != nil {

		if err == domain.ErrQuestionNotFound {
			a.logger.Errorw("handleCheckResponse(): question not found", "id", event.QuestionId)
			selfClient.writeError(context.TODO(), "question not found")
			return
		}

		selfClient.writeError(context.TODO(), "internal server error")
		a.logger.Errorw("handleCheckResponse(): server encountered a problem", "err", err)
		return
	}

	if isCorrect {
		duel.incrementScore(selfClient.user.ID.String())
		peerId := getPeerId(selfClient.user.ID.String(), duel.user1, duel.user2)

		peer, ok := a.wsClients.Get(peerId)
		if ok {
			peer.writeEvent(context.TODO(), serverSentEvent{
				Type: responseCheckedType,
				Body: &responseChecked{
					IsCorrect: isCorrect,
					QuestionId: questionId.String(),
					IsPeer: true,
				},
			})
		}
	}


	selfClient.writeEvent(context.TODO(), serverSentEvent{
		Type: responseCheckedType,
		Body: &responseChecked{
			IsCorrect: isCorrect,
			QuestionId: questionId.String(),
			IsPeer: false,
		},
	})

}

func (a *api) handleDuelDone(selfClient *wsClient, event duelUserDone) {
	userId := selfClient.user.ID.String()
	duel, ok := a.duels.Get(event.DuelId)
	if !ok {
		a.logger.Errorw("done request to a non-existing duel", "id", event.DuelId)
		selfClient.writeError(context.TODO(), "duel not found")
		return
	}

	if ok := duel.isUserDone(userId); ok {
		a.logger.Errorw("user sent redundant done event to duel", "id", userId)
		selfClient.writeError(context.TODO(), "redundant done event sent")
		return
	}

	if count := duel.doneUsersLen(); count < 1 {
		duel.setDoneUser(userId)
		peerId := duel.getPeer(userId)
		peer, ok := a.wsClients.Get(peerId)

		if !ok {
			duel.exitCh <- struct{}{}
			selfClient.writeEvent(context.TODO(), serverSentEvent{
				Type: duelPeerQuitType,
				Body: nil,
			})
			return
		}

		peer.writeEvent(context.TODO(), serverSentEvent{
			Type: duelPeerDoneType,
			Body: &duelPeerDone{},
		})
		return
	}

	peerId := duel.getPeer(userId)
	peer, ok := a.wsClients.Get(peerId)

	if !ok {
		duel.exitCh <- struct{}{}
		selfClient.writeEvent(context.TODO(), serverSentEvent{
			Type: duelPeerQuitType,
			Body: nil,
		})
		return
	}

	winnerId := duel.getWinnerId()

	go peer.writeEvent(context.TODO(), serverSentEvent{
		Type: duelFinishedType,
		Body: &duelFinished{
			WinnerId: winnerId,
			IsDraw: winnerId == "",
		},
	})

	selfClient.writeEvent(context.TODO(), serverSentEvent{
		Type: duelFinishedType,
		Body: &duelFinished{
			WinnerId: winnerId,
			IsDraw: winnerId == "",
		},
	})
	

}


func getPeerId(selfId string, user1, user2 string) string {
	if user1 == selfId {
		return user2
	}

	return user1
}
