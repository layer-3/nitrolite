package metrics

import (
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
)

// labelSet renders a metric's labels as a stable "k=v,k=v" string for assertion.
func labelSet(m *dto.Metric) string {
	pairs := make([]string, 0, len(m.Label))
	for _, lp := range m.Label {
		pairs = append(pairs, lp.GetName()+"="+lp.GetValue())
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}

// gatherLabelSets returns the sorted label-set strings of every series published
// under the given fully-qualified metric name.
func gatherLabelSets(t *testing.T, reg *prometheus.Registry, name string) []string {
	t.Helper()
	mfs, err := reg.Gather()
	require.NoError(t, err)
	var out []string
	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		for _, m := range mf.Metric {
			out = append(out, labelSet(m))
		}
	}
	sort.Strings(out)
	return out
}

// gatherSeriesValues returns label-set → value for every series under the given name.
// For counters/gauges the value comes from Counter/Gauge; for histograms it returns sample count.
func gatherSeriesValues(t *testing.T, reg *prometheus.Registry, name string) map[string]float64 {
	t.Helper()
	mfs, err := reg.Gather()
	require.NoError(t, err)
	out := map[string]float64{}
	for _, mf := range mfs {
		if mf.GetName() != name {
			continue
		}
		for _, m := range mf.Metric {
			switch {
			case m.Counter != nil:
				out[labelSet(m)] = m.Counter.GetValue()
			case m.Gauge != nil:
				out[labelSet(m)] = m.Gauge.GetValue()
			case m.Histogram != nil:
				out[labelSet(m)] = float64(m.Histogram.GetSampleCount())
			}
		}
	}
	return out
}

// --- Store gauge reset semantics ---

func TestStoreMetricExporter_ResetChannels_DropsStaleLabelTuples(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp, err := NewStoreMetricExporter(reg)
	require.NoError(t, err)

	// First tick: two assets open.
	exp.SetChannels("usdc", core.ChannelStatusOpen, 3)
	exp.SetChannels("eth", core.ChannelStatusOpen, 1)
	require.ElementsMatch(t,
		[]string{"asset=eth,status=open", "asset=usdc,status=open"},
		gatherLabelSets(t, reg, "nitronode_channels_total"),
	)

	// Second tick: eth last channel closes — only usdc reported.
	exp.ResetChannels()
	exp.SetChannels("usdc", core.ChannelStatusOpen, 2)
	assert.Equal(t,
		[]string{"asset=usdc,status=open"},
		gatherLabelSets(t, reg, "nitronode_channels_total"),
		"stale eth tuple must drop after Reset+republish",
	)
}

func TestStoreMetricExporter_ResetAppSessions_DropsStaleLabelTuples(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp, err := NewStoreMetricExporter(reg)
	require.NoError(t, err)

	exp.SetAppSessions("app-a", app.AppSessionStatusOpen, 2)
	exp.SetAppSessions("app-b", app.AppSessionStatusOpen, 1)

	exp.ResetAppSessions()
	exp.SetAppSessions("app-a", app.AppSessionStatusOpen, 4)

	assert.Equal(t,
		[]string{"application_id=app-a,status=open"},
		gatherLabelSets(t, reg, "nitronode_app_sessions_total"),
	)
}

func TestStoreMetricExporter_ResetTotalValueLocked_DropsStaleLabelTuples(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp, err := NewStoreMetricExporter(reg)
	require.NoError(t, err)

	exp.SetTotalValueLocked("dom", "usdc", 100)
	exp.SetTotalValueLocked("dom", "eth", 5)

	exp.ResetTotalValueLocked()
	exp.SetTotalValueLocked("dom", "usdc", 80)

	assert.Equal(t,
		[]string{"asset=usdc,domain=dom"},
		gatherLabelSets(t, reg, "nitronode_total_value_locked"),
	)
}

func TestStoreMetricExporter_ResetNodeBalance_DropsStaleLabelTuples(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp, err := NewStoreMetricExporter(reg)
	require.NoError(t, err)

	exp.SetNodeBalance("1", "usdc", 100)
	exp.SetNodeBalance("137", "matic", 50)

	exp.ResetNodeBalance()
	exp.SetNodeBalance("1", "usdc", 90)

	assert.Equal(t,
		[]string{"asset=usdc,blockchain_id=1"},
		gatherLabelSets(t, reg, "nitronode_node_balance"),
	)
}

