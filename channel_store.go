package ircd

import "sync"

type ChannelStoreable interface {
	Size() int
	Add(id string, channel *Channel)
	GetByName(string) (*Channel, bool)
	MemberOf(*Client) []*Channel
}

type ChannelStore struct {
	mu       *sync.RWMutex
	id       string
	channels map[string]*Channel
}

func NewChannelStore(id string) *ChannelStore {
	return &ChannelStore{
		id:       id,
		mu:       &sync.RWMutex{},
		channels: make(map[string]*Channel),
	}
}

func (cs *ChannelStore) Size() int {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	size := len(cs.channels)
	return size
}

// Adds channel to store
func (cs *ChannelStore) Add(id string, channel *Channel) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.channels[id] = channel
}

// Returns a pointer to the channel by name.
// Boolean returns true if thee channel exists
func (cs *ChannelStore) GetByName(channelName string) (*Channel, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	channel, exists := cs.channels[channelName]
	if !exists {
		return nil, false
	}
	return channel, true
}

func (cs *ChannelStore) MemberOf(client *Client) []*Channel {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	var memberOf []*Channel

	for _, v := range cs.channels {
		for _, c := range v.clients {
			if c.id == client.id {
				memberOf = append(memberOf, v)
			}
		}
	}

	return memberOf
}
