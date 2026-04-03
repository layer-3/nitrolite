package core

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

type StatePackerV1 struct {
	assetStore AssetStore
}

func NewStatePackerV1(assetStore AssetStore) *StatePackerV1 {
	return &StatePackerV1{
		assetStore: assetStore,
	}
}

// PackState is a convenience function that creates a StatePackerV1 and packs the state.
// For production use, create a StatePackerV1 instance and reuse it.
func PackState(state State, assetStore AssetStore) ([]byte, error) {
	packer := NewStatePackerV1(assetStore)
	return packer.PackState(state)
}

// packSigningData computes the inner signing data and channelID for a state.
// Returns abi.encode(version, intent, metadata, homeLedger, nonHomeLedger).
func (p *StatePackerV1) packSigningData(state State) (common.Hash, []byte, error) {
	if state.HomeChannelID == nil {
		return common.Hash{}, nil, fmt.Errorf("state.HomeChannelID is required for packing")
	}

	channelID := common.HexToHash(*state.HomeChannelID)

	metadata, err := GetStateTransitionHash(state.Transition)
	if err != nil {
		return common.Hash{}, nil, fmt.Errorf("failed to generate state transitions hash: %w", err)
	}

	ledgerType, err := abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "chainId", Type: "uint64"},
		{Name: "token", Type: "address"},
		{Name: "decimals", Type: "uint8"},
		{Name: "userAllocation", Type: "uint256"},
		{Name: "userNetFlow", Type: "int256"},
		{Name: "nodeAllocation", Type: "uint256"},
		{Name: "nodeNetFlow", Type: "int256"},
	})
	if err != nil {
		return common.Hash{}, nil, err
	}

	bytes32Type := abi.Type{T: abi.FixedBytesTy, Size: 32}

	signingDataArgs := abi.Arguments{
		{Type: uint64Type},  // version
		{Type: uint8Type},   // intent
		{Type: bytes32Type}, // metadata
		{Type: ledgerType},  // homeState
		{Type: ledgerType},  // nonHomeState
	}

	type contractLedger struct {
		ChainId        uint64
		Token          common.Address
		Decimals       uint8
		UserAllocation *big.Int
		UserNetFlow    *big.Int
		NodeAllocation *big.Int
		NodeNetFlow    *big.Int
	}

	homeDecimals, err := p.assetStore.GetTokenDecimals(state.HomeLedger.BlockchainID, state.HomeLedger.TokenAddress)
	if err != nil {
		return common.Hash{}, nil, err
	}

	userBalanceBI, err := DecimalToBigInt(state.HomeLedger.UserBalance, homeDecimals)
	if err != nil {
		return common.Hash{}, nil, err
	}
	userNetFlowBI, err := DecimalToBigInt(state.HomeLedger.UserNetFlow, homeDecimals)
	if err != nil {
		return common.Hash{}, nil, err
	}
	nodeBalanceBI, err := DecimalToBigInt(state.HomeLedger.NodeBalance, homeDecimals)
	if err != nil {
		return common.Hash{}, nil, err
	}
	nodeNetFlowBI, err := DecimalToBigInt(state.HomeLedger.NodeNetFlow, homeDecimals)
	if err != nil {
		return common.Hash{}, nil, err
	}

	homeLedger := contractLedger{
		ChainId:        state.HomeLedger.BlockchainID,
		Token:          common.HexToAddress(state.HomeLedger.TokenAddress),
		Decimals:       homeDecimals,
		UserAllocation: userBalanceBI,
		UserNetFlow:    userNetFlowBI,
		NodeAllocation: nodeBalanceBI,
		NodeNetFlow:    nodeNetFlowBI,
	}

	var nonHomeLedger contractLedger

	if state.EscrowLedger != nil {
		escrowDecimals, err := p.assetStore.GetTokenDecimals(state.EscrowLedger.BlockchainID, state.EscrowLedger.TokenAddress)
		if err != nil {
			return common.Hash{}, nil, err
		}

		escrowUserBalanceBI, err := DecimalToBigInt(state.EscrowLedger.UserBalance, escrowDecimals)
		if err != nil {
			return common.Hash{}, nil, err
		}
		escrowUserNetFlowBI, err := DecimalToBigInt(state.EscrowLedger.UserNetFlow, escrowDecimals)
		if err != nil {
			return common.Hash{}, nil, err
		}
		escrowNodeBalanceBI, err := DecimalToBigInt(state.EscrowLedger.NodeBalance, escrowDecimals)
		if err != nil {
			return common.Hash{}, nil, err
		}
		escrowNodeNetFlowBI, err := DecimalToBigInt(state.EscrowLedger.NodeNetFlow, escrowDecimals)
		if err != nil {
			return common.Hash{}, nil, err
		}

		nonHomeLedger = contractLedger{
			ChainId:        state.EscrowLedger.BlockchainID,
			Token:          common.HexToAddress(state.EscrowLedger.TokenAddress),
			Decimals:       escrowDecimals,
			UserAllocation: escrowUserBalanceBI,
			UserNetFlow:    escrowUserNetFlowBI,
			NodeAllocation: escrowNodeBalanceBI,
			NodeNetFlow:    escrowNodeNetFlowBI,
		}
	} else {
		nonHomeLedger = contractLedger{
			ChainId:        0,
			Token:          common.Address{},
			Decimals:       0,
			UserAllocation: big.NewInt(0),
			UserNetFlow:    big.NewInt(0),
			NodeAllocation: big.NewInt(0),
			NodeNetFlow:    big.NewInt(0),
		}
	}

	intent := TransitionToIntent(state.Transition)

	signingData, err := signingDataArgs.Pack(
		state.Version,
		intent,
		metadata,
		homeLedger,
		nonHomeLedger,
	)
	if err != nil {
		return common.Hash{}, nil, fmt.Errorf("failed to pack signing data: %w", err)
	}

	return channelID, signingData, nil
}

