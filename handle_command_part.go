package ircd

import (
	"fmt"
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handlePart(s *server, c *client, m message) {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return
	}

	targets := strings.Split(m.params[0], ",")

	reason := "no reason given"
	if len(m.params) >= 1 {
		reason = strings.Join(m.params[0:len(m.params)], " ")
	}

	for _, target := range targets {
		if !m.isTargetChannel() {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		// try to get channel
		channel, exists := s.channels.get(target)
		if !exists {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		// remove client
		channel.removeClient(c)

		// broadcast that user has left the channel
		channel.broadcast(
			fmt.Sprintf(
				":%s PART %s :Part: %s",
				c.prefix(), target, reason,
			),
			c.id, false)

		if channel.clients.count() == 0 {
			s.channels.delete(channel.name)
			metrics.Channels.Dec()
		}
	}
}