func TestStoreMetricExporter_ResetUserBalances_ClearsAllThreeGauges(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp, err := NewStoreMetricExporter(reg)
	require.NoError(t, err)

	exp.SetUserBalanceTotal("1", "usdc", 100)
	exp.SetUserBalanceUnderfunded("1", "usdc", 10)
	exp.SetUserBalanceReleasable("1", "usdc", 5)
	exp.SetUserBalanceTotal("137", "matic", 50)

	exp.ResetUserBalances()
	exp.SetUserBalanceTotal("1", "usdc", 90)
	// underfunded / releasable intentionally not re-set this tick.

	assert.Equal(t,
		[]string{"asset=usdc,blockchain_id=1"},
		gatherLabelSets(t, reg, "nitronode_user_balance_total"),
	)
	assert.Empty(t,
		gatherLabelSets(t, reg, "nitronode_user_balance_underfunded"),
		"underfunded must be cleared with the family",
	)
	assert.Empty(t,
		gatherLabelSets(t, reg, "nitronode_user_balance_releasable"),
		"releasable must be cleared with the family",
	)
}

// ResetActiveUsers / ResetActiveAppSessions use DeletePartialMatch(timespan=...),
// so a failure in one timespan must not blank the others.

func TestStoreMetricExporter_ResetActiveUsers_ClearsOnlyTargetTimespan(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp, err := NewStoreMetricExporter(reg)
	require.NoError(t, err)

	exp.SetActiveUsers("usdc", "day", 10)
	exp.SetActiveUsers("usdc", "week", 50)
	exp.SetActiveUsers("usdc", "month", 200)
	exp.SetActiveUsers("eth", "day", 1)
	exp.SetActiveUsers("eth", "week", 4)

	exp.ResetActiveUsers("day")

	got := gatherLabelSets(t, reg, "nitronode_users_active")
	assert.NotContains(t, got, "asset=usdc,timespan=day")
	assert.NotContains(t, got, "asset=eth,timespan=day")
	assert.Contains(t, got, "asset=usdc,timespan=week")
	assert.Contains(t, got, "asset=usdc,timespan=month")
	assert.Contains(t, got, "asset=eth,timespan=week")
}

func TestStoreMetricExporter_ResetActiveAppSessions_ClearsOnlyTargetTimespan(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp, err := NewStoreMetricExporter(reg)
	require.NoError(t, err)

	exp.SetActiveAppSessions("app-a", "day", 3)
	exp.SetActiveAppSessions("app-a", "week", 10)
	exp.SetActiveAppSessions("app-b", "day", 1)
	exp.SetActiveAppSessions("app-b", "month", 7)

	exp.ResetActiveAppSessions("day")

	got := gatherLabelSets(t, reg, "nitronode_app_sessions_active")
	assert.NotContains(t, got, "application_id=app-a,timespan=day")
	assert.NotContains(t, got, "application_id=app-b,timespan=day")
	assert.Contains(t, got, "application_id=app-a,timespan=week")
	assert.Contains(t, got, "application_id=app-b,timespan=month")
}

// --- Runtime exporter cold-start seeding ---

func TestNewRuntimeMetricExporter_SeedsChannelStateSigValidations(t *testing.T) {
	reg := prometheus.NewRegistry()
	_, err := NewRuntimeMetricExporter(reg)
	require.NoError(t, err)

	got := gatherSeriesValues(t, reg, "nitronode_channel_state_sig_validations_total")
	// One series per (signer_type × result) pair, all at value 0.
	expected := map[string]float64{}
	for _, st := range core.ChannelSignerTypes {
		for _, res := range allActionResults {
			expected["result="+res.String()+",sig_type="+st.String()] = 0
		}
	}
	assert.Equal(t, expected, got)
}

