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

	require.True(t, b.Admit(base, 4096), "full burst admitted")
	require.False(t, b.Admit(base, 1), "empty bucket rejects")
}

func TestByteTokenBucket_Refill(t *testing.T) {
	t.Parallel()

	b := rpc.NewByteTokenBucket(1024, 4096)
	base := time.Unix(0, 0)

	require.True(t, b.Admit(base, 4096))
	require.False(t, b.Admit(base, 1))
	require.True(t, b.Admit(base.Add(time.Second), 1024), "1s of refill admits 1024 bytes")
	require.False(t, b.Admit(base.Add(time.Second), 1), "bucket emptied again")
}

func TestByteTokenBucket_BurstCap(t *testing.T) {
	t.Parallel()

	b := rpc.NewByteTokenBucket(1024, 4096)
	base := time.Unix(0, 0)

	// Long idle must not let tokens grow past burst.
	require.True(t, b.Admit(base.Add(time.Hour), 4096))
	require.False(t, b.Admit(base.Add(time.Hour), 1))
}

func TestByteTokenBucket_FrameLargerThanBurstRejected(t *testing.T) {
	t.Parallel()

	b := rpc.NewByteTokenBucket(1024, 4096)
	require.False(t, b.Admit(time.Unix(0, 0), 4097),
		"frame larger than burst is always rejected")
}

func TestByteTokenBucket_PartialRefill(t *testing.T) {
	t.Parallel()

	b := rpc.NewByteTokenBucket(1000, 1000)
	base := time.Unix(0, 0)

	require.True(t, b.Admit(base, 1000))
	require.False(t, b.Admit(base.Add(500*time.Millisecond), 501),
		"500ms refills 500 bytes, not enough for 501")
	require.True(t, b.Admit(base.Add(500*time.Millisecond), 500),
		"500ms refills exactly 500 bytes")
}

func TestNoopFrameRateLimiter_AdmitsAll(t *testing.T) {
	t.Parallel()

	var lim rpc.FrameRateLimiter = rpc.NoopFrameRateLimiter{}
	require.True(t, lim.Admit(time.Now(), 1<<30))
	require.True(t, lim.Admit(time.Time{}, 0))
}

func TestRequestTokenBucket_BurstThenEmpty(t *testing.T) {
	t.Parallel()

	b := rpc.NewRequestTokenBucket(10, 3)
	base := time.Unix(0, 0)

	// Frame size is ignored: each frame costs exactly one token.
	require.True(t, b.Admit(base, 9999))
	require.True(t, b.Admit(base, 0))
	require.True(t, b.Admit(base, 1))
	require.False(t, b.Admit(base, 1), "fourth frame exceeds burst of 3")
}

func TestRequestTokenBucket_Refill(t *testing.T) {
	t.Parallel()

	b := rpc.NewRequestTokenBucket(10, 2) // 10 frames/sec => 1 token per 100ms
	base := time.Unix(0, 0)

	require.True(t, b.Admit(base, 1))
	require.True(t, b.Admit(base, 1))
	require.False(t, b.Admit(base, 1), "bucket emptied")
	require.False(t, b.Admit(base.Add(50*time.Millisecond), 1), "50ms < 100ms, no token yet")
	require.True(t, b.Admit(base.Add(100*time.Millisecond), 1), "100ms refills one token")
}

func TestRequestTokenBucket_BurstCap(t *testing.T) {
	t.Parallel()

	b := rpc.NewRequestTokenBucket(10, 3)
	base := time.Unix(0, 0)

	// Long idle must not let tokens grow past burst.
	require.True(t, b.Admit(base.Add(time.Hour), 1))
	require.True(t, b.Admit(base.Add(time.Hour), 1))
	require.True(t, b.Admit(base.Add(time.Hour), 1))
	require.False(t, b.Admit(base.Add(time.Hour), 1), "capped at burst of 3")
}

func TestRequestTokenBucket_SubUnitBurstRejectsAll(t *testing.T) {
	t.Parallel()

	b := rpc.NewRequestTokenBucket(10, 0.5)
	require.False(t, b.Admit(time.Unix(0, 0), 1), "burst < 1 admits no frame")
}

func TestCompositeFrameRateLimiter_RejectsWhenAnyMemberRejects(t *testing.T) {
	t.Parallel()

	base := time.Unix(0, 0)
	// Byte budget is generous; request budget is the binding constraint.
	lim := rpc.CompositeFrameRateLimiter{
		rpc.NewByteTokenBucket(1<<20, 1<<20),
		rpc.NewRequestTokenBucket(10, 2),
	}

	require.True(t, lim.Admit(base, 50))
	require.True(t, lim.Admit(base, 50))
	require.False(t, lim.Admit(base, 50), "request bucket exhausted closes the frame out")
}

func TestCompositeFrameRateLimiter_EmptyAndNilMembers(t *testing.T) {
	t.Parallel()

	require.True(t, rpc.CompositeFrameRateLimiter{}.Admit(time.Unix(0, 0), 100),
		"no members admits everything")

	lim := rpc.CompositeFrameRateLimiter{nil, rpc.NewRequestTokenBucket(10, 1)}
	require.True(t, lim.Admit(time.Unix(0, 0), 1), "nil members are skipped")
	require.False(t, lim.Admit(time.Unix(0, 0), 1), "non-nil member still enforced")
}
