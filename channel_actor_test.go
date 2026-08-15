package ircd

import (
	"strings"
	"testing"
)

func newTestChannelActor(name string, owner clientID, cd *clientDirectory) *channelActor {
	return newChannelActor(name, owner, newChannelDirectory(), channelActorDeps{serverName: "test.server", clients: cd})
}

func TestChannelActorJoinBasic(t *testing.T) {
	cd := newClientDirectory()
	joiner := newTestHandle("j1", "alice", "user", "host", false)
	registerAndClaim(cd, joiner)

	ch := newTestChannelActor("#chan", "owner-id", cd)
	ch.handleJoin(evJoin{who: joiner.sessionHandle})

	if _, ok := ch.members["j1"]; !ok {
		t.Fatal("expected alice to be a member after join")
	}

	lines := joiner.drain()
	if !containsLine(lines, "JOIN #chan") {
		t.Errorf("expected a JOIN broadcast to the joiner, got %v", lines)
	}
	if !containsLine(lines, "331") {
		t.Errorf("expected RPL_NOTOPIC for an empty topic, got %v", lines)
	}
	if !containsLine(lines, "353") {
		t.Errorf("expected RPL_NAMREPLY, got %v", lines)
	}
}

func TestChannelActorJoinTLSOnlyRejectsPlaintext(t *testing.T) {
	cd := newClientDirectory()
	joiner := newTestHandle("j1", "alice", "user", "host", false)
	registerAndClaim(cd, joiner)

	ch := newTestChannelActor("#chan", "owner-id", cd)
	ch.modes |= modeChannelTLSOnly

	ch.handleJoin(evJoin{who: joiner.sessionHandle})

	if _, ok := ch.members["j1"]; ok {
		t.Fatal("expected join to be rejected on a TLS-only channel")
	}
	if !containsLine(joiner.drain(), "+z") {
		t.Error("expected a (+z) rejection notice")
	}
}

func TestChannelActorJoinInviteOnly(t *testing.T) {
	cd := newClientDirectory()
	joiner := newTestHandle("j1", "alice", "user", "host", false)
	registerAndClaim(cd, joiner)

	ch := newTestChannelActor("#chan", "owner-id", cd)
	ch.modes |= modeChannelInviteOnly

	ch.handleJoin(evJoin{who: joiner.sessionHandle})
	if _, ok := ch.members["j1"]; ok {
		t.Fatal("expected join to be rejected without an invite")
	}
	if !containsLine(joiner.drain(), "473") {
		t.Error("expected ERR_INVITEONLYCHAN")
	}

	ch.invites["j1"] = true
	ch.handleJoin(evJoin{who: joiner.sessionHandle})
	if _, ok := ch.members["j1"]; !ok {
		t.Fatal("expected join to succeed once invited")
	}
	if ch.invites["j1"] {
		t.Error("expected invite to be consumed on successful join")
	}
}

func TestChannelActorJoinBadKey(t *testing.T) {
	cd := newClientDirectory()
	joiner := newTestHandle("j1", "alice", "user", "host", false)
	registerAndClaim(cd, joiner)

	ch := newTestChannelActor("#chan", "owner-id", cd)
	ch.modes |= modeChannelKey
	ch.key = "secret"

	ch.handleJoin(evJoin{who: joiner.sessionHandle, key: "wrong"})
	if _, ok := ch.members["j1"]; ok {
		t.Fatal("expected join with wrong key to be rejected")
	}
	if !containsLine(joiner.drain(), "475") {
		t.Error("expected ERR_BADCHANNELKEY")
	}

	ch.handleJoin(evJoin{who: joiner.sessionHandle, key: "secret"})
	if _, ok := ch.members["j1"]; !ok {
		t.Fatal("expected join with correct key to succeed")
	}
}

func TestChannelActorJoinOwnerGetsAutoOp(t *testing.T) {
	cd := newClientDirectory()
	owner := newTestHandle("owner-id", "alice", "user", "host", false)
	registerAndClaim(cd, owner)

	ch := newTestChannelActor("#chan", "owner-id", cd)
	ch.handleJoin(evJoin{who: owner.sessionHandle})

	m := ch.members["owner-id"]
	if m == nil || m.modes&modeMemberOwner == 0 {
		t.Fatal("expected channel owner to be auto-opped with modeMemberOwner on join")
	}
}

func TestChannelActorPartIsNoOpForNonMember(t *testing.T) {
	cd := newClientDirectory()
	who := newTestHandle("x", "bob", "user", "host", false)
	registerAndClaim(cd, who)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handlePart(evPart{who: who.sessionHandle, reason: "bye"})

	if len(who.drain()) != 0 {
		t.Error("expected no broadcast for a part from a non-member")
	}
}

