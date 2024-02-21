package ircd

import "sync"

type ClientID string

type ClientStorer interface {
	// Number of clients in store.
	Count() int
	// Add client to store.
	Add(id ClientID, client *Client)
	// Remove client from store.
	Delete(id ClientID)
	// Get client from store by nickname.
	Get(nickname string) (client *Client, ok bool)
}

type ClientStore struct {
	mu      *sync.RWMutex
	id      string
	clients map[ClientID]*Client
}

func NewClientStore(id string) *ClientStore {
	return &ClientStore{
		mu:      &sync.RWMutex{},
		id:      id,
		clients: make(map[ClientID]*Client),
	}
}

// Get number of clients in store.
func (s *ClientStore) Count() int {
	clients := 0

	s.mu.RLock()
	clients = len(s.clients)
	s.mu.RUnlock()

	return clients
}

// Get client from store by nickname.
func (s *ClientStore) Get(nickname string) (*Client, bool) {
	var client *Client

	s.mu.RLock()
	for _, c := range s.clients {
		if c.nickname == nickname {
			client = c
			break
		}
	}
	s.mu.RUnlock()

	if client == nil {
		return nil, false
	}

	return client, true
}

// Add client to store.
func (s *ClientStore) Add(id ClientID, client *Client) {
	s.mu.Lock()
	s.clients[id] = client
	s.mu.Unlock()
}

// Delete client from store.
func (s *ClientStore) Delete(id ClientID) {
	s.mu.Lock()
	delete(s.clients, id)
	s.mu.Unlock()
}
