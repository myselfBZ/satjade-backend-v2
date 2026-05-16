package clients

import (
	"errors"
	"sync"
)

var(
	ErrClientNotFound = errors.New("client connection does not exist")
)

type State string

const (
	Idle State = "idle"
	InMatch State = "inMatch"
)


type Manager interface {
	Get(id string) (*Client, bool)
	Set(c *Client)
}

type manager struct {
	clients map[string]*Client
	mu 		sync.Mutex
}



func (m *manager) Get(id string) (*Client, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.clients[id]

	return c, ok
} 


func (m *manager) Set(c *Client) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.clients[c.User.ID.String()] = c
}
