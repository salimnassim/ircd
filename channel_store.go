package ircd

import "sync"

type ChannelStoreable interface {
	Size() int
	Add(id string, channel *Channel)
	IsMember(*Client, *Channel) bool
	GetByName(string) (*Channel, bool)
	MemberOf(*Client) []*Channel
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
