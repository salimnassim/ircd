package ircd

import (
	"errors"
	"sync"
)

type Channel struct {
	name     string
	mutex    *sync.Mutex
	clients  map[*Client]bool
	password string
}

func NewChannel() *Channel {
	channel := &Channel{
		mutex:    &sync.Mutex{},
		clients:  make(map[*Client]bool),
		password: "",
	}
	return channel
}

func (ch *Channel) join(client *Client, password string) error {
	if password != "" && ch.password != password {
		return errors.New("incorrect password")
	}

	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	ch.clients[client] = true

	return nil
}

func (ch *Channel) leave(client *Client) error {
	if !ch.clients[client] {
		return errors.New("not a channel member")
	}
	ch.mutex.Lock()
	defer ch.mutex.Unlock()
	delete(ch.clients, client)
	return nil
}
