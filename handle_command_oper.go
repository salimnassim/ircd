package ircd

func handleOper(s *server, c clienter, m message) {
	user := m.params[0]
	password := m.params[1]

	// if not successful
	if !s.Operators.auth(user, password) {
		c.sendRPL(s.name, errPasswdMismatch{
			client: c.nickname(),
		})
		return
	}

	c.addMode(modeClientOperator)
	c.sendCommand(modeCommand{
		target:     c.nickname(),
		modestring: c.modestring(),
		args:       "",
	})

	c.sendRPL(s.name, rplYoureOper{
		client: c.nickname(),
	})
}
