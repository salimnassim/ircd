package ircd

func handleUser(s *server, c clienter, m message) {
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

	c.setUser(username, realname)

	if !c.handshake() && c.nickname() != "" && c.username() != "" {
		handleHandshake(s, c)
	}
}
