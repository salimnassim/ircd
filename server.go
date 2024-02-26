package ircd

import (
	"fmt"
	"net"
	"regexp"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd/metrics"
)

type Server interface {
	Run(listener net.Listener, isTLS bool)
}

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

type server struct {
	mu        *sync.RWMutex
	router    router
	name      string
	network   string
	version   string
	Clients   ClientStorer
	Channels  ChannelStorer
	Operators OperatorStorer
	motd      *[]string
	ports     []string

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
		Operators:      NewOperatorStore(),
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
	rgxNick, err := regexp.Compile(`([a-zA-Z0-9\[\]\{\}\\\|]{2,31})`)
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
	router.registerGlobalMiddleware(func(s *server, c clienter, m message, next handlerFunc) handlerFunc {
		metrics.Command.WithLabelValues(m.command).Inc()
		return next
	})

	router.registerHandler("PING", handlePing)
	router.registerHandler("PONG", handlePong)
	router.registerHandler("NICK", handleNick, middlewareNeedParams(1))
	router.registerHandler("USER", handleUser, middlewareNeedParams(4))
	router.registerHandler("LUSERS", handleLusers, middlewareNeedHandshake)
	router.registerHandler("JOIN", handleJoin, middlewareNeedHandshake, middlewareNeedParams(1))
	router.registerHandler("PART", handlePart, middlewareNeedHandshake, middlewareNeedParams(1))
	router.registerHandler("TOPIC", handleTopic, middlewareNeedHandshake, middlewareNeedParams(1))
	router.registerHandler("PRIVMSG", handlePrivmsg, middlewareNeedHandshake, middlewareNeedParams(1))
	router.registerHandler("WHOIS", handleWhois, middlewareNeedHandshake, middlewareNeedParams(1))
	router.registerHandler("WHO", handleWho, middlewareNeedHandshake)
	router.registerHandler("MODE", handleMode, middlewareNeedHandshake, middlewareNeedParams(1))
	router.registerHandler("AWAY", handleAway, middlewareNeedHandshake)
	router.registerHandler("QUIT", handleQuit)
	router.registerHandler("OPER", handleOper, middlewareNeedHandshake, middlewareNeedParams(2))
	router.registerHandler("VERSION", handleVersion, middlewareNeedHandshake)
	router.registerHandler("LIST", handleList, middlewareNeedHandshake)
	router.registerHandler("DEBUG", func(s *server, c clienter, m message) {
		func() {}() // breakpoint here
	}, middlewareNeedHandshake, middlewareNeedOper)

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
func (s *server) removeClient(c clienter) {
	memberOf := s.Channels.memberOf(c)
	for _, ch := range memberOf {
		ch.removeClient(c)
	}

	s.Clients.delete(c.id())
}

func (s *server) MOTD() []string {
	var motd []string
	s.mu.RLock()
	motd = *s.motd
	s.mu.RUnlock()
	return motd
}