func TestChannelActorPartRemovesMember(t *testing.T) {
	cd := newClientDirectory()
	a := newTestHandle("a", "alice", "user", "host", false)
	b := newTestHandle("b", "bob", "user", "host", false)
	registerAndClaim(cd, a)
	registerAndClaim(cd, b)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handleJoin(evJoin{who: a.sessionHandle})
	ch.handleJoin(evJoin{who: b.sessionHandle})
	a.drain()
	b.drain()

	ch.handlePart(evPart{who: a.sessionHandle, reason: "later"})

	if _, ok := ch.members["a"]; ok {
		t.Fatal("expected alice to be removed after part")
	}
	if !containsLine(b.drain(), "PART #chan :later") {
		t.Error("expected bob to see alice's PART")
	}
}

func TestChannelActorKickRequiresPrivileges(t *testing.T) {
	cd := newClientDirectory()
	a := newTestHandle("a", "alice", "user", "host", false)
	b := newTestHandle("b", "bob", "user", "host", false)
	registerAndClaim(cd, a)
	registerAndClaim(cd, b)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handleJoin(evJoin{who: a.sessionHandle})
	ch.handleJoin(evJoin{who: b.sessionHandle})
	a.drain()
	b.drain()

	ch.handleKick(evKick{by: a.sessionHandle, target: b.sessionHandle, reason: "no"})

	if _, ok := ch.members["b"]; !ok {
		t.Fatal("expected kick to fail without channel privileges")
	}
	if !containsLine(a.drain(), "482") {
		t.Error("expected ERR_CHANOPRIVSNEEDED")
	}
}

func TestChannelActorKickSucceedsWithPrivileges(t *testing.T) {
	cd := newClientDirectory()
	op := newTestHandle("op", "alice", "user", "host", false)
	target := newTestHandle("t", "bob", "user", "host", false)
	registerAndClaim(cd, op)
	registerAndClaim(cd, target)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handleJoin(evJoin{who: op.sessionHandle})
	ch.handleJoin(evJoin{who: target.sessionHandle})
	ch.members["op"].modes |= modeMemberOperator
	op.drain()
	target.drain()

	ch.handleKick(evKick{by: op.sessionHandle, target: target.sessionHandle, reason: "bye"})

	if _, ok := ch.members["t"]; ok {
		t.Fatal("expected bob to be removed")
	}
	if !containsLine(target.drain(), "KICK #chan bob :bye") {
		t.Error("expected bob to see the KICK")
	}
}

func TestChannelActorKickUnknownMemberError(t *testing.T) {
	cd := newClientDirectory()
	op := newTestHandle("op", "alice", "user", "host", false)
	target := newTestHandle("t", "bob", "user", "host", false)
	registerAndClaim(cd, op)
	registerAndClaim(cd, target)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handleJoin(evJoin{who: op.sessionHandle})
	ch.members["op"].modes |= modeMemberOperator
	op.drain()

	ch.handleKick(evKick{by: op.sessionHandle, target: target.sessionHandle, reason: "bye"})
	if !containsLine(op.drain(), "441") {
		t.Error("expected ERR_USERNOTINCHANNEL for a target that isn't a member")
	}
}

func TestChannelActorTopicRestricted(t *testing.T) {
	cd := newClientDirectory()
	a := newTestHandle("a", "alice", "user", "host", false)
	registerAndClaim(cd, a)

	ch := newTestChannelActor("#chan", "", cd)
	ch.modes |= modeChannelRestrictTopic
	ch.handleJoin(evJoin{who: a.sessionHandle})
	a.drain()

	ch.handleSetTopic(evSetTopic{by: a.sessionHandle, text: "new topic"})
	if ch.topic.text != "" {
		t.Fatal("expected topic change to be rejected without ops on a +t channel")
	}
	if !containsLine(a.drain(), "482") {
		t.Error("expected ERR_CHANOPRIVSNEEDED")
	}

	ch.members["a"].modes |= modeMemberOperator
	ch.handleSetTopic(evSetTopic{by: a.sessionHandle, text: "new topic"})
	if ch.topic.text != "new topic" {
		t.Fatal("expected topic change to succeed with ops")
	}
}

