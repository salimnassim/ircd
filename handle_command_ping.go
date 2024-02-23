package ircd

import (
	"strings"
)

func handlePing(s *server, c *client, m message) {
	c.send <- strings.Replace(m.raw, "PING", "PONG", 1)
}
