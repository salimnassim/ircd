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

	// try to get channel
	channel, exists := s.Channels.get(target)
	if !exists {
		c.sendRPL(s.name, errNoSuchChannel{
			client:  c.nickname(),
			channel: target,
		})
		return
	}

	// set topic
	remainder := strings.Join(m.params[1:len(m.params)], " ")
	channel.setTopic(remainder, c.nickname())

	// get topic
	topic := channel.topic()

	// broadcast new topic to clients on channel
	channel.broadcastRPL(
		rplTopic{
			client:  c.nickname(),
			channel: channel.name(),
			topic:   topic.text,
		}, c.id(), false,
	)
}
