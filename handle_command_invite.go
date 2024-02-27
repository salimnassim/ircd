package ircd

func handleInvite(s *server, c clienter, m message) {
	nickname := m.params[0]
	channel := m.params[1]

	// get target client
	tc, ok := s.Clients.get(nickname)
	if !ok {
		c.sendRPL(s.name, errNoSuchNick{
			client: c.nickname(),
			nick:   nickname,
		})
		return
	}

	// get channel
	ch, ok := s.Channels.get(channel)
	if !ok {
		c.sendRPL(s.name, errNoSuchChannel{
			client:  c.nickname(),
			channel: channel,
		})
	}

	// if channel has invite only, client has to be hop or greater
	if ch.hasMode(modeChannelInviteOnly) && !ch.clients().hasMode(c, modeMemberHalfOperator, modeMemberOperator, modeMemberAdmin, modeMemberOwner) {
		c.sendRPL(s.name, errChanoPrivsNeeded{
			client:  c.nickname(),
			channel: ch.name(),
		})
		return
	}

	// inviter has to be a member of the channel
	if !ch.clients().isMember(c) {
		c.sendRPL(s.name, errNotOnChannel{
			client:  c.nickname(),
			channel: ch.name(),
		})
		return
	}

	// user cant be already a member of the channel
	if ch.clients().isMember(tc) {
		c.sendRPL(s.name, errUserOnChannel{
			client:  c.nickname(),
			nick:    tc.nickname(),
			channel: ch.name(),
		})
		return
	}

	// add to channel invites map
	ch.addInvite(tc.id())

	// send rpl to inviter
	c.sendRPL(s.name, rplInviting{
		client:  c.nickname(),
		nick:    tc.nickname(),
		channel: ch.name(),
	})

	// send invite command to target client
	tc.sendCommand(inviteCommand{
		prefix:  c.prefix(),
		target:  tc.nickname(),
		channel: ch.name(),
	})
}
