package ircd

import "fmt"

func handleMode(s *server, c clienter, m message) {
	target := m.params[0]

	modestring := ""
	if len(m.params) >= 2 {
		modestring = m.params[1]
	}

	// is channel
	if m.isTargetChannel() {
		// get channel
		ch, ok := s.Channels.get(target)
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
				channel:    ch.name(),
				modestring: ch.modestring(),
				modeargs:   "",
			})
			return
		}

		// client must be a member of the channel
		if !ch.clients().isMember(c) {
			c.sendRPL(s.name, errNotOnChannel{
				client:  c.nickname(),
				channel: ch.name(),
			})
			return
		}

		// client has to be hop or higher
		if !ch.clients().hasMode(c, modeHalfOperator, modeOperator, modeAdmin, modeOwner) {
			c.sendRPL(s.name, errChanoPrivsNeeded{
				client:  c.nickname(),
				channel: ch.name(),
			})
			return
		}

		before := ch.mode()
		// parse modestring
		add, del := parseModestring[channelMode](modestring, channelModeMap)
		for _, a := range add {
			switch a {
			case modeChannelModerated:
				ch.addMode(a)
			case modeChannelTLSOnly:
				ch.addMode(a)
			case modeChannelSecret:
				ch.addMode(a)
			case modeChannelRestrictTopic:
				ch.addMode(a)
			}
		}
		for _, d := range del {
			switch d {
			case modeChannelModerated:
				ch.removeMode(d)
			case modeChannelTLSOnly:
				ch.removeMode(d)
			case modeChannelSecret:
				ch.removeMode(d)
			case modeChannelRestrictTopic:
				ch.removeMode(d)
			}
		}
		after := ch.mode()

		plus := []rune{}
		minus := []rune{}

		// diff before and after, add +- if there are changes
		da, dd := diffModes[channelMode](before, after, channelModeMap)
		if len(da) == 0 && len(dd) == 0 {
			return
		}

		if len(da) > 0 {
			plus = append(plus, '+')
		}
		if len(dd) > 0 {
			minus = append(minus, '-')
		}

		// refactor this o-no bueno
		for _, m := range da {
			for r, mm := range channelModeMap {
				if m == mm {
					plus = append(plus, r)
				}
			}
		}
		for _, m := range dd {
			for r, mm := range channelModeMap {
				if m == mm {
					minus = append(minus, r)
				}
			}
		}

		diff := fmt.Sprintf("%s%s", string(minus), string(plus))
		ch.broadcastCommand(modeCommand{
			source:     c.prefix(),
			target:     ch.name(),
			modestring: diff,
			args:       "",
		}, c.id(), false)
		return
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
}
