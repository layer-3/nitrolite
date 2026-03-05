package action_gateway

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
	"go.yaml.in/yaml/v2"
)

const (
	actionGatewayFileName = "action_gateway.yaml"
	defaultTimeWindow     = 24 * time.Hour
)

type ActionLimitConfig struct {
	LevelStepTokens decimal.Decimal                       `yaml:"level_step_tokens"`
	AppCost         decimal.Decimal                       `yaml:"app_cost"`
	ActionGates     map[core.GatedAction]ActionGateConfig `yaml:"action_gates"`
}

type ActionGateConfig struct {
	FreeActionsAllowance uint64 `yaml:"free_actions_allowance"`
	IncreasePerLevel     uint64 `yaml:"increase_per_level"`
}

type ActionGateway struct {
	cfg ActionLimitConfig
}

func NewActionGateway(cfg ActionLimitConfig) (*ActionGateway, error) {
	if !cfg.LevelStepTokens.IsPositive() {
		return nil, errors.New("LevelStepTokens must be greater than zero")
	}
	if !cfg.AppCost.IsPositive() {
		return nil, errors.New("AppCost must be greater than zero")
	}

	seenIDs := make(map[uint8]core.GatedAction, len(cfg.ActionGates))
	for action := range cfg.ActionGates {
		id := action.ID()
		if id == 0 {
			return nil, fmt.Errorf("unknown action_gates key: %q", action)
		}
		if prev, exists := seenIDs[id]; exists && prev != action {
			return nil, fmt.Errorf("duplicate gated action id %d for %q and %q", id, prev, action)
		}
		seenIDs[id] = action
	}

	return &ActionGateway{
		cfg: cfg,
	}, nil
}

func NewActionGatewayFromYaml(configDirPath string) (*ActionGateway, error) {
	assetsPath := filepath.Join(configDirPath, actionGatewayFileName)
	f, err := os.Open(assetsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cfg ActionLimitConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	return NewActionGateway(cfg)
}

type Store interface {
	// GetAppCount returns the total number of applications owned by a specific wallet.
	GetAppCount(ownerWallet string) (uint64, error)

	// GetTotalUserStaked returns the total staked amount for a user across all blockchains.
	GetTotalUserStaked(wallet string) (decimal.Decimal, error)

	// RecordAction inserts a new action log entry for a user.
	RecordAction(wallet string, gatedAction core.GatedAction) error

	// GetUserActionCount returns the number of actions matching the given wallet and gated action
	// within the specified time window (counting backwards from now).
	GetUserActionCount(wallet string, gatedAction core.GatedAction, window time.Duration) (uint64, error)

	// GetUserActionCounts returns a map of gated actions to their respective counts for a user within the specified time window.
	GetUserActionCounts(userWallet string, window time.Duration) (map[core.GatedAction]uint64, error)
}

func (a *ActionGateway) AllowAction(tx Store, userAddress string, gatedAction core.GatedAction) error {
	if _, ok := a.cfg.ActionGates[gatedAction]; !ok {
		return nil
	}

	stakedTokens, err := tx.GetTotalUserStaked(userAddress)
	if err != nil {
		return fmt.Errorf("failed to get user staked amount: %w", err)
	}

	remainingStaked := stakedTokens
	if stakedTokens.IsPositive() {
		appCount, err := tx.GetAppCount(userAddress)
		if err != nil {
			return fmt.Errorf("failed to get app count: %w", err)
		}

		maintenanceCost := a.cfg.AppCost.Mul(decimal.NewFromUint64(appCount))
		remainingStaked = stakedTokens.Sub(maintenanceCost)
	}

	allowance := a.stakedTo24hActionsAllowance(gatedAction, remainingStaked)

	usedCount, err := tx.GetUserActionCount(userAddress, gatedAction, defaultTimeWindow)
	if err != nil {
		return fmt.Errorf("failed to get user action count: %w", err)
	}

	if usedCount >= allowance {
		return fmt.Errorf("action %s limit reached: used %d of %d allowed in 24h", gatedAction, usedCount, allowance)
	}

	if err := tx.RecordAction(userAddress, gatedAction); err != nil {
		return fmt.Errorf("failed to record action: %w", err)
	}

	return nil
}

func (v *ActionGateway) AllowAppRegistration(tx Store, userAddress string) error {
	stakedTokens, err := tx.GetTotalUserStaked(userAddress)
	if err != nil {
		return fmt.Errorf("failed to get user staked amount: %w", err)
	}
	if stakedTokens.IsZero() {
		return errors.New("cannot register an app with zero staked tokens")
	}

	appCount, err := tx.GetAppCount(userAddress)
	if err != nil {
		return fmt.Errorf("failed to get app count: %w", err)
	}

	maintenanceCost := v.cfg.AppCost.Mul(decimal.NewFromUint64(appCount + 1))
	if stakedTokens.LessThan(maintenanceCost) {
		return fmt.Errorf("not enough staked tokens to register a new app: staked %s, required %s", stakedTokens.String(), maintenanceCost.String())
	}

	return nil
}

// stakedTo24hActionsAllowance returns the number of executions allowed in 24 hours for a specific gated action.
func (a *ActionGateway) stakedTo24hActionsAllowance(gatedAction core.GatedAction, remainingStaked decimal.Decimal) uint64 {
	actionLinitsConfig, ok := a.cfg.ActionGates[gatedAction]
	if !ok {
		return 0
	}
	actionAllowance := uint64(actionLinitsConfig.FreeActionsAllowance)
	if remainingStaked.IsPositive() {
		levels := uint64(remainingStaked.Div(a.cfg.LevelStepTokens).IntPart())
		actionAllowance += actionLinitsConfig.IncreasePerLevel * levels
	}
	return actionAllowance
}

// GetUserAllowances returns user allowance for every gated action.
func (a *ActionGateway) GetUserAllowances(tx Store, userAddress string) ([]core.ActionAllowance, error) {
	stakedTokens, err := tx.GetTotalUserStaked(userAddress)
	if err != nil {
		return nil, fmt.Errorf("failed to get user staked amount: %w", err)
	}

	actionCounts, err := tx.GetUserActionCounts(userAddress, defaultTimeWindow)
	if err != nil {
		return nil, err
	}

	remainingStaked := stakedTokens
	if stakedTokens.IsPositive() {
		appCount, err := tx.GetAppCount(userAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to get app count: %w", err)
		}

		maintenanceCost := a.cfg.AppCost.Mul(decimal.NewFromUint64(appCount))
		remainingStaked = stakedTokens.Sub(maintenanceCost)
	}

	timeWindowStr := defaultTimeWindow.String()
	result := make([]core.ActionAllowance, 0, len(a.cfg.ActionGates))
	for action := range a.cfg.ActionGates {
		result = append(result, core.ActionAllowance{
			GatedAction: action,
			TimeWindow:  timeWindowStr,
			Allowance:   a.stakedTo24hActionsAllowance(action, remainingStaked),
			Used:        actionCounts[action],
		})
	}

	slices.SortFunc(result, func(a, b core.ActionAllowance) int {
		return int(a.GatedAction.ID()) - int(b.GatedAction.ID())
	})

	return result, nil
}
