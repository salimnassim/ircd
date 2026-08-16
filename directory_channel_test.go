package ircd

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func testChannelActorDeps(cd *clientDirectory) channelActorDeps {
	return channelActorDeps{serverName: "test.server", clients: cd}
}

func TestChannelDirectoryDispatchCreatesOnce(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	th := newTestHandle("a", "nick", "user", "host", false)
	cd.register(th.sessionHandle)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := chd.dispatch(ctx, testChannelActorDeps(cd), "#chan", evJoin{who: th.sessionHandle}, true, th.id)
	if err != nil {
		t.Fatalf("dispatch: %v", err)
	}

	if chd.count() != 1 {
		t.Fatalf("expected 1 channel, got %d", chd.count())
	}
}

func TestChannelDirectoryDispatchNoCreateReturnsNotFound(t *testing.T) {
	chd := newChannelDirectory()
	err := chd.dispatch(context.Background(), channelActorDeps{}, "#nope", evPart{}, false, "")
	if err != errChannelNotFound {
		t.Fatalf("expected errChannelNotFound, got %v", err)
	}
}

func TestChannelDirectoryDispatchEnforcesMaxChannels(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()
	chd.maxChannels = 1
	th := newTestHandle("a", "nick", "user", "host", false)
	cd.register(th.sessionHandle)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := chd.dispatch(ctx, testChannelActorDeps(cd), "#a", evJoin{who: th.sessionHandle}, true, th.id); err != nil {
		t.Fatalf("dispatch #a: %v", err)
	}

	err := chd.dispatch(ctx, testChannelActorDeps(cd), "#b", evJoin{who: th.sessionHandle}, true, th.id)
	if err != errChannelLimitReached {
		t.Fatalf("expected errChannelLimitReached, got %v", err)
	}
	if chd.count() != 1 {
		t.Fatalf("expected channel count to stay at 1, got %d", chd.count())
	}
}

func TestChannelDirectoryConcurrentFirstJoinCreatesExactlyOneChannel(t *testing.T) {
	cd := newClientDirectory()
	chd := newChannelDirectory()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	const n = 32
	handles := make([]*testHandle, n)
	for i := 0; i < n; i++ {
		id := clientID(fmt.Sprintf("id%d", i))
		handles[i] = newTestHandle(id, fmt.Sprintf("nick%d", i), "user", "host", false)
		cd.register(handles[i].sessionHandle)
	}

	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err := chd.dispatch(ctx, testChannelActorDeps(cd), "#race", evJoin{who: handles[i].sessionHandle}, true, handles[i].id)
			if err != nil {
				t.Errorf("dispatch %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()

	if got := chd.count(); got != 1 {
		t.Fatalf("expected exactly 1 channel, got %d", got)
	}

	deadline := time.After(2 * time.Second)
	for {
		snap, ok := chd.snapshot("#race")
		if ok && snap.count == n {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for all joins to land; got count=%d", snap.count)
		case <-time.After(5 * time.Millisecond):
		}
	}
}

func TestChannelDirectoryPartRaceWithJoinNeverLosesJoiner(t *testing.T) {
	cd := newClientDirectory()

	for iter := 0; iter < 200; iter++ {
		chd := newChannelDirectory()
		ctx, cancel := context.WithCancel(context.Background())

		first := newTestHandle(clientID(fmt.Sprintf("first%d", iter)), "first", "user", "host", false)
		second := newTestHandle(clientID(fmt.Sprintf("second%d", iter)), "second", "user", "host", false)
		cd.register(first.sessionHandle)
		cd.register(second.sessionHandle)

		if err := chd.dispatch(ctx, testChannelActorDeps(cd), "#race2", evJoin{who: first.sessionHandle}, true, first.id); err != nil {
			t.Fatalf("iter %d: initial join: %v", iter, err)
		}
		waitForMemberCount(t, chd, "#race2", 1)

		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			_ = chd.dispatch(ctx, testChannelActorDeps(cd), "#race2", evPart{who: first.sessionHandle}, false, "")
		}()
		go func() {
			defer wg.Done()
			_ = chd.dispatch(ctx, testChannelActorDeps(cd), "#race2", evJoin{who: second.sessionHandle}, true, second.id)
		}()
		wg.Wait()

		deadline := time.After(2 * time.Second)
		ok := false
		for !ok {
			snap, exists := chd.snapshot("#race2")
			if exists && isMemberSnapshot(snap, second.id) {
				ok = true
				break
			}
			select {
			case <-deadline:
				t.Fatalf("iter %d: second never became a visible member of #race2", iter)
			case <-time.After(2 * time.Millisecond):
			}
		}

		cancel()
	}
}

func waitForMemberCount(t *testing.T, chd *channelDirectory, name string, want int) {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		snap, ok := chd.snapshot(name)
		if ok && snap.count == want {
			return
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for %s member count to reach %d", name, want)
		case <-time.After(2 * time.Millisecond):
		}
	}
}
