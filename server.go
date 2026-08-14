package ircd

import (
	"fmt"
	"net"
	"regexp"
	"sync"

	"github.com/rs/zerolog/log"
	"github.com/salimnassim/ircd/metrics"
)

type Serverer interface {
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
	MaxAwayLength int

	CaseMapping string

	ChannelLimit string

	ChannelModes string

	MaxChannelLength int

	ChannelTypes string

	EList string

	Excepts string

	MaxHostnameLength int

	MaxKickLength int

	MaxList string

	MaxModes int

	Network string

	MaxNickLength int

	ChannelPrefixes string

	StatusMessage string

	MaxTargets string

	MaxTopicLength int

	MaxUserLength int
}

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

	p []string

	pingFrequency  int
	pongMaxLatency int

	params string

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
		p:              []string{},
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
		go handleConnection(connection, s)
	}
}

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
		func() {}()
	}, middlewareNeedHandshake)

	s.router = router
}

func (s *server) addPort(port string) {
	s.mu.Lock()
	s.p = append(s.p, port)
	s.mu.Unlock()
}

func (s *server) ports() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.p
}

func (s *server) Stats() (visible int, invisible, channels int) {
	visible, invisible = s.Clients.count()
	return visible, invisible, s.Channels.count()
}

func (s *server) cleanup(c clienter) {
	for _, ch := range s.Channels.memberOf(c) {
		ch.broadcastCommand(quitCommand{
			prefix: c.prefix(),
			text:   fmt.Sprintf("Quit: %s", c.quitReason()),
		}, c.id(), true)
		ch.clients().remove(c)
	}
	s.Clients.delete(c.id())
	metrics.Clients.Dec()
}

func (s *server) MOTD() []string {
	var motd []string
	s.mu.RLock()
	motd = *s.motd
	s.mu.RUnlock()
	return motd
}
