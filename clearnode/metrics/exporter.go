package metrics

import (
	"fmt"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

const (
	// MetricNamespace is the common namespace for all ClearNode metrics.
	MetricNamespace = "clearnode"
)

var (
	_ RuntimeMetricExporter = (*runtimeMetricExporter)(nil)
	_ StoreMetricExporter   = (*storeMetricExporter)(nil)
)

type storeMetricExporter struct {
	appSessionsTotal       *prometheus.GaugeVec
	channelsTotal          *prometheus.GaugeVec
	usersActive            *prometheus.GaugeVec
	appSessionsActive      *prometheus.GaugeVec
	totalValueLocked       *prometheus.GaugeVec
	nodeBalance            *prometheus.GaugeVec
	userBalanceTotal       *prometheus.GaugeVec
	userBalanceUnderfunded *prometheus.GaugeVec
	userBalanceReleasable  *prometheus.GaugeVec
}

func NewStoreMetricExporter(reg prometheus.Registerer) (StoreMetricExporter, error) {
	m := &storeMetricExporter{
		appSessionsTotal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: MetricNamespace,
			Name:      "app_sessions_total",
			Help:      "Current total number of app sessions",
		}, []string{"application", "status"}),
		channelsTotal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: MetricNamespace,
			Name:      "channels_total",
			Help:      "Current total number of channels",
		}, []string{"asset", "status"}),
		usersActive: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: MetricNamespace,
			Name:      "users_active",
			Help:      "Current total active users",
		}, []string{"asset", "timespan"}),
		appSessionsActive: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: MetricNamespace,
			Name:      "app_sessions_active",
			Help:      "Current total active app sessions",
		}, []string{"application", "timespan"}),
		totalValueLocked: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: MetricNamespace,
			Name:      "total_value_locked",
			Help:      "Total value locked by domain and asset",
		}, []string{"domain", "asset"}),
		nodeBalance: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: MetricNamespace,
			Name:      "node_balance",
			Help:      "Node's available on-chain balance by blockchain and asset",
		}, []string{"blockchain_id", "asset"}),
		userBalanceTotal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: MetricNamespace,
			Name:      "user_balance_total",
			Help:      "Total user balance obligations by blockchain and asset",
		}, []string{"blockchain_id", "asset"}),
		userBalanceUnderfunded: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: MetricNamespace,
			Name:      "user_balance_underfunded",
			Help:      "User balance exceeding on-chain locked amount by blockchain and asset",
		}, []string{"blockchain_id", "asset"}),
		userBalanceReleasable: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: MetricNamespace,
			Name:      "user_balance_releasable",
			Help:      "On-chain locked amount exceeding user balance by blockchain and asset",
		}, []string{"blockchain_id", "asset"}),
	}

	if reg != nil {
		reg.MustRegister(
			m.appSessionsTotal,
			m.channelsTotal,
			m.usersActive,
			m.appSessionsActive,
			m.totalValueLocked,
			m.nodeBalance,
			m.userBalanceTotal,
			m.userBalanceUnderfunded,
			m.userBalanceReleasable,
		)
	} else {
		return nil, fmt.Errorf("prometheus registerer not provided")
	}

	return m, nil
}

func (m *storeMetricExporter) SetAppSessions(applicationID string, status app.AppSessionStatus, count uint64) {
	m.appSessionsTotal.WithLabelValues(applicationID, status.String()).Set(float64(count))
}

func (m *storeMetricExporter) SetChannels(asset string, status core.ChannelStatus, count uint64) {
	m.channelsTotal.WithLabelValues(asset, status.String()).Set(float64(count))
}

func (m *storeMetricExporter) SetActiveUsers(asset, timeSpanLabel string, count uint64) {
	m.usersActive.WithLabelValues(asset, timeSpanLabel).Set(float64(count))
}

func (m *storeMetricExporter) SetActiveAppSessions(applicationID, timeSpanLabel string, count uint64) {
	m.appSessionsActive.WithLabelValues(applicationID, timeSpanLabel).Set(float64(count))
}

func (m *storeMetricExporter) SetTotalValueLocked(domain, asset string, value float64) {
	m.totalValueLocked.WithLabelValues(domain, asset).Set(value)
}

func (m *storeMetricExporter) SetNodeBalance(blockchainID, asset string, value float64) {
	m.nodeBalance.WithLabelValues(blockchainID, asset).Set(value)
}

func (m *storeMetricExporter) SetUserBalanceTotal(blockchainID, asset string, value float64) {
	m.userBalanceTotal.WithLabelValues(blockchainID, asset).Set(value)
}

func (m *storeMetricExporter) SetUserBalanceUnderfunded(blockchainID, asset string, value float64) {
	m.userBalanceUnderfunded.WithLabelValues(blockchainID, asset).Set(value)
}

