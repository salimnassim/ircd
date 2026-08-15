package ircd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

func (s *session) cmdPass(m message) {
	if s.deps.password == "" {
		return
	}
	if s.handshakeDone {
		s.handle.deliver(s.formatRPL(errAlreadyRegistered{client: s.nick}))
		return
	}
	if m.params[0] == s.deps.password {
		s.passwordOK = true
	}
}

func (s *session) cmdPing(m message) {
	s.handle.deliver(strings.Replace(m.raw, "PING", "PONG", 1))
}

func (s *session) cmdPong() {
	select {
	case s.ponged <- struct{}{}:
	default:
	}
}

func (s *session) cmdNick(ctx context.Context, m message) {
	nick := m.params[0]

	if !s.deps.nickRegex.MatchString(nick) {
		s.handle.deliver(s.formatRPL(errErroneusNickname{client: s.nick, nick: nick}))
		return
	}

	oldPrefix := s.prefix()
	wasRegistered := s.handshakeDone

	if err := s.deps.clients.claimNick(s.id, nick); err != nil {
		s.handle.deliver(s.formatRPL(errNicknameInUse{client: nick, nick: nick}))
		return
	}

	s.nick = nick
	s.publishSnapshot()

	if wasRegistered {
		for name := range s.memberOf {
			s.deps.channels.dispatch(ctx, s.channelDeps(), name,
				evMemberRenamed{id: s.id, oldPrefix: oldPrefix, newNick: nick}, false, "")
		}
		return
	}

	if s.nick != "" && s.user != "" {
		if s.deps.password != "" && !s.passwordOK {
			s.handle.deliver(s.formatRPL(errPasswdMismatch{client: s.nick}))
			s.handle.terminate(errors.New("Quit: Wrong server password."))
			return
		}
		s.completeHandshake(ctx)
	}
}

func (s *session) cmdUser(ctx context.Context, m message) {
	if s.user != "" {
		s.handle.deliver(s.formatRPL(errAlreadyRegistered{client: s.nick}))
		return
	}

	s.user = m.params[0]
	s.real = m.params[3]
	s.publishSnapshot()

	if !s.handshakeDone && s.nick != "" && s.user != "" {
		if s.deps.password != "" && !s.passwordOK {
			s.handle.deliver(s.formatRPL(errPasswdMismatch{client: s.nick}))
			return
		}
		s.completeHandshake(ctx)
	}
}

func (s *session) completeHandshake(ctx context.Context) {
	if s.handshakeDone {
		return
	}

	s.handle.deliver(noticeCommand{client: s.nick, message: fmt.Sprintf("AUTH :*** Your ID is: %s", s.id)}.command())
	s.handle.deliver(noticeCommand{client: s.nick, message: fmt.Sprintf("AUTH :*** Your IP address is: %s", s.address)}.command())
	s.handle.deliver(noticeCommand{client: s.nick, message: "AUTH :*** Looking up your hostname..."}.command())

	lookupCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	addr, err := net.DefaultResolver.LookupAddr(lookupCtx, s.address)
	cancel()
	if err != nil || len(addr) == 0 {
		s.host = s.address
	} else {
		s.host = addr[0]
	}

	s.handle.deliver(s.formatRPL(rplWelcome{client: s.nick, network: s.deps.network, hostname: s.prefix()}))
	s.handle.deliver(s.formatRPL(rplYourHost{client: s.nick, serverName: s.deps.serverName, version: s.deps.version}))

	prefix := 4
	if strings.Count(s.address, ":") > 1 {
		prefix = 6
	}
	tlsLabel := "plain"
	if s.tls {
		tlsLabel = "tls"
	}
	s.host = fmt.Sprintf("ipv%d-%s-%s.vhost", prefix, tlsLabel, s.id)
	s.handle.deliver(noticeCommand{client: s.nick, message: fmt.Sprintf("AUTH :*** Your hostname has been cloaked to %s", s.host)}.command())
	s.handle.deliver(noticeCommand{
		client:  s.nick,
		message: fmt.Sprintf("NOTICE :*** Server is listening on ports %s", strings.Join(s.deps.ports(), ", ")),
	}.command())

	s.handle.deliver(s.formatRPL(rplMotdStart{client: s.nick, server: s.deps.serverName, text: "MOTD -"}))
	for _, line := range s.deps.motd {
		s.handle.deliver(s.formatRPL(rplMotd{client: s.nick, text: line}))
	}
	s.handle.deliver(s.formatRPL(rplEndOfMotd{client: s.nick}))
	s.handle.deliver(s.formatRPL(rplISupport{client: s.nick, tokens: s.deps.isupport}))

	s.modes |= modeClientVhost
	if s.tls {
		s.modes |= modeClientTLS
	}
	s.handle.deliver(modeCommand{target: s.nick, modestring: s.modestring(), args: ""}.command())

	s.handshakeDone = true
	s.publishSnapshot()
}

