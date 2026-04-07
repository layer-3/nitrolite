package core

import (
	"fmt"
)

var _ StateAdvancer = &StateAdvancerV1{}

// StateAdvancerV1 provides basic validation for state transitions
type StateAdvancerV1 struct {
	assetStore AssetStore
}

// NewStateAdvancerV1 creates a new simple transition validator
func NewStateAdvancerV1(assetStore AssetStore) *StateAdvancerV1 {
	return &StateAdvancerV1{
		assetStore: assetStore,
	}
}

// ValidateAdvancement validates that the proposed state is a valid advancement of the current state
//
// NOTE: User signature is not validated here
//
// TODO: Add shared JSON fixture suite consumed by both Go and TS test suites to guarantee validation parity
func (v *StateAdvancerV1) ValidateAdvancement(currentState, proposedState State) error {
	expectedState := currentState.NextState()
	if proposedState.HomeChannelID == nil {
		return fmt.Errorf("home channel ID cannot be nil")
	}

	// A non-nil state with a nil HomeChannelID is valid — it represents a user who received
	// funds but has not yet opened a channel. In that case, adopt the proposed state's channel info.
	if expectedState.HomeChannelID == nil {
		expectedState.HomeChannelID = proposedState.HomeChannelID
		expectedState.HomeLedger.BlockchainID = proposedState.HomeLedger.BlockchainID
		expectedState.HomeLedger.TokenAddress = proposedState.HomeLedger.TokenAddress
	}

	if *proposedState.HomeChannelID != *expectedState.HomeChannelID {
		return fmt.Errorf("home channel ID mismatch: expected=%s, proposed=%s", *expectedState.HomeChannelID, *proposedState.HomeChannelID)
	}

	// Version must increment
	if proposedState.Version != expectedState.Version {
		return fmt.Errorf("version mismatch: expected=%d, proposed=%d", expectedState.Version, proposedState.Version)
	}

	// User wallet must match
	if proposedState.UserWallet != expectedState.UserWallet {
		return fmt.Errorf("user wallet mismatch: expected=%s, proposed=%s", expectedState.UserWallet, proposedState.UserWallet)
	}

	// Asset must match
	if proposedState.Asset != expectedState.Asset {
		return fmt.Errorf("asset mismatch: expected=%s, proposed=%s", expectedState.Asset, proposedState.Asset)
	}

	// Epoch must match
	if proposedState.Epoch != expectedState.Epoch {
		return fmt.Errorf("epoch mismatch: expected=%d, proposed=%d", expectedState.Epoch, proposedState.Epoch)
	}

	// State ID must match
	if proposedState.ID != expectedState.ID {
		return fmt.Errorf("state ID mismatch: expected=%s, proposed=%s", expectedState.ID, proposedState.ID)
	}

	newTransition := proposedState.Transition

	decimals, err := v.assetStore.GetAssetDecimals(proposedState.Asset)
	if err != nil {
		return fmt.Errorf("failed to get asset decimals: %w", err)
	}

	if err := ValidateDecimalPrecision(newTransition.Amount, decimals); err != nil {
		return fmt.Errorf("invalid amount for asset %s: %w", proposedState.Asset, err)
	}

	switch newTransition.Type {
	case TransitionTypeAcknowledgement:
		if !newTransition.Amount.IsZero() {
			return fmt.Errorf("transition amount must be zero, got %s", newTransition.Amount.String())
		}
	case TransitionTypeFinalize:
		if newTransition.Amount.IsNegative() {
			return fmt.Errorf("transition amount must not be negative, got %s", newTransition.Amount.String())
		}
	default:
		if !newTransition.Amount.IsPositive() {
			return fmt.Errorf("transition amount must be positive, got %s", newTransition.Amount.String())
		}
	}

	lastTransition := currentState.Transition

	switch newTransition.Type {
	case TransitionTypeVoid:
		return fmt.Errorf("cannot apply void transition as new transition")
	case TransitionTypeAcknowledgement:
		if currentState.UserSig != nil {
			return fmt.Errorf("current state is already acknowledged")
		}
		_, err = expectedState.ApplyAcknowledgementTransition()
	case TransitionTypeHomeDeposit:
		_, err = expectedState.ApplyHomeDepositTransition(newTransition.Amount)
	case TransitionTypeHomeWithdrawal:
		_, err = expectedState.ApplyHomeWithdrawalTransition(newTransition.Amount)
	case TransitionTypeTransferSend:
		_, err = expectedState.ApplyTransferSendTransition(newTransition.AccountID, newTransition.Amount)
	case TransitionTypeCommit:
		_, err = expectedState.ApplyCommitTransition(newTransition.AccountID, newTransition.Amount)
	case TransitionTypeMutualLock:
		if proposedState.EscrowLedger == nil {
			return fmt.Errorf("proposed state escrow ledger is nil")
		}
		_, err = expectedState.ApplyMutualLockTransition(
			proposedState.EscrowLedger.BlockchainID,
			proposedState.EscrowLedger.TokenAddress,
			newTransition.Amount)
	case TransitionTypeEscrowDeposit:
		if lastTransition.Type == TransitionTypeMutualLock {
			if !lastTransition.Amount.Equal(newTransition.Amount) {
				return fmt.Errorf("escrow deposit amount must be the same as mutual lock amount")
			}
			_, err = expectedState.ApplyEscrowDepositTransition(newTransition.Amount)
		} else {
			return fmt.Errorf("escrow deposit transition must follow a mutual lock transition")
		}
	case TransitionTypeEscrowLock:
		if proposedState.EscrowLedger == nil {
			return fmt.Errorf("proposed state escrow ledger is nil")
		}
		_, err = expectedState.ApplyEscrowLockTransition(
			proposedState.EscrowLedger.BlockchainID,
			proposedState.EscrowLedger.TokenAddress,
			newTransition.Amount)
	case TransitionTypeEscrowWithdraw:
		if lastTransition.Type == TransitionTypeEscrowLock {
			if !lastTransition.Amount.Equal(newTransition.Amount) {
				return fmt.Errorf("escrow withdraw amount must be the same as escrow lock amount")
			}
			_, err = expectedState.ApplyEscrowWithdrawTransition(newTransition.Amount)
		} else {
			return fmt.Errorf("escrow withdraw transition must follow an escrow lock transition")
		}
	case TransitionTypeMigrate:
		_, err = expectedState.ApplyMigrateTransition(newTransition.Amount)
	case TransitionTypeFinalize:
		_, err = expectedState.ApplyFinalizeTransition()

	default:
		return fmt.Errorf("unsupported type for new transition: %d", newTransition.Type)
	}
	if err != nil {
		return fmt.Errorf("failed to apply new transition: %w", err)
	}

	expectedTransition := expectedState.Transition
	if err := expectedTransition.Equal(newTransition); err != nil {
		return fmt.Errorf("new transition does not match expected: %w", err)
	}

	if err := proposedState.HomeLedger.Equal(expectedState.HomeLedger); err != nil {
		return fmt.Errorf("home ledger mismatch: %w", err)
	}
	if err := proposedState.HomeLedger.Validate(); err != nil {
		return fmt.Errorf("invalid home ledger: %w", err)
	}

	if (expectedState.EscrowChannelID == nil) != (proposedState.EscrowChannelID == nil) {
		return fmt.Errorf("escrow channel ID presence mismatch")
	}

	if expectedState.EscrowChannelID != nil && proposedState.EscrowChannelID != nil {
		if *expectedState.EscrowChannelID != *proposedState.EscrowChannelID {
			return fmt.Errorf("escrow channel ID mismatch: expected=%s, proposed=%s", *expectedState.EscrowChannelID, *proposedState.EscrowChannelID)
		}
	}

	if (expectedState.EscrowLedger == nil) != (proposedState.EscrowLedger == nil) {
		return fmt.Errorf("escrow ledger presence mismatch")
	}

	if expectedState.EscrowLedger != nil && proposedState.EscrowLedger != nil {
		if err := proposedState.EscrowLedger.Equal(*expectedState.EscrowLedger); err != nil {
			return fmt.Errorf("escrow ledger mismatch: %w", err)
		}
		if err := proposedState.EscrowLedger.Validate(); err != nil {
			return fmt.Errorf("invalid escrow ledger: %w", err)
		}

		if proposedState.EscrowLedger.BlockchainID == proposedState.HomeLedger.BlockchainID {
			return fmt.Errorf("escrow ledger blockchain ID cannot match home ledger blockchain ID")
		}
	}

	return nil
}
