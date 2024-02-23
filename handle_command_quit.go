package ircd

import (
	"fmt"
	"strings"
)

func handleQuit(s *server, c *client, m message) {
	reason := "no reason given"
	if len(m.params) >= 1 {
		reason = strings.Join(m.params[0:len(m.params)], " ")
	}

	for _, ch := range s.channels.memberOf(c) {
		ch.broadcast(
			fmt.Sprintf(":%s PART %s :Quit: %s", c.prefix(), ch.name, reason),
			clientID(c.nickname()),
			false,
		)
	}

	c.stop <- true
}
