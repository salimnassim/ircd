package ircd

import "sync"

type ChannelStorer interface {
	// Number of channels in store.
	Count() int
	// Add channel to store. ID is most likely channel name.
	Add(name string, ch *channel)
	Delete(name string)
	// Check if client is a member of channel.
	IsMember(c *client, ch *channel) (ok bool)
	// Get channel by name.
	Get(name string) (ch *channel, ok bool)
	// Get which channels a client belongs to.
	MemberOf(c *client) (chs []*channel)
}

type ChannelStore struct {
	mu       *sync.RWMutex
	id       string
	channels map[string]*channel
}

func NewChannelStore(id string) *ChannelStore {
	return &ChannelStore{
		mu:       &sync.RWMutex{},
		id:       id,
		channels: make(map[string]*channel),
	}
}

func (s *ChannelStore) Count() int {
	channels := 0
	s.mu.RLock()
	channels = len(s.channels)
	s.mu.RUnlock()

	return channels
}

func (s *ChannelStore) Get(name string) (*channel, bool) {
	s.mu.RLock()
	channel, ok := s.channels[name]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return channel, true
}

func (s *ChannelStore) Add(name string, ch *channel) {
	s.mu.Lock()
	s.channels[name] = ch
	s.mu.Unlock()
}

func (s *ChannelStore) Delete(name string) {
	s.mu.Lock()
	delete(s.channels, name)
	s.mu.Unlock()
}

func (s *ChannelStore) IsMember(c *client, ch *channel) bool {
	return ch.clients.IsMember(c.id)
}

func (s *ChannelStore) MemberOf(c *client) []*channel {
	channels := []*channel{}

	s.mu.RLock()
	for _, ch := range s.channels {
		if ch.clients.IsMember(clientID(c.id)) {
			channels = append(channels, ch)
		}
	}
	s.mu.RUnlock()

	return channels
}
