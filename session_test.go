package ircd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"testing"
	"time"
)

func TestSessionNickClaimAndCollision(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	a, aTH := newTestSession("a", deps)
	cd.register(aTH.sessionHandle)
	b, bTH := newTestSession("b", deps)
	cd.register(bTH.sessionHandle)

	a.cmdNick(message{params: []string{"alice"}})
	if a.nick != "alice" {
		t.Fatalf("expected alice to claim her nick, got %q", a.nick)
	}

	b.cmdNick(message{params: []string{"alice"}})
	if b.nick == "alice" {
		t.Fatal("expected bob to be rejected claiming an in-use nick")
	}
	if !containsLine(bTH.drain(), "433") {
		t.Error("expected ERR_NICKNAMEINUSE")
	}

	b.cmdNick(message{params: []string{"bob"}})
	if b.nick != "bob" {
		t.Fatalf("expected bob to claim his own nick, got %q", b.nick)
	}
}

func TestSessionHandshakeCompletesOnNickAndUser(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	s, th := newTestSession("a", deps)
	cd.register(th.sessionHandle)

	s.cmdNick(message{params: []string{"alice"}})
	if s.handshakeDone {
		t.Fatal("handshake should not complete until USER is also set")
	}

	s.cmdUser(context.Background(), message{params: []string{"aliceuser", "0", "0", "Alice Realname"}})

	if !s.handshakeDone {
		t.Fatal("expected handshake to complete once NICK and USER are both set")
	}

	lines := th.drain()
	if !containsLine(lines, "001 alice") {
		t.Errorf("expected RPL_WELCOME in the handshake burst, got %v", lines)
	}
	if !containsLine(lines, "376 alice") {
		t.Errorf("expected RPL_ENDOFMOTD, got %v", lines)
	}
}

func TestSessionUserTwiceRejected(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	s, th := newTestSession("a", deps)
	cd.register(th.sessionHandle)
	s.cmdNick(message{params: []string{"alice"}})
	s.cmdUser(context.Background(), message{params: []string{"u", "0", "0", "R"}})
	th.drain()

	s.cmdUser(context.Background(), message{params: []string{"u2", "0", "0", "R2"}})
	if !containsLine(th.drain(), "462") {
		t.Error("expected ERR_ALREADYREGISTERED on a second USER")
	}
	if s.user != "u" {
		t.Fatal("expected the original username to be preserved")
	}
}

func registerSession(cd *clientDirectory, s *session, th *testHandle, nick string) {
	cd.register(th.sessionHandle)
	s.nick = nick
	_ = cd.claimNick(s.id, nick)
	s.handshakeDone = true
	s.publishSnapshot()
}

func TestSessionJoinDispatchesAndTracksMembership(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	s, th := newTestSession("a", deps)
	registerSession(cd, s, th, "alice")

	s.cmdJoin(context.Background(), message{params: []string{"#chan"}})

	if _, ok := s.memberOf["#chan"]; !ok {
		t.Fatal("expected #chan to be recorded in the session's local membership cache")
	}

	deadline := time.After(2 * time.Second)
	for {
		snap, ok := chd.snapshot("#chan")
		if ok && snap.count == 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timed out waiting for the channel actor to process the join")
		case <-time.After(2 * time.Millisecond):
		}
	}

	if !containsLine(th.drain(), "JOIN #chan") {
		t.Error("expected alice to see her own JOIN")
	}
}

func TestSessionPrivmsgDeliversDirectlyToUser(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	a, aTH := newTestSession("a", deps)
	registerSession(cd, a, aTH, "alice")
	b, bTH := newTestSession("b", deps)
	registerSession(cd, b, bTH, "bob")

	a.cmdPrivmsg(context.Background(), message{params: []string{"bob", "hello there"}})

	if !containsLine(bTH.drain(), "PRIVMSG bob :hello there") {
		t.Error("expected bob to receive alice's direct message")
	}
}

func TestSessionModeClientTogglesInvisible(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	s, th := newTestSession("a", deps)
	registerSession(cd, s, th, "alice")

	s.cmdModeClient(message{params: []string{"alice", "+i"}})
	if s.modes&modeClientInvisible == 0 {
		t.Fatal("expected +i to set modeClientInvisible")
	}
	if !containsLine(th.drain(), "221 alice +i") {
		t.Error("expected RPL_UMODEIS reflecting +i")
	}

	s.cmdModeClient(message{params: []string{"alice", "-i"}})
	if s.modes&modeClientInvisible != 0 {
		t.Fatal("expected -i to clear modeClientInvisible")
	}
}

