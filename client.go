package ircd

import (
	"fmt"
	"net"
	"strings"
)

type Client struct {
	ID         string
	Nickname   string
	Username   string
	Realname   string
	hostname   string
	Invisible  bool
	connection net.Conn
	Handshake  bool
	In         chan string
	Out        chan string
}

func NewClient(connection net.Conn, id string) (*Client, error) {
	return &Client{
		ID:         id,
		Nickname:   "",
		Username:   "",
		Realname:   "",
		hostname:   "",
		connection: connection,
		Handshake:  false,
		In:         make(chan string),
		Out:        make(chan string, 1),
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
	return fmt.Sprintf("%s!%s@%s", client.Nickname, client.Username, client.Hostname())
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
	close(client.In)
	close(client.Out)
	return nil
}
