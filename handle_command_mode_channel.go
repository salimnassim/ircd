package ircd

import (
	"fmt"
	"strings"
)

func handleModeChannel(s *server, c clienter, m message) {
	target := m.params[0]

	modestring := ""
	if len(m.params) >= 2 {
		modestring = m.params[1]
	}

	modeargs := ""
	if len(m.params) >= 3 {
		modeargs = strings.Join(m.params[2:len(m.params)], " ")
	}

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
	if !ch.clients().hasMode(c, modeMemberHalfOperator, modeMemberOperator, modeMemberAdmin, modeMemberOwner) {
		c.sendRPL(s.name, errChanoPrivsNeeded{
			client:  c.nickname(),
			channel: ch.name(),
		})
		return
	}

	// settings channel mode
	if modeargs == "" {
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
			case modeChannelInviteOnly:
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
			case modeChannelInviteOnly:
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

		diff := ""
		if len(minus) > 0 {
			diff = fmt.Sprintf("%s%s", diff, string(minus))
		}
		if len(plus) > 0 {
			diff = fmt.Sprintf("%s%s", diff, string(plus))
		}

		ch.broadcastCommand(modeCommand{
			source:     c.prefix(),
			target:     ch.name(),
			modestring: diff,
			args:       "",
		}, c.id(), false)
		return
	}

	type tmc struct {
		nick string
		mode string
	}

	if modeargs != "" {
		// split target clients from modeargs
		tcs := strings.Split(modeargs, " ")

		tcm := []tmc{}

		// todo: client can change mode only for user that is below their bitmask
		// todo: refactor this

		// parse modestring
		add, del := parseModestring[channelMembershipMode](modestring, channelMembershipModeMap)
		for i, mode := range add {
			if len(tcs) < i {
				c.sendRPL(s.name, errNeedMoreParams{
					client:  c.nickname(),
					command: m.command,
				})
				return
			}
			tc, exists := s.Clients.get(tcs[i])
			if !exists {
				c.sendRPL(s.name, errNoSuchNick{
					client: c.nickname(),
					nick:   tcs[i],
				})
				return
			}
			if !ch.clients().isMember(tc) {
				c.sendRPL(s.name, errUserNotInChannel{
					client:  c.nickname(),
					nick:    tcs[i],
					channel: ch.name(),
				})
				return
			}
			switch mode {
			case modeMemberVoice:
				ch.clients().addMode(tc, modeMemberVoice)
				tcm = append(tcm, tmc{
					nick: tc.nickname(),
					mode: fmt.Sprintf("+%c", runeByMode[channelMembershipMode](mode, channelMembershipModeMap)),
				})
			case modeMemberHalfOperator:
				ch.clients().addMode(tc, modeMemberHalfOperator)
				tcm = append(tcm, tmc{
					nick: tc.nickname(),
					mode: fmt.Sprintf("+%c", runeByMode[channelMembershipMode](mode, channelMembershipModeMap)),
				})
			case modeMemberOperator:
				ch.clients().addMode(tc, modeMemberOperator)
				tcm = append(tcm, tmc{
					nick: tc.nickname(),
					mode: fmt.Sprintf("+%c", runeByMode[channelMembershipMode](mode, channelMembershipModeMap)),
				})
			case modeMemberAdmin:
				ch.clients().addMode(tc, modeMemberAdmin)
				tcm = append(tcm, tmc{
					nick: tc.nickname(),
					mode: fmt.Sprintf("+%c", runeByMode[channelMembershipMode](mode, channelMembershipModeMap)),
				})
			}
		}

		for i, mode := range del {
			if len(tcs) < i {
				c.sendRPL(s.name, errNeedMoreParams{
					client:  c.nickname(),
					command: m.command,
				})
				return
			}
			tc, exists := s.Clients.get(tcs[i])
			if !exists {
				c.sendRPL(s.name, errNoSuchNick{
					client: c.nickname(),
					nick:   tcs[i],
				})
				return
			}
			if !ch.clients().isMember(tc) {
				c.sendRPL(s.name, errUserNotInChannel{
					client:  c.nickname(),
					nick:    tcs[i],
					channel: ch.name(),
				})
				return
			}
			switch mode {
			case modeMemberVoice:
				ch.clients().removeMode(tc, modeMemberVoice)
				tcm = append(tcm, tmc{
					nick: tc.nickname(),
					mode: fmt.Sprintf("-%c", runeByMode[channelMembershipMode](mode, channelMembershipModeMap)),
				})
			case modeMemberHalfOperator:
				ch.clients().removeMode(tc, modeMemberHalfOperator)
				tcm = append(tcm, tmc{
					nick: tc.nickname(),
					mode: fmt.Sprintf("-%c", runeByMode[channelMembershipMode](mode, channelMembershipModeMap)),
				})
			case modeMemberOperator:
				ch.clients().removeMode(tc, modeMemberOperator)
				tcm = append(tcm, tmc{
					nick: tc.nickname(),
					mode: fmt.Sprintf("-%c", runeByMode[channelMembershipMode](mode, channelMembershipModeMap)),
				})
			case modeMemberAdmin:
				ch.clients().removeMode(tc, modeMemberAdmin)
				tcm = append(tcm, tmc{
					nick: tc.nickname(),
					mode: fmt.Sprintf("-%c", runeByMode[channelMembershipMode](mode, channelMembershipModeMap)),
				})
			}
		}

		modeNicknames := []string{}
		modeModes := []string{}

		for _, t := range tcm {
			modeNicknames = append(modeNicknames, t.nick)
			modeModes = append(modeModes, t.mode)
		}

		ch.broadcastCommand(modeCommand{
			source:     c.prefix(),
			target:     ch.name(),
			modestring: strings.Join(modeModes, ""),
			args:       strings.Join(modeNicknames, " "),
		}, c.id(), false)
		// for i, d := range del {

		// }
	}

}
