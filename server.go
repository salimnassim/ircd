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
	regexNick    = regexKey(0)
	regexChannel = regexKey(1)
)

type ServerConfig struct {
	Name    string
	Network string
	Version string
	MOTD    []string

	TLS             bool
	CertificateFile string
	CertificateKey  string

	PingFrequency  int
	PongMaxLatency int
}

type Server interface {
	Run(listener net.Listener, isTLS bool)
}

type server struct {
	mu       *sync.RWMutex
	name     string
	network  string
	version  string
	clients  ClientStorer
	channels ChannelStorer
	message  *[]string

	pingFrequency  int
	pongMaxLatency int

	// regex cache
	regex map[regexKey]*regexp.Regexp
}

func NewServer(config ServerConfig) *server {
	server := &server{
		mu:             &sync.RWMutex{},
		name:           config.Name,
		network:        config.Network,
		version:        config.Version,
		clients:        NewClientStore("clients"),
		channels:       NewChannelStore("channels"),
		regex:          make(map[regexKey]*regexp.Regexp),
		message:        &config.MOTD,
		pingFrequency:  config.PingFrequency,
		pongMaxLatency: config.PongMaxLatency,
	}

	compileRegexp(server)

	return server
}

func (s *server) Run(listener net.Listener, isTLS bool) {
	for {
		connection, err := listener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("unable to accept connection")
			continue
		}
		log.Info().Msgf("accepted connection from %s", connection.RemoteAddr())
		go handleConnection(connection, s)
	}
}

// Compiles expressions and caches them to a map.
func compileRegexp(s *server) {
	rgxNick, err := regexp.Compile(`([a-zA-Z0-9\[\]\{\}\\\|]{2,16})`)
	if err != nil {
		log.Panic().Err(err).Msg("unable to compile nickname validation regex")
	}
	s.regex[regexNick] = rgxNick

	rgxChannel, err := regexp.Compile(`[#!&][^\x00\x07\x0a\x0d\x20\x2C\x3A]{1,50}`)
	if err != nil {
		log.Panic().Err(err).Msg("unable to compile channel validation regex")
	}
	s.regex[regexChannel] = rgxChannel
}

// Returns the number of connected clients and open channels.
func (s *server) stats() (visible int, invisible, channels int) {
	visible, invisible = s.clients.count()
	return visible, invisible, s.channels.count()
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