func (s *session) cmdLusers() {
	visible, invisible := 0, 0
	for _, h := range s.deps.clients.all() {
		snap := h.snapshot.Load()
		if snap == nil {
			continue
		}
		if snap.modes&modeClientInvisible != 0 {
			invisible++
		} else {
			visible++
		}
	}

	s.handle.deliver(s.formatRPL(rplLuserClient{client: s.nick, users: visible + invisible, invisible: invisible, servers: 1}))
	s.handle.deliver(s.formatRPL(rplLuserOp{client: s.nick, ops: 0}))
	s.handle.deliver(s.formatRPL(rplLuserChannels{client: s.nick, channels: s.deps.channels.count()}))
}

func (s *session) cmdJoin(ctx context.Context, m message) {
	targets := strings.Split(m.params[0], ",")

	var keys []string
	if len(m.params) >= 2 {
		keys = strings.Split(m.params[1], ",")
		if len(targets) != len(keys) {
			s.handle.deliver(s.formatRPL(errNeedMoreParams{client: s.nick, command: m.command}))
			return
		}
	}

	for i, target := range targets {
		if !isChannelName(target) || !s.deps.channelRegex.MatchString(target) {
			s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: target}))
			continue
		}

		key := ""
		if i < len(keys) {
			key = keys[i]
		}

		s.memberOf[target] = struct{}{}
		s.publishSnapshot()

		s.deps.channels.dispatch(ctx, s.channelDeps(), target, evJoin{who: s.handle, key: key}, true, s.id)
	}
}

func (s *session) cmdPart(ctx context.Context, m message) {
	targets := strings.Split(m.params[0], ",")

	reason := ""
	if len(m.params) >= 2 {
		reason = strings.Join(m.params[1:], " ")
	}

	for _, target := range targets {
		if !isChannelName(target) {
			s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: target}))
			continue
		}
		if _, ok := s.deps.channels.get(target); !ok {
			s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: target}))
			continue
		}

		delete(s.memberOf, target)
		s.publishSnapshot()

		s.deps.channels.dispatch(ctx, s.channelDeps(), target, evPart{who: s.handle, reason: reason}, false, "")
	}
}

func (s *session) cmdKick(ctx context.Context, m message) {
	channel := m.params[0]

	if _, ok := s.deps.channels.get(channel); !ok {
		s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: channel}))
		return
	}

	reason := "No reason given."
	if len(m.params) >= 3 {
		reason = strings.Join(m.params[2:], " ")
	}

	for target := range strings.SplitSeq(m.params[1], ",") {
		tc, ok := s.deps.clients.lookupNick(target)
		if !ok {
			s.handle.deliver(s.formatRPL(errUserNotInChannel{client: s.nick, nick: target, channel: channel}))
			continue
		}
		s.deps.channels.dispatch(ctx, s.channelDeps(), channel, evKick{by: s.handle, target: tc, reason: reason}, false, "")
	}
}

func (s *session) cmdTopic(ctx context.Context, m message) {
	target := m.params[0]

	if !isChannelName(target) {
		s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: target}))
		return
	}
	if _, ok := s.deps.channels.get(target); !ok {
		s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: target}))
		return
	}

	text := ""
	if len(m.params) >= 2 {
		text = strings.Join(m.params[1:], " ")
	}

	s.deps.channels.dispatch(ctx, s.channelDeps(), target, evSetTopic{by: s.handle, text: text}, false, "")
}

