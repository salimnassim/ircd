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
	clients  sync.Map
	owner    string
	password string
}

type channelTopic struct {
	text      string
	timestamp int
	author    string
}

func NewChannel(channelName string, owner string) *Channel {
	channel := &Channel{
		mu:   &sync.RWMutex{},
		name: channelName,
		topic: channelTopic{
			text:      "",
			timestamp: 0,
			author:    "",
		},
		clients:  sync.Map{},
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

	ch.clients.Store(client.id, client)

	return nil
}

// Remove client from channel.
func (ch *Channel) RemoveClient(client *Client) {
	ch.clients.Delete(client.id)
}

// Returns channel users delimited by a space for RPL_NAMREPLY.
func (ch *Channel) Names() []string {
	var names []string

	ch.clients.Range(func(key, value any) bool {
		client := value.(*Client)
		if ch.owner == client.id {
			names = append(names, fmt.Sprintf("@%s", client.Nickname()))
		} else {
			names = append(names, client.Nickname())
		}
		return true
	})

	return names
}

func (ch *Channel) Who() []string {
	var who []string

	ch.clients.Range(func(key, value any) bool {
		client := value.(*Client)
		who = append(who, fmt.Sprintf("%s %s %s %s %s %s :%s %s",
			ch.name, client.username, client.hostname, "ircd", client.Nickname(), "H", "0", client.realname))
		return true
	})

	return who
}

// Send message to all clients on the channel.
// If skip is true, the client in source will not receive the message
func (ch *Channel) Broadcast(message string, sourceId string, skip bool) {
	ch.clients.Range(func(key, value any) bool {
		client := value.(*Client)
		if sourceId != key {
			client.send <- message
		}
		if sourceId == key && !skip {
			client.send <- message
		}
		return true
	})
}
