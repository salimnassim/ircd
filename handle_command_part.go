package ircd

import (
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handlePart(s *server, c clienter, m message) {
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

		ch, exists := s.Channels.get(target)
		if !exists {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		ch.clients().remove(c)

		ch.broadcastCommand(partCommand{
			prefix:  c.prefix(),
			channel: ch.name(),
			text:    reason,
		}, c.id(), false)

		if ch.clients().count() == 0 {
			s.Channels.delete(ch.name())
			metrics.Channels.Dec()
		}
	}
}
