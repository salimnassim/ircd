package ircd

import (
	"sync"
)

type ChannelClientStorer interface {
	count() int
	add(ID clientID, c clienter)
	delete(ID clientID)
	all() []clienter
	isMember(ID clientID) bool
}

type channelClientStore struct {
	mu      *sync.RWMutex
	clients map[clientID]clienter
}

func newChannelClientStore() *channelClientStore {
	return &channelClientStore{
		mu:      &sync.RWMutex{},
		clients: make(map[clientID]clienter),
	}
}

func (s *channelClientStore) count() int {
	clients := 0
	s.mu.RLock()
	clients = len(s.clients)
	s.mu.RUnlock()
	return clients
}

func (s *channelClientStore) add(ID clientID, c clienter) {
	s.mu.Lock()
	s.clients[ID] = c
	s.mu.Unlock()
}

func (s *channelClientStore) delete(ID clientID) {
	s.mu.Lock()
	delete(s.clients, ID)
	s.mu.Unlock()
}

func (s *channelClientStore) all() []clienter {
	clients := []clienter{}

	s.mu.RLock()
	for _, c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.RUnlock()

	return clients
}

func (s *channelClientStore) isMember(ID clientID) bool {
	s.mu.RLock()
	_, ok := s.clients[ID]
	s.mu.RUnlock()
	return ok
}
