package ircd

func handleInvite(s *server, c clienter, m message) {
	nickname := m.params[0]
	channel := m.params[1]

	tc, ok := s.Clients.get(nickname)
	if !ok {
		c.sendRPL(s.name, errNoSuchNick{
			client: c.nickname(),
			nick:   nickname,
		})
		return
	}

	ch, ok := s.Channels.get(channel)
	if !ok {
		c.sendRPL(s.name, errNoSuchChannel{
			client:  c.nickname(),
			channel: channel,
		})
	}

	if ch.hasMode(modeChannelInviteOnly) && !ch.clients().hasMode(c, modeMemberHalfOperator, modeMemberOperator, modeMemberAdmin, modeMemberOwner) {
		c.sendRPL(s.name, errChanoPrivsNeeded{
			client:  c.nickname(),
			channel: ch.name(),
		})
		return
	}

	if !ch.clients().isMember(c) {
		c.sendRPL(s.name, errNotOnChannel{
			client:  c.nickname(),
			channel: ch.name(),
		})
		return
	}

	if ch.clients().isMember(tc) {
		c.sendRPL(s.name, errUserOnChannel{
			client:  c.nickname(),
			nick:    tc.nickname(),
			channel: ch.name(),
		})
		return
	}

	ch.addInvite(tc.id())

	c.sendRPL(s.name, rplInviting{
		client:  c.nickname(),
		nick:    tc.nickname(),
		channel: ch.name(),
	})

	tc.sendCommand(inviteCommand{
		prefix:  c.prefix(),
		target:  tc.nickname(),
		channel: ch.name(),
	})
}
