package ircd

import (
	"errors"
	"net/http"
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
	clients  ClientStoreable
	channels map[*Channel]interface{}

	// regex cache
	regex map[string]*regexp.Regexp

	// metrics
	gauges   map[string]prometheus.Gauge
	counters map[string]prometheus.Counter
}

func (server *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func NewServer(config ServerConfig) *Server {
	server := &Server{
		mu:       &sync.RWMutex{},
		name:     config.Name,
		clients:  NewClientStore(),
		channels: make(map[*Channel]interface{}),
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
	return server.clients.Size(), len(server.channels)
}

// Removes client from channels and client map
func (server *Server) RemoveClient(client *Client) error {
	server.mu.Lock()
	defer server.mu.Unlock()

	for channel := range server.channels {
		for v, c := range channel.clients {
			if v == client.id {
				err := channel.RemoveClient(c)
				if err != nil {
					return err
				}
			}
		}
	}

	server.clients.Remove(client)
	server.gauges["ircd_clients"].Dec()

	return nil
}

func (server *Server) Channels() []Channel {
	server.mu.RLock()
	defer server.mu.RUnlock()

	var channels []Channel
	for c := range server.channels {
		channels = append(channels, *c)
	}

	return channels
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
func (server *Server) CreateChannel(name string, owner *Client) *Channel {
	server.mu.Lock()
	defer server.mu.Unlock()

	channel := NewChannel(name, owner)
	server.channels[channel] = true
	server.gauges["ircd_channels"].Inc()

	return channel
}

// Removes client from channel map
func (server *Server) RemoveChannel(channel *Channel) error {
	server.mu.Lock()
	defer server.mu.Unlock()

	if _, ok := server.channels[channel]; ok {
		delete(server.channels, channel)
		server.gauges["ircd_channels"].Dec()
		return nil
	}

	return errors.New("channel not found")
}
