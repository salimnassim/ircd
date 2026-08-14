package ircd

import (
	"cmp"
	"context"
	"fmt"
	"slices"
	"strings"
	"sync/atomic"
	"time"
)

type banMask string

type channelMember struct {
	handle *sessionHandle
	modes  channelMembershipMode
}

type topicInfo struct {
	text      string
	timestamp int64
	author    string
}

type channelMemberSnapshot struct {
	id       clientID
	nick     string
	username string
	host     string
	realname string
	away     string
	modes    channelMembershipMode
}

type channelSnapshot struct {
	name        string
	topicText   string
	topicAuthor string
	topicSetAt  int64
	modes       channelMode
	modestring  string
	key         string
	owner       clientID
	members     []channelMemberSnapshot
	count       int
}

type channelActor struct {
	name    string
	owner   clientID
	members map[clientID]*channelMember
	topic   topicInfo
	modes   channelMode
	bans    map[banMask]bool
	invites map[clientID]bool
	key     string

	inbox    chan channelEvent
	snapshot atomic.Pointer[channelSnapshot]

	directory  *channelDirectory
	clients    *clientDirectory
	serverName string
}

func newChannelActor(name string, owner clientID, directory *channelDirectory, clients *clientDirectory, serverName string) *channelActor {
	ch := &channelActor{
		name:       name,
		owner:      owner,
		members:    make(map[clientID]*channelMember),
		bans:       make(map[banMask]bool),
		invites:    make(map[clientID]bool),
		modes:      modeChannelNoExternal | modeChannelRestrictTopic,
		inbox:      make(chan channelEvent, 64),
		directory:  directory,
		clients:    clients,
		serverName: serverName,
	}
	ch.publishSnapshot()
	channelGauge.Inc()
	return ch
}
func (ch *channelActor) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch.inbox:
			if !ok {
				return
			}
			ch.handle(ev)
			ch.directory.doneProcessing(ch.name)
			if len(ch.members) == 0 && ch.directory.removeIfIdle(ch.name, ch) {
				channelGauge.Dec()
				return
			}
		}
	}
}

func (ch *channelActor) handle(ev channelEvent) {
	switch e := ev.(type) {
	case evJoin:
		ch.handleJoin(e)
	case evPart:
		ch.handlePart(e)
	case evKick:
		ch.handleKick(e)
	case evSetTopic:
		ch.handleSetTopic(e)
	case evSetMode:
		ch.handleSetMode(e)
	case evBroadcast:
		ch.handleBroadcast(e)
	case evInvite:
		ch.handleInvite(e)
	case evQuit:
		ch.handleQuit(e)
	case evMemberRenamed:
		ch.handleMemberRenamed(e)
	}
}

func (ch *channelActor) formatRPL(r rpl) string {
	return fmt.Sprintf(":%s %s", ch.serverName, r.rpl())
}

func (ch *channelActor) broadcastCommand(cmd command, skipID clientID, skip bool) {
	line := cmd.command()
	for id, m := range ch.members {
		if skip && id == skipID {
			continue
		}
		m.handle.deliver(line)
	}
}

func (ch *channelActor) broadcastRPL(r rpl, skipID clientID, skip bool) {
	line := ch.formatRPL(r)
	for id, m := range ch.members {
		if skip && id == skipID {
			continue
		}
		m.handle.deliver(line)
	}
}

func (ch *channelActor) names() []string {
	names := make([]string, 0, len(ch.members))
	for _, m := range ch.members {
		if snap := m.handle.snapshot.Load(); snap != nil {
			names = append(names, snap.nick)
		}
	}
	return names
}

func (ch *channelActor) modestring() string {
	modes := []rune{}
	for r, m := range channelModeMap {
		if ch.modes&m != 0 {
			modes = append(modes, r)
		}
	}
	slices.SortFunc(modes, func(a, b rune) int { return cmp.Compare(a, b) })
	return fmt.Sprintf("+%s", string(modes))
}

