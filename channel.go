package ircd

import (
	"errors"
	"sync"
)

type Channel struct {
	Name     string
	mutex    *sync.Mutex
	clients  map[*Client]bool
	password string
}

func NewChannel(name string) *Channel {
	channel := &Channel{
		Name:     name,
		mutex:    &sync.Mutex{},
		clients:  make(map[*Client]bool),
		password: "",
	}
	return channel
}

func (channel *Channel) Join(client *Client, password string) error {
	if password != "" && channel.password != password {
		return errors.New("incorrect password")
	}

	channel.mutex.Lock()
	defer channel.mutex.Unlock()

	channel.clients[client] = true

	return nil
}

func (ch *Channel) Part(client *Client) error {
	if !ch.clients[client] {
		return errors.New("not a channel member")
	}
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	delete(ch.clients, client)
	return nil
}

func (ch *Channel) Broadcast(message string) {
	for c := range ch.clients {
		c.Out <- message
	}
}
