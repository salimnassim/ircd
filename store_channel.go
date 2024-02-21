package ircd

import "sync"

type ChannelStorer interface {
	// Number of channels in store.
	Count() int
	// Add channel to store. ID is most likely channel name.
	Add(name string, channel *Channel)
	Delete(name string)
	// Check if client is a member of channel.
	IsMember(client *Client, channel *Channel) (ok bool)
	// Get channel by name.
	Get(name string) (channel *Channel, ok bool)
	// Get which channels a client belongs to.
	MemberOf(client *Client) (channels []*Channel)
}

type ChannelStore struct {
	mu       *sync.RWMutex
	id       string
	channels map[string]*Channel
}

func NewChannelStore(id string) *ChannelStore {
	return &ChannelStore{
		mu:       &sync.RWMutex{},
		id:       id,
		channels: make(map[string]*Channel),
	}
}

func (s *ChannelStore) Count() int {
	channels := 0
	s.mu.RLock()
	channels = len(s.channels)
	s.mu.RUnlock()

	return channels
}

func (s *ChannelStore) Get(name string) (*Channel, bool) {
	s.mu.RLock()
	channel, ok := s.channels[name]
	s.mu.RUnlock()
	if !ok {
		return nil, false
	}
	return channel, true
}

func (s *ChannelStore) Add(name string, channel *Channel) {
	s.mu.Lock()
	s.channels[name] = channel
	s.mu.Unlock()
}

func (s *ChannelStore) Delete(name string) {
	s.mu.Lock()
	delete(s.channels, name)
	s.mu.Unlock()
}

func (s *ChannelStore) IsMember(client *Client, channel *Channel) bool {
	return channel.clients.IsMember(client.id)
}

func (s *ChannelStore) MemberOf(client *Client) []*Channel {
	channels := []*Channel{}

	s.mu.RLock()
	for _, c := range s.channels {
		if c.clients.IsMember(ClientID(client.id)) {
			channels = append(channels, c)
		}
	}
	s.mu.RUnlock()

	return channels
}
