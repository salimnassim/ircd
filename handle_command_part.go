package ircd

import (
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handlePart(s *server, c *client, m message) {
	targets := strings.Split(m.params[0], ",")

	reason := "no reason given"
	if len(m.params) >= 1 {
		reason = strings.Join(m.params[1:len(m.params)], " ")
	}

	for _, target := range targets {
		if !m.isTargetChannel() {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		// try to get ch
		ch, exists := s.Channels.get(target)
		if !exists {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		// remove client
		ch.removeClient(c)

		// broadcast that user has left the channel
		ch.broadcastCommand(partCommand{
			prefix:  c.prefix(),
			channel: ch.name,
			text:    reason,
		}, c.clientID, false)

		if ch.clients.count() == 0 {
			s.Channels.delete(ch.name)
			metrics.Channels.Dec()
		}
	}
}