func (ch *channelActor) membershipModestring(id clientID) string {
	m, ok := ch.members[id]
	if !ok {
		return "+"
	}
	modes := []rune{}
	for r, mm := range channelMembershipModeMap {
		if m.modes&mm != 0 {
			modes = append(modes, r)
		}
	}
	slices.SortFunc(modes, func(a, b rune) int { return cmp.Compare(a, b) })
	return fmt.Sprintf("+%s", string(modes))
}

func (ch *channelActor) publishSnapshot() {
	members := make([]channelMemberSnapshot, 0, len(ch.members))
	for id, m := range ch.members {
		s := m.handle.snapshot.Load()
		if s == nil {
			continue
		}
		members = append(members, channelMemberSnapshot{
			id:       id,
			nick:     s.nick,
			username: s.user,
			host:     s.host,
			realname: s.real,
			away:     s.away,
			modes:    m.modes,
		})
	}
	ch.snapshot.Store(&channelSnapshot{
		name:        ch.name,
		topicText:   ch.topic.text,
		topicAuthor: ch.topic.author,
		topicSetAt:  ch.topic.timestamp,
		modes:       ch.modes,
		modestring:  ch.modestring(),
		key:         ch.key,
		owner:       ch.owner,
		members:     members,
		count:       len(ch.members),
	})
}

const privsMask = modeMemberHalfOperator | modeMemberOperator | modeMemberAdmin | modeMemberOwner

func (ch *channelActor) handleJoin(ev evJoin) {
	snap := ev.who.snapshot.Load()
	if snap == nil {
		return
	}

	if _, already := ch.members[ev.who.id]; already {
		return
	}

	if ch.modes&modeChannelTLSOnly != 0 && !snap.tls {
		ev.who.deliver(noticeCommand{client: snap.nick, message: "Cannot join channel (+z)"}.command())
		return
	}

	if ch.modes&modeChannelInviteOnly != 0 && !ch.invites[ev.who.id] {
		ev.who.deliver(ch.formatRPL(errInviteOnlyChan{client: snap.nick, channel: ch.name}))
		return
	}
	delete(ch.invites, ev.who.id)

	if ch.modes&modeChannelKey != 0 {
		if ev.key == "" || ev.key != ch.key {
			ev.who.deliver(ch.formatRPL(errBadChannelKey{client: snap.nick, channel: ch.name}))
			return
		}
	}

	ch.members[ev.who.id] = &channelMember{handle: ev.who}

	ch.broadcastCommand(joinCommand{prefix: snap.prefix(), channel: ch.name}, "", false)

	if ch.owner == ev.who.id {
		ch.members[ev.who.id].modes |= modeMemberOwner
		ch.broadcastCommand(modeCommand{
			source:     ch.serverName,
			target:     ch.name,
			modestring: ch.membershipModestring(ev.who.id),
			args:       snap.nick,
		}, "", false)
	}

	if ch.topic.text == "" {
		ev.who.deliver(ch.formatRPL(rplNoTopic{client: snap.nick, channel: ch.name}))
	} else {
		ev.who.deliver(ch.formatRPL(rplTopic{client: snap.nick, channel: ch.name, topic: ch.topic.text}))
		ev.who.deliver(ch.formatRPL(rplTopicWhoTime{
			client: snap.nick, channel: ch.name, nick: ch.topic.author, setat: int(ch.topic.timestamp),
		}))
	}

	ev.who.deliver(ch.formatRPL(rplNamReply{client: snap.nick, symbol: "=", channel: ch.name, nicks: ch.names()}))
	ev.who.deliver(ch.formatRPL(rplEndOfNames{client: snap.nick, channel: ch.name}))

	ev.who.deliver(modeCommand{target: ch.name, modestring: ch.modestring(), args: ""}.command())

	ch.publishSnapshot()
}

