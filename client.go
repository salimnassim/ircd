package ircd

import (
	"fmt"
	"net"
	"strings"
	"sync"
)

type Client struct {
	mu         *sync.RWMutex
	id         string
	nickname   string
	username   string
	realname   string
	hostname   string
	connection net.Conn
	handshake  bool
	recv       chan string
	send       chan string
}

func (client *Client) String() string {
	return fmt.Sprintf("id: %s, nickname: %s, username: %s, realname: %s, hostname: %s, handshake: %t",
		client.id, client.nickname, client.username, client.realname, client.hostname, client.handshake)
}

func NewClient(connection net.Conn, id string) (*Client, error) {
	return &Client{
		mu:         &sync.RWMutex{},
		id:         id,
		nickname:   "",
		username:   "",
		realname:   "",
		hostname:   "",
		connection: connection,
		handshake:  false,
		recv:       make(chan string),
		send:       make(chan string, 1),
	}, nil
}

func (client *Client) IP() string {
	return strings.Split(client.connection.RemoteAddr().String(), ":")[0]
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

func (client *Client) GetNickname() string {
	client.mu.RLock()
	defer client.mu.RUnlock()

	nickname := client.nickname
	return nickname
}

func (client *Client) Prefix() string {
	client.mu.RLock()
	defer client.mu.RUnlock()

	return fmt.Sprintf("%s!%s@%s", client.nickname, client.username, client.hostname)
}

func (client *Client) Write(message string) (int, error) {
	n, err := client.connection.Write([]byte(message))
	return n, err
}

func (client *Client) Close() error {
	err := client.connection.Close()
	if err != nil {
		return err
	}

	return nil
}
