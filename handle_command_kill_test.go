package ircd

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestKillBroadcastsKilledByReasonToSharedChannels(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	deps := testSessionDeps(cd, chd)

	killer, killerTH := newTestSession("k", deps)
	registerSession(cd, killer, killerTH, "root")
	killer.modes |= modeClientOperator

	target, targetTH := newTestSession("t", deps)
	registerSession(cd, target, targetTH, "victim")

	witnessTH := newTestHandle("w", "witness", "user", "host", false)
	registerAndClaim(cd, witnessTH)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := chd.dispatch(ctx, target.channelDeps(), "#chan", evJoin{who: target.handle}, true, target.id); err != nil {
		t.Fatalf("target join: %v", err)
	}
	if err := chd.dispatch(ctx, target.channelDeps(), "#chan", evJoin{who: witnessTH.sessionHandle}, false, ""); err != nil {
		t.Fatalf("witness join: %v", err)
	}
	waitForChannelMemberCount(t, chd, "#chan", 2)
	targetTH.drain()
	witnessTH.drain()
	target.memberOf["#chan"] = struct{}{}

	killer.cmdKill(message{params: []string{"victim", "spamming"}})

	ok, cause := targetTH.wasTerminated()
	if !ok {
		t.Fatal("expected KILL to terminate the target")
	}
	if !errors.Is(cause, errKilled) {
		t.Fatalf("expected errors.Is(cause, errKilled), got %v", cause)
	}

	target.cleanup(cause)

	deadline := time.After(2 * time.Second)
	for {
		lines := witnessTH.drain()
		if containsLine(lines, "QUIT :Killed by root (spamming)") {
			break
		}
		select {
		case <-deadline:
			t.Fatal("witness never saw the Killed-by QUIT notice")
		case <-time.After(2 * time.Millisecond):
		}
	}

	if _, ok := cd.lookupID("t"); ok {
		t.Fatal("expected the killed session to be unregistered after cleanup")
	}
}

func waitForChannelMemberCount(t *testing.T, chd *channelDirectory, name string, want int) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		snap, ok := chd.snapshot(name)
		if ok && snap.count == want {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for %s to reach %d members", name, want)
		case <-time.After(2 * time.Millisecond):
		}
	}
}
