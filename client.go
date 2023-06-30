package ircd

import (
	"fmt"
	"net"
)

type Client struct {
	Nickname   string
	Username   string
	Hostname   string
	Invisible  bool
	Connection net.Conn
	Handshake  bool
	In         chan string
	Out        chan string
}

func NewClient(connection net.Conn) (*Client, error) {
	return &Client{
		Nickname:   "",
		Username:   "",
		Hostname:   "",
		Connection: connection,
		Handshake:  false,
		In:         make(chan string),
		Out:        make(chan string, 1),
	}, nil
}

func (client *Client) Target() string {
	return fmt.Sprintf("%s!%s@%s", client.Nickname, client.Username, client.Hostname)
}

func (client *Client) Close() error {
	err := client.Connection.Close()
	if err != nil {
		return err
	}
	close(client.In)
	close(client.Out)
	return nil
}
