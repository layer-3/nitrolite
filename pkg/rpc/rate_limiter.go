package rpc

import "time"

// FrameRateLimiter decides whether an inbound WebSocket frame is admitted.
// Implementations attached to a single connection may assume serial access
// from the connection's read goroutine; implementations shared across
// connections must be safe for concurrent use.
//
// Returning false causes the connection to close.
type FrameRateLimiter interface {
	// Allow reports whether a frame of size bytes is permitted at now.
	Allow(now time.Time, size int) bool
}

// NoopFrameRateLimiter accepts every frame. Default when no limiter is
// configured; useful for tests and dev environments.
type NoopFrameRateLimiter struct{}

// Allow always returns true.
func (NoopFrameRateLimiter) Allow(time.Time, int) bool { return true }

// ByteTokenBucket is a token bucket on bytes read. One bucket per connection.
// Not safe for concurrent use; the connection's read goroutine is the sole
// caller of Allow.
type ByteTokenBucket struct {
	bytesPerSec float64
	burst       float64
	tokens      float64
	last        time.Time
}

// NewByteTokenBucket returns a bucket pre-filled to burst capacity.
//
// bytesPerSec is the steady-state refill rate in bytes per second.
// burst is the maximum bucket size; a single frame larger than burst is
// always rejected.
func NewByteTokenBucket(bytesPerSec, burst float64) *ByteTokenBucket {
	return &ByteTokenBucket{
		bytesPerSec: bytesPerSec,
		burst:       burst,
		tokens:      burst,
	}
}

// Allow refills tokens for elapsed time, caps at burst, then debits size.
// Returns false if size exceeds available tokens.
func (b *ByteTokenBucket) Allow(now time.Time, size int) bool {
	if !b.last.IsZero() {
		b.tokens += now.Sub(b.last).Seconds() * b.bytesPerSec
		if b.tokens > b.burst {
			b.tokens = b.burst
		}
	}
	b.last = now

	cost := float64(size)
	if b.tokens < cost {
		return false
	}
	b.tokens -= cost
	return true
}
