package ircd

import (
	"sync"
)

type ChannelClientStorer interface {
	Count() int
	Add(ID clientID, c *client)
	Delete(ID clientID)
	All() []*client
	IsMember(ID clientID) bool
}

type ChannelClientStore struct {
	mu      *sync.RWMutex
	clients map[clientID]*client
}

func NewChannelClientStore() *ChannelClientStore {
	return &ChannelClientStore{
		mu:      &sync.RWMutex{},
		clients: make(map[clientID]*client),
	}
}

func (s *ChannelClientStore) Count() int {
	clients := 0
	s.mu.RLock()
	clients = len(s.clients)
	s.mu.RUnlock()
	return clients
}

func (s *ChannelClientStore) Add(ID clientID, c *client) {
	s.mu.Lock()
	s.clients[ID] = c
	s.mu.Unlock()
}

func (s *ChannelClientStore) Delete(ID clientID) {
	s.mu.Lock()
	delete(s.clients, ID)
	s.mu.Unlock()
}

func (s *ChannelClientStore) All() []*client {
	clients := []*client{}

	s.mu.RLock()
	for _, c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.RUnlock()

	return clients
}

func (s *ChannelClientStore) IsMember(ID clientID) bool {
	s.mu.RLock()
	_, ok := s.clients[ID]
	s.mu.RUnlock()
	return ok
}
