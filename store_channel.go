package ircd

import "sync"

type ChannelStorer interface {
	// Number of channels in store.
	count() int
	// Add channel to store.
	add(name string, ch channeler)
	// Delete channel.
	delete(name string)
	// Get channel by name.
	get(name string) (ch channeler, exists bool)
	// Get which channels a client belongs to.
	memberOf(c clienter) (chs []channeler)
	// Get all channels.
	all() []channeler
}

type channelStore struct {
	mu       *sync.RWMutex
	id       string
	channels map[string]channeler
}

func NewChannelStore(id string) *channelStore {
	return &channelStore{
		mu:       &sync.RWMutex{},
		id:       id,
		channels: make(map[string]channeler),
	}
}

func (s *channelStore) count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.channels)
}

func (s *channelStore) get(name string) (channeler, bool) {
	s.mu.RLock()
	channel, exists := s.channels[name]
	s.mu.RUnlock()
	if !exists {
		return nil, false
	}
	return channel, true
}

func (s *channelStore) add(name string, ch channeler) {
	s.mu.Lock()
	s.channels[name] = ch
	s.mu.Unlock()
}

func (s *channelStore) delete(name string) {
	s.mu.Lock()
	delete(s.channels, name)
	s.mu.Unlock()
}

func (s *channelStore) memberOf(c clienter) []channeler {
	channels := []channeler{}

	s.mu.RLock()
	for _, ch := range s.channels {
		if ch.clients().isMember(c) {
			channels = append(channels, ch)
		}
	}
	s.mu.RUnlock()

	return channels
}

func (s *channelStore) all() []channeler {
	channels := []channeler{}

	s.mu.RLock()
	for _, ch := range s.channels {
		channels = append(channels, ch)
	}
	s.mu.RUnlock()

	return channels
}
