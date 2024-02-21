package ircd

import (
	"fmt"
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handlePrivmsg(s *server, c *client, m Message) {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return
	}

	targets := strings.Split(m.Params[0], ",")
	text := strings.Join(m.Params[1:len(m.Params)], " ")

	for _, target := range targets {
		// is channel
		if strings.HasPrefix(target, "#") || strings.HasPrefix(target, "&") {
			channel, exists := s.channels.Get(target)
			if !exists {
				c.sendRPL(s.name, errNoSuchChannel{
					client:  c.nickname(),
					channel: target,
				})
				continue
			}

			// is user a member of the channel?
			if !s.channels.IsMember(c, channel) {
				c.sendRPL(s.name, errNotOnChannel{
					client:  c.nickname(),
					channel: channel.name,
				})
				continue
			}

			// send message to channel
			channel.broadcast(fmt.Sprintf(":%s PRIVMSG %s :%s",
				c.prefix(), channel.name, text), c.id, true)
			metrics.PrivmsgChannel.Inc()
			continue
		}
		// is user
		dest, exists := s.clients.Get(target)
		if dest == nil || !exists {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		dest.send <- fmt.Sprintf(":%s PRIVMSG %s :%s",
			c.nick, dest.nick, text)
		metrics.PrivmsgClient.Inc()
		continue
	}
}