func TestChannelActorMembershipModeRemoveActuallyClears(t *testing.T) {
	cd := newClientDirectory()
	op := newTestHandle("op", "alice", "user", "host", false)
	target := newTestHandle("t", "bob", "user", "host", false)
	registerAndClaim(cd, op)
	registerAndClaim(cd, target)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handleJoin(evJoin{who: op.sessionHandle})
	ch.handleJoin(evJoin{who: target.sessionHandle})
	ch.members["op"].modes |= modeMemberOperator
	ch.members["t"].modes |= modeMemberVoice
	op.drain()
	target.drain()

	ch.handleSetMode(evSetMode{by: op.sessionHandle, modestring: "-v", args: "bob"})

	if ch.members["t"].modes&modeMemberVoice != 0 {
		t.Fatal("expected -v to clear modeMemberVoice, but it is still set (ANALYSIS.md #7)")
	}
}

func TestChannelActorChannelModeAddRemoveBroadcastsDiff(t *testing.T) {
	cd := newClientDirectory()
	op := newTestHandle("op", "alice", "user", "host", false)
	registerAndClaim(cd, op)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handleJoin(evJoin{who: op.sessionHandle})
	ch.members["op"].modes |= modeMemberOperator
	op.drain()

	ch.handleSetMode(evSetMode{by: op.sessionHandle, modestring: "+m"})
	if ch.modes&modeChannelModerated == 0 {
		t.Fatal("expected +m to be set")
	}
	lines := op.drain()
	if !containsLine(lines, "MODE #chan +m") {
		t.Errorf("expected a MODE +m broadcast, got %v", lines)
	}

	ch.handleSetMode(evSetMode{by: op.sessionHandle, modestring: "-m"})
	if ch.modes&modeChannelModerated != 0 {
		t.Fatal("expected -m to clear moderated")
	}
}

func TestChannelActorKeySetAndCleared(t *testing.T) {
	cd := newClientDirectory()
	op := newTestHandle("op", "alice", "user", "host", false)
	registerAndClaim(cd, op)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handleJoin(evJoin{who: op.sessionHandle})
	ch.members["op"].modes |= modeMemberOperator
	op.drain()

	ch.handleSetMode(evSetMode{by: op.sessionHandle, modestring: "+k", args: "hunter2"})
	if ch.key != "hunter2" || ch.modes&modeChannelKey == 0 {
		t.Fatal("expected +k to set the channel key")
	}

	ch.handleSetMode(evSetMode{by: op.sessionHandle, modestring: "-k"})
	if ch.key != "" || ch.modes&modeChannelKey != 0 {
		t.Fatal("expected -k to clear the channel key")
	}
}

func TestChannelActorBanSetAndCleared(t *testing.T) {
	cd := newClientDirectory()
	op := newTestHandle("op", "alice", "user", "host", false)
	registerAndClaim(cd, op)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handleJoin(evJoin{who: op.sessionHandle})
	ch.members["op"].modes |= modeMemberOperator
	op.drain()

	ch.handleSetMode(evSetMode{by: op.sessionHandle, modestring: "+b", args: "mallory!*@*"})
	if !ch.bans["mallory!*@*"] {
		t.Fatal("expected +b to set the ban mask")
	}
	if !containsLine(op.drain(), "MODE #chan +b mallory!*@*") {
		t.Error("expected a MODE +b broadcast")
	}

	ch.handleSetMode(evSetMode{by: op.sessionHandle, modestring: "-b", args: "mallory!*@*"})
	if ch.bans["mallory!*@*"] {
		t.Fatal("expected -b to clear the ban mask")
	}
	if !containsLine(op.drain(), "MODE #chan -b mallory!*@*") {
		t.Error("expected a MODE -b broadcast")
	}
}

func TestChannelActorJoinRejectsBannedMask(t *testing.T) {
	cd := newClientDirectory()
	joiner := newTestHandle("j1", "alice", "user", "host", false)
	registerAndClaim(cd, joiner)

	ch := newTestChannelActor("#chan", "owner-id", cd)
	ch.bans["alice!*@*"] = true

	ch.handleJoin(evJoin{who: joiner.sessionHandle})

	if _, ok := ch.members["j1"]; ok {
		t.Fatal("expected join to be rejected for a banned mask")
	}
	if !containsLine(joiner.drain(), "474") {
		t.Error("expected ERR_BANNEDFROMCHAN")
	}
}

