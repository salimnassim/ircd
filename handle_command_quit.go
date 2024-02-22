package ircd

import (
	"fmt"
	"strings"
)

func handleQuit(s *server, c *client, m message) {
	reason := "leaving"
	if len(m.params) >= 1 {
		reason = strings.Join(m.params[0:len(m.params)], " ")
	}

	for _, ch := range s.channels.memberOf(c) {
		ch.broadcast(
			fmt.Sprintf(":%s QUIT :%s", c.prefix(), reason),
			clientID(c.nickname()),
			false,
		)
	}

	c.stop <- true
}
