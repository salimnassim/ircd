package ircd

import (
	"net/http"
	"regexp"

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
	name     string
	clients  ClientStoreable
	channels ChannelStoreable

	// regex cache
	regex map[string]*regexp.Regexp

	// metrics
	gauges   map[string]prometheus.Gauge
	counters map[string]prometheus.Counter
}

func (server *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func NewServer(config ServerConfig) *Server {
	server := &Server{
		name:     config.Name,
		clients:  NewClientStore("clients"),
		channels: NewChannelStore("channels"),
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
	rgxNick, err := regexp.Compile(`([a-zA-Z0-9\[\]\|]{2,9})`)
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

	server.counters["ircd_ping"] =
		prometheus.NewCounter(prometheus.CounterOpts{
			Name: "ircd_ping",
			Help: "Number of PING messages",
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
	return server.clients.Size(), server.channels.Size()
}

// Removes client from channels and client map
func (server *Server) RemoveClient(client *Client) error {
	memberOf := server.channels.MemberOf(client)
	for _, ch := range memberOf {
		ch.RemoveClient(client)
	}

	server.clients.Remove(client)
	server.gauges["ircd_clients"].Dec()

	return nil
}
