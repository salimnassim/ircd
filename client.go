package ircd

import (
	"fmt"
	"net"
	"strings"
)

type Client struct {
	id         string
	nickname   string
	username   string
	realname   string
	hostname   string
	connection net.Conn
	handshake  bool
	in         chan string
	out        chan string
}

func NewClient(connection net.Conn, id string) (*Client, error) {
	return &Client{
		id:         id,
		nickname:   "",
		username:   "",
		realname:   "",
		hostname:   "",
		connection: connection,
		handshake:  false,
		in:         make(chan string),
		out:        make(chan string, 1),
	}, nil
}

func (client *Client) IP() string {
	return strings.Split(client.connection.RemoteAddr().Network(), ":")[0]
}

func (client *Client) SetHostname(hostname string) {
	client.hostname = hostname
}

func (client *Client) Hostname() string {
	return client.hostname
}

func (client *Client) Target() string {
	return fmt.Sprintf("%s!%s@%s", client.nickname, client.username, client.Hostname())
}

func (client *Client) Write(message string) (int, error) {
	n, err := client.connection.Write([]byte(message + "\r\n"))
	return n, err
}

func (client *Client) Close() error {
	err := client.connection.Close()
	if err != nil {
		return err
	}
	close(client.in)
	close(client.out)
	return nil
}
