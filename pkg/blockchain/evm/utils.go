package evm

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
	"github.com/pkg/errors"
	"github.com/shopspring/decimal"
)

// signerTxOpts creates a TransactOpts object for signing transactions
// using the provided signer and blockchain ID.
func signerTxOpts(signer sign.Signer, blockchainID uint64) *bind.TransactOpts {
	bigChainID := big.NewInt(int64(blockchainID))
	signingMethod := types.LatestSignerForChainID(bigChainID)
	signerAddress := common.HexToAddress(signer.PublicKey().Address().String())
	signerFn := func(address common.Address, tx *types.Transaction) (*types.Transaction, error) {
		if address != signerAddress {
			return nil, bind.ErrNotAuthorized
		}

		hash := signingMethod.Hash(tx).Bytes()
		sig, err := signer.Sign(hash)
		if err != nil {
			return nil, err
		}

		if sig[64] >= 27 {
			sig[64] -= 27
		}

		return tx.WithSignature(signingMethod, sig)
	}

	return &bind.TransactOpts{
		From:    signerAddress,
		Signer:  signerFn,
		Context: context.Background(),
	}
}

// backOffDuration returns the exponential backoff delay for the given attempt count.
// Returns 0 when backOffCount is 0 (no delay).
// Returns -1 when backOffCount exceeds maxBackOffCount (caller should abort).
func backOffDuration(backOffCount int) time.Duration {
	if backOffCount > maxBackOffCount {
		return -1
	}
	if backOffCount == 0 {
		return 0
	}
	return time.Duration((1<<backOffCount)-1) * time.Second
}

// ========= Client Helper Functions =========

func hexToBytes32(s string) ([32]byte, error) {
	var arr [32]byte
	b, err := hexutil.Decode(s)
	if err != nil {
		return arr, errors.Wrap(err, "failed to decode hex string")
	}
	if len(b) != 32 {
		return arr, errors.Errorf("invalid length: expected 32 bytes, got %d", len(b))
	}
	copy(arr[:], b)
	return arr, nil
}

func coreDefToContractDef(def core.ChannelDefinition, asset, userWallet string, nodeAddress common.Address) (ChannelDefinition, error) {
	approvedSigValidators := new(big.Int)
	if def.ApprovedSigValidators != "" {
		var ok bool
		approvedSigValidators, ok = approvedSigValidators.SetString(def.ApprovedSigValidators, 0)
		if !ok {
			return ChannelDefinition{}, errors.Errorf("failed to parse approved sig validators: %s", def.ApprovedSigValidators)
		}
	}

	return ChannelDefinition{
		ChallengeDuration:           def.Challenge,
		User:                        common.HexToAddress(userWallet),
		Node:                        nodeAddress,
		Nonce:                       def.Nonce,
		ApprovedSignatureValidators: approvedSigValidators,
		Metadata:                    core.GenerateChannelMetadata(asset),
	}, nil
}

func coreStateToContractState(state core.State, tokenGetter func(blockchainID uint64, tokenAddress string) (uint8, error)) (State, error) {
	homeDecimals, err := tokenGetter(state.HomeLedger.BlockchainID, state.HomeLedger.TokenAddress)
	if err != nil {
		return State{}, errors.Wrap(err, "failed to get home token decimals")
	}

	homeLedger, err := coreLedgerToContractLedger(state.HomeLedger, homeDecimals)
	if err != nil {
		return State{}, errors.Wrap(err, "failed to convert home ledger")
	}

	var nonHomeLedger Ledger
	if state.EscrowLedger != nil {
		escrowDecimals, err := tokenGetter(state.EscrowLedger.BlockchainID, state.EscrowLedger.TokenAddress)
		if err != nil {
			return State{}, errors.Wrap(err, "failed to get escrow token decimals")
		}

		nonHomeLedger, err = coreLedgerToContractLedger(*state.EscrowLedger, escrowDecimals)
		if err != nil {
			return State{}, errors.Wrap(err, "failed to convert escrow ledger")
		}
	} else {
		// Initialize empty ledger with non-nil big.Int pointers to prevent ABI encoding panic
		nonHomeLedger = Ledger{
			ChainId:        0,
			Token:          common.Address{},
			Decimals:       0,
			UserAllocation: big.NewInt(0),
			UserNetFlow:    big.NewInt(0),
			NodeAllocation: big.NewInt(0),
			NodeNetFlow:    big.NewInt(0),
		}
	}

	var userSig, nodeSig []byte
	if state.UserSig != nil {
		userSig, err = hexutil.Decode(*state.UserSig)
		if err != nil {
			return State{}, errors.Wrap(err, "failed to decode user signature")
		}
	}
	if state.NodeSig != nil {
		nodeSig, err = hexutil.Decode(*state.NodeSig)
		if err != nil {
			return State{}, errors.Wrap(err, "failed to decode node signature")
		}
	}

	intent := core.TransitionToIntent(state.Transition)
	metadata, err := core.GetStateTransitionHash(state.Transition)
	if err != nil {
		return State{}, errors.Wrap(err, "failed to compute state transitions hash")
	}

	return State{
		Version:       state.Version,
		Intent:        intent,
		Metadata:      metadata,
		HomeLedger:    homeLedger,
		NonHomeLedger: nonHomeLedger,
		UserSig:       userSig,
		NodeSig:       nodeSig,
	}, nil
}

