package ircd

import (
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestConnLimiterGlobalCap(t *testing.T) {
	l := newConnLimiter(2, 0, 0, 0)

	if ok, reason := l.acquire("1.1.1.1"); !ok {
		t.Fatalf("acquire(1) = %v, %q, want ok", ok, reason)
	}
	if ok, reason := l.acquire("2.2.2.2"); !ok {
		t.Fatalf("acquire(2) = %v, %q, want ok", ok, reason)
	}
	ok, reason := l.acquire("3.3.3.3")
	if ok || reason != "global" {
		t.Fatalf("acquire(3) = %v, %q, want false, \"global\"", ok, reason)
	}

	l.release("1.1.1.1")
	if ok, reason := l.acquire("3.3.3.3"); !ok {
		t.Fatalf("acquire after release = %v, %q, want ok", ok, reason)
	}
}

func TestConnLimiterPerIPCap(t *testing.T) {
	l := newConnLimiter(0, 2, 0, 0)

	if ok, reason := l.acquire("1.1.1.1"); !ok {
		t.Fatalf("acquire(1) = %v, %q, want ok", ok, reason)
	}
	if ok, reason := l.acquire("1.1.1.1"); !ok {
		t.Fatalf("acquire(2) = %v, %q, want ok", ok, reason)
	}
	ok, reason := l.acquire("1.1.1.1")
	if ok || reason != "per_ip" {
		t.Fatalf("acquire(3) = %v, %q, want false, \"per_ip\"", ok, reason)
	}

	// A different IP is unaffected by the first IP's cap.
	if ok, reason := l.acquire("2.2.2.2"); !ok {
		t.Fatalf("acquire(other ip) = %v, %q, want ok", ok, reason)
	}
}

func TestConnLimiterRate(t *testing.T) {
	l := newConnLimiter(0, 0, rate.Every(time.Hour), 2)

	for i := range 2 {
		if ok, reason := l.acquire("1.1.1.1"); !ok {
			t.Fatalf("acquire(%d) = %v, %q, want ok", i, ok, reason)
		}
		l.release("1.1.1.1")
	}

	ok, reason := l.acquire("1.1.1.1")
	if ok || reason != "rate" {
		t.Fatalf("acquire after burst exhausted = %v, %q, want false, \"rate\"", ok, reason)
	}
}

func TestConnLimiterSweep(t *testing.T) {
	l := newConnLimiter(0, 0, 0, 0)
	l.idleTTL = time.Minute

	now := time.Now()
	l.perIP["idle-stale"] = &ipState{active: 0, lastSeen: now.Add(-2 * time.Minute)}
	l.perIP["idle-fresh"] = &ipState{active: 0, lastSeen: now}
	l.perIP["active-stale"] = &ipState{active: 1, lastSeen: now.Add(-2 * time.Minute)}

	l.sweep(now)

	if _, ok := l.perIP["idle-stale"]; ok {
		t.Error("sweep left idle-stale entry, want evicted")
	}
	if _, ok := l.perIP["idle-fresh"]; !ok {
		t.Error("sweep evicted idle-fresh entry, want kept")
	}
	if _, ok := l.perIP["active-stale"]; !ok {
		t.Error("sweep evicted active-stale entry, want kept (still active)")
	}
}

func TestConnLimiterNilReceiver(t *testing.T) {
	var l *connLimiter

	if ok, reason := l.acquire("1.1.1.1"); !ok {
		t.Fatalf("nil.acquire() = %v, %q, want ok", ok, reason)
	}

	l.release("1.1.1.1") // must not panic
}