// packWithChannelID wraps signing data with the channel ID: abi.encode(channelId, signingData)
func packWithChannelID(channelID common.Hash, signingData []byte) ([]byte, error) {
	bytes32Type := abi.Type{T: abi.FixedBytesTy, Size: 32}
	bytesType, err := abi.NewType("bytes", "", nil)
	if err != nil {
		return nil, err
	}

	outerArgs := abi.Arguments{
		{Type: bytes32Type},
		{Type: bytesType},
	}

	packed, err := outerArgs.Pack(channelID, signingData)
	if err != nil {
		return nil, fmt.Errorf("failed to pack outer message: %w", err)
	}

	return packed, nil
}

// PackState encodes a channel ID and state into ABI-packed bytes for on-chain submission.
// This matches the Solidity contract's two-step encoding:
//
//	Step 1: signingData = abi.encode(version, intent, metadata, homeLedger, nonHomeLedger)
//	Step 2: message = abi.encode(channelId, signingData)
func (p *StatePackerV1) PackState(state State) ([]byte, error) {
	channelID, signingData, err := p.packSigningData(state)
	if err != nil {
		return nil, err
	}
	return packWithChannelID(channelID, signingData)
}

// PackChallengeState is a convenience function that creates a StatePackerV1 and packs the challenge state.
func PackChallengeState(state State, assetStore AssetStore) ([]byte, error) {
	packer := NewStatePackerV1(assetStore)
	return packer.PackChallengeState(state)
}

// PackChallengeState encodes a state for challenge signature verification.
// This matches the Solidity contract's challenge validation:
//
//	challengerSigningData = abi.encodePacked(abi.encode(version, intent, metadata, homeLedger, nonHomeLedger), "challenge")
//	message = abi.encode(channelId, challengerSigningData)
func (p *StatePackerV1) PackChallengeState(state State) ([]byte, error) {
	channelID, signingData, err := p.packSigningData(state)
	if err != nil {
		return nil, err
	}
	challengeSigningData := append(signingData, []byte("challenge")...)
	return packWithChannelID(channelID, challengeSigningData)
}