func (s *session) cmdPrivmsg(ctx context.Context, m message) {
	targets := strings.Split(m.params[0], ",")

	text := ""
	if len(m.params) >= 2 {
		text = strings.Join(m.params[1:], " ")
	}

	for _, target := range targets {
		if isChannelName(target) {
			if _, ok := s.deps.channels.get(target); !ok {
				s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: target}))
				continue
			}
			s.deps.channels.dispatch(ctx, s.channelDeps(), target, evBroadcast{by: s.handle, text: text, kind: broadcastPrivmsg}, false, "")
			continue
		}

		tc, ok := s.deps.clients.lookupNick(target)
		if !ok {
			s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: target}))
			continue
		}

		tsnap := tc.snapshot.Load()
		if tsnap == nil {
			continue
		}
		if tsnap.away != "" {
			s.handle.deliver(s.formatRPL(rplAway{client: s.nick, nick: tsnap.nick, message: tsnap.away}))
		}

		tc.deliver(privmsgCommand{prefix: s.nick, target: tsnap.nick, text: text}.command())
	}
}

func (s *session) cmdWhois(m message) {
	target := m.params[0]

	who, ok := s.deps.clients.lookupNick(target)
	if !ok {
		s.handle.deliver(s.formatRPL(errNoSuchNick{client: s.nick, nick: target}))
		return
	}
	wsnap := who.snapshot.Load()
	if wsnap == nil {
		s.handle.deliver(s.formatRPL(errNoSuchNick{client: s.nick, nick: target}))
		return
	}

	s.handle.deliver(s.formatRPL(rplWhoisUser{
		client: s.nick, nick: wsnap.nick, username: wsnap.user, host: wsnap.host, realname: wsnap.real,
	}))

	channels := []string{}
	for _, name := range wsnap.channels {
		snap, ok := s.deps.channels.snapshot(name)
		if !ok {
			continue
		}
		if snap.modes&modeChannelSecret != 0 {
			continue
		}
		if snap.modes&modeChannelPrivate != 0 && !isMemberSnapshot(snap, s.id) {
			continue
		}
		channels = append(channels, name)
	}
	s.handle.deliver(s.formatRPL(rplWhoisChannels{client: s.nick, nick: wsnap.nick, channels: channels}))

	if wsnap.away != "" {
		s.handle.deliver(s.formatRPL(rplWhoisSpecial{client: s.nick, nick: wsnap.nick, text: "User is away."}))
	}
}

func (s *session) cmdWho(m message) {
	if len(m.params) == 0 {
		s.handle.deliver(s.formatRPL(errNeedMoreParams{client: s.nick, command: m.command}))
		return
	}

	target := m.params[0]
	if !isChannelName(target) {
		return
	}

	snap, ok := s.deps.channels.snapshot(target)
	if !ok {
		s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: target}))
		return
	}

	for _, mem := range snap.members {
		flags := "H"
		if mem.away != "" {
			flags = "G"
		}
		s.handle.deliver(s.formatRPL(rplWhoReply{
			client: s.nick, channel: target, username: mem.username, host: mem.host,
			server: s.deps.serverName, nick: mem.nick, flags: flags, hopcount: 0, realname: mem.realname,
		}))
	}
	s.handle.deliver(s.formatRPL(rplEndOfWho{client: s.nick, mask: ""}))
}

func (s *session) cmdMode(ctx context.Context, m message) {
	if !isChannelName(m.params[0]) {
		s.cmdModeClient(m)
		return
	}
	s.cmdModeChannel(ctx, m)
}

func (s *session) cmdModeClient(m message) {
	target := m.params[0]

	modestring := ""
	if len(m.params) >= 2 {
		modestring = m.params[1]
	}

	if target != s.nick {
		s.handle.deliver(s.formatRPL(errUsersDontMatch{client: s.nick}))
		return
	}

	if modestring == "" {
		s.handle.deliver(s.formatRPL(rplUModeIs{client: s.nick, modestring: s.modestring()}))
		return
	}

	add, del := parseModestring(modestring, clientModeMap)
	for _, a := range add {
		switch a {
		case modeClientInvisible, modeClientWallops:
			s.modes |= a
		}
	}
	for _, d := range del {
		switch d {
		case modeClientInvisible, modeClientWallops:
			s.modes &^= d
		}
	}
	s.publishSnapshot()

	s.handle.deliver(s.formatRPL(rplUModeIs{client: s.nick, modestring: s.modestring()}))
}

