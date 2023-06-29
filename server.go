package ircd

import (
	"errors"
	"math/rand"
	"sync"
	"time"
)

type Server struct {
	sync.Mutex
	Name     string
	clients  map[*Client]bool
	channels map[*Channel]bool
	random   *rand.Rand
}

type ServerClients struct {
	Active    int
	Invisible int
	Channels  int
}

func NewServer(name string) *Server {
	server := &Server{
		Name:     name,
		clients:  make(map[*Client]bool),
		channels: make(map[*Channel]bool),
		random:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	return server
}

func (server *Server) GetRandom() int {
	return server.random.Intn(1000)
}

// Returns the number active users and channels on the server
func (server *Server) Counts() ServerClients {
	active := len(server.clients)
	channels := len(server.channels)
	invisible := 0
	for client := range server.clients {
		if client.Invisible {
			invisible++
		}
	}
	return ServerClients{
		Active:    active,
		Invisible: invisible,
		Channels:  channels,
	}
}

// Adds client to client list
func (server *Server) AddClient(client *Client) {
	server.Mutex.Lock()
	defer server.Mutex.Unlock()

	server.clients[client] = true
}

// Removes client from client list
func (server *Server) RemoveClient(client *Client) {
	server.Mutex.Lock()
	defer server.Mutex.Unlock()

	for channel := range server.channels {
		for c := range channel.clients {
			if c.Nickname == client.Nickname {
				channel.RemoveClient(c)
			}
		}
	}

	delete(server.clients, client)
}

// Returns a pointer to client by nickname
func (server *Server) GetClient(nickname string) (*Client, bool) {
	for client := range server.clients {
		if client.Nickname == nickname {
			return client, true
		}
	}
	return nil, false
}

// Returns a pointer to channel by name
func (server *Server) GetChannel(name string) (*Channel, error) {
	for channel := range server.channels {
		if channel.Name == name {
			return channel, nil
		}
	}
	return nil, errors.New("channel not found")
}

// Creates a channel and returns a pointer to it
func (server *Server) CreateChannel(name string) *Channel {
	server.Lock()
	defer server.Unlock()

	channel := NewChannel(name)
	server.channels[channel] = true

	return channel
}
