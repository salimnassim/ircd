package ircd

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Client struct {
	mu       *sync.RWMutex
	id       string
	ip       string
	nickname string
	username string
	realname string
	hostname string

	handshake bool
	ping      int64

	conn   net.Conn
	reader io.Reader

	recv chan string
	send chan string
	stop chan interface{}
}

func (client *Client) String() string {
	return fmt.Sprintf("id: %s, nickname: %s, username: %s, realname: %s, hostname: %s, handshake: %t",
		client.id, client.nickname, client.username, client.realname, client.hostname, client.handshake)
}

func NewClient(connection net.Conn, id string) (*Client, error) {
	ip := strings.Split(connection.RemoteAddr().String(), ":")[0]

	return &Client{
		mu:        &sync.RWMutex{},
		id:        id,
		ip:        ip,
		nickname:  "",
		username:  "",
		realname:  "",
		hostname:  "",
		ping:      time.Now().Unix(),
		handshake: false,
		conn:      connection,
		reader:    bufio.NewReader(connection),
		recv:      make(chan string, 1),
		send:      make(chan string, 1),
		stop:      make(chan interface{}),
	}, nil
}

func (client *Client) SetPing(ping int64) {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.ping = ping
}

func (client *Client) SetHostname(hostname string) {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.hostname = hostname
}

func (client *Client) SetNickname(nickname string) {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.nickname = nickname
}

func (client *Client) SetUsername(username string, realname string) {
	client.mu.Lock()
	defer client.mu.Unlock()

	client.username = username
	client.realname = realname
}

func (client *Client) Nickname() string {
	client.mu.RLock()
	defer client.mu.RUnlock()

	return client.nickname
}

func (client *Client) Prefix() string {
	client.mu.RLock()
	defer client.mu.RUnlock()

	return fmt.Sprintf("%s!%s@%s", client.nickname, client.username, client.hostname)
}

func (client *Client) Write(message string) (int, error) {
	return client.conn.Write([]byte(message))
}
