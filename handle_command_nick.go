package ircd

func handleNick(s *server, c clienter, m message) {
	// nick params should be 1
	if len(m.params) < 1 {
		c.sendRPL(s.name, errNoNicknameGiven{
			client: c.nickname(),
		})
		return
	}

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
		handleHandshake(s, c)
	}
}
