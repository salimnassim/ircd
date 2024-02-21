package ircd

import (
	"sync"
)

type ChannelClientStorer interface {
	Count() int
	Add(ID ClientID, client *Client)
	Delete(ID ClientID)
	All() []*Client
	IsMember(ID ClientID) bool
}

type ChannelClientStore struct {
	mu      *sync.RWMutex
	clients map[ClientID]*Client
}

func NewChannelClientStore() *ChannelClientStore {
	return &ChannelClientStore{
		mu:      &sync.RWMutex{},
		clients: make(map[ClientID]*Client),
	}
}

func (s *ChannelClientStore) Count() int {
	clients := 0
	s.mu.RLock()
	clients = len(s.clients)
	s.mu.RUnlock()
	return clients
}

func (s *ChannelClientStore) Add(ID ClientID, client *Client) {
	s.mu.Lock()
	s.clients[ID] = client
	s.mu.Unlock()
}

func (s *ChannelClientStore) Delete(ID ClientID) {
	s.mu.Lock()
	delete(s.clients, ID)
	s.mu.Unlock()
}

func (s *ChannelClientStore) All() []*Client {
	clients := []*Client{}

	s.mu.RLock()
	for _, c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.RUnlock()

	return clients
}

func (s *ChannelClientStore) IsMember(ID ClientID) bool {
	s.mu.RLock()
	_, ok := s.clients[ID]
	s.mu.RUnlock()
	return ok
}
