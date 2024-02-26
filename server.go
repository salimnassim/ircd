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
	Name     string
	Password string
	Network  string
	Version  string
	MOTD     []string

	TLS             bool
	CertificateFile string
	CertificateKey  string

	PingFrequency  int
	PongMaxLatency int

	Parameters ServerConfigParameters
}

type ServerConfigParameters struct {
	// https://modern.ircdocs.horse/#awaylen-parameter
	MaxAwayLength int
	// https://modern.ircdocs.horse/#casemapping-parameter
	CaseMapping string
	// https://modern.ircdocs.horse/#casemapping-parameter
	ChannelLimit string
	// https://modern.ircdocs.horse/#chanmodes-parameter
	ChannelModes string
	// https://modern.ircdocs.horse/#channellen-parameter
	MaxChannelLength int
	// https://modern.ircdocs.horse/#chantypes-parameter
	ChannelTypes string
	// https://modern.ircdocs.horse/#elist-parameter
	EList string
	// https://modern.ircdocs.horse/#excepts-parameter
	Excepts string
	// https://modern.ircdocs.horse/#hostlen-parameter
	MaxHostnameLength int
	// https://modern.ircdocs.horse/#kicklen-parameter
	MaxKickLength int
	// https://modern.ircdocs.horse/#maxlist-parameter
	MaxList string
	// https://modern.ircdocs.horse/#modes-parameter
	MaxModes int
	// https://modern.ircdocs.horse/#network-parameter
	//
	// NOTE: Use ASCII codes for characters such as space (\x20)
	Network string
	// https://modern.ircdocs.horse/#nicklen-parameter
	MaxNickLength int
	// https://modern.ircdocs.horse/#prefix-parameter
	ChannelPrefixes string
	// https://modern.ircdocs.horse/#statusmsg-parameter
	StatusMessage string
	// https://modern.ircdocs.horse/#targmax-parameter
	MaxTargets string
	// https://modern.ircdocs.horse/#topiclen-parameter
	MaxTopicLength int
	// https://modern.ircdocs.horse/#userlen-parameter
	MaxUserLength int
}

// https://modern.ircdocs.horse/#elist-parameter
//
// Returns a ELIST compatible string
func (s ServerConfigParameters) build() string {
	return fmt.Sprintf(
		`AWAYLEN=%d CASEMAPPING=%s CHANLIMIT=%s CHANMODES=%s CHANTYPES=%s ELIST=%s HOSTLEN=%d KICKLEN=%d MAXLIST=%s MODES=%d NETWORK=%s NICKLEN=%d PREFIX=%s TARGMAX=%s TOPICLEN=%d USERLEN=%d`,
		s.MaxAwayLength, s.CaseMapping, s.ChannelLimit,
		s.ChannelModes, s.ChannelTypes, s.EList,
		s.MaxHostnameLength, s.MaxKickLength, s.MaxList,
		s.MaxModes, s.Network, s.MaxNickLength,
		s.ChannelPrefixes, s.MaxTargets, s.MaxTopicLength,
		s.MaxUserLength,
	)
}

type server struct {
	mu        *sync.RWMutex
	router    router
	name      string
	password  string
	network   string
	version   string
	Clients   ClientStorer
	Channels  ChannelStorer
	Operators OperatorStorer
	motd      *[]string
	ports     []string

	pingFrequency  int
	pongMaxLatency int

	params string

	// regex cache
	regex map[regexKey]*regexp.Regexp
}

func NewServer(config ServerConfig) *server {
	server := &server{
		mu:             &sync.RWMutex{},
		name:           config.Name,
		password:       config.Password,
		network:        config.Network,
		version:        config.Version,
		Clients:        NewClientStore("clients"),
		Channels:       NewChannelStore("channels"),
		Operators:      NewOperatorStore(),
		motd:           &config.MOTD,
		ports:          []string{},
		pingFrequency:  config.PingFrequency,
		pongMaxLatency: config.PongMaxLatency,
		params:         config.Parameters.build(),
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

	router.registerHandler("PASS", handlePass, middlewareNeedParams(1))
	router.registerHandler("PING", handlePing)
	router.registerHandler("PONG", handlePong)
	router.registerHandler("NICK", handleNick, middlewareNeedParams(1))
	router.registerHandler("USER", handleUser, middlewareNeedParams(4))
	router.registerHandler("LUSERS", handleLusers, middlewareNeedHandshake)
	router.registerHandler("JOIN", handleJoin, middlewareNeedHandshake, middlewareNeedParams(1))
	router.registerHandler("PART", handlePart, middlewareNeedHandshake, middlewareNeedParams(1))
	router.registerHandler("KICK", handleKick, middlewareNeedHandshake, middlewareNeedParams(2))
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
	router.registerHandler("INVITE", handleInvite, middlewareNeedHandshake, middlewareNeedParams(2))
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