func (ch *channelActor) handlePart(ev evPart) {
	if _, ok := ch.members[ev.who.id]; !ok {
		return
	}
	delete(ch.members, ev.who.id)

	snap := ev.who.snapshot.Load()
	reason := ev.reason
	if reason == "" {
		reason = "No reason given"
	}
	ch.broadcastCommand(partCommand{prefix: snap.prefix(), channel: ch.name, text: reason}, "", false)
	ch.publishSnapshot()
}

func (ch *channelActor) handleKick(ev evKick) {
	bySnap := ev.by.snapshot.Load()

	by, isMember := ch.members[ev.by.id]
	if !isMember {
		if ch.modes&modeChannelSecret != 0 {
			ev.by.deliver(ch.formatRPL(errNoSuchChannel{client: bySnap.nick, channel: ch.name}))
		} else {
			ev.by.deliver(ch.formatRPL(errNotOnChannel{client: bySnap.nick, channel: ch.name}))
		}
		return
	}

	if by.modes&privsMask == 0 {
		ev.by.deliver(ch.formatRPL(errChanoPrivsNeeded{client: bySnap.nick, channel: ch.name}))
		return
	}

	targetSnap := ev.target.snapshot.Load()
	if _, ok := ch.members[ev.target.id]; !ok {
		ev.by.deliver(ch.formatRPL(errUserNotInChannel{client: bySnap.nick, nick: targetSnap.nick, channel: ch.name}))
		return
	}

	reason := ev.reason
	if reason == "" {
		reason = "No reason given."
	}

	ch.broadcastCommand(kickCommand{
		prefix: bySnap.prefix(), channel: ch.name, target: targetSnap.nick, reason: reason,
	}, "", false)
	delete(ch.members, ev.target.id)
	ch.publishSnapshot()
}

func (ch *channelActor) handleSetTopic(ev evSetTopic) {
	snap := ev.by.snapshot.Load()

	if ch.modes&modeChannelRestrictTopic != 0 {
		m, isMember := ch.members[ev.by.id]
		if !isMember || m.modes&privsMask == 0 {
			ev.by.deliver(ch.formatRPL(errChanoPrivsNeeded{client: snap.nick, channel: ch.name}))
			return
		}
	}

	ch.topic = topicInfo{text: ev.text, timestamp: time.Now().Unix(), author: snap.nick}
	ch.broadcastRPL(rplTopic{client: snap.nick, channel: ch.name, topic: ch.topic.text}, "", false)
	ch.publishSnapshot()
}

func (ch *channelActor) handleInvite(ev evInvite) {
	bySnap := ev.by.snapshot.Load()
	targetSnap := ev.target.snapshot.Load()

	by, isMember := ch.members[ev.by.id]

	if ch.modes&modeChannelInviteOnly != 0 {
		if !isMember || by.modes&privsMask == 0 {
			ev.by.deliver(ch.formatRPL(errChanoPrivsNeeded{client: bySnap.nick, channel: ch.name}))
			return
		}
	}

	if !isMember {
		ev.by.deliver(ch.formatRPL(errNotOnChannel{client: bySnap.nick, channel: ch.name}))
		return
	}

	if _, alreadyMember := ch.members[ev.target.id]; alreadyMember {
		ev.by.deliver(ch.formatRPL(errUserOnChannel{client: bySnap.nick, nick: targetSnap.nick, channel: ch.name}))
		return
	}

	ch.invites[ev.target.id] = true

	ev.by.deliver(ch.formatRPL(rplInviting{client: bySnap.nick, nick: targetSnap.nick, channel: ch.name}))
	ev.target.deliver(inviteCommand{prefix: bySnap.prefix(), target: targetSnap.nick, channel: ch.name}.command())
}

type channelMessageCommand struct {
	prefix string
	verb   string
	target string
	text   string
}

func (cmd channelMessageCommand) command() string {
	return fmt.Sprintf(":%s %s %s :%s", cmd.prefix, cmd.verb, cmd.target, cmd.text)
}

