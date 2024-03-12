package ircd

import (
	"cmp"
	"fmt"
	"slices"
	"sync"
)

type channelClientStorer interface {
	// Number of clients on the channel.
	count() int
	// Add client to channel.
	add(c clienter)
	// Remove client from channel.
	remove(c clienter)
	// Get all channel clients.
	all() []clienter
	// Is client member of the channel?
	isMember(c clienter) bool
	// Add channel membership mode to client.
	addMode(c clienter, m channelMembershipMode)
	// Remove channel membership mode from client.
	removeMode(c clienter, m channelMembershipMode)
	// Does client have any of the roles in m?
	//
	// Can be used to match against multiple membership modes.
	hasMode(c clienter, m ...channelMembershipMode) bool
	// Channel membership modestring for client.
	modestring(c clienter) string
}

type channelClientStore struct {
	mu      *sync.RWMutex
	clients map[clienter]channelMembershipMode
}

func newChannelClientStore() *channelClientStore {
	return &channelClientStore{
		mu:      &sync.RWMutex{},
		clients: make(map[clienter]channelMembershipMode),
	}
}

func (s *channelClientStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

func (s *channelClientStore) add(c clienter) {
	s.mu.Lock()
	s.clients[c] = 0
	s.mu.Unlock()
}

func (s *channelClientStore) remove(c clienter) {
	s.mu.Lock()
	delete(s.clients, c)
	s.mu.Unlock()
}

func (s *channelClientStore) all() []clienter {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]clienter, len(s.clients))
	i := 0
	for k := range s.clients {
		clients[i] = k
		i++
	}

	return clients
}

func (s *channelClientStore) isMember(c clienter) bool {
	s.mu.RLock()
	_, ok := s.clients[c]
	s.mu.RUnlock()
	return ok
}

func (s *channelClientStore) addMode(c clienter, mode channelMembershipMode) {
	if s.hasMode(c, mode) {
		return
	}

	s.mu.RLock()
	current, ok := s.clients[c]
	s.mu.RUnlock()
	if !ok {
		return
	}

	current |= mode

	s.mu.Lock()
	s.clients[c] = current
	s.mu.Unlock()
}

func (s *channelClientStore) removeMode(c clienter, mode channelMembershipMode) {
	if !s.hasMode(c, mode) {
		return
	}

	s.mu.RLock()
	current, ok := s.clients[c]
	s.mu.RUnlock()
	if !ok {
		return
	}

	s.mu.Lock()
	current |= mode
	s.clients[c] = current
	s.mu.Unlock()
}

func (s *channelClientStore) hasMode(c clienter, modes ...channelMembershipMode) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	current, ok := s.clients[c]
	if !ok {
		return false
	}

	has := false
	for _, mode := range modes {
		if current&mode != 0 {
			has = true
			break
		}
	}

	return has
}

func (s *channelClientStore) modestring(c clienter) string {
	modes := []rune{}
	for m, r := range channelMembershipModeMap {
		if s.hasMode(c, r) {
			modes = append(modes, m)
		}
	}
	slices.SortFunc[[]rune, rune](modes, func(a rune, b rune) int {
		return cmp.Compare(a, b)
	})
	return fmt.Sprintf("+%s", string(modes))
}
