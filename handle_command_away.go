package ircd

import "strings"

func handleAway(s *server, c *client, m message) {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return
	}

	// unaway
	if len(m.params) == 0 {
		if c.away() != "" {
			c.setAway("")
		}
		c.sendRPL(s.name, rplUnAway{
			client: c.nickname(),
		})
		return
	}

	// away
	text := strings.Join(m.params[0:len(m.params)], " ")
	c.setAway(text)

	c.sendRPL(s.name, rplNowAway{
		c.nickname(),
	})
}
