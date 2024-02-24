package ircd

import (
	"fmt"
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
	router   router
	name     string
	network  string
	version  string
	Clients  ClientStorer
	Channels ChannelStorer
	motd     *[]string
	ports    []string

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
		Clients:        NewClientStore("clients"),
		Channels:       NewChannelStore("channels"),
		motd:           &config.MOTD,
		pingFrequency:  config.PingFrequency,
		pongMaxLatency: config.PongMaxLatency,
		regex:          make(map[regexKey]*regexp.Regexp),
	}

	compileRegexp(server)
	registerHandlers(server)

	return server
}

func (s *server) Run(listener net.Listener, isTLS bool) {
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		log.Error().Err(err).Msgf("cant split net host port")
	}

	if isTLS {
		s.addPort(fmt.Sprintf("+%s", port))
	} else {
		s.addPort(port)
	}

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

func registerHandlers(s *server) {
	router := NewCommandRouter(s)
	router.registerHandler("PING", handlePing)
	router.registerHandler("PONG", handlePong)
	router.registerHandler("NICK", handleNick)
	router.registerHandler("USER", handleUser)
	router.registerHandler("LUSERS", handleLusers, middlewareNeedHandshake)
	router.registerHandler("JOIN", handleJoin, middlewareNeedHandshake)
	router.registerHandler("PART", handlePart, middlewareNeedHandshake)
	router.registerHandler("TOPIC", handleTopic, middlewareNeedHandshake)
	router.registerHandler("PRIVMSG", handlePrivmsg, middlewareNeedHandshake)
	router.registerHandler("WHOIS", handleWhois, middlewareNeedHandshake)
	router.registerHandler("WHO", handleWho, middlewareNeedHandshake)
	router.registerHandler("MODE", handleMode, middlewareNeedHandshake)
	router.registerHandler("AWAY", handleAway, middlewareNeedHandshake)
	router.registerHandler("QUIT", handleQuit)

	s.router = router
}

func (s *server) addPort(port string) {
	s.mu.Lock()
	s.ports = append(s.ports, port)
	s.mu.Unlock()
}

func (s *server) Ports() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ports
}

// Returns the number of connected clients and open channels.
func (s *server) Stats() (visible int, invisible, channels int) {
	visible, invisible = s.Clients.count()
	return visible, invisible, s.Channels.count()
}

// Removes client from channels and client map.
func (s *server) removeClient(c *client) error {
	log.Info().Msgf("removing client '%s' from store.", c.id)

	memberOf := s.Channels.memberOf(c)
	for _, ch := range memberOf {
		ch.removeClient(c)
	}

	s.Clients.delete(clientID(c.id))
	metrics.Clients.Dec()

	return nil
}

func (s *server) MOTD() []string {
	var motd []string
	s.mu.RLock()
	motd = *s.motd
	s.mu.RUnlock()
	return motd
}
