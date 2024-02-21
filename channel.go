package ircd

import (
	"fmt"
	"sync"
	"time"
)

type Channel struct {
	mu       *sync.RWMutex
	name     string
	topic    channelTopic
	clients  ChannelClientStorer
	owner    ClientID
	password string
}

type channelTopic struct {
	text      string
	timestamp int
	author    string
}

func NewChannel(channelName string, owner ClientID) *Channel {
	channel := &Channel{
		mu:   &sync.RWMutex{},
		name: channelName,
		topic: channelTopic{
			text:      "",
			timestamp: 0,
			author:    "",
		},
		clients:  NewChannelClientStore(),
		owner:    owner,
		password: "",
	}

	return channel
}

// Sets channel topic.
func (channel *Channel) SetTopic(topic string, author string) {
	channel.mu.Lock()
	channel.topic.text = topic
	channel.topic.timestamp = int(time.Now().Unix())
	channel.topic.author = author
	channel.mu.Unlock()
}

// Returns current topic.
func (channel *Channel) Topic() channelTopic {
	channel.mu.RLock()
	defer channel.mu.RUnlock()
	return channel.topic
}

// Adds client to channel. If password does not match, an error is returned.
func (ch *Channel) AddClient(client *Client, password string) error {
	if password != "" && ch.password != password {
		return errorBadChannelKey
	}

	ch.clients.Add(ClientID(client.id), client)

	return nil
}

// Remove client from channel.
func (ch *Channel) RemoveClient(client *Client) {
	ch.clients.Delete(ClientID(client.id))
}

// Returns channel users delimited by a space for RPL_NAMREPLY.
func (ch *Channel) Names() []string {
	var names []string

	for _, c := range ch.clients.All() {
		if ch.owner == c.id {
			names = append(names, fmt.Sprintf("@%s", c.Nickname()))
		} else {
			names = append(names, c.Nickname())
		}
	}

	return names
}

func (ch *Channel) Who() []string {
	var who []string

	for _, c := range ch.clients.All() {
		who = append(who, fmt.Sprintf("%s %s %s %s %s %s :%s %s",
			ch.name, c.username, c.hostname, "ircd", c.Nickname(), "H", "0", c.realname))
	}
	return who
}

// Send message to all clients on the channel.
// If skip is true, the client in source will not receive the message.
func (ch *Channel) Broadcast(message string, sourceId ClientID, skip bool) {
	for _, c := range ch.clients.All() {
		if c.id == sourceId && skip {
			continue
		}
		c.send <- message
	}
}
