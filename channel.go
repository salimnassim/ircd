package ircd

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

type Channel struct {
	mu       *sync.RWMutex
	name     string
	topic    *ChannelTopic
	clients  map[*Client]bool
	password string
}

type ChannelTopic struct {
	text      string
	timestamp int
	author    string
}

func NewChannel(name string) *Channel {
	channel := &Channel{
		mu:   &sync.RWMutex{},
		name: name,
		topic: &ChannelTopic{
			text:      "",
			timestamp: 0,
			author:    "",
		},

		clients:  make(map[*Client]bool),
		password: "",
	}
	return channel
}

func (channel *Channel) SetTopic(topic string, author string) {
	channel.mu.Lock()
	defer channel.mu.Unlock()
	channel.topic.text = topic
	channel.topic.timestamp = int(time.Now().Unix())
	channel.topic.author = author
}

func (channel *Channel) Topic() ChannelTopic {
	channel.mu.RLock()
	defer channel.mu.RUnlock()
	return *channel.topic
}

func (channel *Channel) AddClient(client *Client, password string) error {
	if password != "" && channel.password != password {
		return errors.New("incorrect password")
	}

	channel.mu.Lock()
	defer channel.mu.Unlock()

	channel.clients[client] = true

	return nil
}

func (ch *Channel) RemoveClient(client *Client) error {
	if !ch.clients[client] {
		return errors.New("not a channel member")
	}
	ch.mu.Lock()
	defer ch.mu.Unlock()
	delete(ch.clients, client)
	return nil
}

// Returns channel users delimited by a space for RPL_NAMREPLY
func (ch *Channel) Names() string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	var names string
	clients := ch.clients
	for client := range clients {
		// todo: add modes
		names = names + fmt.Sprintf("%s ", client.nickname)
	}

	return names
}

// Send message to all clients on the channel.
// If skip is true, the client in source will not receive the message
func (ch *Channel) Broadcast(message string, source *Client, skip bool) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	for c, alive := range ch.clients {
		if !alive {
			continue
		}
		if skip && c == source {
			continue
		}
		c.send <- message
	}
}
