package ircd

import "strings"

func handleKick(s *server, c clienter, m message) {
	channel := m.params[0]

	ch, ok := s.Channels.get(channel)
	if !ok {
		c.sendRPL(s.name, errNoSuchChannel{
			client:  c.nickname(),
			channel: channel,
		})
		return
	}

	if !ch.clients().isMember(c) {
		// send 403 if channel is secret
		if ch.hasMode(modeChannelSecret) {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: channel,
			})
			return
		}

		c.sendRPL(s.name, errNotOnChannel{
			client:  c.nickname(),
			channel: channel,
		})
		return
	}

	// user has to be halfop, op, admin or owner
	if !ch.clients().hasMode(c, modeHalfOperator, modeOperator, modeAdmin, modeOwner) {
		c.sendRPL(s.name, errChanoPrivsNeeded{
			client:  c.nickname(),
			channel: ch.name(),
		})
		return
	}

	targets := strings.Split(m.params[1], ",")
	for _, target := range targets {
		tc, ok := s.Clients.get(target)
		// no matching client found
		if !ok {
			c.sendRPL(s.name, errUserNotInChannel{
				client:  c.nickname(),
				nick:    target,
				channel: ch.name(),
			})
			continue
		}

		// client must be in channel
		if !ch.clients().isMember(c) {
			c.sendRPL(s.name, errUserNotInChannel{
				client:  c.nickname(),
				nick:    target,
				channel: ch.name(),
			})
			continue
		}

		reason := "No reason given."
		if len(m.params) >= 2 {
			reason = strings.Join(m.params[1:len(m.params)], " ")
		}

		ch.broadcastCommand(kickCommand{
			prefix:  c.prefix(),
			channel: ch.name(),
			target:  tc.nickname(),
			reason:  reason,
		}, c.id(), false)
		ch.clients().delete(tc)
	}
}
