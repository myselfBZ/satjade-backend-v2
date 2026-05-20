package clients

import (
	"errors"
	"sync"
)

var(
	ErrClientNotFound = errors.New("client connection does not exist")
)

func NewManager() Manager {
	return &manager{
		clients: make(map[string]*Client),
		mu: sync.Mutex{},
	}
}


type Manager interface {
	Get(id string) (*Client, bool)
	Set(c *Client)
	Del(id string)
}

type manager struct {
	clients map[string]*Client
	mu 		sync.Mutex
}

func (m *manager) Del(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	client, ok := m.clients[id]
	if ok {
		close(client.writeCh)
		delete(m.clients, id)
	}

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
