package ircd

import "sync"

type clientID string

type ClientStorer interface {
	// Number of clients in store.
	Count() int
	// Add client to store.
	Add(id clientID, c *client)
	// Remove client from store.
	Delete(id clientID)
	// Get client from store by nickname.
	Get(nickname string) (c *client, ok bool)
}

type ClientStore struct {
	mu      *sync.RWMutex
	id      string
	clients map[clientID]*client
}

func NewClientStore(id string) *ClientStore {
	return &ClientStore{
		mu:      &sync.RWMutex{},
		id:      id,
		clients: make(map[clientID]*client),
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
func (s *ClientStore) Get(nickname string) (*client, bool) {
	var client *client

	s.mu.RLock()
	for _, c := range s.clients {
		if c.nick == nickname {
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
func (s *ClientStore) Add(id clientID, c *client) {
	s.mu.Lock()
	s.clients[id] = c
	s.mu.Unlock()
}

// Delete client from store.
func (s *ClientStore) Delete(id clientID) {
	s.mu.Lock()
	delete(s.clients, id)
	s.mu.Unlock()
}
