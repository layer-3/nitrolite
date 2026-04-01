package action_gateway

import (
	"time"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/shopspring/decimal"
)

// ActionAllower defines the interface for action gating and allowance checks.
type ActionAllower interface {
	// AllowAction checks if a user is allowed to perform a specific gated action
	// based on their past activity and allowances.
	AllowAction(tx Store, userAddress string, gatedAction core.GatedAction) error

	// AllowAppRegistration checks if a user is allowed to register a new application
	// based on their staked tokens and existing app count.
	AllowAppRegistration(tx Store, userAddress string) error

	// GetUserAllowances returns user allowance for every gated action.
	// An empty slice indicates the user has no limits.
	GetUserAllowances(tx Store, userAddress string) ([]core.ActionAllowance, error)
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
