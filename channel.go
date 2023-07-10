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
	clients  map[string]*Client
	owner    *Client
	password string
}

type ChannelTopic struct {
	text      string
	timestamp int
	author    string
}

func NewChannel(channelName string, owner *Client) *Channel {
	channel := &Channel{
		mu:   &sync.RWMutex{},
		name: channelName,
		topic: &ChannelTopic{
			text:      "",
			timestamp: 0,
			author:    "",
		},
		clients:  map[string]*Client{},
		owner:    owner,
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
	channel.mu.Lock()
	defer channel.mu.Unlock()

	if password != "" && channel.password != password {
		return errors.New("incorrect password")
	}

	channel.clients[client.id] = client

	return nil
}

func (ch *Channel) RemoveClient(client *Client) error {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	_, ok := ch.clients[client.id]
	if !ok {
		return errors.New("not a channel member")
	}

	delete(ch.clients, client.id)
	return nil
}

// Returns channel users delimited by a space for RPL_NAMREPLY
func (ch *Channel) Names() string {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	var names string
	clients := ch.clients
	for _, c := range clients {
		// todo: add modes
		names = names + fmt.Sprintf("%s ", c.nickname)
	}

	return names
}

// Send message to all clients on the channel.
// If skip is true, the client in source will not receive the message
func (ch *Channel) Broadcast(message string, source *Client, skip bool) {
	ch.mu.RLock()
	defer ch.mu.RUnlock()

	for id, c := range ch.clients {
		if skip && id == source.id {
			continue
		}
		c.send <- message
	}
}
