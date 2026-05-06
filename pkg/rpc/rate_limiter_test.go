package rpc_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/rpc"
)

func TestByteTokenBucket_BurstThenEmpty(t *testing.T) {
	t.Parallel()

	b := rpc.NewByteTokenBucket(1024, 4096)
	base := time.Unix(0, 0)

	require.True(t, b.Allow(base, 4096), "full burst admitted")
	require.False(t, b.Allow(base, 1), "empty bucket rejects")
}

func TestByteTokenBucket_Refill(t *testing.T) {
	t.Parallel()

	b := rpc.NewByteTokenBucket(1024, 4096)
	base := time.Unix(0, 0)

	require.True(t, b.Allow(base, 4096))
	require.False(t, b.Allow(base, 1))
	require.True(t, b.Allow(base.Add(time.Second), 1024), "1s of refill admits 1024 bytes")
	require.False(t, b.Allow(base.Add(time.Second), 1), "bucket emptied again")
}

func TestByteTokenBucket_BurstCap(t *testing.T) {
	t.Parallel()

	b := rpc.NewByteTokenBucket(1024, 4096)
	base := time.Unix(0, 0)

	// Long idle must not let tokens grow past burst.
	require.True(t, b.Allow(base.Add(time.Hour), 4096))
	require.False(t, b.Allow(base.Add(time.Hour), 1))
}

func TestByteTokenBucket_FrameLargerThanBurstRejected(t *testing.T) {
	t.Parallel()

	b := rpc.NewByteTokenBucket(1024, 4096)
	require.False(t, b.Allow(time.Unix(0, 0), 4097),
		"frame larger than burst is always rejected")
}

func TestByteTokenBucket_PartialRefill(t *testing.T) {
	t.Parallel()

	b := rpc.NewByteTokenBucket(1000, 1000)
	base := time.Unix(0, 0)

	require.True(t, b.Allow(base, 1000))
	require.False(t, b.Allow(base.Add(500*time.Millisecond), 501),
		"500ms refills 500 bytes, not enough for 501")
	require.True(t, b.Allow(base.Add(500*time.Millisecond), 500),
		"500ms refills exactly 500 bytes")
}

func TestNoopFrameRateLimiter_AllowsAll(t *testing.T) {
	t.Parallel()

	var lim rpc.FrameRateLimiter = rpc.NoopFrameRateLimiter{}
	require.True(t, lim.Allow(time.Now(), 1<<30))
	require.True(t, lim.Allow(time.Time{}, 0))
}
