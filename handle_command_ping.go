package ircd

import (
	"strings"
)

func handlePing(s *server, c *client, m message) {
	c.out <- strings.Replace(m.raw, "PING", "PONG", 1)
}