func (ch *channelActor) handleBroadcast(ev evBroadcast) {
	snap := ev.by.snapshot.Load()

	m, isMember := ch.members[ev.by.id]
	if !isMember {
		ev.by.deliver(ch.formatRPL(errNotOnChannel{client: snap.nick, channel: ch.name}))
		return
	}

	if ch.modes&modeChannelModerated != 0 && m.modes&(modeMemberVoice|privsMask) == 0 {
		ev.by.deliver(ch.formatRPL(errCannotSendToChan{client: snap.nick, channel: ch.name, text: "Channel is moderated."}))
		return
	}

	verb := "PRIVMSG"
	if ev.kind == broadcastNotice {
		verb = "NOTICE"
	}
	ch.broadcastCommand(channelMessageCommand{prefix: snap.prefix(), verb: verb, target: ch.name, text: ev.text}, ev.by.id, true)
}

func (ch *channelActor) handleQuit(ev evQuit) {
	if _, ok := ch.members[ev.who.id]; !ok {
		return
	}
	snap := ev.who.snapshot.Load()
	delete(ch.members, ev.who.id)

	ch.broadcastCommand(quitCommand{prefix: snap.prefix(), text: ev.reason}, "", false)
	ch.publishSnapshot()
}

type nickCommand struct {
	prefix string
	nick   string
}

func (cmd nickCommand) command() string {
	return fmt.Sprintf(":%s NICK :%s", cmd.prefix, cmd.nick)
}

func (ch *channelActor) handleMemberRenamed(ev evMemberRenamed) {
	if _, ok := ch.members[ev.id]; !ok {
		return
	}
	ch.broadcastCommand(nickCommand{prefix: ev.oldPrefix, nick: ev.newNick}, "", false)
	ch.publishSnapshot()
}

func isMemberSnapshot(snap channelSnapshot, id clientID) bool {
	for _, m := range snap.members {
		if m.id == id {
			return true
		}
	}
	return false
}

func isSettableMembershipMode(mode channelMembershipMode) bool {
	switch mode {
	case modeMemberVoice, modeMemberHalfOperator, modeMemberOperator, modeMemberAdmin:
		return true
	default:
		return false
	}
}

func (ch *channelActor) handleSetMode(ev evSetMode) {
	snap := ev.by.snapshot.Load()

	by, isMember := ch.members[ev.by.id]
	if !isMember {
		ev.by.deliver(ch.formatRPL(errNotOnChannel{client: snap.nick, channel: ch.name}))
		return
	}
	if by.modes&privsMask == 0 {
		ev.by.deliver(ch.formatRPL(errChanoPrivsNeeded{client: snap.nick, channel: ch.name}))
		return
	}

	var tcs []string
	if ev.args != "" {
		tcs = strings.Split(ev.args, " ")
	}

	if len(tcs) == 0 {
		ch.applyChannelModeChange(ev.by, ev.modestring)
		return
	}

	add, _ := parseModestring(ev.modestring, channelModeMap)
	if slices.Contains(add, modeChannelKey) {
		ch.applyKeyChange(ev.by, add, tcs)
		return
	}

	ch.applyMembershipModeChange(ev.by, ev.modestring, tcs)
}

