package ircd

import "sync"

type clientID string

type ClientStorer interface {
	// Number of clients in store.
	count() (visible int, invisible int)
	// add client to store.
	add(c clienter)
	// Remove client from store.
	delete(id clientID)
	// Get client from store by nickname.
	get(nickname string) (c clienter, exists bool)
}

type clientStore struct {
	mu      *sync.RWMutex
	id      string
	clients map[clientID]clienter
}

func NewClientStore(id string) *clientStore {
	return &clientStore{
		mu:      &sync.RWMutex{},
		id:      id,
		clients: make(map[clientID]clienter),
	}
}

// Get number of clients in store.
func (s *clientStore) count() (visible int, invisible int) {
	s.mu.RLock()
	for _, c := range s.clients {
		if c.hasMode(modeClientInvisible) {
			invisible++
		} else {
			visible++
		}
	}
	s.mu.RUnlock()

	return visible, invisible
}

// get client from store by nickname.
func (s *clientStore) get(nickname string) (clienter, bool) {
	var client clienter

	s.mu.RLock()
	for _, c := range s.clients {
		if c.nickname() == nickname {
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
func (s *clientStore) add(c clienter) {
	s.mu.Lock()
	s.clients[c.id()] = c
	s.mu.Unlock()
}

// delete client from store.
func (s *clientStore) delete(id clientID) {
	s.mu.Lock()
	delete(s.clients, id)
	s.mu.Unlock()
}
