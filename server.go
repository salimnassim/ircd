package ircd

import (
	"net/http"
	"regexp"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd/metrics"
)

type regexKey string

const (
	regexKeyNick    = regexKey("nick")
	regexKeyChannel = regexKey("channel")
)

type ServerConfig struct {
	Name string
	MOTD []string
}

type Server struct {
	mu       *sync.RWMutex
	name     string
	clients  ClientStorer
	channels ChannelStorer
	motd     *[]string

	// regex cache
	regex map[regexKey]*regexp.Regexp
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
		regex:    make(map[regexKey]*regexp.Regexp),
		motd:     &config.MOTD,
	}

	compileRegexp(server)

	return server
}

// Compiles expressions and caches them to a map.
func compileRegexp(server *Server) {
	rgxNick, err := regexp.Compile(`([a-zA-Z0-9\[\]\|]{2,9})`)
	if err != nil {
		log.Panic().Err(err).Msg("unable to compile nickname validation regex")
	}
	server.regex[regexKeyNick] = rgxNick

	rgxChannel, err := regexp.Compile(`[#!&][^\x00\x07\x0a\x0d\x20\x2C\x3A]{1,50}`)
	if err != nil {
		log.Panic().Err(err).Msg("unable to compile channel validation regex")
	}
	server.regex[regexKeyChannel] = rgxChannel
}

// Returns the number of connected clients and open channels.
func (server *Server) Stats() (clients int, channels int) {
	return server.clients.Count(), server.channels.Count()
}

// Removes client from channels and client map.
func (server *Server) RemoveClient(client *Client) error {
	memberOf := server.channels.MemberOf(client)
	for _, ch := range memberOf {
		ch.RemoveClient(client)
	}

	server.clients.Delete(ClientID(client.id))
	metrics.Clients.Dec()

	return nil
}

func (server *Server) MOTD() []string {
	var motd []string
	server.mu.RLock()
	motd = *server.motd
	server.mu.RUnlock()
	return motd
}
