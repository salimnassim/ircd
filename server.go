package ircd

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type Server struct {
	mu       sync.Mutex
	Name     string
	clients  map[*Client]bool
	channels map[*Channel]bool
	Gauges   map[string]prometheus.Gauge
	Counters map[string]prometheus.Counter
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
		Gauges:   make(map[string]prometheus.Gauge),
		Counters: make(map[string]prometheus.Counter),
	}

	server.Gauges["ircd_clients"] =
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ircd_clients",
			Help: "Number of connected clients",
		})

	server.Gauges["ircd_channels"] =
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ircd_channels",
			Help: "Number of channels",
		})

	server.Counters["ircd_channels_privmsg"] =
		prometheus.NewCounter(prometheus.CounterOpts{
			Name: "ircd_channels_privmsg",
			Help: "Number of PRIVMSG sent to channels",
		})

	server.Counters["ircd_clients_privmsg"] =
		prometheus.NewCounter(prometheus.CounterOpts{
			Name: "ircd_clients_privmsg",
			Help: "Number of PRIVMSG sent to clients",
		})

	for _, v := range server.Gauges {
		prometheus.MustRegister(v)
	}

	for _, v := range server.Counters {
		prometheus.MustRegister(v)
	}

	return server
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
	server.mu.Lock()
	defer server.mu.Unlock()

	server.clients[client] = true
	server.Gauges["ircd_clients"].Inc()
}

// Removes client from client list
func (server *Server) RemoveClient(client *Client) {
	server.mu.Lock()
	defer server.mu.Unlock()

	for channel := range server.channels {
		for c := range channel.clients {
			if c.Nickname == client.Nickname {
				channel.RemoveClient(c)
			}
		}
	}

	delete(server.clients, client)
	server.Gauges["ircd_clients"].Dec()
}

// Returns a pointer to client by nickname
func (server *Server) GetClient(nickname string) (*Client, error) {
	server.mu.Lock()
	defer server.mu.Unlock()
	for client := range server.clients {
		if client.Nickname == nickname {
			return client, nil
		}
	}
	return nil, errors.New("client not found")
}

// Returns a pointer to channel by name
func (server *Server) GetChannel(name string) (*Channel, error) {
	server.mu.Lock()
	defer server.mu.Unlock()
	for channel := range server.channels {
		if channel.Name == name {
			return channel, nil
		}
	}
	return nil, errors.New("channel not found")
}

// Creates a channel and returns a pointer to it
func (server *Server) CreateChannel(name string) *Channel {
	server.mu.Lock()
	defer server.mu.Unlock()

	channel := NewChannel(name)
	server.channels[channel] = true
	server.Gauges["ircd_channels"].Inc()

	return channel
}

func (server *Server) RemoveChannel(channel *Channel) error {
	server.mu.Lock()
	defer server.mu.Unlock()

	if _, ok := server.channels[channel]; ok {
		delete(server.channels, channel)
		server.Gauges["ircd_channels"].Dec()
		return nil
	}

	return errors.New("channel not found")
}
