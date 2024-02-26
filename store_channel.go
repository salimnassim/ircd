package ircd

import "sync"

type ChannelStorer interface {
	// Number of channels in store.
	count() int
	// add channel to store. ID is most likely channel name.
	add(name string, ch channeler)
	delete(name string)
	// Check if client is a member of channel.
	isMember(c clienter, ch channeler) (ok bool)
	// get channel by name.
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
	channels := 0
	s.mu.RLock()
	channels = len(s.channels)
	s.mu.RUnlock()

	return channels
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

func (s *channelStore) isMember(c clienter, ch channeler) bool {
	return ch.clients().isMember(c.id())
}

func (s *channelStore) memberOf(c clienter) []channeler {
	channels := []channeler{}

	s.mu.RLock()
	for _, ch := range s.channels {
		if ch.clients().isMember(c.id()) {
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
