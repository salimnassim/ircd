package ircd

import "sync"

type ChannelStoreable interface {
	Add(id string, channel *Channel)
	GetByName(string) (*Channel, bool)
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

// Adds channel to store
func (cs *ChannelStore) Add(id string, channel *Channel) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.channels[id] = channel
}

// Returns a pointer to the channel by name.
// Boolean returns true if thee channel exists
func (cs *ChannelStore) GetByName(channelName string) (*Channel, bool) {
	channel, exists := cs.channels[channelName]
	if !exists {
		return nil, false
	}
	return channel, true
}
