package ircd

import (
	"errors"
	"regexp"
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
	name     string
	clients  map[*Client]bool
	channels map[*Channel]bool

	// regex cache
	regex map[string]*regexp.Regexp

	// metrics
	gauges   map[string]prometheus.Gauge
	counters map[string]prometheus.Counter
}

func NewServer(config ServerConfig) *Server {
	server := &Server{
		mu:       &sync.RWMutex{},
		name:     config.Name,
		clients:  make(map[*Client]bool),
		channels: make(map[*Channel]bool),
		regex:    make(map[string]*regexp.Regexp),
		gauges:   make(map[string]prometheus.Gauge),
		counters: make(map[string]prometheus.Counter),
	}

	compileRegexp(server)
	registerMetrics(server)

	return server
}

// Compiles expressions and caches them to a map
func compileRegexp(server *Server) {
	rgxNick, err := regexp.Compile(`([a-zA-Z0-9\[\]\|]{2,16})`)
	if err != nil {
		log.Panic().Err(err).Msg("unable to compile nickname validation regex")
	}
	server.regex["nick"] = rgxNick
}

// Register prometheus metrics
func registerMetrics(server *Server) {
	server.gauges["ircd_clients"] =
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ircd_clients",
			Help: "Number of connected clients",
		})

	server.gauges["ircd_channels"] =
		prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "ircd_channels",
			Help: "Number of channels",
		})

	server.counters["ircd_channels_privmsg"] =
		prometheus.NewCounter(prometheus.CounterOpts{
			Name: "ircd_channels_privmsg",
			Help: "Number of PRIVMSG sent to channels",
		})

	server.counters["ircd_clients_privmsg"] =
		prometheus.NewCounter(prometheus.CounterOpts{
			Name: "ircd_clients_privmsg",
			Help: "Number of PRIVMSG sent to clients",
		})

	for _, v := range server.gauges {
		prometheus.MustRegister(v)
	}

	for _, v := range server.counters {
		prometheus.MustRegister(v)
	}
}

// Returns the number of connected clients, and channels
func (server *Server) Stats() (int, int) {
	server.mu.RLock()
	defer server.mu.RUnlock()

	return len(server.clients), len(server.channels)
}

// Adds client to client map
func (server *Server) AddClient(client *Client) error {
	log.Info().Msgf("adding client %s", client.nickname)
	server.mu.Lock()
	defer server.mu.Unlock()

	server.clients[client] = true
	server.gauges["ircd_clients"].Inc()

	return nil
}

// Removes client from channels and client map
func (server *Server) RemoveClient(client *Client) error {
	log.Info().Msgf("removing client %s", client.nickname)
	server.mu.Lock()
	defer server.mu.Unlock()

	for channel := range server.channels {
		for c := range channel.clients {
			if c == client {
				err := channel.RemoveClient(c)
				if err != nil {
					return err
				}
			}
		}
	}

	delete(server.clients, client)
	server.gauges["ircd_clients"].Dec()

	return nil
}

// Returns a pointer to client by nickname
func (server *Server) ClientByNickname(nickname string) (*Client, bool) {
	server.mu.Lock()
	defer server.mu.Unlock()

	for client := range server.clients {
		if client.nickname == nickname {
			return client, true
		}
	}
	return nil, false
}

func (server *Server) Clients() []Client {
	server.mu.RLock()
	defer server.mu.RUnlock()

	var clients []Client
	for c := range server.clients {
		clients = append(clients, *c)
	}

	return clients
}

func (server *Server) Channels() map[*Channel]bool {
	server.mu.RLock()
	defer server.mu.RUnlock()

	return server.channels
}

// Returns a pointer to channel by name. bool will be true if channel exists
func (server *Server) Channel(name string) (*Channel, bool) {
	server.mu.RLock()
	defer server.mu.RUnlock()

	for channel := range server.channels {
		if channel.name == name {
			return channel, true
		}
	}
	return nil, false
}

// Creates a channel and returns a pointer to it
func (server *Server) CreateChannel(name string) *Channel {
	log.Info().Msgf("creating channel %s", name)

	server.mu.Lock()
	defer server.mu.Unlock()

	channel := NewChannel(name)
	server.channels[channel] = true
	server.gauges["ircd_channels"].Inc()

	return channel
}

// Removes client from channel map
func (server *Server) RemoveChannel(channel *Channel) error {
	log.Info().Msgf("removing channel %s", channel.name)

	server.mu.Lock()
	defer server.mu.Unlock()

	if _, ok := server.channels[channel]; ok {
		delete(server.channels, channel)
		server.gauges["ircd_channels"].Dec()
		return nil
	}

	return errors.New("channel not found")
}
