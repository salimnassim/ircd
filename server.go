// Package ircd implements a subset of the IRC protocol (RFC 1459) as a
// standalone server: client sessions, channels, modes, and the commands
// needed to connect, join, and converse.
package ircd

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

// ServerConfig holds the settings NewServer needs to construct a Server.
type ServerConfig struct {
	Name        string
	Password    string
	Network     string
	Version     string
	MOTD        []string
	CloakSecret []byte

	TLS             bool
	CertificateFile string
	CertificateKey  string

	PingFrequency  int
	PongMaxLatency int

	RateLimit  rate.Limit
	RateBurst  int
	OutboxSize int

	MaxConnectionsGlobal int
	MaxConnectionsPerIP  int
	MaxConnectRatePerIP  int // new connections allowed per IP per minute, 0 = unlimited
	MaxConnectBurstPerIP int

	MaxChannelsGlobal    int
	MaxChannelsPerClient int

	Parameters ServerConfigParameters
}

// ServerConfigParameters are the RPL_ISUPPORT limits and capabilities advertised to clients.
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

func (p ServerConfigParameters) build() string {
	return fmt.Sprintf(
		`AWAYLEN=%d CASEMAPPING=%s CHANLIMIT=%s CHANMODES=%s CHANTYPES=%s ELIST=%s HOSTLEN=%d KICKLEN=%d MAXLIST=%s MODES=%d NETWORK=%s NICKLEN=%d PREFIX=%s TARGMAX=%s TOPICLEN=%d USERLEN=%d`,
		p.MaxAwayLength, p.CaseMapping, p.ChannelLimit,
		p.ChannelModes, p.ChannelTypes, p.EList,
		p.MaxHostnameLength, p.MaxKickLength, p.MaxList,
		p.MaxModes, p.Network, p.MaxNickLength,
		p.ChannelPrefixes, p.MaxTargets, p.MaxTopicLength,
		p.MaxUserLength,
	)
}

var (
	nickRegex    = regexp.MustCompile(`([a-zA-Z0-9\[\]\{\}\\\|]{2,31})`)
	channelRegex = regexp.MustCompile(`[#!&][^\x00\x07\x0a\x0d\x20\x2C\x3A]{1,50}`)
)

// Server is a running IRC server: its configuration plus the client and channel state it manages.
type Server struct {
	config   ServerConfig
	isupport string

	clients   *clientDirectory
	channels  *channelDirectory
	operators *OperatorStore
	limiter   *connLimiter

	ports atomic.Pointer[[]string]

	sweepOnce sync.Once
}

// NewServer constructs a Server from config. Call Serve to begin accepting connections.
func NewServer(config ServerConfig) *Server {
	connRate := rate.Limit(0)
	if config.MaxConnectRatePerIP > 0 {
		connRate = rate.Limit(float64(config.MaxConnectRatePerIP) / 60.0)
	}

	channels := newChannelDirectory()
	channels.maxChannels = config.MaxChannelsGlobal

	return &Server{
		config:    config,
		isupport:  config.Parameters.build(),
		clients:   newClientDirectory(),
		channels:  channels,
		operators: NewOperatorStore(),
		limiter: newConnLimiter(
			config.MaxConnectionsGlobal,
			config.MaxConnectionsPerIP,
			connRate,
			config.MaxConnectBurstPerIP,
		),
	}
}

func (srv *Server) addPort(port string) {
	for {
		old := srv.ports.Load()
		next := make([]string, 0, len(deref(old))+1)
		next = append(next, deref(old)...)
		next = append(next, port)
		if srv.ports.CompareAndSwap(old, &next) {
			return
		}
	}
}

func deref(p *[]string) []string {
	if p == nil {
		return nil
	}
	return *p
}

// Ports returns the addresses the server is currently listening on, TLS ports prefixed with "+".
func (srv *Server) Ports() []string {
	return deref(srv.ports.Load())
}

func (srv *Server) sessionDeps() sessionDeps {
	return sessionDeps{
		serverName:  srv.config.Name,
		network:     srv.config.Network,
		version:     srv.config.Version,
		password:    srv.config.Password,
		motd:        srv.config.MOTD,
		cloakSecret: srv.config.CloakSecret,
		isupport:    srv.isupport,
		ports:       srv.Ports,

		pingFrequency:  time.Duration(srv.config.PingFrequency) * time.Second,
		pongMaxLatency: time.Duration(srv.config.PongMaxLatency) * time.Second,

		rateLimit:  srv.config.RateLimit,
		rateBurst:  srv.config.RateBurst,
		outboxSize: srv.config.OutboxSize,

		clients:      srv.clients,
		channels:     srv.channels,
		operators:    srv.operators,
		limiter:      srv.limiter,
		channelLimit: srv.config.MaxChannelsPerClient,

		nickRegex:    nickRegex,
		channelRegex: channelRegex,
	}
}

// Serve registers listener's port and accepts connections on it until ctx is canceled.
func (srv *Server) Serve(ctx context.Context, listener net.Listener, isTLS bool) {
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		slog.Error("cant split net host port", "err", err)
	}
	if isTLS {
		srv.addPort(fmt.Sprintf("+%s", port))
	} else {
		srv.addPort(port)
	}

	srv.channels.setRunContext(ctx)
	srv.sweepOnce.Do(func() { go srv.limiter.runSweeper(ctx) })

	acceptLoop(ctx, listener, isTLS, srv.sessionDeps())
}

var errServerShutdown = errors.New("Server shutting down")

// Shutdown terminates all connected clients and blocks until they disconnect or timeout elapses.
func (srv *Server) Shutdown(timeout time.Duration) {
	for _, h := range srv.clients.all() {
		h.terminate(errServerShutdown)
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if srv.clients.count() == 0 {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	slog.Warn("shutdown timed out waiting for sessions to drain")
}
