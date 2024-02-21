package ircd

func handleWhois(s *server, c *client, m message) {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return
	}

	target := m.Params[0]
	who, exists := s.clients.get(target)
	if who == nil || !exists {
		c.sendRPL(s.name, errNoSuchNick{
			client: c.nickname(),
			nick:   target,
		})
		return
	}

	// https://modern.ircdocs.horse/#rplwhoisuser-311
	c.sendRPL(s.name, rplWhoisUser{
		client:   c.nick,
		nick:     who.nickname(),
		username: who.username(),
		host:     who.hostname(),
		realname: who.realname(),
	})

	channels := []string{}
	memberOf := s.channels.memberOf(who)
	for _, c := range memberOf {
		if !c.secret {
			channels = append(channels, c.name)
		}
	}

	c.sendRPL(s.name, rplWhoisChannels{
		client:   c.nickname(),
		nick:     who.nickname(),
		channels: channels,
	})
}
