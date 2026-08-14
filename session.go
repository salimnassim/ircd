package ircd

import (
	"bufio"
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"regexp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

var (
	errSendQExceeded    = errors.New("SendQ exceeded")
	errExcessFlood      = errors.New("Quit: Excess Flood")
	errConnectionClosed = errors.New("Quit: EOF")
)

type sessionSnapshot struct {
	id            clientID
	nick          string
	user          string
	real          string
	host          string
	modes         clientMode
	away          string
	tls           bool
	handshakeDone bool

	channels []string
}

func (s *sessionSnapshot) prefix() string {
	return fmt.Sprintf("%s!%s@%s", s.nick, s.user, s.host)
}

type sessionHandle struct {
	id       clientID
	snapshot atomic.Pointer[sessionSnapshot]
	outbox   chan string

	terminate func(cause error)
}

func (h *sessionHandle) deliver(line string) {
	select {
	case h.outbox <- line:
	default:
		h.terminate(errSendQExceeded)
	}
}

type sessionDeps struct {
	serverName string
	network    string
	version    string
	password   string
	motd       []string
	isupport   string
	ports      func() []string

	pingFrequency  time.Duration
	pongMaxLatency time.Duration

	rateLimit  rate.Limit
	rateBurst  int
	rateGrace  time.Duration
	outboxSize int

	clients   *clientDirectory
	channels  *channelDirectory
	operators *OperatorStore

	nickRegex    *regexp.Regexp
	channelRegex *regexp.Regexp
}

type session struct {
	id      clientID
	address string

	nick, user, real, host string
	modes                  clientMode
	away                   string
	tls                    bool
	handshakeDone          bool
	passwordOK             bool

	memberOf map[string]struct{}

	ponged chan struct{}

	handle *sessionHandle
	deps   sessionDeps
}

func (s *session) prefix() string {
	return fmt.Sprintf("%s!%s@%s", s.nick, s.user, s.host)
}

func (s *session) modestring() string {
	modes := []rune{}
	for r, m := range clientModeMap {
		if s.modes&m != 0 {
			modes = append(modes, r)
		}
	}
	slices.SortFunc(modes, func(a, b rune) int { return cmp.Compare(a, b) })
	return fmt.Sprintf("+%s", string(modes))
}

func (s *session) formatRPL(r rpl) string {
	return fmt.Sprintf(":%s %s", s.deps.serverName, r.rpl())
}

func (s *session) channelDeps() channelActorDeps {
	return channelActorDeps{serverName: s.deps.serverName, clients: s.deps.clients}
}

func (s *session) publishSnapshot() {
	channels := make([]string, 0, len(s.memberOf))
	for name := range s.memberOf {
		channels = append(channels, name)
	}
	s.handle.snapshot.Store(&sessionSnapshot{
		id:            s.id,
		nick:          s.nick,
		user:          s.user,
		real:          s.real,
		host:          s.host,
		modes:         s.modes,
		away:          s.away,
		tls:           s.tls,
		handshakeDone: s.handshakeDone,
		channels:      channels,
	})
}

func isChannelName(target string) bool {
	return strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&")
}

func remoteHost(conn net.Conn) string {
	addr := conn.RemoteAddr()
	if addr == nil {
		return ""
	}
	host, _, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String()
	}
	return host
}

func runSession(parent context.Context, conn net.Conn, id clientID, isTLS bool, deps sessionDeps) {
	ctx, cancel := context.WithCancelCause(parent)

	var once sync.Once
	terminate := func(cause error) {
		once.Do(func() {
			cancel(cause)

			if cause != nil {
				conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
				_, _ = io.WriteString(conn, errorCommand{text: cause.Error()}.command()+"\r\n")
			}

			conn.Close()
		})
	}

	outboxSize := deps.outboxSize
	if outboxSize <= 0 {
		outboxSize = 256
	}

	handle := &sessionHandle{
		id:        id,
		outbox:    make(chan string, outboxSize),
		terminate: terminate,
	}

	sess := &session{
		id:       id,
		address:  remoteHost(conn),
		tls:      isTLS,
		memberOf: make(map[string]struct{}),
		ponged:   make(chan struct{}, 1),
		handle:   handle,
		deps:     deps,
	}
	sess.publishSnapshot()

	deps.clients.register(handle)
	clientGauge.Inc()

	go runWriter(ctx, conn, handle.outbox, terminate)
	go runPingPong(ctx, pingPongDeps{
		serverName:     deps.serverName,
		deliver:        handle.deliver,
		ponged:         sess.ponged,
		pingFrequency:  deps.pingFrequency,
		pongMaxLatency: deps.pongMaxLatency,
		terminate:      terminate,
	})

	sess.readLoop(ctx, conn)

	terminate(errConnectionClosed)

	sess.cleanup(context.Cause(ctx))
}

func runWriter(ctx context.Context, conn net.Conn, outbox <-chan string, terminate func(error)) {
	for {
		select {
		case <-ctx.Done():
			return
		case line, ok := <-outbox:
			if !ok {
				return
			}
			if _, err := io.WriteString(conn, line+"\r\n"); err != nil {
				terminate(errors.New("Quit: Broken pipe"))
				return
			}
		}
	}
}

type pingPongDeps struct {
	serverName     string
	deliver        func(string)
	ponged         <-chan struct{}
	pingFrequency  time.Duration
	pongMaxLatency time.Duration
	terminate      func(error)
}

