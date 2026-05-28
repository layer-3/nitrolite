package api

import (
	"testing"
	"time"

	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitMiddleware(t *testing.T) {
	t.Parallel()

	newTestRouter := func(ratePerSec, burst float64) *RPCRouter {
		return &RPCRouter{
			rateLimitPerSec: ratePerSec,
			rateLimitBurst:  burst,
		}
	}

	newTestContext := func(storage *rpc.SafeStorage, requestID uint64) *rpc.Context {
		return &rpc.Context{
			Storage: storage,
			Request: rpc.Message{RequestID: requestID},
		}
	}

	isRateLimited := func(ctx *rpc.Context) bool {
		err := ctx.Response.Error()
		return err != nil && err.Error() == "rate limit exceeded"
	}

	t.Run("allows requests within burst limit", func(t *testing.T) {
		t.Parallel()

		router := newTestRouter(10, 5)
		storage := rpc.NewSafeStorage()

		var allowed int
		for i := 0; i < 5; i++ {
			ctx := newTestContext(storage, uint64(i))
			router.RateLimitMiddleware(ctx)
			if !isRateLimited(ctx) {
				allowed++
			}
		}

		assert.Equal(t, 5, allowed, "all requests within burst should be allowed")
	})

	t.Run("blocks requests exceeding burst", func(t *testing.T) {
		t.Parallel()

		router := newTestRouter(10, 3)
		storage := rpc.NewSafeStorage()

		var allowed, rateLimited int
		for i := 0; i < 5; i++ {
			ctx := newTestContext(storage, uint64(i))
			router.RateLimitMiddleware(ctx)
			if isRateLimited(ctx) {
				rateLimited++
			} else {
				allowed++
			}
		}

		assert.Equal(t, 3, allowed, "only burst amount of requests should be allowed")
		assert.Equal(t, 2, rateLimited, "excess requests should be rate limited")
	})

	t.Run("returns rate limit error message", func(t *testing.T) {
		t.Parallel()

		router := newTestRouter(10, 1)
		storage := rpc.NewSafeStorage()

		// First request - allowed
		ctx1 := newTestContext(storage, 1)
		router.RateLimitMiddleware(ctx1)
		require.False(t, isRateLimited(ctx1), "first request should be allowed")

		// Second request - should be rate limited
		ctx2 := newTestContext(storage, 2)
		router.RateLimitMiddleware(ctx2)

		require.True(t, isRateLimited(ctx2), "second request should be rate limited")
		assert.Equal(t, "rate limit exceeded", ctx2.Response.Error().Error())
	})

	t.Run("tokens refill over time", func(t *testing.T) {
		t.Parallel()

		router := newTestRouter(100, 2) // 100 tokens per second = 1 token per 10ms
		storage := rpc.NewSafeStorage()

		// Exhaust the bucket
		for i := 0; i < 2; i++ {
			ctx := newTestContext(storage, uint64(i))
			router.RateLimitMiddleware(ctx)
			require.False(t, isRateLimited(ctx), "initial requests should be allowed")
		}

		// This should be rate limited
		ctx := newTestContext(storage, 100)
		router.RateLimitMiddleware(ctx)
		require.True(t, isRateLimited(ctx), "should be rate limited after exhausting bucket")

		// Wait for tokens to refill (need 1 token, at 100/sec = 10ms per token)
		time.Sleep(15 * time.Millisecond)

		// Now it should work
		ctx = newTestContext(storage, 101)
		router.RateLimitMiddleware(ctx)
		assert.False(t, isRateLimited(ctx), "should be allowed after refill")
	})

	t.Run("separate storage has separate buckets", func(t *testing.T) {
		t.Parallel()

		router := newTestRouter(10, 2)
		storage1 := rpc.NewSafeStorage()
		storage2 := rpc.NewSafeStorage()

		// Exhaust storage1's bucket
		for i := 0; i < 2; i++ {
			ctx := newTestContext(storage1, uint64(i))
			router.RateLimitMiddleware(ctx)
		}

		// storage1 should be rate limited
		ctx1 := newTestContext(storage1, 100)
		router.RateLimitMiddleware(ctx1)
		assert.True(t, isRateLimited(ctx1), "storage1 should be rate limited")

		// storage2 should still have its own bucket
		ctx2 := newTestContext(storage2, 200)
		router.RateLimitMiddleware(ctx2)
		assert.False(t, isRateLimited(ctx2), "storage2 should have its own bucket")
	})

	t.Run("bucket persists in storage", func(t *testing.T) {
		t.Parallel()

		router := newTestRouter(10, 5)
		storage := rpc.NewSafeStorage()

		// Make one request
		ctx := newTestContext(storage, 1)
		router.RateLimitMiddleware(ctx)

		// Check bucket is stored
		val, ok := storage.Get(rateLimitStorageKey)
		require.True(t, ok, "bucket should be stored")

		bucket, ok := val.(*tokenBucket)
		require.True(t, ok, "stored value should be a tokenBucket")
		assert.Less(t, bucket.tokens, 5.0, "tokens should have been consumed")
	})
}
