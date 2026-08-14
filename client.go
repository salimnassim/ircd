package ircd

import (
	"cmp"
	"fmt"
	"net"
	"os"
	"slices"
	"sync"
)

type clienter interface {
	String() string

	id() clientID

	ip() string

	nickname() string

	setNickname(nickname string)

	username() string

	realname() string

	setUser(username string, realname string)

	hostname() string

	setHostname(hostname string)

	tls() bool

	setTLS(tls bool)

	away() string

	setAway(text string)

	handshake() bool

	setHandshake(handshake bool)

	password() bool

	setPassword(correct bool)

	prefix() string

	modestring() string

	addMode(mode clientMode)

	removeMode(mode clientMode)

	hasMode(mode clientMode) bool

	sendRPL(serverName string, rpl rpl)

	sendCommand(command command)

	quitReason() string

	setQuitreason(reason string)

	send(text string)

	pong(pong bool)

	kill(reason string)
}

type client struct {
	mu *sync.RWMutex

	clientID clientID
	address  string
	nick     string
	user     string
	real     string
	host     string
	modes    clientMode

	secure bool
	afk    string

	o bool

	hs bool

	pw bool

	q string

	conn   net.Conn
	in     chan string
	out    chan string
	ponged chan bool

	killIn   chan bool
	killOut  chan bool
	killPong chan bool
}

func newClient(connection net.Conn, id string) (*client, error) {
	if connection == nil {
		return nil, errorConnectionNil
	}

	if connection.RemoteAddr() == nil {
		return nil, errorConnectionRemoteAddressNil
	}

	host, _, err := net.SplitHostPort(connection.RemoteAddr().String())
	if err != nil {
		return nil, err
	}

	if connection.LocalAddr() == nil {
		return nil, errorConnectionLocalAddressNil
	}

	_, port, err := net.SplitHostPort(connection.LocalAddr().String())
	if err != nil {
		return nil, err
	}

	client := &client{
		mu:       &sync.RWMutex{},
		clientID: clientID(id),
		address:  host,
		nick:     "",
		user:     "",
		real:     "",
		host:     "",
		modes:    0,
		secure:   false,
		afk:      "",
		o:        false,

		hs: false,

		conn: connection,

		in:     make(chan string, 1),
		out:    make(chan string, 1),
		ponged: make(chan bool, 1),

		killIn:   make(chan bool, 1),
		killOut:  make(chan bool, 1),
		killPong: make(chan bool, 1),
	}

	if port == os.Getenv("PORT_TLS") {
		client.secure = true
	}

	return client, nil
}

func (c *client) String() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("id: %s, nickname: %s, username: %s, realname: %s, hostname: %s, handshake: %t",
		c.clientID, c.nick, c.user, c.real, c.host, c.hs)
}

func (c *client) id() clientID {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.clientID
}

func (c *client) ip() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.address
}

func (c *client) nickname() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.nick
}

func (c *client) setNickname(nickname string) {
	c.mu.Lock()
	c.nick = nickname
	c.mu.Unlock()
}

func (c *client) username() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.user
}

func (c *client) realname() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.real
}

func (c *client) setUser(username string, realname string) {
	c.mu.Lock()
	c.user = username
	c.real = realname
	c.mu.Unlock()
}

func (c *client) hostname() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.host
}

func (c *client) setHostname(hostname string) {
	c.mu.Lock()
	c.host = hostname
	c.mu.Unlock()
}

func (c *client) tls() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.secure
}

func (c *client) setTLS(tls bool) {
	c.mu.Lock()
	c.secure = true
	c.mu.Unlock()
}

func (c *client) away() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.afk
}

func (c *client) setAway(text string) {
	c.mu.Lock()
	c.afk = text
	c.mu.Unlock()
}

func (c *client) handshake() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hs
}

func (c *client) setHandshake(handshake bool) {
	c.mu.Lock()
	c.hs = handshake
	c.mu.Unlock()
}

func (c *client) password() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.pw
}

func (c *client) setPassword(correct bool) {
	c.mu.Lock()
	c.pw = correct
	c.mu.Unlock()
}

func (c *client) prefix() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return fmt.Sprintf("%s!%s@%s", c.nick, c.user, c.host)
}

func (c *client) modestring() string {
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

func (c *client) addMode(mode clientMode) {
	if c.hasMode(mode) {
		return
	}
	c.mu.Lock()
	c.modes |= mode
	c.mu.Unlock()
}

func (c *client) removeMode(mode clientMode) {
	if !c.hasMode(mode) {
		return
	}
	c.mu.Lock()
	c.modes &= ^mode
	c.mu.Unlock()
}

func (c *client) hasMode(mode clientMode) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.modes&mode != 0
}

func (c *client) sendRPL(server string, rpl rpl) {
	c.out <- fmt.Sprintf(":%s %s", server, rpl.rpl())
}

func (c *client) sendCommand(cmd command) {
	c.out <- cmd.command()
}

func (c *client) quitReason() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.q
}

func (c *client) setQuitreason(reason string) {
	c.mu.Lock()
	c.q = reason
	c.mu.Unlock()
}

func (c *client) send(text string) {
	c.out <- text
}

func (c *client) pong(pong bool) {
	c.ponged <- pong
}

func (c *client) kill(reason string) {
	c.setQuitreason(reason)

	go func() {
		c.killIn <- true
		c.killOut <- true
		c.killPong <- true
	}()
}