func runPingPong(ctx context.Context, deps pingPongDeps) {
	pingFrequency := deps.pingFrequency
	if pingFrequency <= 0 {
		pingFrequency = 30 * time.Second
	}
	pongMaxLatency := deps.pongMaxLatency
	if pongMaxLatency <= 0 {
		pongMaxLatency = 10 * time.Second
	}

	pingTimer := time.NewTimer(pingFrequency)
	defer pingTimer.Stop()
	var timeout <-chan time.Time

	for {
		select {
		case <-ctx.Done():
			return
		case <-deps.ponged:
			timeout = nil
			if !pingTimer.Stop() {
				select {
				case <-pingTimer.C:
				default:
				}
			}
			pingTimer.Reset(pingFrequency)
		case <-pingTimer.C:
			deps.deliver(pingCommand{text: deps.serverName}.command())
			timeout = time.After(pongMaxLatency)
		case <-timeout:
			deps.terminate(fmt.Errorf("Quit: Timeout after %d seconds", int(pongMaxLatency.Seconds())))
			return
		}
	}
}

func (s *session) readLoop(ctx context.Context, conn net.Conn) {
	rateLimit := s.deps.rateLimit
	if rateLimit <= 0 {
		rateLimit = rate.Every(300 * time.Millisecond)
	}
	rateBurst := s.deps.rateBurst
	if rateBurst <= 0 {
		rateBurst = 10
	}
	rateGrace := s.deps.rateGrace
	if rateGrace <= 0 {
		rateGrace = 10 * time.Second
	}
	limiter := rate.NewLimiter(rateLimit, rateBurst)

	scanner := bufio.NewScanner(bufio.NewReader(conn))

	for scanner.Scan() {
		waitCtx, cancelWait := context.WithTimeout(ctx, rateGrace)
		err := limiter.Wait(waitCtx)
		cancelWait()
		if err != nil {
			s.handle.terminate(errExcessFlood)
			return
		}

		line := strings.Trim(scanner.Text(), "\r\n")
		parsed, err := parseMessage(line)
		if err != nil || parsed.command == "" {
			continue
		}
		s.dispatch(ctx, parsed)
	}

	if err := scanner.Err(); err != nil {
		s.handle.terminate(fmt.Errorf("Quit: %w", err))
	}
}

func (s *session) cleanup(cause error) {
	reason := "Quit: no reason given"
	if cause != nil {
		reason = cause.Error()
	}

	for name := range s.memberOf {
		s.deps.channels.dispatch(context.Background(), s.channelDeps(), name, evQuit{who: s.handle, reason: reason}, false, "")
	}

	s.deps.clients.unregister(s.id)

	clientGauge.Dec()
}

func (s *session) dispatch(ctx context.Context, m message) {
	commandCounter.WithLabelValues(m.command).Inc()

	switch m.command {
	case "PASS":
		if !s.requireParams(m, 1) {
			return
		}
		s.cmdPass(m)
	case "PING":
		s.cmdPing(m)
	case "PONG":
		s.cmdPong()
	case "NICK":
		if !s.requireParams(m, 1) {
			return
		}
		s.cmdNick(ctx, m)
	case "USER":
		if !s.requireParams(m, 4) {
			return
		}
		s.cmdUser(ctx, m)
	case "LUSERS":
		if !s.requireHandshake() {
			return
		}
		s.cmdLusers()
	case "JOIN":
		if !s.requireHandshake() || !s.requireParams(m, 1) {
			return
		}
		s.cmdJoin(ctx, m)
	case "PART":
		if !s.requireHandshake() || !s.requireParams(m, 1) {
			return
		}
		s.cmdPart(ctx, m)
	case "KICK":
		if !s.requireHandshake() || !s.requireParams(m, 2) {
			return
		}
		s.cmdKick(ctx, m)
	case "TOPIC":
		if !s.requireHandshake() || !s.requireParams(m, 1) {
			return
		}
		s.cmdTopic(ctx, m)
	case "PRIVMSG":
		if !s.requireHandshake() || !s.requireParams(m, 1) {
			return
		}
		s.cmdPrivmsg(ctx, m)
	case "WHOIS":
		if !s.requireHandshake() || !s.requireParams(m, 1) {
			return
		}
		s.cmdWhois(m)
	case "WHO":
		if !s.requireHandshake() {
			return
		}
		s.cmdWho(m)
	case "MODE":
		if !s.requireHandshake() || !s.requireParams(m, 1) {
			return
		}
		s.cmdMode(ctx, m)
	case "AWAY":
		if !s.requireHandshake() {
			return
		}
		s.cmdAway(m)
	case "QUIT":
		s.cmdQuit(m)
	case "OPER":
		if !s.requireHandshake() || !s.requireParams(m, 2) {
			return
		}
		s.cmdOper(m)
	case "VERSION":
		if !s.requireHandshake() {
			return
		}
		s.cmdVersion(m)
	case "LIST":
		if !s.requireHandshake() {
			return
		}
		s.cmdList(m)
	case "INVITE":
		if !s.requireHandshake() || !s.requireParams(m, 2) {
			return
		}
		s.cmdInvite(ctx, m)
	case "KILL":
		if !s.requireHandshake() || !s.requireParams(m, 2) {
			return
		}
		s.cmdKill(m)
	}
}

func (s *session) requireParams(m message, n int) bool {
	if len(m.params) < n {
		s.handle.deliver(s.formatRPL(errNeedMoreParams{client: s.nick, command: m.command}))
		return false
	}
	return true
}

func (s *session) requireHandshake() bool {
	if !s.handshakeDone {
		s.handle.deliver(s.formatRPL(errNotRegistered{client: s.nick}))
		return false
	}
	return true
}
