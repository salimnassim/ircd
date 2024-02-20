package ircd

import "sync"

type ChannelStoreable interface {
	// Number of channels in store.
	Size() int
	// Add channel to store. ID is most likely channel name.
	Add(ID string, channel *Channel)
	// Check if client is a member of channel.
	IsMember(client *Client, channel *Channel) (ok bool)
	// Get channel by name.
	GetByName(name string) (channel *Channel, ok bool)
	// Get which channels a client belongs to.
	MemberOf(client *Client) (channels []*Channel)
}

type ChannelStore struct {
	id       string
	channels sync.Map
}

func NewChannelStore(id string) *ChannelStore {
	return &ChannelStore{
		id:       id,
		channels: sync.Map{},
	}
}

func (cs *ChannelStore) Size() int {
	size := 0

	cs.channels.Range(func(key, value any) bool {
		size++
		return true
	})

	return size
}

// Adds channel to store
func (cs *ChannelStore) Add(id string, channel *Channel) {
	cs.channels.Store(id, channel)
}

// Returns a pointer to the channel by name.
// Boolean returns true if thee channel exists
func (cs *ChannelStore) GetByName(channelName string) (*Channel, bool) {
	var channel *Channel

	cs.channels.Range(func(key, value any) bool {
		if value.(*Channel).name == channelName {
			channel = value.(*Channel)
			return false
		}
		return true
	})

	if channel == nil {
		return nil, false
	}

	return channel, true
}

func (cs *ChannelStore) IsMember(client *Client, channel *Channel) bool {
	_, ok := channel.clients.Load(client.id)
	return ok
}

func (cs *ChannelStore) MemberOf(client *Client) []*Channel {
	var memberOf []*Channel

	cs.channels.Range(func(key, value any) bool {
		_, exists := value.(*Channel).clients.Load(client.id)
		if exists {
			memberOf = append(memberOf, value.(*Channel))
		}
		return true
	})

	return memberOf
}
