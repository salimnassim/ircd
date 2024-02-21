package ircd

import (
	"net/http"
	"regexp"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

var (
	// Number of connected clients
	promClients = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ircd_clients",
		Help: "Number of connected clients",
	})
	// Number of existing channels
	promChannels = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "ircd_channels",
		Help: "Number of existing channels",
	})
	// Number of PING messages
	promPings = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ircd_ping",
		Help: "Number of PING messages",
	})
	// Number of PRIVMSG sent to channels
	promPrivmsgChannel = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ircd_channels_privmsg",
		Help: "Number of PRIVMSG sent to channels",
	})
	// Number of PRIVMSG sent to clients
	promPrivmsgClient = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ircd_clients_privmsg",
		Help: "Number of PRIVMSG sent to clients",
	})
)

type ServerConfig struct {
	Name string
}

type Server struct {
	mu       *sync.RWMutex
	name     string
	clients  ClientStorer
	channels ChannelStorer
	motd     *[]string

	// regex cache
	regex map[string]*regexp.Regexp
}

func (server *Server) IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusTeapot)
}

func NewServer(config ServerConfig) *Server {
	server := &Server{
		mu:       &sync.RWMutex{},
		name:     config.Name,
		clients:  NewClientStore("clients"),
		channels: NewChannelStore("channels"),
		regex:    make(map[string]*regexp.Regexp),
		motd:     &[]string{"This is the message of the day.", "It contains multiple lines because the lines could be long.", "🍩🍫🍡🍦🍬🍮"},
	}

	compileRegexp(server)

	return server
}

// Compiles expressions and caches them to a map
func compileRegexp(server *Server) {
	rgxNick, err := regexp.Compile(`([a-zA-Z0-9\[\]\|]{2,9})`)
	if err != nil {
		log.Panic().Err(err).Msg("unable to compile nickname validation regex")
	}
	server.regex["nick"] = rgxNick

	rgxChannel, err := regexp.Compile(`[#!&][^\x00\x07\x0a\x0d\x20\x2C\x3A]{1,50}`)
	if err != nil {
		log.Panic().Err(err).Msg("unable to compile channel validation regex")
	}
	server.regex["channel"] = rgxChannel
}

// Returns the number of connected clients, and channels
func (server *Server) Stats() (clients int, channels int) {
	return server.clients.Count(), server.channels.Count()
}

// Removes client from channels and client map
func (server *Server) RemoveClient(client *Client) error {
	memberOf := server.channels.MemberOf(client)
	for _, ch := range memberOf {
		ch.RemoveClient(client)
	}

	server.clients.Delete(ClientID(client.id))
	promClients.Dec()

	return nil
}

func (server *Server) MOTD() []string {
	var motd []string
	server.mu.RLock()
	motd = *server.motd
	server.mu.RUnlock()
	return motd
}
