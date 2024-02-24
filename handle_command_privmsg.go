package ircd

import (
	"strings"

	"github.com/salimnassim/ircd/metrics"
)

func handlePrivmsg(s *server, c *client, m message) {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return
	}

	targets := strings.Split(m.params[0], ",")
	text := strings.Join(m.params[1:len(m.params)], " ")

	for _, target := range targets {
		// is channel
		if m.isTargetChannel() {
			channel, exists := s.Channels.get(target)
			if !exists {
				c.sendRPL(s.name, errNoSuchChannel{
					client:  c.nickname(),
					channel: target,
				})
				continue
			}

			// is user a member of the channel?
			if !s.Channels.isMember(c, channel) {
				c.sendRPL(s.name, errNotOnChannel{
					client:  c.nickname(),
					channel: channel.name,
				})
				continue
			}

			channel.broadcastCommand(privmsgCommand{
				prefix: c.prefix(),
				target: channel.name,
				text:   text,
			}, c.id, true)
			continue
		}

		// is user
		dest, exists := s.Clients.get(target)
		if dest == nil || !exists {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		// is away?
		if dest.away() != "" {
			dest.sendRPL(s.name, rplAway{
				client:  c.nickname(),
				nick:    dest.nickname(),
				message: dest.away(),
			})
		}

		dest.sendCommand(privmsgCommand{
			prefix: c.nick,
			target: dest.nickname(),
			text:   text,
		})

		metrics.PrivmsgClient.Inc()
		continue
	}
}
