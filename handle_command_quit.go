package ircd

import (
	"fmt"
	"strings"
)

func handleQuit(s *server, c clienter, m message) {
	reason := "no reason given"
	if len(m.params) >= 1 {
		reason = strings.Join(m.params[0:len(m.params)], " ")
		reason = strings.TrimSpace(reason)
	}

	for _, ch := range s.Channels.memberOf(c) {
		ch.broadcastCommand(partCommand{
			prefix:  c.prefix(),
			channel: ch.name(),
			text:    fmt.Sprintf("Quit: %s", reason),
		}, c.id(), false)
	}

	c.kill("quit")
}
