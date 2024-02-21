package ircd

import (
	"strings"
)

func handleWho(s *server, c *client, m message) {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return
	}

	if len(m.Params) == 0 {
		c.sendRPL(s.name, errNeedMoreParams{
			client:  c.nickname(),
			command: m.Command,
		})
		return
	}

	target := m.Params[0]
	if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&") {
		channel, ok := s.channels.get(target)
		if !ok {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			return
		}

		for _, cl := range channel.clients.all() {
			c.sendRPL(s.name, rplWhoReply{
				client:   c.nickname(),
				channel:  channel.name,
				username: cl.username(),
				host:     cl.hostname(),
				server:   s.name,
				nick:     cl.nickname(),
				flags:    "",
				hopcount: 0,
				realname: cl.realname(),
			})
		}

		c.sendRPL(s.name, rplEndOfWho{
			client: c.nickname(),
			mask:   "",
		})
		return
	}

	// todo: support querying users with mask
}
