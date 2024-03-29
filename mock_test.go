package ircd

import (
	"cmp"
	"fmt"
	"net"
	"slices"
	"time"
)

type connMock struct {
	deadline time.Time
}

func (c *connMock) Read(b []byte) (n int, err error)   { return 0, nil }
func (c *connMock) SetDeadline(t time.Time) error      { c.deadline = t; return nil }
func (c *connMock) SetReadDeadline(t time.Time) error  { c.deadline = t; return nil }
func (c *connMock) SetWriteDeadline(t time.Time) error { c.deadline = t; return nil }
func (c *connMock) Write(b []byte) (int, error)        { return 0, nil }
func (c *connMock) Close() error                       { return nil }
func (C *connMock) LocalAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.IPv4(0x7F, 0x00, 0x00, 0x01),
		Port: 5000,
	}
}
func (C *connMock) RemoteAddr() net.Addr {
	return &net.TCPAddr{
		IP:   net.IPv4(0x7F, 0x00, 0x00, 0x01),
		Port: 5001,
	}
}

type clientMock struct {
	messagesIn   []string
	messagesOut  []string
	messagesPong []bool
	messagesKill []string

	clientID clientID
	addr     string

	nick   string
	user   string
	real   string
	host   string
	secure bool
	afk    string
	hs     bool
	pw     bool
	modes  clientMode
	q      string
}

func newMockClient(handshake bool) *clientMock {
	return &clientMock{
		messagesIn:   []string{},
		messagesOut:  []string{},
		messagesPong: []bool{},
		messagesKill: []string{},
		clientID:     "12345",
		addr:         "127.0.0.1",
		nick:         "mocknick",
		user:         "mockuser",
		real:         "mockreal",
		host:         "mockhost",
		secure:       false,
		afk:          "",
		hs:           handshake,
		pw:           false,
		modes:        0,
	}
}

// Reset mock message state to empty slices.
func (c *clientMock) reset() {
	c.messagesIn = []string{}
	c.messagesOut = []string{}
	c.messagesPong = []bool{}
	c.messagesKill = []string{}
}

func (c *clientMock) String() string {
	return string(c.clientID)
}

func (c *clientMock) id() clientID {
	return c.clientID
}

func (c *clientMock) ip() string {
	return c.addr
}

func (c *clientMock) nickname() string {
	return c.nick
}

func (c *clientMock) setNickname(nickname string) {
	c.nick = nickname
}

func (c *clientMock) username() string {
	return c.user
}

func (c *clientMock) realname() string {
	return c.real
}

func (c *clientMock) setUser(username string, realname string) {
	c.user = username
	c.real = realname
}

func (c *clientMock) hostname() string {
	return c.host
}

func (c *clientMock) setHostname(hostname string) {
	c.host = hostname
}

func (c *clientMock) tls() bool {
	return c.secure
}

func (c *clientMock) setTLS(tls bool) {
	c.secure = tls
}

func (c *clientMock) away() string {
	return c.afk
}

func (c *clientMock) setAway(text string) {
	c.afk = text
}

func (c *clientMock) handshake() bool {
	return c.hs
}

func (c *clientMock) setHandshake(handshake bool) {
	c.hs = handshake
}

func (c *clientMock) password() bool {
	return c.pw
}

func (c *clientMock) setPassword(correct bool) {
	c.pw = correct
}

func (c *clientMock) prefix() string {
	return fmt.Sprintf("%s!%s@%s", c.nickname(), c.username(), c.hostname())
}

func (c *clientMock) modestring() string {
	modes := []rune{}
	for m, r := range clientModeMap {
		if c.hasMode(r) {
			modes = append(modes, m)
		}
	}
	slices.SortFunc[[]rune, rune](modes, func(a rune, b rune) int {
		return cmp.Compare(a, b)
	})
	return fmt.Sprintf("+%s", string(modes))
}

func (c *clientMock) addMode(mode clientMode) {
	if c.hasMode(mode) {
		return
	}
	c.modes |= mode
}

func (c *clientMock) removeMode(mode clientMode) {
	if !c.hasMode(mode) {
		return
	}
	c.modes &= ^mode
}

func (c *clientMock) hasMode(mode clientMode) bool {
	return c.modes&mode != 0
}

func (c *clientMock) sendRPL(serverName string, rpl rpl) {
	c.messagesOut = append(c.messagesOut, rpl.rpl())
}

func (c *clientMock) sendCommand(command command) {
	c.messagesOut = append(c.messagesOut, command.command())
}

func (c *clientMock) quitReason() string {
	return c.q
}

func (c *clientMock) setQuitreason(reason string) {
	c.q = reason
}

func (c *clientMock) send(text string) {
	c.messagesOut = append(c.messagesOut, text)
}

func (c *clientMock) pong(pong bool) {
	c.messagesPong = append(c.messagesPong, pong)
}

func (c *clientMock) kill(reason string) {
	c.messagesKill = append(c.messagesKill, reason)
}
