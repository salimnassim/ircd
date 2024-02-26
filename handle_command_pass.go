package ircd

func handlePass(s *server, c clienter, m message) {
	if s.password == "" {
		return
	}

	// handshaked clients cant PASS
	if c.handshake() {
		c.sendRPL(s.name, errAlreadyRegistered{
			client: c.nickname(),
		})
		return
	}

	password := m.params[0]
	if password == s.password {
		c.setPassword(true)
	}
}
