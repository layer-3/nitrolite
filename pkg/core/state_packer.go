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

// contractLedger mirrors the Solidity Ledger struct used inside the signed state
// payload. Allocation fields are encoded as uint256 and must fit [0, 2^256-1];
// net-flow fields are encoded as int256 and must fit [-2^255, 2^255-1].
type contractLedger struct {
	ChainId        uint64
	Token          common.Address
	Decimals       uint8
	UserAllocation *big.Int
	UserNetFlow    *big.Int
	NodeAllocation *big.Int
	NodeNetFlow    *big.Int
}

// Validate guards against any caller that constructs a contractLedger without
// going through the bounds-checking decimal helpers, so values that would be
// silently truncated by ABI encoding are rejected before signing.
func (l contractLedger) Validate() error {
	if err := checkUint256(l.UserAllocation); err != nil {
		return fmt.Errorf("user allocation: %w", err)
	}
	if err := checkUint256(l.NodeAllocation); err != nil {
		return fmt.Errorf("node allocation: %w", err)
	}
	if err := checkInt256(l.UserNetFlow); err != nil {
		return fmt.Errorf("user net flow: %w", err)
	}
	if err := checkInt256(l.NodeNetFlow); err != nil {
		return fmt.Errorf("node net flow: %w", err)
	}
	return nil
}

func checkUint256(v *big.Int) error {
	if v == nil {
		return fmt.Errorf("value is nil")
	}
	if v.Sign() < 0 {
		return fmt.Errorf("value %s is negative", v.String())
	}
	if v.BitLen() > 256 {
		return fmt.Errorf("value %s exceeds uint256 max (2^256-1)", v.String())
	}
	return nil
}

func checkInt256(v *big.Int) error {
	if v == nil {
		return fmt.Errorf("value is nil")
	}
	if v.Cmp(maxInt256) > 0 {
		return fmt.Errorf("value %s exceeds int256 max (2^255-1)", v.String())
	}
	if v.Cmp(minInt256) < 0 {
		return fmt.Errorf("value %s below int256 min (-2^255)", v.String())
	}
	return nil
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

	homeDecimals, err := p.assetStore.GetTokenDecimals(state.HomeLedger.BlockchainID, state.HomeLedger.TokenAddress)
	if err != nil {
		return common.Hash{}, nil, err
	}

	userBalanceBI, err := DecimalToUint256(state.HomeLedger.UserBalance, homeDecimals)
	if err != nil {
		return common.Hash{}, nil, fmt.Errorf("home user balance: %w", err)
	}
	userNetFlowBI, err := DecimalToInt256(state.HomeLedger.UserNetFlow, homeDecimals)
	if err != nil {
		return common.Hash{}, nil, fmt.Errorf("home user net flow: %w", err)
	}
	nodeBalanceBI, err := DecimalToUint256(state.HomeLedger.NodeBalance, homeDecimals)
	if err != nil {
		return common.Hash{}, nil, fmt.Errorf("home node balance: %w", err)
	}
	nodeNetFlowBI, err := DecimalToInt256(state.HomeLedger.NodeNetFlow, homeDecimals)
	if err != nil {
		return common.Hash{}, nil, fmt.Errorf("home node net flow: %w", err)
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

		escrowUserBalanceBI, err := DecimalToUint256(state.EscrowLedger.UserBalance, escrowDecimals)
		if err != nil {
			return common.Hash{}, nil, fmt.Errorf("escrow user balance: %w", err)
		}
		escrowUserNetFlowBI, err := DecimalToInt256(state.EscrowLedger.UserNetFlow, escrowDecimals)
		if err != nil {
			return common.Hash{}, nil, fmt.Errorf("escrow user net flow: %w", err)
		}
		escrowNodeBalanceBI, err := DecimalToUint256(state.EscrowLedger.NodeBalance, escrowDecimals)
		if err != nil {
			return common.Hash{}, nil, fmt.Errorf("escrow node balance: %w", err)
		}
		escrowNodeNetFlowBI, err := DecimalToInt256(state.EscrowLedger.NodeNetFlow, escrowDecimals)
		if err != nil {
			return common.Hash{}, nil, fmt.Errorf("escrow node net flow: %w", err)
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

	if err := homeLedger.Validate(); err != nil {
		return common.Hash{}, nil, fmt.Errorf("invalid home ledger for signing: %w", err)
	}
	if err := nonHomeLedger.Validate(); err != nil {
		return common.Hash{}, nil, fmt.Errorf("invalid escrow ledger for signing: %w", err)
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
