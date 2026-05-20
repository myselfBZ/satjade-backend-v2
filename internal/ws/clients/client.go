/*
FUCKING HELL I can't figure that what the fuck kind of context to use for my fucking writeLoop
OR my ReadEvent().

note, we'll leave it as TODO() and come back later once i see things more clearly
*/
package clients

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
	"github.com/myselfBZ/satjade-backend/internal/domain"
	"github.com/myselfBZ/satjade-backend/internal/ws/events"
)

var(
	ErrConnClosed = errors.New("err: connection closed")
)

type ClientState struct {
	Type StateType
	// if Type == InMatch
	DuelId string
}

type StateType string

const (
	Idle    StateType = "idle"
	InMatch StateType = "inMatch"
)

type ClientParams struct {
	User   *domain.User
	Conn   *websocket.Conn
	ExitCh chan<- string
}

func NewClient(params ClientParams) *Client {
	c := &Client{

		state: ClientState{
			Type: Idle,
			DuelId: "",
		},

		User:    params.User,
		conn:    params.Conn,
		writeCh: make(chan events.ServerSentEventPayload),
		exitCh:  params.ExitCh,
		mu:      sync.RWMutex{},
	}

	go c.writeLoop()

	return c
}


type Client struct {
	User  *domain.User
	conn  *websocket.Conn
	state ClientState
	writeCh chan events.ServerSentEventPayload
	// ID of the user
	exitCh chan<- string
	mu     sync.RWMutex
}

func (c *Client) writeLoop() {
	shouldClose := false

	defer func() {
		if shouldClose {
			c.conn.Close(websocket.StatusNormalClosure, "bye")
			c.exitCh <- c.User.ID.String()
			log.Println("BYE BYE wrteLoop() is done")
		}
	}()

	for e := range c.writeCh {

		if err := wsjson.Write(context.TODO(), c.conn, e); err != nil {
			if isCloseError(err) {
				// zero value when close(c.writeCh)
				shouldClose = e.Type != ""
				return
			}

			// TODO: else do some logging....
		}
	}
}

func (c *Client) GetState() ClientState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

func (c *Client) WriteState(s ClientState) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if s.Type == Idle && s.DuelId != "" {
		panic("state conflict in ws client, you weepy willow shitsack")
	}

	c.state = s
}

func (c *Client) WriteErrEvent(msg string) {
	c.writeCh <- events.ServerSentEventPayload{
		Type: events.ErrEventType,
		Body: &events.ErrEvent{
			Message: msg,
		},
	}
}

func (c *Client) WriteEvent(event events.ServerSentEventPayload) {
	c.writeCh <- event
}


func (c *Client) ReadEvent() (events.ClientSentEvent, error) {
	var payload events.ClientSentEventPayload
	if err := wsjson.Read(context.TODO(), c.conn, &payload); err != nil {
		if isCloseError(err) {
			return nil, ErrConnClosed
		}
		return nil, err
	}

	var event events.ClientSentEvent

	switch payload.Type {

	case events.ChallengeRequestType:
		event = &events.ChallengeRequest{}

	case events.AcceptChallengeType:
		event = &events.AcceptChallenge{}

	case events.QuitDuelType:
		event = &events.QuitDuel{}

	case events.CheckResponseType:
		event = &events.CheckResponse{}

	case events.DuelUserDoneType:
		event = &events.DuelUserDone{}
	}


	if err := json.Unmarshal(payload.Body, event); err != nil {
		return nil, fmt.Errorf("malformed event body")
	}

	return event, nil
}

func isCloseError(err error) bool {
	closeStatus := websocket.CloseStatus(err)

	if closeStatus != -1 {
		return true
	}

	return strings.Contains(err.Error(), "EOF") || strings.Contains(err.Error(), "use of closed network connection")
}