func (m *storeMetricExporter) SetUserBalanceReleasable(blockchainID, asset string, value float64) {
	m.userBalanceReleasable.WithLabelValues(blockchainID, asset).Set(value)
}

// runtimeMetricExporter is the concrete implementation of the Metrics interface.
type runtimeMetricExporter struct {
	// Shared Metrics (Cross-Package)
	userStatesTotal                 *prometheus.CounterVec
	transactionsTotal               *prometheus.CounterVec
	transactionsAmountTotal         *prometheus.CounterVec
	channelStateSigValidationsTotal *prometheus.CounterVec

	// api/rpc_router.go
	rpcMessagesTotal          *prometheus.CounterVec
	rpcRequestsTotal          *prometheus.CounterVec
	rpcRequestDurationSeconds *prometheus.HistogramVec
	rpcConnectionsTotal       *prometheus.GaugeVec

	// api/app_session_v1
	appStateUpdates                     *prometheus.CounterVec
	appSessionUpdateSigValidationsTotal *prometheus.CounterVec

	// Blockchain Worker
	blockchainActionsTotal *prometheus.CounterVec

	// Event Listener
	blockchainEventsTotal *prometheus.CounterVec

	// Metric Worker
	channelSessionKeysTotal prometheus.Counter
	appSessionKeysTotal     prometheus.Counter
}

// RuntimeMetricExporter exposes metrics related to runtime operations, such as API requests, channel state validations, and blockchain interactions.
func NewRuntimeMetricExporter(reg prometheus.Registerer) (RuntimeMetricExporter, error) {
	m := &runtimeMetricExporter{
		// Shared
		userStatesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Name:      "user_states_total",
			Help:      "Total number of user states",
		}, []string{"asset", "home_blockchain_id", "transition"}),
		transactionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Name:      "transactions_total",
			Help:      "Total number of transactions",
		}, []string{"asset", "tx_type"}),
		transactionsAmountTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Name:      "transactions_amount_total",
			Help:      "Total amount of transactions processed",
		}, []string{"asset", "tx_type"}),
		channelSessionKeysTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Name:      "channel_session_keys_total",
			Help:      "Total number of channel session keys issued",
		}),
		appSessionKeysTotal: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Name:      "app_session_keys_total",
			Help:      "Total number of app session keys issued",
		}),
		channelStateSigValidationsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Name:      "channel_state_sig_validations_total",
			Help:      "Total number of channel state signature validations",
		}, []string{"sig_type", "result"}),

		// api/rpc_router
		rpcMessagesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Name:      "rpc_messages_total",
			Help:      "Total number of RPC messages",
		}, []string{"msg_type", "method"}),
		rpcRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Name:      "rpc_requests_total",
			Help:      "Total number of RPC requests",
		}, []string{"method", "path", "status"}),
		rpcRequestDurationSeconds: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: MetricNamespace,
			Name:      "rpc_request_duration_seconds",
			Help:      "Duration of RPC requests in seconds",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
		}, []string{"method", "path", "status"}),
		rpcConnectionsTotal: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: MetricNamespace,
			Name:      "rpc_connections_active",
			Help:      "Current number of active RPC connections",
		}, []string{"region", "origin"}),

		// api/app_session_v1
		appStateUpdates: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Name:      "app_state_updates_total",
			Help:      "Total number of app state updates",
		}, []string{"application"}),
		appSessionUpdateSigValidationsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Name:      "app_session_update_sig_validations_total",
			Help:      "Total number of app session update signature validations",
		}, []string{"application", "sig_type", "result"}),

		// Blockchain Worker
		blockchainActionsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Name:      "blockchain_actions_total",
			Help:      "Total number of blockchain actions",
		}, []string{"asset", "blockchain_id", "action_type", "result"}),

		// Event Listener
		blockchainEventsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: MetricNamespace,
			Name:      "blockchain_events_total",
			Help:      "Total number of blockchain events processed",
		}, []string{"blockchain_id", "process_result"}),
	}

	if reg != nil {
		reg.MustRegister(
			m.userStatesTotal,
			m.transactionsTotal,
			m.transactionsAmountTotal,
			m.channelStateSigValidationsTotal,
			m.rpcMessagesTotal,
			m.rpcRequestsTotal,
			m.rpcRequestDurationSeconds,
			m.rpcConnectionsTotal,
			m.appStateUpdates,
			m.appSessionUpdateSigValidationsTotal,
			m.blockchainActionsTotal,
			m.blockchainEventsTotal,
			m.channelSessionKeysTotal,
			m.appSessionKeysTotal,
		)
	} else {
		return nil, fmt.Errorf("prometheus registerer not provided")
	}

	return m, nil
}

// Shared
func (m *runtimeMetricExporter) IncUserState(asset string, homeBlockchainID uint64, transition core.TransitionType) {
	homeBlockchainIDStr := strconv.FormatUint(homeBlockchainID, 10)
	m.userStatesTotal.WithLabelValues(asset, homeBlockchainIDStr, transition.String()).Inc()
}

