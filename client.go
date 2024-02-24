package ircd

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
)

type client struct {
	mu *sync.RWMutex

	alive bool
	id    clientID
	ip    string
	nick  string
	user  string
	real  string
	host  string
	modes clientMode
	tls   bool
	afk   string

	handshake bool

	conn   net.Conn
	reader io.Reader

	recv chan string
	send chan string
	stop chan string
	pong chan bool
}

func (c *client) String() string {
	return fmt.Sprintf("id: %s, nickname: %s, username: %s, realname: %s, hostname: %s, handshake: %t",
		c.id, c.nick, c.user, c.real, c.host, c.handshake)
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
		mu:    &sync.RWMutex{},
		alive: true,
		id:    clientID(id),
		ip:    host,
		nick:  "",
		user:  "",
		real:  "",
		host:  "",
		modes: 0,
		tls:   false,
		afk:   "",

		handshake: false,

		conn:   connection,
		reader: bufio.NewReader(connection),

		recv: make(chan string),
		send: make(chan string),
		stop: make(chan string),
		pong: make(chan bool),
	}

	if port == os.Getenv("PORT_TLS") {
		client.tls = true
	}

	return client, nil
}

func (c *client) sendRPL(server string, rpl rpl) {
	c.send <- fmt.Sprintf(":%s %s", server, rpl.format())
}

func (c *client) sendCommand(cmd command) {
	c.send <- cmd.command()
}

func (c *client) setAway(text string) {
	c.mu.Lock()
	c.afk = text
	c.mu.Unlock()
}

func (c *client) away() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.afk
}

func (c *client) setHostname(hostname string) {
	c.mu.Lock()
	c.host = hostname
	c.mu.Unlock()
}

func (c *client) setNickname(nickname string) {
	c.mu.Lock()
	c.nick = nickname
	c.mu.Unlock()
}

func (c *client) setUsername(username string, realname string) {
	c.mu.Lock()
	c.user = username
	c.real = realname
	c.mu.Unlock()
}

func (c *client) nickname() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.nick
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

func (c *client) hostname() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.host
}

func (c *client) prefix() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return fmt.Sprintf("%s!%s@%s", c.nickname(), c.username(), c.hostname())
}

func (c *client) modestring() string {
	modes := []rune{}
	for m, r := range clientModeMap {
		if c.hasMode(r) {
			modes = append(modes, m)
		}
	}
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

func (c *client) write(message string) (int, error) {
	return c.conn.Write([]byte(message))
}
