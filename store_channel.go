package ircd

import "sync"

type ChannelStorer interface {
	// Number of channels in store.
	count() int
	// add channel to store. ID is most likely channel name.
	add(name string, ch *channel)
	delete(name string)
	// Check if client is a member of channel.
	isMember(c *client, ch *channel) (ok bool)
	// get channel by name.
	get(name string) (ch *channel, ok bool)
	// Get which channels a client belongs to.
	memberOf(c *client) (chs []*channel)
}

type channelStore struct {
	mu       *sync.RWMutex
	id       string
	channels map[string]*channel
}

func newChannelStore(id string) *channelStore {
	return &channelStore{
		mu:       &sync.RWMutex{},
		id:       id,
		channels: make(map[string]*channel),
	}
}

func (s *channelStore) count() int {
	channels := 0
	s.mu.RLock()
	channels = len(s.channels)
	s.mu.RUnlock()

	return channels
}

func (s *channelStore) get(name string) (*channel, bool) {
	s.mu.RLock()
	channel, ok := s.channels[name]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return channel, true
}

func (s *channelStore) add(name string, ch *channel) {
	s.mu.Lock()
	s.channels[name] = ch
	s.mu.Unlock()
}

func (s *channelStore) delete(name string) {
	s.mu.Lock()
	delete(s.channels, name)
	s.mu.Unlock()
}

func (s *channelStore) isMember(c *client, ch *channel) bool {
	return ch.clients.isMember(c.id)
}

func (s *channelStore) memberOf(c *client) []*channel {
	channels := []*channel{}

	s.mu.RLock()
	for _, ch := range s.channels {
		if ch.clients.isMember(clientID(c.id)) {
			channels = append(channels, ch)
		}
	}
	s.mu.RUnlock()

	return channels
}
