package ircd

func handleNick(s *server, c clienter, m message) {
	// validate nickname
	ok := s.regex[regexNick].MatchString(m.params[0])
	if !ok {
		c.sendRPL(s.name, errErroneusNickname{
			client: c.nickname(),
			nick:   m.params[0],
		})
		return
	}

	// check if nick is already in use
	_, exists := s.Clients.get(m.params[0])
	if exists {
		c.sendRPL(s.name, errNicknameInUse{
			client: m.params[0],
			nick:   m.params[0],
		})
		return
	}

	c.setNickname(m.params[0])

	if !c.handshake() && c.nickname() != "" && c.username() != "" {
		if s.password != "" && !c.password() {
			c.sendRPL(s.name, errPasswdMismatch{
				client: c.nickname(),
			})
			c.kill("Wrong server password.")
			return
		}
		handleHandshake(s, c)
	}
}
