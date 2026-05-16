package clients

import (
	"github.com/coder/websocket"
	"github.com/myselfBZ/satjade-backend/internal/domain"
)


type Client struct {
	User  domain.User
	conn  *websocket.Conn
	State State
	// if State == InMatch
	DuelId string
}


func (c *Client) WriteEvent() {}

func (c *Client) ReadEvent() {}



