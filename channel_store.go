package ircd

import "sync"

type ChannelStoreable interface {
}

type ChannelStore struct {
	mu       *sync.RWMutex
	id       string
	channels map[*Channel]bool
}

func NewChannelStore(id string) *ChannelStore {
	return &ChannelStore{
		id:       id,
		mu:       &sync.RWMutex{},
		channels: map[*Channel]bool{},
	}
}

// Adds channel to store
func (cs *ChannelStore) Add(channel *Channel) {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.channels[channel] = true
}

// Returns a pointer to the channel by name.
// Boolean returns true if thee channel exists
// func (cs *ChannelStore) GetByName(channelName string) (*Channel, bool) {
// 	return
// }
