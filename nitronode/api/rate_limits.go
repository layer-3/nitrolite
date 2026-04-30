package api

import (
	"time"

	"github.com/layer-3/nitrolite/pkg/rpc"
)

const (
	// rateLimitStorageKey is the key used to store the token bucket in connection storage.
	rateLimitStorageKey = "rate_limiter"
)

// tokenBucket holds the mutable state for per-connection rate limiting.
type tokenBucket struct {
	tokens float64
	last   time.Time
}

// RateLimitMiddleware enforces per-connection rate limiting using a token bucket algorithm.
// It stores the token bucket in the connection's Storage for persistence across requests.
func (r *RPCRouter) RateLimitMiddleware(c *rpc.Context) {
	bucket := &tokenBucket{
		tokens: r.rateLimitBurst,
		last:   time.Now().Add(-time.Second),
	}
	if val, ok := c.Storage.Get(rateLimitStorageKey); ok {
		if b, ok := val.(*tokenBucket); ok {
			bucket = b
		}
	}

	now := time.Now()
	elapsed := now.Sub(bucket.last).Seconds()
	bucket.last = now

	// Refill tokens based on elapsed time
	bucket.tokens += elapsed * r.rateLimitPerSec
	if bucket.tokens > r.rateLimitBurst {
		bucket.tokens = r.rateLimitBurst
	}

	if bucket.tokens < 1 {
		c.Fail(nil, "rate limit exceeded")
		return
	}
	bucket.tokens--
	c.Storage.Set(rateLimitStorageKey, bucket)

	c.Next()
}