func (ch *channelActor) applyChannelModeChange(by *sessionHandle, modestring string) {
	before := ch.modes

	add, del := parseModestring(modestring, channelModeMap)
	for _, a := range add {
		switch a {
		case modeChannelModerated, modeChannelTLSOnly, modeChannelSecret, modeChannelRestrictTopic, modeChannelInviteOnly:
			ch.modes |= a
		}
	}
	for _, d := range del {
		switch d {
		case modeChannelModerated, modeChannelTLSOnly, modeChannelSecret, modeChannelRestrictTopic, modeChannelInviteOnly, modeChannelKey:
			ch.modes &^= d
			if d == modeChannelKey {
				ch.key = ""
			}
		}
	}

	after := ch.modes
	da, dd := diffModes(before, after, channelModeMap)
	if len(da) == 0 && len(dd) == 0 {
		return
	}

	diff := ""
	if len(dd) > 0 {
		minus := []rune{'-'}
		for _, m := range dd {
			minus = append(minus, runeByMode(m, channelModeMap))
		}
		diff += string(minus)
	}
	if len(da) > 0 {
		plus := []rune{'+'}
		for _, m := range da {
			plus = append(plus, runeByMode(m, channelModeMap))
		}
		diff += string(plus)
	}

	ch.broadcastCommand(modeCommand{
		source: by.snapshot.Load().prefix(), target: ch.name, modestring: diff, args: "",
	}, "", false)
	ch.publishSnapshot()
}

func (ch *channelActor) applyKeyChange(by *sessionHandle, addModes []channelMode, tcs []string) {
	snap := by.snapshot.Load()

	for i, mode := range addModes {
		if len(tcs) <= i {
			by.deliver(ch.formatRPL(errNeedMoreParams{client: snap.nick, command: "MODE"}))
			return
		}
		if mode == modeChannelKey {
			ch.key = tcs[i]
			ch.modes |= modeChannelKey
		}
		ch.broadcastCommand(modeCommand{
			source:     snap.prefix(),
			target:     ch.name,
			modestring: fmt.Sprintf("+%c", runeByMode(modeChannelKey, channelModeMap)),
			args:       tcs[i],
		}, "", false)
	}
	ch.publishSnapshot()
}

func (ch *channelActor) applyMembershipModeChange(by *sessionHandle, modestring string, tcs []string) {
	snap := by.snapshot.Load()

	type change struct {
		nick string
		mode string
	}
	var changes []change

	resolve := func(i int) (*sessionHandle, *channelMember, bool) {
		if len(tcs) <= i {
			by.deliver(ch.formatRPL(errNeedMoreParams{client: snap.nick, command: "MODE"}))
			return nil, nil, false
		}
		h, exists := ch.clients.lookupNick(tcs[i])
		if !exists {
			by.deliver(ch.formatRPL(errNoSuchNick{client: snap.nick, nick: tcs[i]}))
			return nil, nil, false
		}
		m, isMember := ch.members[h.id]
		if !isMember {
			by.deliver(ch.formatRPL(errUserNotInChannel{client: snap.nick, nick: tcs[i], channel: ch.name}))
			return nil, nil, false
		}
		return h, m, true
	}

	add, del := parseModestring(modestring, channelMembershipModeMap)

	for i, mode := range add {
		if !isSettableMembershipMode(mode) {
			continue
		}
		h, m, ok := resolve(i)
		if !ok {
			return
		}
		m.modes |= mode
		changes = append(changes, change{
			nick: h.snapshot.Load().nick,
			mode: fmt.Sprintf("+%c", runeByMode(mode, channelMembershipModeMap)),
		})
	}

	for i, mode := range del {
		if !isSettableMembershipMode(mode) {
			continue
		}
		h, m, ok := resolve(i)
		if !ok {
			return
		}
		m.modes &^= mode
		changes = append(changes, change{
			nick: h.snapshot.Load().nick,
			mode: fmt.Sprintf("-%c", runeByMode(mode, channelMembershipModeMap)),
		})
	}

	if len(changes) == 0 {
		return
	}

	nicks := make([]string, 0, len(changes))
	modes := make([]string, 0, len(changes))
	for _, c := range changes {
		nicks = append(nicks, c.nick)
		modes = append(modes, c.mode)
	}

	ch.broadcastCommand(modeCommand{
		source:     snap.prefix(),
		target:     ch.name,
		modestring: strings.Join(modes, ""),
		args:       strings.Join(nicks, " "),
	}, "", false)
	ch.publishSnapshot()
}
