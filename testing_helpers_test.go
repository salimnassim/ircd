package ircd

import (
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

var (
	testNickRegex    = regexp.MustCompile(`([a-zA-Z0-9\[\]\{\}\\\|]{2,31})`)
	testChannelRegex = regexp.MustCompile(`[#!&][^\x00\x07\x0a\x0d\x20\x2C\x3A]{1,50}`)
)

func testSessionDeps(cd *clientDirectory, chd *channelDirectory) sessionDeps {
	return sessionDeps{
		serverName:     "test.server",
		network:        "TestNet",
		version:        "0.0.0-test",
		isupport:       "TEST=1",
		motd:           []string{"line one"},
		cloakSecret:    []byte("test-cloak-secret"),
		ports:          func() []string { return []string{"6667"} },
		pingFrequency:  30 * time.Second,
		pongMaxLatency: 10 * time.Second,
		rateLimit:      rate.Every(time.Millisecond),
		rateBurst:      1000,
		outboxSize:     256,
		clients:        cd,
		channels:       chd,
		operators:      NewOperatorStore(),
		nickRegex:      testNickRegex,
		channelRegex:   testChannelRegex,
	}
}

func newTestSession(id clientID, deps sessionDeps) (*session, *testHandle) {
	th := newTestHandle(id, "", "", "", false)
	sess := &session{
		id:       id,
		address:  "127.0.0.1",
		memberOf: make(map[string]struct{}),
		ponged:   make(chan struct{}, 1),
		handle:   th.sessionHandle,
		deps:     deps,
	}
	sess.publishSnapshot()
	return sess, th
}

type testHandle struct {
	*sessionHandle

	mu          sync.Mutex
	terminated  []error
	terminateFn func(error)
}

func newTestHandle(id clientID, nick, user, host string, tls bool) *testHandle {
	th := &testHandle{}
	h := &sessionHandle{
		id:     id,
		outbox: make(chan string, 256),
	}
	h.terminate = func(cause error) {
		th.mu.Lock()
		th.terminated = append(th.terminated, cause)
		th.mu.Unlock()
		if th.terminateFn != nil {
			th.terminateFn(cause)
		}
	}
	h.snapshot.Store(&sessionSnapshot{
		id: id, nick: nick, user: user, host: host, tls: tls,
	})
	th.sessionHandle = h
	return th
}

func (th *testHandle) drain() []string {
	var out []string
	for {
		select {
		case line := <-th.outbox:
			out = append(out, line)
		default:
			return out
		}
	}
}

func (th *testHandle) wasTerminated() (bool, error) {
	th.mu.Lock()
	defer th.mu.Unlock()
	if len(th.terminated) == 0 {
		return false, nil
	}
	return true, th.terminated[len(th.terminated)-1]
}

func registerAndClaim(cd *clientDirectory, th *testHandle) {
	cd.register(th.sessionHandle)
	snap := th.snapshot.Load()
	if snap != nil && snap.nick != "" {
		_ = cd.claimNick(th.id, snap.nick)
	}
}

func containsLine(lines []string, substr string) bool {
	for _, l := range lines {
		if strings.Contains(l, substr) {
			return true
		}
	}
	return false
}

// drainUntil polls th's outbox, accumulating lines, until one contains
// substr or the deadline elapses. Useful for asserting on replies that
// arrive asynchronously via a channel actor's goroutine.
func drainUntil(t *testing.T, th *testHandle, substr string) []string {
	t.Helper()
	var out []string
	deadline := time.After(2 * time.Second)
	for {
		out = append(out, th.drain()...)
		if containsLine(out, substr) {
			return out
		}
		select {
		case <-deadline:
			t.Fatalf("timed out waiting for a line containing %q, got %v", substr, out)
			return out
		case <-time.After(2 * time.Millisecond):
		}
	}
}
