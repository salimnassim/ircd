package ircd

import (
	"strings"
)

func handlePing(s *server, c clienter, m message) {
	c.send(strings.Replace(m.raw, "PING", "PONG", 1))
}
