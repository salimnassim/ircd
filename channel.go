package ircd

import (
	"errors"
	"sync"
	"time"
)

type Channel struct {
	mu       *sync.RWMutex
	Name     string
	topic    *ChannelTopic
	clients  map[*Client]bool
	password string
}

type ChannelTopic struct {
	Topic     string
	Timestamp int
	Author    string
}

func NewChannel(name string) *Channel {
	channel := &Channel{
		Name: name,
		topic: &ChannelTopic{
			Topic:     "",
			Timestamp: 0,
			Author:    "",
		},
		mu:       &sync.RWMutex{},
		clients:  make(map[*Client]bool),
		password: "",
	}
	return channel
}

func (channel *Channel) SetTopic(topic string, author string) {
	channel.mu.Lock()
	defer channel.mu.Unlock()
	channel.topic.Topic = topic
	channel.topic.Timestamp = int(time.Now().Unix())
	channel.topic.Author = author
}

func (channel *Channel) Topic() *ChannelTopic {
	return channel.topic
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

// Send message to all clients on the channel.
// If skip is true, the client in source will not receive the message
func (ch *Channel) Broadcast(message string, source *Client, skip bool) {
	for c := range ch.clients {
		if skip && c.Nickname == source.Nickname {
			continue
		}
		c.Out <- message
	}
}
