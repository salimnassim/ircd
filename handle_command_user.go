package ircd

func handleUser(s *server, c *client, m message) {
	if len(m.params) < 4 {
		c.sendRPL(s.name, errNeedMoreParams{
			client: c.nickname(),
		})
		return
	}

	if c.username() != "" {
		c.sendRPL(s.name, errAlreadyRegistered{
			client: c.nickname(),
		})
		return
	}

	username := m.params[0]
	realname := m.params[3]

	c.setUsername(username, realname)

	if !c.hs && c.nickname() != "" && c.username() != "" {
		handleHandshake(s, c)
	}
}
