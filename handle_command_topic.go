package ircd

import (
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handleTopic(s *server, c *client, m message) {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return
	}

	target := m.params[0]

	// channel name max length is 50, check for allowed channel prefixes
	if !(strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&")) {
		c.sendRPL(s.name, errNoSuchChannel{
			client:  c.nickname(),
			channel: target,
		})
		return
	}

	// try to get channel
	channel, exists := s.channels.get(target)
	if !exists {
		c.sendRPL(s.name, errNoSuchChannel{
			client:  c.nickname(),
			channel: target,
		})
		return
	}

	// set topic
	remainder := strings.Join(m.params[1:len(m.params)], " ")
	channel.setTopic(remainder, c.nick)
	metrics.Topic.Inc()

	// get topic
	topic := channel.topic()

	// broadcast new topic to clients on channel
	channel.broadcastRPL(
		rplTopic{
			client:  c.nickname(),
			channel: channel.name,
			topic:   topic.text,
		},
		c.id,
		false,
	)
}
