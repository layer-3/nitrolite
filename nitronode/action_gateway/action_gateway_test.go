package action_gateway

import (
	"errors"
	"testing"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockAVStore implements AVStore for unit tests.
type mockAVStore struct {
	totalStaked     decimal.Decimal
	stakedErr       error
	appCount        uint64
	appCountErr     error
	actionCount     uint64
	actionErr       error
	actionCounts    map[core.GatedAction]uint64
	actionCSErr     error
	recordedActions []core.GatedAction
	recordErr       error
}

func (m *mockAVStore) GetTotalUserStaked(string) (decimal.Decimal, error) {
	return m.totalStaked, m.stakedErr
}

func (m *mockAVStore) GetAppCount(string) (uint64, error) {
	return m.appCount, m.appCountErr
}

func (m *mockAVStore) GetUserActionCount(string, core.GatedAction, time.Duration) (uint64, error) {
	return m.actionCount, m.actionErr
}

func (m *mockAVStore) GetUserActionCounts(string, time.Duration) (map[core.GatedAction]uint64, error) {
	return m.actionCounts, m.actionCSErr
}

func (m *mockAVStore) RecordAction(_ string, action core.GatedAction) error {
	m.recordedActions = append(m.recordedActions, action)
	return m.recordErr
}

func defaultConfig() ActionLimitConfig {
	return ActionLimitConfig{
		LevelStepTokens: decimal.NewFromInt(100),
		AppCost:         decimal.NewFromInt(50),
		ActionGates: map[core.GatedAction]ActionGateConfig{
			core.GatedActionTransfer: {FreeActionsAllowance: 5, IncreasePerLevel: 10},
		},
	}
}

func mustNewGateway(t *testing.T, cfg ActionLimitConfig) *ActionGateway {
	t.Helper()
	gw, err := NewActionGateway(cfg)
	require.NoError(t, err)
	return gw
}

// --- NewActionGateway ---

func TestNewActionGateway(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		gw, err := NewActionGateway(defaultConfig())
		require.NoError(t, err)
		assert.NotNil(t, gw)
	})

	t.Run("zero LevelStepTokens", func(t *testing.T) {
		cfg := defaultConfig()
		cfg.LevelStepTokens = decimal.Zero
		_, err := NewActionGateway(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "LevelStepTokens")
	})

	t.Run("zero AppCost", func(t *testing.T) {
		cfg := defaultConfig()
		cfg.AppCost = decimal.Zero
		_, err := NewActionGateway(cfg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "AppCost")
	})
}

// --- stakedTo24hActionsAllowance ---

func TestStakedTo24hActionsAllowance(t *testing.T) {
	gw := mustNewGateway(t, defaultConfig())

	t.Run("unknown action returns 0", func(t *testing.T) {
		assert.Equal(t, uint64(0), gw.stakedTo24hActionsAllowance("unknown_action", decimal.NewFromInt(1000)))
	})

	t.Run("zero staked returns free allowance only", func(t *testing.T) {
		assert.Equal(t, uint64(5), gw.stakedTo24hActionsAllowance(core.GatedActionTransfer, decimal.Zero))
	})

	t.Run("negative staked returns free allowance only", func(t *testing.T) {
		assert.Equal(t, uint64(5), gw.stakedTo24hActionsAllowance(core.GatedActionTransfer, decimal.NewFromInt(-100)))
	})

	t.Run("positive staked adds levels", func(t *testing.T) {
		// 250 tokens / 100 step = 2 levels -> 5 + 2*10 = 25
		assert.Equal(t, uint64(25), gw.stakedTo24hActionsAllowance(core.GatedActionTransfer, decimal.NewFromInt(250)))
	})

	t.Run("partial level truncated", func(t *testing.T) {
		// 199 tokens / 100 step = 1 level -> 5 + 1*10 = 15
		assert.Equal(t, uint64(15), gw.stakedTo24hActionsAllowance(core.GatedActionTransfer, decimal.NewFromInt(199)))
	})
}

// --- AllowAction ---

