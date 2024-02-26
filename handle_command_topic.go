package ircd

import (
	"strings"
)

func handleTopic(s *server, c clienter, m message) {
	target := m.params[0]

	if !m.isTargetChannel() {
		c.sendRPL(s.name, errNoSuchChannel{
			client:  c.nickname(),
			channel: target,
		})
		return
	}

	ch, exists := s.Channels.get(target)
	if !exists {
		c.sendRPL(s.name, errNoSuchChannel{
			client:  c.nickname(),
			channel: target,
		})
		return
	}

	if ch.hasMode(modeChannelRestrictTopic) && !ch.clients().hasMode(c, modeHalfOperator, modeOperator, modeAdmin, modeOwner) {
		c.sendRPL(s.name, errChanoPrivsNeeded{
			client:  c.nickname(),
			channel: ch.name(),
		})
		return
	}

	// set topic
	text := strings.Join(m.params[1:len(m.params)], " ")
	ch.setTopic(text, c.nickname())

	// get topic
	topic := ch.topic()

	// broadcast new topic to clients on channel
	ch.broadcastRPL(
		rplTopic{
			client:  c.nickname(),
			channel: ch.name(),
			topic:   topic.text,
		}, c.id(), false,
	)
}
