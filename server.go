package ircd

import (
	"net"
	"regexp"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd/metrics"
)

type regexKey int

const (
	regexKeyNick    = regexKey(0)
	regexKeyChannel = regexKey(1)
)

type ServerConfig struct {
	Name string
	MOTD []string

	TLS             bool
	CertificateFile string
	CertificateKey  string
}

type Server interface {
	Run(listener net.Listener)
}

type server struct {
	mu       *sync.RWMutex
	name     string
	clients  ClientStorer
	channels ChannelStorer
	message  *[]string
	tls      bool

	// regex cache
	regex map[regexKey]*regexp.Regexp
}

func NewServer(config ServerConfig) *server {
	server := &server{
		mu:       &sync.RWMutex{},
		name:     config.Name,
		clients:  newClientStore("clients"),
		channels: newChannelStore("channels"),
		regex:    make(map[regexKey]*regexp.Regexp),
		message:  &config.MOTD,
		tls:      config.TLS,
	}

	compileRegexp(server)

	return server
}

func (s *server) Run(listener net.Listener) {
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("unable to accept connection")
			continue
		}
		log.Info().Msgf("accepted connection from %s", connection.RemoteAddr())
		go HandleConnection(connection, s)
	}
}

// Compiles expressions and caches them to a map.
func compileRegexp(s *server) {
	rgxNick, err := regexp.Compile(`([a-zA-Z0-9\[\]\|]{2,9})`)
	if err != nil {
		log.Panic().Err(err).Msg("unable to compile nickname validation regex")
	}
	s.regex[regexKeyNick] = rgxNick

	rgxChannel, err := regexp.Compile(`[#!&][^\x00\x07\x0a\x0d\x20\x2C\x3A]{1,50}`)
	if err != nil {
		log.Panic().Err(err).Msg("unable to compile channel validation regex")
	}
	s.regex[regexKeyChannel] = rgxChannel
}

// Returns the number of connected clients and open channels.
func (s *server) stats() (c int, channels int) {
	return s.clients.count(), s.channels.count()
}

// Removes client from channels and client map.
func (s *server) removeClient(c *client) error {
	memberOf := s.channels.memberOf(c)
	for _, ch := range memberOf {
		ch.removeClient(c)
	}

	s.clients.delete(clientID(c.id))
	metrics.Clients.Dec()

	return nil
}

func (s *server) motd() []string {
	var motd []string
	s.mu.RLock()
	motd = *s.message
	s.mu.RUnlock()
	return motd
}
