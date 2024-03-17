package ircd

import (
	"strings"
)

func handleQuit(s *server, c clienter, m message) {
	reason := "no reason given"
	if len(m.params) > 0 {
		reason = strings.Join(m.params[0:len(m.params)], " ")
		reason = strings.TrimSpace(reason)
	}
	c.kill(reason)
}
