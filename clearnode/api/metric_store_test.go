package api

import (
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/layer-3/nitrolite/clearnode/store/database"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
)

// ── minimal stubs ────────────────────────────────────────────────────────────

// panicDatabaseStore panics on every method, making any unintended call obvious.
// Embed it in test-specific structs and override only the methods under test.
type panicDatabaseStore struct{}

func (panicDatabaseStore) ExecuteInTransaction(database.StoreTxHandler) error { panic("not impl") }
func (panicDatabaseStore) GetUserBalances(string) ([]core.BalanceEntry, error) {
	panic("not impl")
}
func (panicDatabaseStore) LockUserState(string, string) (decimal.Decimal, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetUserTransactions(string, *string, *core.TransactionType, *uint64, *uint64, *core.PaginationParams) ([]core.Transaction, core.PaginationMetadata, error) {
	panic("not impl")
}
func (panicDatabaseStore) RecordTransaction(core.Transaction) error              { panic("not impl") }
func (panicDatabaseStore) CreateChannel(core.Channel) error                      { panic("not impl") }
func (panicDatabaseStore) GetChannelByID(string) (*core.Channel, error)          { panic("not impl") }
func (panicDatabaseStore) GetActiveHomeChannel(string, string) (*core.Channel, error) {
	panic("not impl")
}
func (panicDatabaseStore) CheckOpenChannel(string, string) (string, bool, error) {
	panic("not impl")
}
func (panicDatabaseStore) UpdateChannel(core.Channel) error { panic("not impl") }
func (panicDatabaseStore) GetUserChannels(string, *core.ChannelStatus, *string, *core.ChannelType, uint32, uint32) ([]core.Channel, uint32, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetLastStateByChannelID(string, bool) (*core.State, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetStateByChannelIDAndVersion(string, uint64) (*core.State, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetLastUserState(string, string, bool) (*core.State, error) {
	panic("not impl")
}
func (panicDatabaseStore) StoreUserState(core.State) error               { panic("not impl") }
func (panicDatabaseStore) EnsureNoOngoingStateTransitions(string, string) error { panic("not impl") }
func (panicDatabaseStore) ScheduleInitiateEscrowWithdrawal(string, uint64) error {
	panic("not impl")
}
func (panicDatabaseStore) ScheduleCheckpoint(string, uint64) error          { panic("not impl") }
func (panicDatabaseStore) ScheduleFinalizeEscrowDeposit(string, uint64) error { panic("not impl") }
func (panicDatabaseStore) ScheduleFinalizeEscrowWithdrawal(string, uint64) error {
	panic("not impl")
}
func (panicDatabaseStore) ScheduleInitiateEscrowDeposit(string, uint64) error { panic("not impl") }
func (panicDatabaseStore) Fail(int64, string) error                            { panic("not impl") }
func (panicDatabaseStore) FailNoRetry(int64, string) error                     { panic("not impl") }
func (panicDatabaseStore) RecordAttempt(int64, string) error                   { panic("not impl") }
func (panicDatabaseStore) Complete(int64, string) error                        { panic("not impl") }
func (panicDatabaseStore) GetActions(uint8, uint64) ([]database.BlockchainAction, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetStateByID(string) (*core.State, error)              { panic("not impl") }
func (panicDatabaseStore) CreateApp(app.AppV1) error                              { panic("not impl") }
func (panicDatabaseStore) GetApp(string) (*app.AppInfoV1, error)                  { panic("not impl") }
func (panicDatabaseStore) GetApps(*string, *string, *core.PaginationParams) ([]app.AppInfoV1, core.PaginationMetadata, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetAppCount(string) (uint64, error)                                    { panic("not impl") }
func (panicDatabaseStore) CreateAppSession(app.AppSessionV1) error                               { panic("not impl") }
func (panicDatabaseStore) GetAppSession(string) (*app.AppSessionV1, error)                       { panic("not impl") }
func (panicDatabaseStore) GetAppSessions(*string, *string, app.AppSessionStatus, *core.PaginationParams) ([]app.AppSessionV1, core.PaginationMetadata, error) {
	panic("not impl")
}
func (panicDatabaseStore) UpdateAppSession(app.AppSessionV1) error { panic("not impl") }
func (panicDatabaseStore) GetAppSessionBalances(string) (map[string]decimal.Decimal, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetParticipantAllocations(string) (map[string]map[string]decimal.Decimal, error) {
	panic("not impl")
}
func (panicDatabaseStore) RecordLedgerEntry(string, string, string, decimal.Decimal) error {
	panic("not impl")
}
func (panicDatabaseStore) StoreAppSessionKeyState(app.AppSessionKeyStateV1) error { panic("not impl") }
func (panicDatabaseStore) GetAppSessionKeyOwner(string, string) (string, error)   { panic("not impl") }
func (panicDatabaseStore) GetLastAppSessionKeyVersion(string, string) (uint64, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetLastAppSessionKeyState(string, string) (*app.AppSessionKeyStateV1, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetLastAppSessionKeyStates(string, *string) ([]app.AppSessionKeyStateV1, error) {
	panic("not impl")
}
func (panicDatabaseStore) StoreChannelSessionKeyState(core.ChannelSessionKeyStateV1) error {
	panic("not impl")
}
func (panicDatabaseStore) GetLastChannelSessionKeyVersion(string, string) (uint64, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetLastChannelSessionKeyStates(string, *string) ([]core.ChannelSessionKeyStateV1, error) {
	panic("not impl")
}
func (panicDatabaseStore) ValidateChannelSessionKeyForAsset(string, string, string, string) (bool, error) {
	panic("not impl")
}
func (panicDatabaseStore) CountActiveUsers(time.Duration) ([]database.ActiveCountByLabel, error) {
	panic("not impl")
}
func (panicDatabaseStore) CountActiveAppSessions(time.Duration) ([]database.ActiveCountByLabel, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetLifetimeMetricLastTimestamp(string) (time.Time, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetAppSessionsCountByLabels() ([]database.AppSessionCount, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetChannelsCountByLabels() ([]database.ChannelCount, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetTotalValueLocked() ([]database.TotalValueLocked, error) {
	panic("not impl")
}
func (panicDatabaseStore) SetNodeBalance(uint64, string, decimal.Decimal) error { panic("not impl") }
func (panicDatabaseStore) RefreshUserEnforcedBalance(string, string) error      { panic("not impl") }
func (panicDatabaseStore) GetNodeBalance() ([]database.NodeBalance, error)       { panic("not impl") }
func (panicDatabaseStore) GetUserBalanceSummary() ([]database.UserBalanceSummary, error) {
	panic("not impl")
}
func (panicDatabaseStore) UpdateUserStaked(string, uint64, decimal.Decimal) error { panic("not impl") }
func (panicDatabaseStore) GetTotalUserStaked(string) (decimal.Decimal, error)     { panic("not impl") }
func (panicDatabaseStore) RecordAction(string, core.GatedAction) error             { panic("not impl") }
func (panicDatabaseStore) GetUserActionCount(string, core.GatedAction, time.Duration) (uint64, error) {
	panic("not impl")
}
func (panicDatabaseStore) GetUserActionCounts(string, time.Duration) (map[core.GatedAction]uint64, error) {
	panic("not impl")
}
func (panicDatabaseStore) StoreContractEvent(core.BlockchainEvent) error { panic("not impl") }
func (panicDatabaseStore) GetLatestContractEventBlockNumber(string, uint64) (uint64, error) {
	panic("not impl")
}
func (panicDatabaseStore) IsContractEventPresent(uint64, uint64, string, uint32) (bool, error) {
	panic("not impl")
}

// recordingDatabaseStore embeds panicDatabaseStore and records RecordTransaction /
// StoreUserState / UpdateAppSession / StoreChannelSessionKeyState / StoreAppSessionKeyState calls.
type recordingDatabaseStore struct {
	panicDatabaseStore
	recordTransactionFn       func(core.Transaction) error
	storeUserStateFn          func(core.State) error
	updateAppSessionFn        func(app.AppSessionV1) error
	storeChannelSessionKeyFn  func(core.ChannelSessionKeyStateV1) error
	storeAppSessionKeyStateFn func(app.AppSessionKeyStateV1) error
}

func (r *recordingDatabaseStore) RecordTransaction(tx core.Transaction) error {
	if r.recordTransactionFn != nil {
		return r.recordTransactionFn(tx)
	}
	return nil
}
func (r *recordingDatabaseStore) StoreUserState(state core.State) error {
	if r.storeUserStateFn != nil {
		return r.storeUserStateFn(state)
	}
	return nil
}
func (r *recordingDatabaseStore) UpdateAppSession(session app.AppSessionV1) error {
	if r.updateAppSessionFn != nil {
		return r.updateAppSessionFn(session)
	}
	return nil
}
func (r *recordingDatabaseStore) StoreChannelSessionKeyState(state core.ChannelSessionKeyStateV1) error {
	if r.storeChannelSessionKeyFn != nil {
		return r.storeChannelSessionKeyFn(state)
	}
	return nil
}
func (r *recordingDatabaseStore) StoreAppSessionKeyState(state app.AppSessionKeyStateV1) error {
	if r.storeAppSessionKeyStateFn != nil {
		return r.storeAppSessionKeyStateFn(state)
	}
	return nil
}

// recordingMetricExporter tracks calls to IncUserState and RecordTransaction.
type recordingMetricExporter struct {
	incUserStateCalls       []incUserStateArgs
	recordTransactionCalls  []recordTransactionArgs
	incAppStateUpdateCalls  []string
	incChannelSessionKeys   int
	incAppSessionKeys       int
}

type incUserStateArgs struct {
	asset           string
	homeBlockchainID uint64
	transition      core.TransitionType
}
type recordTransactionArgs struct {
	asset  string
	txType core.TransactionType
	amount decimal.Decimal
}

func (r *recordingMetricExporter) IncUserState(asset string, homeBlockchainID uint64, transition core.TransitionType) {
	r.incUserStateCalls = append(r.incUserStateCalls, incUserStateArgs{asset, homeBlockchainID, transition})
}
func (r *recordingMetricExporter) RecordTransaction(asset string, txType core.TransactionType, amount decimal.Decimal) {
	r.recordTransactionCalls = append(r.recordTransactionCalls, recordTransactionArgs{asset, txType, amount})
}
func (r *recordingMetricExporter) IncAppStateUpdate(applicationID string) {
	r.incAppStateUpdateCalls = append(r.incAppStateUpdateCalls, applicationID)
}
func (r *recordingMetricExporter) IncChannelSessionKeys()                                             { r.incChannelSessionKeys++ }
func (r *recordingMetricExporter) IncAppSessionKeys()                                                 { r.incAppSessionKeys++ }
func (r *recordingMetricExporter) IncChannelStateSigValidation(core.ChannelSignerType, bool)          {}
func (r *recordingMetricExporter) IncRPCMessage(rpc.MsgType, string)                                  {}
func (r *recordingMetricExporter) IncRPCRequest(string, string, bool)                                 {}
func (r *recordingMetricExporter) ObserveRPCDuration(string, string, bool, time.Duration)             {}
func (r *recordingMetricExporter) SetRPCConnections(string, string, uint32)                           {}
func (r *recordingMetricExporter) IncAppSessionUpdateSigValidation(string, app.AppSessionSignerTypeV1, bool) {}
func (r *recordingMetricExporter) IncBlockchainAction(string, uint64, string, bool)                   {}
func (r *recordingMetricExporter) IncBlockchainEvent(uint64, bool)                                    {}

// ── helpers ───────────────────────────────────────────────────────────────────

func newMetricStore(db database.DatabaseStore, m *recordingMetricExporter) *metricStore {
	return &metricStore{DatabaseStore: db, m: m}
}

// ── RecordTransaction tests ───────────────────────────────────────────────────

func TestMetricStore_RecordTransaction_DelegatesAndQueuesCallback(t *testing.T) {
	m := &recordingMetricExporter{}
	var capturedTx core.Transaction
	db := &recordingDatabaseStore{
		recordTransactionFn: func(tx core.Transaction) error {
			capturedTx = tx
			return nil
		},
	}
	ms := newMetricStore(db, m)

	tx := core.Transaction{
		Asset:  "USDC",
		TxType: core.TransactionTypeHomeDeposit,
		Amount: decimal.NewFromInt(500),
	}

	err := ms.RecordTransaction(tx)
	require.NoError(t, err)

	// DB was called with the right transaction.
	assert.Equal(t, tx, capturedTx)

	// Metric must NOT have been recorded yet (callbacks are buffered).
	assert.Empty(t, m.recordTransactionCalls, "metric must not be emitted before flush")

	// After flush the metric is emitted exactly once.
	ms.flush()
	require.Len(t, m.recordTransactionCalls, 1)
	call := m.recordTransactionCalls[0]
	assert.Equal(t, "USDC", call.asset)
	assert.Equal(t, core.TransactionTypeHomeDeposit, call.txType)
	assert.True(t, decimal.NewFromInt(500).Equal(call.amount))
}

func TestMetricStore_RecordTransaction_ErrorPreventsCallback(t *testing.T) {
	m := &recordingMetricExporter{}
	db := &recordingDatabaseStore{
		recordTransactionFn: func(core.Transaction) error {
			return errors.New("db failure")
		},
	}
	ms := newMetricStore(db, m)

	err := ms.RecordTransaction(core.Transaction{Asset: "ETH"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "db failure")

	// flush should do nothing — callback was never queued.
	ms.flush()
	assert.Empty(t, m.recordTransactionCalls)
}

func TestMetricStore_RecordTransaction_MultipleCallsAllFlushed(t *testing.T) {
	m := &recordingMetricExporter{}
	db := &recordingDatabaseStore{}
	ms := newMetricStore(db, m)

	txTypes := []core.TransactionType{
		core.TransactionTypeHomeDeposit,
		core.TransactionTypeTransfer,
		core.TransactionTypeCommit,
	}

	for _, txType := range txTypes {
		err := ms.RecordTransaction(core.Transaction{
			Asset:  "USDC",
			TxType: txType,
			Amount: decimal.NewFromInt(1),
		})
		require.NoError(t, err)
	}

	// Still no metrics emitted before flush.
	assert.Empty(t, m.recordTransactionCalls)

	ms.flush()
	assert.Len(t, m.recordTransactionCalls, len(txTypes))
}

// ── StoreUserState tests ──────────────────────────────────────────────────────

func TestMetricStore_StoreUserState_DelegatesAndQueuesCallback(t *testing.T) {
	m := &recordingMetricExporter{}
	var capturedState core.State
	db := &recordingDatabaseStore{
		storeUserStateFn: func(state core.State) error {
			capturedState = state
			return nil
		},
	}
	ms := newMetricStore(db, m)

	state := core.State{
		Asset: "ETH",
		Transition: core.Transition{
			Type: core.TransitionTypeHomeDeposit,
		},
		HomeLedger: core.Ledger{
			BlockchainID: 42,
		},
	}

	err := ms.StoreUserState(state)
	require.NoError(t, err)

	// DB was called with the right state.
	assert.Equal(t, state.Asset, capturedState.Asset)

	// Metric not yet emitted.
	assert.Empty(t, m.incUserStateCalls, "metric must not be emitted before flush")

	ms.flush()
	require.Len(t, m.incUserStateCalls, 1)
	call := m.incUserStateCalls[0]
	assert.Equal(t, "ETH", call.asset)
	assert.Equal(t, uint64(42), call.homeBlockchainID)
	assert.Equal(t, core.TransitionTypeHomeDeposit, call.transition)
}

func TestMetricStore_StoreUserState_ErrorPreventsCallback(t *testing.T) {
	m := &recordingMetricExporter{}
	db := &recordingDatabaseStore{
		storeUserStateFn: func(core.State) error {
			return errors.New("write failed")
		},
	}
	ms := newMetricStore(db, m)

	err := ms.StoreUserState(core.State{Asset: "USDC"})
	require.Error(t, err)

	ms.flush()
	assert.Empty(t, m.incUserStateCalls)
}

// ── UpdateAppSession tests ────────────────────────────────────────────────────

func TestMetricStore_UpdateAppSession_DelegatesAndQueuesCallback(t *testing.T) {
	m := &recordingMetricExporter{}
	db := &recordingDatabaseStore{}
	ms := newMetricStore(db, m)

	session := app.AppSessionV1{ApplicationID: "my-app", SessionID: "0xabc"}

	err := ms.UpdateAppSession(session)
	require.NoError(t, err)

	assert.Empty(t, m.incAppStateUpdateCalls)

	ms.flush()
	require.Len(t, m.incAppStateUpdateCalls, 1)
	assert.Equal(t, "my-app", m.incAppStateUpdateCalls[0])
}

func TestMetricStore_UpdateAppSession_ErrorPreventsCallback(t *testing.T) {
	m := &recordingMetricExporter{}
	db := &recordingDatabaseStore{
		updateAppSessionFn: func(app.AppSessionV1) error {
			return errors.New("update failed")
		},
	}
	ms := newMetricStore(db, m)

	err := ms.UpdateAppSession(app.AppSessionV1{ApplicationID: "x"})
	require.Error(t, err)

	ms.flush()
	assert.Empty(t, m.incAppStateUpdateCalls)
}

// ── StoreChannelSessionKeyState tests ────────────────────────────────────────

func TestMetricStore_StoreChannelSessionKeyState_DelegatesAndQueuesCallback(t *testing.T) {
	m := &recordingMetricExporter{}
	db := &recordingDatabaseStore{}
	ms := newMetricStore(db, m)

	err := ms.StoreChannelSessionKeyState(core.ChannelSessionKeyStateV1{})
	require.NoError(t, err)

	assert.Equal(t, 0, m.incChannelSessionKeys)

	ms.flush()
	assert.Equal(t, 1, m.incChannelSessionKeys)
}

func TestMetricStore_StoreChannelSessionKeyState_ErrorPreventsCallback(t *testing.T) {
	m := &recordingMetricExporter{}
	db := &recordingDatabaseStore{
		storeChannelSessionKeyFn: func(core.ChannelSessionKeyStateV1) error {
			return errors.New("key store failed")
		},
	}
	ms := newMetricStore(db, m)

	err := ms.StoreChannelSessionKeyState(core.ChannelSessionKeyStateV1{})
	require.Error(t, err)

	ms.flush()
	assert.Equal(t, 0, m.incChannelSessionKeys)
}

// ── StoreAppSessionKeyState tests ─────────────────────────────────────────────

func TestMetricStore_StoreAppSessionKeyState_DelegatesAndQueuesCallback(t *testing.T) {
	m := &recordingMetricExporter{}
	db := &recordingDatabaseStore{}
	ms := newMetricStore(db, m)

	err := ms.StoreAppSessionKeyState(app.AppSessionKeyStateV1{})
	require.NoError(t, err)

	assert.Equal(t, 0, m.incAppSessionKeys)

	ms.flush()
	assert.Equal(t, 1, m.incAppSessionKeys)
}

func TestMetricStore_StoreAppSessionKeyState_ErrorPreventsCallback(t *testing.T) {
	m := &recordingMetricExporter{}
	db := &recordingDatabaseStore{
		storeAppSessionKeyStateFn: func(app.AppSessionKeyStateV1) error {
			return errors.New("key state failed")
		},
	}
	ms := newMetricStore(db, m)

	err := ms.StoreAppSessionKeyState(app.AppSessionKeyStateV1{})
	require.Error(t, err)

	ms.flush()
	assert.Equal(t, 0, m.incAppSessionKeys)
}

// ── flush behaviour ───────────────────────────────────────────────────────────

func TestMetricStore_Flush_EmptyCallbacksIsNoop(t *testing.T) {
	m := &recordingMetricExporter{}
	db := &recordingDatabaseStore{}
	ms := newMetricStore(db, m)

	// Should not panic or do anything.
	ms.flush()

	assert.Empty(t, m.recordTransactionCalls)
	assert.Empty(t, m.incUserStateCalls)
	assert.Empty(t, m.incAppStateUpdateCalls)
	assert.Equal(t, 0, m.incChannelSessionKeys)
	assert.Equal(t, 0, m.incAppSessionKeys)
}

// TestMetricStore_Flush_MixedCallbacks verifies that a metricStore that has accumulated
// callbacks from multiple different operations fires them all — and in order — on flush.
func TestMetricStore_Flush_MixedCallbacks(t *testing.T) {
	m := &recordingMetricExporter{}
	db := &recordingDatabaseStore{}
	ms := newMetricStore(db, m)

	tx := core.Transaction{Asset: "USDC", TxType: core.TransactionTypeTransfer, Amount: decimal.NewFromInt(10)}
	state := core.State{Asset: "USDC", Transition: core.Transition{Type: core.TransitionTypeTransferSend}}
	session := app.AppSessionV1{ApplicationID: "test-app"}
	keyState := core.ChannelSessionKeyStateV1{}
	appKeyState := app.AppSessionKeyStateV1{}

	require.NoError(t, ms.RecordTransaction(tx))
	require.NoError(t, ms.StoreUserState(state))
	require.NoError(t, ms.UpdateAppSession(session))
	require.NoError(t, ms.StoreChannelSessionKeyState(keyState))
	require.NoError(t, ms.StoreAppSessionKeyState(appKeyState))

	// Nothing emitted yet.
	assert.Empty(t, m.recordTransactionCalls)
	assert.Empty(t, m.incUserStateCalls)
	assert.Empty(t, m.incAppStateUpdateCalls)
	assert.Equal(t, 0, m.incChannelSessionKeys)
	assert.Equal(t, 0, m.incAppSessionKeys)

	ms.flush()

	assert.Len(t, m.recordTransactionCalls, 1)
	assert.Len(t, m.incUserStateCalls, 1)
	assert.Len(t, m.incAppStateUpdateCalls, 1)
	assert.Equal(t, 1, m.incChannelSessionKeys)
	assert.Equal(t, 1, m.incAppSessionKeys)
}

// TestMetricStore_RecordTransaction_CorrectArgsPassedToMetric exercises the full argument
// forwarding for RecordTransaction with non-trivial values, acting as a regression guard.
func TestMetricStore_RecordTransaction_CorrectArgsPassedToMetric(t *testing.T) {
	m := &recordingMetricExporter{}
	db := &recordingDatabaseStore{}
	ms := newMetricStore(db, m)

	amount := decimal.RequireFromString("12345.678900")
	tx := core.Transaction{
		Asset:  "ETH",
		TxType: core.TransactionTypeRebalance,
		Amount: amount,
	}

	require.NoError(t, ms.RecordTransaction(tx))
	ms.flush()

	require.Len(t, m.recordTransactionCalls, 1)
	c := m.recordTransactionCalls[0]
	assert.Equal(t, "ETH", c.asset)
	assert.Equal(t, core.TransactionTypeRebalance, c.txType)
	assert.True(t, amount.Equal(c.amount), "amount mismatch: got %s want %s", c.amount, amount)
}
