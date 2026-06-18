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
// caller of Admit.
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

// RequestTokenBucket is a token bucket on frame COUNT: one token per frame,
// regardless of size. One bucket per connection; not safe for concurrent use,
// the connection's read goroutine is the sole caller of Admit.
//
// It complements ByteTokenBucket: bytes guard bandwidth, request count guards
// RPC throughput so a flood of tiny frames — including malformed or
// unknown-method frames that never reach the handler chain — cannot drive CPU
// past the intended rate while staying under the byte cap.
type RequestTokenBucket struct {
	perSec float64
	burst  float64
	tokens float64
	last   time.Time
}

// NewRequestTokenBucket returns a bucket pre-filled to burst capacity.
//
// perSec is the steady-state refill rate in frames per second.
// burst is the maximum bucket size; with burst < 1 no frame is ever admitted.
func NewRequestTokenBucket(perSec, burst float64) *RequestTokenBucket {
	return &RequestTokenBucket{
		perSec: perSec,
		burst:  burst,
		tokens: burst,
	}
}

// Admit refills tokens for elapsed time, caps at burst, then debits one token.
// The frame size is ignored. Returns false when fewer than one token remains.
func (b *RequestTokenBucket) Admit(now time.Time, _ int) bool {
	if !b.last.IsZero() {
		b.tokens += now.Sub(b.last).Seconds() * b.perSec
		if b.tokens > b.burst {
			b.tokens = b.burst
		}
	}
	b.last = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// CompositeFrameRateLimiter admits a frame only when every member admits it.
// Members are consulted in order and short-circuit on the first rejection;
// because a rejection closes the connection, a member debited before a later
// member rejects is harmless. nil members are skipped.
type CompositeFrameRateLimiter []FrameRateLimiter

// Admit returns false as soon as any member rejects the frame.
func (c CompositeFrameRateLimiter) Admit(now time.Time, size int) bool {
	for _, l := range c {
		if l == nil {
			continue
		}
		if !l.Admit(now, size) {
			return false
		}
	}
	return true
}