func (s *session) cmdModeChannel(ctx context.Context, m message) {
	target := m.params[0]

	modestring := ""
	if len(m.params) >= 2 {
		modestring = m.params[1]
	}

	if modestring == "" {
		snap, ok := s.deps.channels.snapshot(target)
		if !ok {
			s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: target}))
			return
		}
		s.handle.deliver(s.formatRPL(rplChannelModeIs{client: s.nick, channel: target, modestring: snap.modestring, modeargs: ""}))
		return
	}

	if (modestring == "b" || modestring == "+b") && len(m.params) < 3 {
		snap, ok := s.deps.channels.snapshot(target)
		if !ok {
			s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: target}))
			return
		}
		for _, mask := range snap.bans {
			s.handle.deliver(s.formatRPL(rplBanList{client: s.nick, channel: target, mask: mask}))
		}
		s.handle.deliver(s.formatRPL(rplEndOfBanList{client: s.nick, channel: target}))
		return
	}

	if _, ok := s.deps.channels.get(target); !ok {
		s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: target}))
		return
	}

	modeargs := ""
	if len(m.params) >= 3 {
		modeargs = strings.Join(m.params[2:], " ")
	}

	s.deps.channels.dispatch(ctx, s.channelDeps(), target, evSetMode{by: s.handle, modestring: modestring, args: modeargs}, false, "")
}

func (s *session) cmdAway(m message) {
	if len(m.params) == 0 {
		if s.away != "" {
			s.away = ""
			s.publishSnapshot()
		}
		s.handle.deliver(s.formatRPL(rplUnAway{client: s.nick}))
		return
	}

	s.away = strings.Join(m.params, " ")
	s.publishSnapshot()
	s.handle.deliver(s.formatRPL(rplNowAway{client: s.nick}))
}

func (s *session) cmdQuit(m message) {
	reason := "no reason given"
	if len(m.params) > 0 {
		reason = strings.TrimSpace(strings.Join(m.params, " "))
	}
	s.handle.terminate(fmt.Errorf("Quit: %s", reason))
}

func (s *session) cmdOper(m message) {
	user := m.params[0]
	password := m.params[1]

	if !s.deps.operators.auth(user, password) {
		s.handle.deliver(s.formatRPL(errPasswdMismatch{client: s.nick}))
		return
	}

	s.modes |= modeClientOperator
	s.publishSnapshot()

	s.handle.deliver(modeCommand{target: s.nick, modestring: s.modestring(), args: ""}.command())
	s.handle.deliver(s.formatRPL(rplYoureOper{client: s.nick}))
}

func (s *session) cmdVersion(m message) {
	if len(m.params) == 0 {
		s.handle.deliver(s.formatRPL(rplVersion{client: s.nick, version: fmt.Sprintf("ircd-%s", s.deps.version), server: s.deps.serverName, comments: ""}))
		s.handle.deliver(s.formatRPL(rplISupport{
			client: s.nick,
			tokens: "AWAYLEN=128 CASEMAPPING=ascii CHANLIMIT=#&:64 CHANNELLEN=50 CHANTYPES=#& HOSTLEN=128 KICKLEN=128 MODES=24 NICKLEN=31 PREFIX=(qaohv)~&@%+ TOPICLEN=307 USERLEN=18",
		}))
		return
	}
	s.handle.deliver(s.formatRPL(errNoSuchServer{client: s.nick, server: m.params[0]}))
}

func (s *session) cmdList(m message) {
	if len(m.params) != 0 {
		return
	}

	s.handle.deliver(s.formatRPL(rplListStart{client: s.nick}))
	for _, ch := range s.deps.channels.all() {
		snap := ch.snapshot.Load()
		if snap == nil || snap.modes&modeChannelSecret != 0 {
			continue
		}
		s.handle.deliver(s.formatRPL(rplList{client: s.nick, channel: snap.name, count: snap.count, topic: snap.topicText}))
	}
	s.handle.deliver(s.formatRPL(rplListEnd{client: s.nick}))
}

func (s *session) cmdInvite(ctx context.Context, m message) {
	nickname := m.params[0]
	channel := m.params[1]

	tc, ok := s.deps.clients.lookupNick(nickname)
	if !ok {
		s.handle.deliver(s.formatRPL(errNoSuchNick{client: s.nick, nick: nickname}))
		return
	}
	if _, ok := s.deps.channels.get(channel); !ok {
		s.handle.deliver(s.formatRPL(errNoSuchChannel{client: s.nick, channel: channel}))
		return
	}

	s.deps.channels.dispatch(ctx, s.channelDeps(), channel, evInvite{by: s.handle, target: tc}, false, "")
}
