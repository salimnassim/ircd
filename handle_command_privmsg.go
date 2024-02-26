package ircd

import (
	"strings"
)

func handlePrivmsg(s *server, c clienter, m message) {
	targets := strings.Split(m.params[0], ",")
	text := strings.Join(m.params[1:len(m.params)], " ")

	for _, target := range targets {
		// is channel
		if m.isTargetChannel() {
			ch, exists := s.Channels.get(target)
			if !exists {
				c.sendRPL(s.name, errNoSuchChannel{
					client:  c.nickname(),
					channel: target,
				})
				continue
			}

			// is user a member of the channel?
			if !s.Channels.isMember(c, ch) {
				c.sendRPL(s.name, errNotOnChannel{
					client:  c.nickname(),
					channel: ch.name(),
				})
				continue
			}

			if ch.hasMode(modeChannelModerated) && !ch.clients().hasMode(c, modeVoice, modeHalfOperator, modeOperator, modeAdmin, modeOwner) {
				c.sendRPL(s.name, errCannotSendToChan{
					client:  c.nickname(),
					channel: ch.name(),
					text:    "Channel is moderated.",
				})
				return
			}

			ch.broadcastCommand(privmsgCommand{
				prefix: c.prefix(),
				target: ch.name(),
				text:   text,
			}, c.id(), true)
			continue
		}

		// is user
		tc, exists := s.Clients.get(target)
		if tc == nil || !exists {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			continue
		}

		// is away?
		if tc.away() != "" {
			tc.sendRPL(s.name, rplAway{
				client:  c.nickname(),
				nick:    tc.nickname(),
				message: tc.away(),
			})
		}

		tc.sendCommand(privmsgCommand{
			prefix: c.nickname(),
			target: tc.nickname(),
			text:   text,
		})
		continue
	}
}
