package ircd

import (
	"net"
)

type Client struct {
	Nickname   string
	Username   string
	Connection net.Conn
	Handshake  bool
	channels   map[string]bool
	In         chan string
	Out        chan string
}

func NewClient(connection net.Conn) (*Client, error) {
	return &Client{
		Nickname:   "",
		Username:   "",
		Connection: connection,
		Handshake:  false,
		channels:   make(map[string]bool),
		In:         make(chan string),
		Out:        make(chan string, 1),
	}, nil
}
