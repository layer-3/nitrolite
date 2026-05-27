package server

import (
	"sync"
	"time"
)

type rateLimiter struct {
	mu       sync.Mutex
	cooldown time.Duration
	seen     map[string]time.Time
	calls    uint64
}

func newRateLimiter(cooldown time.Duration) *rateLimiter {
	return &rateLimiter{
		cooldown: cooldown,
		seen:     make(map[string]time.Time),
	}
}

// checkAndRecord atomically checks if key is allowed and, if so, records the attempt.
// Returns true if allowed (cooldown slot consumed), false if on cooldown.
func (r *rateLimiter) checkAndRecord(key string) bool {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls++
	r.evictExpiredLocked(now)
	last, exists := r.seen[key]
	if exists && now.Sub(last) < r.cooldown {
		return false
	}
	r.seen[key] = now
	return true
}

// checkAndRecordBoth atomically checks both addr and ip. Only records both if
// both pass — prevents a blocked IP from burning the wallet's cooldown slot.
// Returns (false, "address") or (false, "ip") if either is blocked.
func (r *rateLimiter) checkAndRecordBoth(addr, ip string) (bool, string) {
	now := time.Now()
	r.mu.Lock()
	defer r.mu.Unlock()

	r.calls++
	r.evictExpiredLocked(now)

	if last, ok := r.seen[addr]; ok && now.Sub(last) < r.cooldown {
		return false, "address"
	}
	if last, ok := r.seen[ip]; ok && now.Sub(last) < r.cooldown {
		return false, "ip"
	}

	r.seen[addr] = now
	if ip != addr {
		r.seen[ip] = now
	}
	return true, ""
}

// evictExpiredLocked removes entries that are past the cooldown window.
// Must be called with r.mu held. Runs every 1024 calls to amortise cost.
func (r *rateLimiter) evictExpiredLocked(now time.Time) {
	if r.calls%1024 != 0 {
		return
	}
	for k, ts := range r.seen {
		if now.Sub(ts) >= r.cooldown {
			delete(r.seen, k)
		}
	}
}
