package ircd

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type client struct {
	mu    *sync.RWMutex
	id    clientID
	ip    string
	nick  string
	user  string
	real  string
	host  string
	modes clientMode

	handshake bool
	ping      int64

	conn   net.Conn
	reader io.Reader

	recv chan string
	send chan string
	stop chan interface{}
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

	return &client{
		mu:        &sync.RWMutex{},
		id:        clientID(id),
		ip:        host,
		nick:      "",
		user:      "",
		real:      "",
		host:      "",
		modes:     0,
		ping:      time.Now().Unix(),
		handshake: false,
		conn:      connection,
		reader:    bufio.NewReader(connection),
		recv:      make(chan string, 1),
		send:      make(chan string, 1),
		stop:      make(chan interface{}),
	}, nil
}

func (c *client) sendRPL(server string, rpl rpl) {
	c.send <- fmt.Sprintf(":%s %s", server, rpl.format())
}

func (c *client) sendNotice(n notice) {
	c.send <- n.format()
}

func (c *client) setPing(ping int64) {
	c.mu.Lock()
	c.ping = ping
	c.mu.Unlock()
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

	return fmt.Sprintf("%s!%s@%s", c.nick, c.user, c.host)
}

func (c *client) addMode(mode clientMode) {
	c.modes |= mode
}

func (c *client) removeMode(mode clientMode) {
	c.modes &= ^mode
}

func (c *client) hasMode(mode clientMode) bool {
	return c.modes&mode != 0
}

func (c *client) write(message string) (int, error) {
	return c.conn.Write([]byte(message))
}
