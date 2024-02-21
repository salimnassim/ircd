package ircd

import (
	"fmt"
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handlePart(s *server, c *client, m Message) {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return
	}

	targets := strings.Split(m.Params[0], ",")

	for _, target := range targets {
		// try to get channel
		channel, exists := s.channels.Get(target)
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
		channel.broadcast(fmt.Sprintf(":%s PART %s :<no reason given>",
			c.prefix(), target), c.id, false)

		if channel.clients.Count() == 0 {
			s.channels.Delete(channel.name)
			metrics.Channels.Dec()
		}
	}
}
