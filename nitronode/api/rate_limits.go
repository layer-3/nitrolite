package api

import (
	"time"

	"github.com/layer-3/nitrolite/pkg/rpc"
)

// rateLimitStorageKey is the per-connection SafeStorage key for the request-rate
// token bucket. The bucket is allocated on the first request and mutated in
// place on subsequent ones — processRequests dispatches the middleware chain
// serially per connection, so a single load is sufficient.
const rateLimitStorageKey = "rate_limiter"

// tokenBucket holds the mutable state for per-connection rate limiting.
type tokenBucket struct {
	tokens float64
	last   time.Time
}

// RateLimitMiddleware enforces a per-connection request-count token bucket.
// It complements the per-frame byte budget enforced by FrameRateLimiter at the
// connection layer: bytes guard bandwidth, this guards RPC throughput so a
// flood of small requests cannot bypass the byte cap.
//
// On overrun the request fails with an RPC error and the connection stays
// open; the byte limiter is the layer that closes connections.
func (r *RPCRouter) RateLimitMiddleware(c *rpc.Context) {
	bucket := loadOrInitBucket(c, r.rateLimitBurst)

	now := time.Now()
	bucket.tokens += now.Sub(bucket.last).Seconds() * r.rateLimitPerSec
	if bucket.tokens > r.rateLimitBurst {
		bucket.tokens = r.rateLimitBurst
	}
	bucket.last = now

	if bucket.tokens < 1 {
		c.Fail(rpc.Errorf("rate limit exceeded"), "")
		return
	}
	bucket.tokens--

	c.Next()
}

// loadOrInitBucket returns the bucket stored on the connection, allocating a
// fresh one pre-filled to burst on first use. The bucket is stored as a
// pointer; later mutations are visible without re-Set.
func loadOrInitBucket(c *rpc.Context, burst float64) *tokenBucket {
	if v, ok := c.Storage.Get(rateLimitStorageKey); ok {
		if b, ok := v.(*tokenBucket); ok {
			return b
		}
	}
	b := &tokenBucket{tokens: burst, last: time.Now()}
	c.Storage.Set(rateLimitStorageKey, b)
	return b
}
