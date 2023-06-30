package ircd

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog/log"
)

type Messager interface {
	message(string)
}

type ServerConfig struct {
	Name string
}

type Server struct {
	mu       *sync.RWMutex
	Name     string
	clients  map[*Client]bool
	channels map[*Channel]bool
	Gauges   map[string]prometheus.Gauge
	Counters map[string]prometheus.Counter
}

func NewServer(config ServerConfig) *Server {
	server := &Server{
		mu:       &sync.RWMutex{},
		Name:     config.Name,
		clients:  make(map[*Client]bool),
		channels: make(map[*Channel]bool),
		Gauges:   make(map[string]prometheus.Gauge),
		Counters: make(map[string]prometheus.Counter),
	}

	registerMetrics(server)

	return server
}

// Register prometheus metrics
func registerMetrics(server *Server) {
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
}

// Adds client to client map
func (server *Server) AddClient(client *Client) error {
	log.Info().Msgf("adding client %s", client.Nickname)
	server.mu.Lock()
	defer server.mu.Unlock()

	server.clients[client] = true
	server.Gauges["ircd_clients"].Inc()

	return nil
}

// Removes client from client map
func (server *Server) RemoveClient(client *Client) error {
	log.Info().Msgf("removing client %s", client.Nickname)
	server.mu.Lock()
	defer server.mu.Unlock()

	for channel := range server.channels {
		for c := range channel.clients {
			if c.Nickname == client.Nickname {
				err := channel.RemoveClient(c)
				if err != nil {
					return err
				}
			}
		}
	}

	delete(server.clients, client)
	server.Gauges["ircd_clients"].Dec()

	return nil
}

// Returns a pointer to client by nickname
func (server *Server) ClientByNickname(nickname string) (*Client, bool) {
	server.mu.RLock()
	defer server.mu.RUnlock()

	for client := range server.clients {
		if client.Nickname == nickname {
			return client, true
		}
	}
	return nil, false
}

func (server *Server) Clients() map[*Client]bool {
	server.mu.RLock()
	defer server.mu.RUnlock()

	return server.clients
}

func (server *Server) Channels() map[*Channel]bool {
	server.mu.RLock()
	defer server.mu.RUnlock()

	return server.channels
}

// Returns a pointer to channel by name
func (server *Server) Channel(name string) (*Channel, error) {
	server.mu.RLock()
	defer server.mu.RUnlock()

	for channel := range server.channels {
		if channel.Name == name {
			return channel, nil
		}
	}
	return nil, errors.New("channel not found")
}

// Creates a channel and returns a pointer to it
func (server *Server) CreateChannel(name string) *Channel {
	log.Info().Msgf("creating channel %s", name)

	server.mu.Lock()
	defer server.mu.Unlock()

	channel := NewChannel(name)
	server.channels[channel] = true
	server.Gauges["ircd_channels"].Inc()

	return channel
}

// Removes client from channel map
func (server *Server) RemoveChannel(channel *Channel) error {
	log.Info().Msgf("removing channel %s", channel.Name)

	server.mu.Lock()
	defer server.mu.Unlock()

	if _, ok := server.channels[channel]; ok {
		delete(server.channels, channel)
		server.Gauges["ircd_channels"].Dec()
		return nil
	}

	return errors.New("channel not found")
}