func TestSessionModeClientRejectsOtherTargets(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	s, th := newTestSession("a", deps)
	registerSession(cd, s, th, "alice")

	s.cmdModeClient(message{params: []string{"someoneelse", "+i"}})
	if !containsLine(th.drain(), "502") {
		t.Error("expected ERR_USERSDONTMATCH")
	}
}

func TestSessionQuitTerminatesWithReason(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	s, th := newTestSession("a", deps)
	registerSession(cd, s, th, "alice")

	s.cmdQuit(message{params: []string{"goodbye", "world"}})

	ok, cause := th.wasTerminated()
	if !ok {
		t.Fatal("expected QUIT to call terminate")
	}
	if cause.Error() != "Quit: goodbye world" {
		t.Errorf("unexpected quit cause: %q", cause.Error())
	}
}

func TestSessionKillRequiresOper(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	killer, killerTH := newTestSession("k", deps)
	registerSession(cd, killer, killerTH, "mallory")
	target, targetTH := newTestSession("t", deps)
	registerSession(cd, target, targetTH, "victim")

	killer.cmdKill(message{params: []string{"victim", "just because"}})

	if !containsLine(killerTH.drain(), "481") {
		t.Error("expected ERR_NOPRIVILEGES for a non-oper KILL")
	}
	if ok, _ := targetTH.wasTerminated(); ok {
		t.Fatal("expected the target to survive a KILL from a non-oper")
	}
}

func TestSessionKillTerminatesTarget(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	killer, killerTH := newTestSession("k", deps)
	registerSession(cd, killer, killerTH, "root")
	killer.modes |= modeClientOperator

	target, targetTH := newTestSession("t", deps)
	registerSession(cd, target, targetTH, "victim")

	killer.cmdKill(message{params: []string{"victim", "spamming"}})

	ok, cause := targetTH.wasTerminated()
	if !ok {
		t.Fatal("expected KILL from an oper to terminate the target")
	}
	if !errors.Is(cause, errKilled) {
		t.Errorf("expected errors.Is(cause, errKilled), cause was: %v", cause)
	}
	if cause.Error() != "Killed by root (spamming)" {
		t.Errorf("unexpected kill cause text: %q", cause.Error())
	}
}

func drainDiscard(conn net.Conn) {
	go io.Copy(io.Discard, conn)
}

func waitForRegistration(t *testing.T, cd *clientDirectory, id clientID) *sessionHandle {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		if h, ok := cd.lookupID(id); ok {
			return h
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for session %s to register", id)
			return nil
		case <-time.After(2 * time.Millisecond):
		}
	}
}

func TestRunSessionEOFTriggersCleanup(t *testing.T) {
	serverConn, clientConn := net.Pipe()

	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	drainDiscard(clientConn)

	done := make(chan struct{})
	go func() {
		runSession(ctx, serverConn, "eof1", false, deps)
		close(done)
	}()

	waitForRegistration(t, cd, "eof1")

	clientConn.Close()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("expected runSession to return after the connection closed")
	}

	if _, ok := cd.lookupID("eof1"); ok {
		t.Fatal("expected the session to be unregistered after cleanup")
	}
}

func TestRunSessionOutboxOverflowDisconnects(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	defer clientConn.Close()

	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)
	deps.outboxSize = 1

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		runSession(ctx, serverConn, "ov1", false, deps)
		close(done)
	}()

	h := waitForRegistration(t, cd, "ov1")

	for i := 0; i < 20; i++ {
		h.deliver(fmt.Sprintf("line %d", i))
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("expected SendQ overflow to terminate the session")
	}
}

func TestRunSessionPingTimeoutDisconnects(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	drainDiscard(clientConn)
	defer clientConn.Close()

	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)
	deps.pingFrequency = 20 * time.Millisecond
	deps.pongMaxLatency = 20 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		runSession(ctx, serverConn, "pt1", false, deps)
		close(done)
	}()

	waitForRegistration(t, cd, "pt1")

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("expected a ping timeout to terminate the session")
	}
}

func TestSessionTerminateIsIdempotent(t *testing.T) {
	serverConn, clientConn := net.Pipe()
	drainDiscard(clientConn)
	defer clientConn.Close()

	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		runSession(ctx, serverConn, "idem1", false, deps)
		close(done)
	}()

	h := waitForRegistration(t, cd, "idem1")

	for i := 0; i < 10; i++ {
		go h.terminate(fmt.Errorf("Quit: cause %d", i))
	}

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("expected runSession to return")
	}
}