func (m *runtimeMetricExporter) RecordTransaction(asset string, txType core.TransactionType, amount decimal.Decimal) {
	m.transactionsTotal.WithLabelValues(asset, txType.String()).Inc()
	m.transactionsAmountTotal.WithLabelValues(asset, txType.String()).Add(amount.InexactFloat64())
}

// api/rpc_router
func (m *runtimeMetricExporter) IncRPCMessage(msgType rpc.MsgType, method string) {
	m.rpcMessagesTotal.WithLabelValues(msgType.String(), method).Inc()
}

func (m *runtimeMetricExporter) IncRPCRequest(method, path string, success bool) {
	result := ActionResultFailed
	if success {
		result = ActionResultSuccess
	}
	m.rpcRequestsTotal.WithLabelValues(method, path, result.String()).Inc()
}

func (m *runtimeMetricExporter) ObserveRPCDuration(method, path string, success bool, duration time.Duration) {
	result := ActionResultFailed
	if success {
		result = ActionResultSuccess
	}
	m.rpcRequestDurationSeconds.WithLabelValues(method, path, result.String()).Observe(duration.Seconds())
}

func (m *runtimeMetricExporter) SetRPCConnections(region, origin string, count uint32) {
	m.rpcConnectionsTotal.WithLabelValues(region, origin).Set(float64(count))
}

// api/app_session_v1
func (m *runtimeMetricExporter) IncAppStateUpdate(applicationID string) {
	m.appStateUpdates.WithLabelValues(applicationID).Inc()
}

func (m *runtimeMetricExporter) IncAppSessionUpdateSigValidation(applicationID string, signerType app.AppSessionSignerTypeV1, result bool) {
	res := ActionResultFailed
	if result {
		res = ActionResultSuccess
	}
	m.appSessionUpdateSigValidationsTotal.WithLabelValues(applicationID, signerType.String(), res.String()).Inc()
}

func (m *runtimeMetricExporter) IncChannelStateSigValidation(sigType core.ChannelSignerType, result bool) {
	res := ActionResultFailed
	if result {
		res = ActionResultSuccess
	}
	m.channelStateSigValidationsTotal.WithLabelValues(sigType.String(), res.String()).Inc()
}

// Blockchain Worker
func (m *runtimeMetricExporter) IncBlockchainAction(asset string, blockchainID uint64, actionType string, result bool) {
	stringBlockchainID := strconv.FormatUint(blockchainID, 10)
	res := ActionResultFailed
	if result {
		res = ActionResultSuccess
	}
	m.blockchainActionsTotal.WithLabelValues(asset, stringBlockchainID, actionType, res.String()).Inc()
}

// Event Listener
func (m *runtimeMetricExporter) IncBlockchainEvent(blockchainID uint64, result bool) {
	stringBlockchainID := strconv.FormatUint(blockchainID, 10)
	res := ActionResultFailed
	if result {
		res = ActionResultSuccess
	}
	m.blockchainEventsTotal.WithLabelValues(stringBlockchainID, res.String()).Inc()
}

// Metric Worker
func (m *runtimeMetricExporter) IncChannelSessionKeys() {
	m.channelSessionKeysTotal.Inc()
}

func (m *runtimeMetricExporter) IncAppSessionKeys() {
	m.appSessionKeysTotal.Inc()
}

type ActionResult string

const (
	ActionResultSuccess ActionResult = "success"
	ActionResultFailed  ActionResult = "failed"
)

func (res ActionResult) String() string {
	return string(res)
}

// api/channel_v1
// -* `user_states_total{asset, home_blockchain_id, transition}`
// -* `transactions_total{asset, tx_type}`
// -* `transactions_amount_total{asset, tx_type}`
// - `channel_state_validations_total{sig_type, result}`

// api/rpc_router.go
// - `rpc_messages_total{msg_type, method}`
// - `rpc_requests_total{method, status}`
// - `rpc_request_duration_seconds{method, path, status}`
// - `rpc_connections_total{region}`

// api/app_session_v1
// -* `app_state_updates{application}`
// - `app_session_update_sig_validations_total{application, sig_type, result}`
// -* `user_states_total{asset, home_blockchain_id, transition}`
// -* `transactions_total{asset, tx_type}`
// -* `transactions_amount_total{asset, tx_type}`
// - `channel_state_sig_validations_total{sig_type, result}`

// Blockchain Worker
// -* `blockchain_actions_total{asset, blockchain_id, action_type, result}`

// Event Listener
// -* `blockchain_events_total{blockchain_id, process_result}`

// By the end of this story a separate metric worker should expose the following metrics:

// metric worker
// - `channel_session_keys_total`
// - `app_session_keys_total`
// - `app_sessions_total{application,status}`
// - `channels_total{asset,status}`