func TestChannelActorBroadcastModeratedRequiresVoice(t *testing.T) {
	cd := newClientDirectory()
	speaker := newTestHandle("s", "alice", "user", "host", false)
	listener := newTestHandle("l", "bob", "user", "host", false)
	registerAndClaim(cd, speaker)
	registerAndClaim(cd, listener)

	ch := newTestChannelActor("#chan", "", cd)
	ch.modes |= modeChannelModerated
	ch.handleJoin(evJoin{who: speaker.sessionHandle})
	ch.handleJoin(evJoin{who: listener.sessionHandle})
	speaker.drain()
	listener.drain()

	ch.handleBroadcast(evBroadcast{by: speaker.sessionHandle, text: "hi", kind: broadcastPrivmsg})
	if !containsLine(speaker.drain(), "404") {
		t.Error("expected ERR_CANNOTSENDTOCHAN without voice on a +m channel")
	}
	if len(listener.drain()) != 0 {
		t.Error("expected no broadcast to have gone out")
	}

	ch.members["s"].modes |= modeMemberVoice
	ch.handleBroadcast(evBroadcast{by: speaker.sessionHandle, text: "hi", kind: broadcastPrivmsg})
	if !containsLine(listener.drain(), "PRIVMSG #chan :hi") {
		t.Error("expected the message to reach bob once alice has voice")
	}
}

func TestChannelActorInviteRequiresMembership(t *testing.T) {
	cd := newClientDirectory()
	inviter := newTestHandle("i", "alice", "user", "host", false)
	target := newTestHandle("t", "bob", "user", "host", false)
	registerAndClaim(cd, inviter)
	registerAndClaim(cd, target)

	ch := newTestChannelActor("#chan", "", cd)

	ch.handleInvite(evInvite{by: inviter.sessionHandle, target: target.sessionHandle})
	if !containsLine(inviter.drain(), "442") {
		t.Error("expected ERR_NOTONCHANNEL for a non-member inviter")
	}
	if ch.invites["t"] {
		t.Error("expected no invite to be recorded")
	}
}

func TestChannelActorInviteSucceeds(t *testing.T) {
	cd := newClientDirectory()
	inviter := newTestHandle("i", "alice", "user", "host", false)
	target := newTestHandle("t", "bob", "user", "host", false)
	registerAndClaim(cd, inviter)
	registerAndClaim(cd, target)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handleJoin(evJoin{who: inviter.sessionHandle})
	inviter.drain()

	ch.handleInvite(evInvite{by: inviter.sessionHandle, target: target.sessionHandle})
	if !ch.invites["t"] {
		t.Fatal("expected bob to be recorded as invited")
	}
	if !containsLine(target.drain(), "INVITE bob #chan") {
		t.Error("expected bob to receive the INVITE")
	}
}

func TestChannelActorQuitUsesQuitNotPart(t *testing.T) {
	cd := newClientDirectory()
	a := newTestHandle("a", "alice", "user", "host", false)
	b := newTestHandle("b", "bob", "user", "host", false)
	registerAndClaim(cd, a)
	registerAndClaim(cd, b)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handleJoin(evJoin{who: a.sessionHandle})
	ch.handleJoin(evJoin{who: b.sessionHandle})
	a.drain()
	b.drain()

	ch.handleQuit(evQuit{who: a.sessionHandle, reason: "Killed by root (bye)"})

	if _, ok := ch.members["a"]; ok {
		t.Fatal("expected alice to be removed")
	}
	lines := b.drain()
	if !containsLine(lines, "QUIT :Killed by root (bye)") {
		t.Errorf("expected a QUIT-shaped broadcast, got %v", lines)
	}
	for _, l := range lines {
		if strings.HasPrefix(l, "PART") || strings.Contains(l, " PART ") {
			t.Errorf("did not expect a PART-shaped broadcast for evQuit, got %v", lines)
		}
	}
}

func TestChannelActorQuitIsNoOpForNonMember(t *testing.T) {
	cd := newClientDirectory()
	who := newTestHandle("x", "ghost", "user", "host", false)
	registerAndClaim(cd, who)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handleQuit(evQuit{who: who.sessionHandle, reason: "Quit: EOF"})

	if len(who.drain()) != 0 {
		t.Error("expected no broadcast for a quit from a non-member")
	}
}

func TestChannelActorMemberRenamedBroadcastsToSelfAndOthers(t *testing.T) {
	cd := newClientDirectory()
	a := newTestHandle("a", "alice", "user", "host", false)
	b := newTestHandle("b", "bob", "user", "host", false)
	registerAndClaim(cd, a)
	registerAndClaim(cd, b)

	ch := newTestChannelActor("#chan", "", cd)
	ch.handleJoin(evJoin{who: a.sessionHandle})
	ch.handleJoin(evJoin{who: b.sessionHandle})
	a.drain()
	b.drain()

	ch.handleMemberRenamed(evMemberRenamed{id: "a", oldPrefix: "alice!user@host", newNick: "alice2"})

	if !containsLine(a.drain(), "NICK :alice2") {
		t.Error("expected alice to see her own NICK change echoed back")
	}
	if !containsLine(b.drain(), "NICK :alice2") {
		t.Error("expected bob to see alice's NICK change")
	}
}
