package ircd

import (
	"fmt"
)

type Server struct {
	Clients  map[*Client]bool
	Channels map[*Client]bool
}

func NewServer() *Server {
	server := &Server{
		Clients:  make(map[*Client]bool),
		Channels: make(map[*Client]bool),
	}
	return server
}

func Line(code string, message string) string {
	return fmt.Sprintf(":%s %d %s\r\n", "localhost", code, message)
}
