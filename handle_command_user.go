package ircd

func handleUser(s *server, c *client, m Message) {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return
	}

	if len(m.Params) < 4 {
		c.sendRPL(s.name, errNeedMoreParams{
			client: c.nick,
		})
		return
	}

	if c.user != "" {
		c.sendRPL(s.name, errAlreadyRegistered{
			client: c.nickname(),
		})
		return
	}

	username := m.Params[0]
	realname := m.Params[3]

	c.setUsername(username, realname)
}