func TestSeedRPCMethodMetrics_PublishesZeroValuedSeries(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp, err := NewRuntimeMetricExporter(reg)
	require.NoError(t, err)

	rt := exp.(*runtimeMetricExporter)
	methods := []string{"channels.v1.submit_state", "app_sessions.v1.create_app_session"}
	rt.SeedRPCMethodMetrics(methods)

	// rpc_messages_emitted_total: 3 msg_types × N methods, all at 0.
	msgs := gatherSeriesValues(t, reg, "nitronode_rpc_messages_emitted_total")
	require.Len(t, msgs, 3*len(methods))
	for _, m := range methods {
		for _, mt := range []string{"req", "resp", "error"} {
			key := "method=" + m + ",msg_type=" + mt
			v, ok := msgs[key]
			assert.True(t, ok, "missing series %s", key)
			assert.Zero(t, v)
		}
	}

	// rpc_requests_total: N methods × {path=default} × 2 results, all at 0.
	reqs := gatherSeriesValues(t, reg, "nitronode_rpc_requests_total")
	require.Len(t, reqs, 2*len(methods))
	for _, m := range methods {
		for _, res := range allActionResults {
			key := "method=" + m + ",path=default,result=" + res.String()
			v, ok := reqs[key]
			assert.True(t, ok, "missing series %s", key)
			assert.Zero(t, v)
		}
	}

	// rpc_request_duration_seconds (histogram): same label space, sample count 0.
	durs := gatherSeriesValues(t, reg, "nitronode_rpc_request_duration_seconds")
	assert.Len(t, durs, 2*len(methods))
	for _, v := range durs {
		assert.Zero(t, v)
	}

	// rpc_inflight: one series per method (bounded `method` label), gauge at 0.
	inflight := gatherSeriesValues(t, reg, "nitronode_rpc_inflight")
	require.Len(t, inflight, len(methods))
	for _, m := range methods {
		v, ok := inflight["method="+m]
		assert.True(t, ok, "missing rpc_inflight series for method %s", m)
		assert.Zero(t, v)
	}
}

func TestSeedRPCMethodMetrics_EmptyMethodsIsNoop(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp, err := NewRuntimeMetricExporter(reg)
	require.NoError(t, err)

	rt := exp.(*runtimeMetricExporter)
	rt.SeedRPCMethodMetrics(nil)

	assert.Empty(t, gatherSeriesValues(t, reg, "nitronode_rpc_messages_emitted_total"))
	assert.Empty(t, gatherSeriesValues(t, reg, "nitronode_rpc_requests_total"))
	assert.Empty(t, gatherSeriesValues(t, reg, "nitronode_rpc_inflight"))
}

func TestSeedBlockchainEventMetrics_PublishesZeroValuedSeries(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp, err := NewRuntimeMetricExporter(reg)
	require.NoError(t, err)

	rt := exp.(*runtimeMetricExporter)
	ids := []uint64{1, 137, 8453}
	rt.SeedBlockchainEventMetrics(ids)

	got := gatherSeriesValues(t, reg, "nitronode_blockchain_events_total")
	require.Len(t, got, len(ids)*len(allActionResults))
	for _, id := range ids {
		for _, res := range allActionResults {
			key := "blockchain_id=" + strconv.FormatUint(id, 10) + ",result=" + res.String()
			v, ok := got[key]
			assert.True(t, ok, "missing series %s", key)
			assert.Zero(t, v)
		}
	}
}

func TestSeedBlockchainEventMetrics_PreservesPriorIncrements(t *testing.T) {
	reg := prometheus.NewRegistry()
	exp, err := NewRuntimeMetricExporter(reg)
	require.NoError(t, err)

	exp.IncBlockchainEvent(1, true)
	exp.IncBlockchainEvent(1, true)

	rt := exp.(*runtimeMetricExporter)
	rt.SeedBlockchainEventMetrics([]uint64{1, 137})

	got := gatherSeriesValues(t, reg, "nitronode_blockchain_events_total")
	assert.Equal(t, 2.0, got["blockchain_id=1,result=success"],
		"WithLabelValues without Inc must not reset existing counter")
	assert.Equal(t, 0.0, got["blockchain_id=1,result=failed"])
	assert.Equal(t, 0.0, got["blockchain_id=137,result=success"])
	assert.Equal(t, 0.0, got["blockchain_id=137,result=failed"])
}