func coreLedgerToContractLedger(ledger core.Ledger, decimals uint8) (Ledger, error) {
	tokenAddr := common.HexToAddress(ledger.TokenAddress)

	userAllocation, err := core.DecimalToBigInt(ledger.UserBalance, decimals)
	if err != nil {
		return Ledger{}, errors.Wrap(err, "failed to convert user balance to big.Int")
	}

	userNetFlow, err := core.DecimalToBigInt(ledger.UserNetFlow, decimals)
	if err != nil {
		return Ledger{}, errors.Wrap(err, "failed to convert user net flow to big.Int")
	}

	nodeAllocation, err := core.DecimalToBigInt(ledger.NodeBalance, decimals)
	if err != nil {
		return Ledger{}, errors.Wrap(err, "failed to convert node balance to big.Int")
	}

	nodeNetFlow, err := core.DecimalToBigInt(ledger.NodeNetFlow, decimals)
	if err != nil {
		return Ledger{}, errors.Wrap(err, "failed to convert node net flow to big.Int")
	}

	return Ledger{
		ChainId:        ledger.BlockchainID,
		Token:          tokenAddr,
		Decimals:       decimals,
		UserAllocation: userAllocation,
		UserNetFlow:    userNetFlow,
		NodeAllocation: nodeAllocation,
		NodeNetFlow:    nodeNetFlow,
	}, nil
}

func contractStateToCoreState(contractState State, homeChannelID string, escrowChannelID *string) (*core.State, error) {
	homeLedger := contractLedgerToCoreLedger(contractState.HomeLedger)

	var escrowLedger *core.Ledger
	if contractState.NonHomeLedger.ChainId != 0 {
		el := contractLedgerToCoreLedger(contractState.NonHomeLedger)
		escrowLedger = &el
	}

	var homeChannelIDPtr *string
	if homeChannelID != "" {
		homeChannelIDPtr = &homeChannelID
	}

	var userSig, nodeSig *string
	if len(contractState.UserSig) > 0 {
		sig := hexutil.Encode(contractState.UserSig)
		userSig = &sig
	}
	if len(contractState.NodeSig) > 0 {
		sig := hexutil.Encode(contractState.NodeSig)
		nodeSig = &sig
	}

	return &core.State{
		Version:         contractState.Version,
		HomeChannelID:   homeChannelIDPtr,
		EscrowChannelID: escrowChannelID,
		HomeLedger:      homeLedger,
		EscrowLedger:    escrowLedger,
		UserSig:         userSig,
		NodeSig:         nodeSig,
		// Note: ID, Transitions, Asset, UserWallet, Epoch are not available from contract state
		// These may need to be populated separately or passed as parameters
	}, nil
}

func contractLedgerToCoreLedger(ledger Ledger) core.Ledger {
	// NOTE: consider YN decimals when using
	exp := -int32(ledger.Decimals)
	return core.Ledger{
		BlockchainID: ledger.ChainId,
		TokenAddress: ledger.Token.Hex(),
		UserBalance:  decimal.NewFromBigInt(ledger.UserAllocation, exp),
		UserNetFlow:  decimal.NewFromBigInt(ledger.UserNetFlow, exp),
		NodeBalance:  decimal.NewFromBigInt(ledger.NodeAllocation, exp),
		NodeNetFlow:  decimal.NewFromBigInt(ledger.NodeNetFlow, exp),
	}
}
