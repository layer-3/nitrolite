package rpc

import "time"

// FrameRateLimiter decides whether an inbound WebSocket frame is admitted.
// Implementations attached to a single connection may assume serial access
// from the connection's read goroutine; implementations shared across
// connections must be safe for concurrent use.
//
// Returning false causes the connection to close.
//
// Allocation note: Admit runs after the frame has been read off the wire and
// allocated on the Go heap. Per-frame size is bounded by SetReadLimit (see
// WebsocketConnectionConfig.MaxMessageSize), but a burst of N back-to-back
// max-sized frames can briefly hold up to N * MaxMessageSize bytes per
// connection before this hook closes it. Burst capacity should be sized with
// that ceiling in mind.
type FrameRateLimiter interface {
	// Admit reports whether a frame of size bytes is permitted at now.
	Admit(now time.Time, size int) bool
}

// NoopFrameRateLimiter accepts every frame. Default when no limiter is
// configured; useful for tests and dev environments.
type NoopFrameRateLimiter struct{}

// Admit always returns true.
func (NoopFrameRateLimiter) Admit(time.Time, int) bool { return true }

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

// Admit refills tokens for elapsed time, caps at burst, then debits size.
// Returns false if size exceeds available tokens.
func (b *ByteTokenBucket) Admit(now time.Time, size int) bool {
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
