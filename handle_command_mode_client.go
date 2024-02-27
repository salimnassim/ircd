package ircd

func handleModeClient(s *server, c clienter, m message) {
	target := m.params[0]

	modestring := ""
	if len(m.params) >= 2 {
		modestring = m.params[1]
	}

	// target has to be client
	if target != c.nickname() {
		c.sendRPL(s.name, errUsersDontMatch{
			client: c.nickname(),
		})
		return
	}

	// send modes if modestring is not set
	if modestring == "" {
		c.sendRPL(s.name, rplUModeIs{
			client:     c.nickname(),
			modestring: c.modestring(),
		})
		return
	}

	add, del := parseModestring[clientMode](modestring, clientModeMap)
	for _, a := range add {
		switch a {
		case modeClientInvisible:
			c.addMode(a)
		case modeClientWallops:
			c.addMode(a)
		}
	}
	for _, d := range del {
		switch d {
		case modeClientInvisible:
			c.removeMode(d)
		case modeClientWallops:
			c.removeMode(d)
		}
	}

	c.sendRPL(s.name, rplUModeIs{
		client:     c.nickname(),
		modestring: c.modestring(),
	})
}
