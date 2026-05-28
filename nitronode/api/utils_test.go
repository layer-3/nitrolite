package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// TestMethodPathDomains_MatchesEnums fails if the bounded enums grow but
// MethodPathDomains is not updated, which would silently re-introduce the
// cold-start gap that the metric-seeding fix closes.
func TestMethodPathDomains_MatchesEnums(t *testing.T) {
	domains := MethodPathDomains()

	intentMethod := rpc.AppSessionsV1SubmitAppStateMethod.String()
	require.Contains(t, domains, intentMethod)
	require.Len(t, domains[intentMethod], len(app.AllAppStateUpdateIntents))
	wantIntents := make([]string, 0, len(app.AllAppStateUpdateIntents))
	for _, i := range app.AllAppStateUpdateIntents {
		wantIntents = append(wantIntents, i.String())
	}
	assert.ElementsMatch(t, wantIntents, domains[intentMethod])

	wantTransitions := make([]string, 0, len(core.AllTransitionTypes))
	for _, ttype := range core.AllTransitionTypes {
		wantTransitions = append(wantTransitions, ttype.String())
	}
	for _, m := range []string{
		rpc.ChannelsV1RequestCreationMethod.String(),
		rpc.ChannelsV1SubmitStateMethod.String(),
	} {
		require.Contains(t, domains, m)
		require.Len(t, domains[m], len(core.AllTransitionTypes))
		assert.ElementsMatch(t, wantTransitions, domains[m])
	}

	// No "unknown" sentinel ever leaks in — enum String() should cover every
	// listed value, and any new enum value without a String() arm would land
	// here as "unknown" and fail the test.
	for _, paths := range domains {
		for _, p := range paths {
			assert.NotEqual(t, "unknown", p, "enum without String() arm in MethodPathDomains output")
		}
	}
}

// TestMethodPathDomains_CoversGetMethodPathSwitch is a hand-maintained pairing
// between getMethodPath's switch arms and MethodPathDomains' map keys. If a
// new method gains payload-derived path logic in getMethodPath, this test
// fails until the same method is added to MethodPathDomains (and vice versa).
func TestMethodPathDomains_CoversGetMethodPathSwitch(t *testing.T) {
	expected := []string{
		rpc.AppSessionsV1SubmitAppStateMethod.String(),
		rpc.ChannelsV1RequestCreationMethod.String(),
		rpc.ChannelsV1SubmitStateMethod.String(),
	}
	got := MethodPathDomains()
	keys := make([]string, 0, len(got))
	for k := range got {
		keys = append(keys, k)
	}
	assert.ElementsMatch(t, expected, keys)
}
