package ircd

func handleUser(s *server, c *client, m message) {
	if !c.handshake {
		c.sendRPL(s.name, errNotRegistered{
			client: c.nickname(),
		})
		return
	}

	if len(m.params) < 4 {
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

	username := m.params[0]
	realname := m.params[3]

	c.setUsername(username, realname)
}