func TestAllowAction(t *testing.T) {
	t.Run("allowed with free allowance, zero staked", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		store := &mockAVStore{
			totalStaked: decimal.Zero,
			actionCount: 0,
		}
		err := gw.AllowAction(store, "0xuser", core.GatedActionTransfer)
		require.NoError(t, err)
		assert.Equal(t, []core.GatedAction{core.GatedActionTransfer}, store.recordedActions)
	})

	t.Run("allowed with staked tokens and apps", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		// 300 staked - 2 apps * 50 cost = 200 remaining -> 2 levels -> 5 + 20 = 25
		store := &mockAVStore{
			totalStaked: decimal.NewFromInt(300),
			appCount:    2,
			actionCount: 24, // under 25
		}
		err := gw.AllowAction(store, "0xuser", core.GatedActionTransfer)
		require.NoError(t, err)
	})

	t.Run("rejected at limit", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		store := &mockAVStore{
			totalStaked: decimal.Zero,
			actionCount: 5, // equals free allowance
		}
		err := gw.AllowAction(store, "0xuser", core.GatedActionTransfer)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "limit reached")
	})

	t.Run("rejected over limit", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		store := &mockAVStore{
			totalStaked: decimal.Zero,
			actionCount: 10,
		}
		err := gw.AllowAction(store, "0xuser", core.GatedActionTransfer)
		assert.Error(t, err)
	})

	t.Run("unknown gated action returns nil without store calls", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		store := &mockAVStore{
			stakedErr: errors.New("should not be called"),
		}
		err := gw.AllowAction(store, "0xuser", "nonexistent")
		require.NoError(t, err)
		assert.Empty(t, store.recordedActions)
	})

	t.Run("staked error propagated", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		store := &mockAVStore{stakedErr: errors.New("db down")}
		err := gw.AllowAction(store, "0xuser", core.GatedActionTransfer)
		assert.ErrorContains(t, err, "db down")
	})

	t.Run("app count error propagated", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		store := &mockAVStore{
			totalStaked: decimal.NewFromInt(100),
			appCountErr: errors.New("db down"),
		}
		err := gw.AllowAction(store, "0xuser", core.GatedActionTransfer)
		assert.ErrorContains(t, err, "db down")
	})

	t.Run("action count error propagated", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		store := &mockAVStore{
			totalStaked: decimal.Zero,
			actionErr:   errors.New("db down"),
		}
		err := gw.AllowAction(store, "0xuser", core.GatedActionTransfer)
		assert.ErrorContains(t, err, "db down")
	})

	t.Run("record error propagated", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		store := &mockAVStore{
			totalStaked: decimal.Zero,
			actionCount: 0,
			recordErr:   errors.New("write fail"),
		}
		err := gw.AllowAction(store, "0xuser", core.GatedActionTransfer)
		assert.ErrorContains(t, err, "write fail")
	})

	t.Run("skips GetAppCount when staked is zero", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		// If GetAppCount were called it would return an error
		store := &mockAVStore{
			totalStaked: decimal.Zero,
			appCountErr: errors.New("should not be called"),
			actionCount: 0,
		}
		err := gw.AllowAction(store, "0xuser", core.GatedActionTransfer)
		require.NoError(t, err)
	})
}

// --- AllowAppRegistration ---

func TestAllowAppRegistration(t *testing.T) {
	t.Run("allowed with enough stake", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		// 0 existing apps, cost for 1 = 50, staked = 100
		store := &mockAVStore{totalStaked: decimal.NewFromInt(100), appCount: 0}
		err := gw.AllowAppRegistration(store, "0xuser")
		require.NoError(t, err)
	})

	t.Run("allowed exact cost", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		// 1 existing app, cost for 2 = 100, staked = 100
		store := &mockAVStore{totalStaked: decimal.NewFromInt(100), appCount: 1}
		err := gw.AllowAppRegistration(store, "0xuser")
		require.NoError(t, err)
	})

	t.Run("rejected not enough stake", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		// 1 existing app, cost for 2 = 100, staked = 99
		store := &mockAVStore{totalStaked: decimal.NewFromInt(99), appCount: 1}
		err := gw.AllowAppRegistration(store, "0xuser")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not enough staked tokens")
	})

	t.Run("rejected zero staked", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		store := &mockAVStore{totalStaked: decimal.Zero}
		err := gw.AllowAppRegistration(store, "0xuser")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "zero staked tokens")
	})
}

