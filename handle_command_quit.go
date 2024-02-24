package ircd

import (
	"strings"
)

func handleQuit(s *server, c *client, m message) {
	reason := "no reason given"
	if len(m.params) >= 1 {
		reason = strings.Join(m.params[0:len(m.params)], " ")
	}

	for _, ch := range s.Channels.memberOf(c) {
		ch.broadcastCommand(partCommand{
			prefix:  c.prefix(),
			channel: ch.name,
			text:    reason,
		}, c.clientID, false)
	}

	c.stop <- "quit"
}
