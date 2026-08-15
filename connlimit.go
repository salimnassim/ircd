package ircd

import (
	"context"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type connLimiter struct {
	maxGlobal int
	maxPerIP  int
	connRate  rate.Limit
	connBurst int

	sweepEvery time.Duration
	idleTTL    time.Duration

	mu     sync.Mutex
	global int
	perIP  map[string]*ipState
}

type ipState struct {
	active   int
	limiter  *rate.Limiter
	lastSeen time.Time
}

func newConnLimiter(maxGlobal, maxPerIP int, connRate rate.Limit, connBurst int) *connLimiter {
	return &connLimiter{
		maxGlobal:  maxGlobal,
		maxPerIP:   maxPerIP,
		connRate:   connRate,
		connBurst:  connBurst,
		sweepEvery: 5 * time.Minute,
		idleTTL:    10 * time.Minute,
		perIP:      make(map[string]*ipState),
	}
}

func (l *connLimiter) acquire(ip string) (ok bool, reason string) {
	if l == nil {
		return true, ""
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.maxGlobal > 0 && l.global >= l.maxGlobal {
		return false, "global"
	}

	st, exists := l.perIP[ip]
	if !exists {
		st = &ipState{}
		if l.connRate > 0 {
			st.limiter = rate.NewLimiter(l.connRate, l.connBurst)
		}
		l.perIP[ip] = st
	}
	st.lastSeen = time.Now()

	if l.maxPerIP > 0 && st.active >= l.maxPerIP {
		return false, "per_ip"
	}
	if st.limiter != nil && !st.limiter.Allow() {
		return false, "rate"
	}

	l.global++
	st.active++
	return true, ""
}

func (l *connLimiter) release(ip string) {
	if l == nil {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.global > 0 {
		l.global--
	}
	if st, ok := l.perIP[ip]; ok && st.active > 0 {
		st.active--
		st.lastSeen = time.Now()
	}
}

func (l *connLimiter) sweep(now time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	for ip, st := range l.perIP {
		if st.active == 0 && now.Sub(st.lastSeen) > l.idleTTL {
			delete(l.perIP, ip)
		}
	}
}

func (l *connLimiter) runSweeper(ctx context.Context) {
	ticker := time.NewTicker(l.sweepEvery)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			l.sweep(now)
		}
	}
}