// --- GetUserAllowances ---

func TestGetUserAllowances(t *testing.T) {
	multiActionConfig := ActionLimitConfig{
		LevelStepTokens: decimal.NewFromInt(100),
		AppCost:         decimal.NewFromInt(50),
		ActionGates: map[core.GatedAction]ActionGateConfig{
			core.GatedActionTransfer:            {FreeActionsAllowance: 5, IncreasePerLevel: 10},
			core.GatedActionAppSessionOperation: {FreeActionsAllowance: 20, IncreasePerLevel: 5},
		},
	}

	t.Run("zero staked returns free allowances", func(t *testing.T) {
		gw := mustNewGateway(t, multiActionConfig)
		store := &mockAVStore{
			totalStaked:  decimal.Zero,
			actionCounts: map[core.GatedAction]uint64{core.GatedActionTransfer: 3},
		}
		result, err := gw.GetUserAllowances(store, "0xuser")
		require.NoError(t, err)
		require.Len(t, result, 2)

		// Results should be sorted by GatedAction ID
		assert.Equal(t, core.GatedActionTransfer, result[0].GatedAction)
		assert.Equal(t, uint64(5), result[0].Allowance)
		assert.Equal(t, uint64(3), result[0].Used)

		assert.Equal(t, core.GatedActionAppSessionOperation, result[1].GatedAction)
		assert.Equal(t, uint64(20), result[1].Allowance)
		assert.Equal(t, uint64(0), result[1].Used)
	})

	t.Run("staked tokens increase allowances", func(t *testing.T) {
		gw := mustNewGateway(t, multiActionConfig)
		// 500 staked - 2 apps * 50 = 400 remaining -> 4 levels
		store := &mockAVStore{
			totalStaked:  decimal.NewFromInt(500),
			appCount:     2,
			actionCounts: map[core.GatedAction]uint64{},
		}
		result, err := gw.GetUserAllowances(store, "0xuser")
		require.NoError(t, err)

		// Transfer: 5 + 4*10 = 45
		assert.Equal(t, uint64(45), result[0].Allowance)
		// AppSessionOperation: 20 + 4*5 = 40
		assert.Equal(t, uint64(40), result[1].Allowance)
	})

	t.Run("maintenance exceeds staked gives free allowance", func(t *testing.T) {
		gw := mustNewGateway(t, multiActionConfig)
		// 50 staked - 2 apps * 50 = -50 remaining
		store := &mockAVStore{
			totalStaked:  decimal.NewFromInt(50),
			appCount:     2,
			actionCounts: map[core.GatedAction]uint64{},
		}
		result, err := gw.GetUserAllowances(store, "0xuser")
		require.NoError(t, err)

		assert.Equal(t, uint64(5), result[0].Allowance)
		assert.Equal(t, uint64(20), result[1].Allowance)
	})

	t.Run("time window is set", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		store := &mockAVStore{
			totalStaked:  decimal.Zero,
			actionCounts: map[core.GatedAction]uint64{},
		}
		result, err := gw.GetUserAllowances(store, "0xuser")
		require.NoError(t, err)
		assert.Equal(t, "24h0m0s", result[0].TimeWindow)
	})

	t.Run("staked error propagated", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		store := &mockAVStore{stakedErr: errors.New("db down")}
		_, err := gw.GetUserAllowances(store, "0xuser")
		assert.ErrorContains(t, err, "db down")
	})

	t.Run("action counts error propagated", func(t *testing.T) {
		gw := mustNewGateway(t, defaultConfig())
		store := &mockAVStore{
			totalStaked: decimal.Zero,
			actionCSErr: errors.New("db down"),
		}
		_, err := gw.GetUserAllowances(store, "0xuser")
		assert.ErrorContains(t, err, "db down")
	})
}
