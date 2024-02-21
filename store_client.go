package ircd

import "sync"

type clientID string

type ClientStorer interface {
	// Number of clients in store.
	count() int
	// add client to store.
	add(id clientID, c *client)
	// Remove client from store.
	delete(id clientID)
	// get client from store by nickname.
	get(nickname string) (c *client, ok bool)
}

type clientStore struct {
	mu      *sync.RWMutex
	id      string
	clients map[clientID]*client
}

func newClientStore(id string) *clientStore {
	return &clientStore{
		mu:      &sync.RWMutex{},
		id:      id,
		clients: make(map[clientID]*client),
	}
}

// Get number of clients in store.
func (s *clientStore) count() int {
	clients := 0

	s.mu.RLock()
	clients = len(s.clients)
	s.mu.RUnlock()

	return clients
}

// get client from store by nickname.
func (s *clientStore) get(nickname string) (*client, bool) {
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

// add client to store.
func (s *clientStore) add(id clientID, c *client) {
	s.mu.Lock()
	s.clients[id] = c
	s.mu.Unlock()
}

// delete client from store.
func (s *clientStore) delete(id clientID) {
	s.mu.Lock()
	delete(s.clients, id)
	s.mu.Unlock()
}
