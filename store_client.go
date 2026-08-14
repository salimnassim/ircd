package ircd

import "sync"

type clientID string

type ClientStorer interface {
	count() (visible int, invisible int)

	add(c clienter)

	delete(id clientID)

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

func (s *clientStore) add(c clienter) {
	s.mu.Lock()
	s.clients[c.id()] = c
	s.mu.Unlock()
}

func (s *clientStore) delete(id clientID) {
	s.mu.Lock()
	delete(s.clients, id)
	s.mu.Unlock()
}
