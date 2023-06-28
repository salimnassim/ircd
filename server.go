package ircd

import (
	"errors"
	"fmt"
	"sync"
)

type Server struct {
	sync.Mutex
	Name     string
	Clients  map[*Client]bool
	Channels map[*Channel]bool
}

func NewServer(name string) *Server {
	server := &Server{
		Name:     name,
		Clients:  make(map[*Client]bool),
		Channels: make(map[*Channel]bool),
	}
	return server
}

func (server *Server) AddClient(client *Client) {
	server.Mutex.Lock()
	defer server.Mutex.Unlock()

	server.Clients[client] = true
}

func (server *Server) RemoveClient(client *Client) {
	server.Mutex.Lock()
	defer server.Mutex.Unlock()

	delete(server.Clients, client)
}

func (server *Server) GetClient(nickname string) (*Client, bool) {
	for client := range server.Clients {
		if client.Nickname == nickname {
			return client, true
		}
	}
	return nil, false
}

func (server *Server) GetChannel(name string) (*Channel, error) {
	for channel := range server.Channels {
		if channel.Name == name {
			return channel, nil
		}
	}
	return nil, errors.New("channel not found")
}

func (server *Server) CreateChannel(name string) *Channel {
	server.Lock()
	defer server.Unlock()

	channel := NewChannel(name)
	server.Channels[channel] = true

	return channel
}

func Line(nickname string, code string, message string) string {
	if nickname == "" {
		nickname = "*"
	}
	return fmt.Sprintf("%s: %s %s\r\n", nickname, code, message)
}
