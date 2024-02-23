package ircd

func handleMode(s *server, c *client, m message) {
	// command needs target
	if len(m.params) < 1 {
		c.sendRPL(s.name, errNeedMoreParams{
			client:  c.nickname(),
			command: m.command,
		})
		return
	}

	target := m.params[0]

	modestring := ""
	if len(m.params) >= 2 {
		modestring = m.params[1]
	}

	// is channel
	if m.isTargetChannel() {
		// get channel
		ch, ok := s.channels.get(target)
		// does it exist?
		if !ok {
			c.sendRPL(s.name, errNoSuchChannel{
				client:  c.nickname(),
				channel: target,
			})
			return
		}

		// return modes if modestring is not set
		if modestring == "" {
			c.sendRPL(s.name, rplChannelModeIs{
				client:     c.nickname(),
				channel:    ch.name,
				modestring: ch.modestring(),
				modeargs:   "",
			})
			return
		}

		add, del := parseModestring[channelMode](modestring, channelModeMap)
		for _, a := range add {
			switch a {
			case modeChannelModerated:
				ch.addMode(a)
			case modeChannelTLSOnly:
				ch.addMode(a)
			}
		}
		for _, d := range del {
			switch d {
			case modeChannelModerated:
				ch.removeMode(d)
			case modeChannelTLSOnly:
				ch.removeMode(d)
			}
		}
		return
	}

	// is user

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
}
