// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package evm

import (
	"errors"
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = errors.New
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// ChannelDefinition is an auto generated low-level Go binding around an user-defined struct.
type ChannelDefinition struct {
	ChallengeDuration           uint32
	User                        common.Address
	Node                        common.Address
	Nonce                       uint64
	ApprovedSignatureValidators *big.Int
	Metadata                    [32]byte
}

// Ledger is an auto generated low-level Go binding around an user-defined struct.
type Ledger struct {
	ChainId        uint64
	Token          common.Address
	Decimals       uint8
	UserAllocation *big.Int
	UserNetFlow    *big.Int
	NodeAllocation *big.Int
	NodeNetFlow    *big.Int
}

// State is an auto generated low-level Go binding around an user-defined struct.
type State struct {
	Version       uint64
	Intent        uint8
	Metadata      [32]byte
	HomeLedger    Ledger
	NonHomeLedger Ledger
	UserSig       []byte
	NodeSig       []byte
}

// ChannelHubMetaData contains all meta data concerning the ChannelHub contract.
var ChannelHubMetaData = &bind.MetaData{
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_defaultSigValidator\",\"type\":\"address\",\"internalType\":\"contractISignatureValidator\"},{\"name\":\"_node\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DEFAULT_SIG_VALIDATOR\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractISignatureValidator\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ESCROW_DEPOSIT_UNLOCK_DELAY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_CHALLENGE_DURATION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_DEPOSIT_ESCROW_STEPS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MIN_CHALLENGE_DURATION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"NODE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"TRANSFER_GAS_LIMIT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VALIDATOR_ACTIVATION_DELAY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"challengeChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"challengerIdx\",\"type\":\"uint8\",\"internalType\":\"enumParticipantIndex\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"challengeEscrowDeposit\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"challengerIdx\",\"type\":\"uint8\",\"internalType\":\"enumParticipantIndex\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"challengeEscrowWithdrawal\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"challengerIdx\",\"type\":\"uint8\",\"internalType\":\"enumParticipantIndex\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"checkpointChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"claimFunds\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"destination\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"closeChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"createChannel\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"depositToChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"depositToNode\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"escrowHead\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"finalizeEscrowDeposit\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"finalizeEscrowWithdrawal\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"finalizeMigration\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getChannelData\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumChannelStatus\"},{\"name\":\"definition\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"lastState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"lockedFunds\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getChannelIds\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowDepositData\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumEscrowStatus\"},{\"name\":\"unlockAt\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"challengeExpiry\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"lockedAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowDepositIds\",\"inputs\":[{\"name\":\"page\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"pageSize\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"ids\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowWithdrawalData\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumEscrowStatus\"},{\"name\":\"challengeExpiry\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"lockedAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNodeBalance\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNodeValidator\",\"inputs\":[{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[{\"name\":\"validator\",\"type\":\"address\",\"internalType\":\"contractISignatureValidator\"},{\"name\":\"registeredAt\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOpenChannels\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"openChannels\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getReclaimBalance\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getUnlockableEscrowDepositStats\",\"inputs\":[],\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"totalAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initiateEscrowDeposit\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"initiateEscrowWithdrawal\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initiateMigration\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"purgeEscrowDeposits\",\"inputs\":[{\"name\":\"maxSteps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registerNodeValidator\",\"inputs\":[{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"validator\",\"type\":\"address\",\"internalType\":\"contractISignatureValidator\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"withdrawFromChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"withdrawFromNode\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"ChannelChallenged\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelCheckpointed\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelClosed\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"finalState\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelCreated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"definition\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"initialState\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelDeposited\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelWithdrawn\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Deposited\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositChallenged\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositFinalized\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositFinalizedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositInitiated\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositInitiatedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositsPurged\",\"inputs\":[{\"name\":\"escrowIds\",\"type\":\"bytes32[]\",\"indexed\":false,\"internalType\":\"bytes32[]\"},{\"name\":\"purgedCount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalChallenged\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalFinalized\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalFinalizedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalInitiated\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalInitiatedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"FundsClaimed\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"destination\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationInFinalized\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationInInitiated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationOutFinalized\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationOutInitiated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"NodeBalanceUpdated\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TransferFailed\",\"inputs\":[{\"name\":\"recipient\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ValidatorRegistered\",\"inputs\":[{\"name\":\"validatorId\",\"type\":\"uint8\",\"indexed\":true,\"internalType\":\"uint8\"},{\"name\":\"validator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"contractISignatureValidator\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Withdrawn\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressCollision\",\"inputs\":[{\"name\":\"collision\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ChallengerVersionTooLow\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureS\",\"inputs\":[{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"EmptySignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectAmount\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectChallengeDuration\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectChannelId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectChannelStatus\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectMsgSender\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectNode\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectStateIntent\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectValue\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientBalance\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidValidatorId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NativeTransferFailed\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"NoChannelIdFoundForEscrow\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeCastOverflowedIntToUint\",\"inputs\":[{\"name\":\"value\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ValidatorAlreadyRegistered\",\"inputs\":[{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}]},{\"type\":\"error\",\"name\":\"ValidatorNotActive\",\"inputs\":[{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"activatesAt\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"type\":\"error\",\"name\":\"ValidatorNotApproved\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ValidatorNotRegistered\",\"inputs\":[{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}]}]",
	Bin: "0x60c03461010b57601f615ee238819003918201601f19168301916001600160401b0383118484101761010f57808492604094855283398101031261010b5780516001600160a01b0381169182820361010b5760200151916001600160a01b0383169081840361010b5760017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055156100fc57156100fc5760805260a052604051615dbe908161012482396080518181816111910152613ed3015260a051818181610c5c01528181610d790152818161145001528181611a3e0152818161207d0152818161361d015281816140800152818161464901526147510152f35b63e6c4247b60e01b5f5260045ffd5b5f80fd5b634e487b7160e01b5f52604160045260245ffdfe60806040526004361015610011575f80fd5b5f3560e01c806307f241ce1461027f57806316b390b11461027a578063187576d8146102755780633115f6301461027057806338a66be21461026b5780633c684f921461026657806341b660ef1461026157806347de477a1461025c57806351bfcdbd1461025757806353269198146102525780635a0745b41461024d5780635ae2accc146102485780635b9acbf9146102435780635dc46a741461023e5780636840dbd2146102395780636898234b1461023457806371a471411461022f578063735181f01461022a57806382d3e15d146102255780638d0b12a5146102205780638e31c7351461021b57806394191051146102115780639691b46814610216578063a459463114610211578063a5c826801461020c578063b25a1d3814610207578063b65b78d114610202578063b9f4420d146101fd578063c74a2d10146101f8578063c9408398146101f3578063d888ccae146101ee578063d91a1283146101e9578063dc23f29e146101e4578063dd73d494146101df578063e617208c146101da578063f4ac51f5146101d5578063f766f8d6146101d0578063ff5bc09e146101cb5763ffa1ad74146101c6575f80fd5b6126ae565b612697565b612578565b6124fd565b61245f565b6122e5565b61212e565b612012565b611f09565b611c7a565b611bfa565b611bdd565b611aee565b611770565b611611565b6114e7565b611504565b611384565b61123d565b611220565b6111da565b611172565b611093565b61107c565b611031565b610ffb565b610fe0565b610fc4565b610dcc565b610d5a565b610b96565b610870565b6107ad565b610772565b61057b565b6104f5565b610351565b610299565b6001600160a01b0381160361029557565b5f80fd5b34610295576020366003190112610295576001600160a01b036004356102be81610284565b165f526006602052602060405f2054604051908152f35b9181601f84011215610295578235916001600160401b038311610295576020838186019501011161029557565b60643590600282101561029557565b9060606003198301126102955760043591602435906001600160401b03821161029557610340916004016102d5565b909160443560028110156102955790565b34610295576103b36103ed61036536610311565b9294916103c8610380879693965f52600260205260405f2090565b9485549261038f8415156126c9565b600187015460059060081c6001600160a01b031696879260028a01549a8b91613eb2565b9192909901986103c28a6128e3565b87613fe3565b60c06103d3876140d5565b604051809481926301999b9360e61b835260048301612a53565b038173__$682d6198b4eca5bc7e038b912a26498e7e$__5af480156104a9577fba075bd445233f7cad862c72f0343b3503aad9c8e704a2295f122b82abf8e80195610461946080945f93610476575b5082610453939461044c896128e3565b908b614149565b01516001600160401b031690565b9061047160405192839283612b8e565b0390a2005b610453935061049c9060c03d60c0116104a2575b610494818361275f565b810190612991565b9261043c565b503d61048a565b612a64565b90602080835192838152019201905f5b8181106104cb5750505090565b82518452602093840193909201916001016104be565b9060206104f29281815201906104ae565b90565b34610295576020366003190112610295576001600160a01b0360043561051a81610284565b165f52600160205260405f206040519081602082549182815201915f5260205f20905f5b81811061056557610561856105558187038261275f565b604051918291826104e1565b0390f35b825484526020909301926001928301920161053e565b3461029557602036600319011261029557600354600480545f92918390358284111561076c576105ab838561332c565b8082101561075e57506105c28195949392956132ed565b925b80831080610755575b15610748576105e86105de84613145565b90549060031b1c90565b6106036105fd825f52600260205260405f2090565b966139b6565b9561060d81615559565b6107335761061a81615589565b156106e3576001600160a01b036106cb6105fd600198999a6106ab955f866106ba610661600c5f516020615d695f395f51905f529a01546001600160a01b039060401c1690565b9d8e9261067f846001600160a01b03165f52600660205260405f2090565b5493610691600483019586549061331f565b9c8d916001600160a01b03165f52600660205260405f2090565b5501805460ff19166003179055565b556106c5828d613339565b526139b6565b604051938452961691602090a25b94939291946105c4565b505050506106f391939250600455565b806106fa57005b81817f8fac6141d748dc9c9bc16cc25f636385597618190a44c03d33be5656e01b3642935261072e60405192839283614462565b0390a1005b505092939491610742906139b6565b926106d9565b50506004559190506106f3565b508185106105cd565b6105c29095949392956132ed565b5f6105ab565b34610295575f366003190112610295576020604051620186a08152f35b6004359060ff8216820361029557565b359060ff8216820361029557565b346102955760203660031901126102955760ff6107c861078f565b165f52600760205260405f2060405160408101918183106001600160401b03841117610826576040928352546001600160a01b03811680835260a09190911c6001600160401b03166020928301819052835191825291810191909152f35b6126de565b90816102609103126102955790565b90600319820160e081126102955760c0136102955760049160c435906001600160401b038211610295576104f29160040161082b565b6108793661083a565b60208101600261088882612bbf565b61089181611d68565b148015610b7b575b8015610b5d575b6108a990612bc9565b60026108b482612bbf565b6108bd81611d68565b03610b4e575b6109a36109016108d33686612c0e565b60c090207effffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff16600160f81b1790565b9261092f610920610919865f525f60205260405f2090565b5460ff1690565b610929816123bb565b15612c82565b61093b60208601612c98565b906109458661460e565b610955608087013583838861470f565b60a08161098861098161096a60808401612c98565b6001600160a01b03165f52600660205260405f2090565b5488614776565b604051632a2d120f60e21b8152958692839260048401612ec0565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49081156104a9577fb0d099feaab5034d04a1c610e86b8832343f2127b3c667b705834dafdf96e9e494610a18610a99936001600160a01b03965f91610b1f575b50610a07368b612c0e565b610a113686612fc4565b908a6148c2565b610a3c87610a37866001600160a01b03165f52600160205260405f2090565b61598d565b506002610a4882612bbf565b610a5181611d68565b03610a9e5750857f6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f4177869620669660405180610a878582613070565b0390a25b604051938493169683613081565b0390a3005b610aa9600391612bbf565b610ab281611d68565b03610aef57857f188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf98660405180610ae78582613070565b0390a2610a8b565b857f567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc60405180610ae78582613070565b610b41915060a03d60a011610b47575b610b39818361275f565b810190612ca2565b5f6109fc565b503d610b2f565b610b583415612bdf565b6108c3565b506108a9610b6a82612bbf565b610b7381611d68565b1590506108a0565b506003610b8782612bbf565b610b9081611d68565b14610899565b610b9f3661083a565b90610bc06004610bb160208501612bbf565b610bba81611d68565b14612bc9565b610bc98161460e565b610bd66108d33683612c0e565b916080610be560208401612c98565b92013591610bf58382848761470f565b610c19610c0183613110565b85906001600160401b03915f521660205260405f2090565b92610c23856149d5565b15610ca3575050610a997f471c4ebe4e57d25ef7117e141caac31c6b98f067b8098a7a7bbd38f637c2f98091610c836001600160a01b037f00000000000000000000000000000000000000000000000000000000000000001633146131e2565b610c8d3415612bdf565b610c978186614a31565b60405191829182613070565b9091610ccf60c082610cb4876140d5565b604051632ef10bcd60e21b815293849283926004840161311a565b038173__$682d6198b4eca5bc7e038b912a26498e7e$__5af49283156104a9577fede7867afa7cdb9c443667efd8244d98bf9df1dce68e60dc94dca6605125ca7694610a9994610d32935f91610d3b575b50610d2b3686612fc4565b8989614149565b610c9784613194565b610d54915060c03d60c0116104a257610494818361275f565b5f610d20565b34610295575f3660031901126102955760206040516001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000168152f35b9060406003198301126102955760043591602435906001600160401b038211610295576104f29160040161082b565b3461029557610dda36610d9d565b610deb6009610bb160208401612bbf565b610e076001610e01845f525f60205260405f2090565b016131f8565b610ea2610e1e60208301516001600160a01b031690565b91610e2f608082015184868861470f565b610e393685612fc4565b61014085019386610e4986613110565b6001600160401b031646149586610f5c575b50505060a081610e87610e8061096a60206060850151016001600160a01b0390511690565b5489614776565b604051632a2d120f60e21b8152958692839260048401613282565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49182156104a957610ed4935f93610f3b575b50866148c2565b15610f0a576104717f9a6f675cc94b83b55f1ecc0876affd4332a30c92e6faa2aca0199b1b6df922c39160405191829182613070565b6104717f7b20773c41402791c5f18914dbbeacad38b1ebcc4c55d8eb3bfe0a4cde26c8269160405191829182613070565b610f5591935060a03d60a011610b4757610b39818361275f565b915f610ecd565b610fbb92610f6e610fb6923690612ee5565b6060860152610f803660608b01612ee5565b6080860152610f8d61326e565b60a0860152610f9a61326e565b60c08601526001600160a01b03165f52600160205260405f2090565b615a37565b505f8681610e5b565b34610295575f366003190112610295576020604051612a308152f35b34610295575f36600319011261029557602060405160408152f35b346102955760403660031901126102955761056161101d60243560043561334d565b6040519182916020835260208301906104ae565b346102955761104861104236610d9d565b90613406565b005b6060600319820112610295576004359160243591604435906001600160401b038211610295576104f29160040161082b565b346102955761104861108d3661104a565b91613756565b34610295576020366003190112610295576001600160a01b036004356110b881610284565b165f5260016020526110cc60405f20615901565b5f905f5b815181101561115f576110f76109196110e98385613339565b515f525f60205260405f2090565b611100816123bb565b6003811415908161114a575b5061111a575b6001016110d0565b9161112d818460019310611135576139b6565b929050611112565b61113f8585613339565b516106c58286613339565b60059150611157816123bb565b14155f61110c565b50610561918152604051918291826104e1565b34610295575f3660031901126102955760206040516001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000168152f35b6040906003190112610295576004356111cd81610284565b906024356104f281610284565b346102955760206112176001600160a01b036111f5366111b5565b91165f526008835260405f20906001600160a01b03165f5260205260405f2090565b54604051908152f35b34610295575f366003190112610295576020600454604051908152f35b346102955761124b36610311565b611297611263859493945f52600560205260405f2090565b918254946112728615156126c9565b60a061127d88614c71565b604051809581926312031f5d60e11b8352600483016139c4565b038173__$b69fb814c294bfc16f92e50d7aeced4bde$__5af49081156104a9577fb8568a1f475f3c76759a620e08a653d28348c5c09e2e0bc91d533339801fefd8966103c296610461966060965f95611341575b50916113318596610453969385600561131560016113259901546001600160a01b039060081c1690565b97889360028401549a8b91613eb2565b92909193019e8f6128e3565b61133a896128e3565b908b614d2b565b6104539550611325939192966113716113319260a03d60a01161137d575b611369818361275f565b8101906136a5565b965096929193506112eb565b503d61135f565b346102955760603660031901126102955761139d61078f565b6024356113a981610284565b6044356001600160401b038111610295576114bc916113cf6114c19236906004016102d5565b93909461148261147d60ff8316966113e88815156139d5565b6001600160a01b038616986113fe8a15156139eb565b61143f8561143961142d61142d6114208460ff165f52600760205260405f2090565b546001600160a01b031690565b6001600160a01b031690565b15613a01565b61147761144d8b8730614e62565b917f0000000000000000000000000000000000000000000000000000000000000000933691612f73565b90614e9a565b613a1f565b61149c61148d612780565b6001600160a01b039094168452565b426001600160401b0316602084015260ff165f52600760205260405f2090565b613a35565b7f9ee792368f12db92ad66335fa19df35feaec025c86445fea202ab5412a180e055f80a3005b34610295575f366003190112610295576020604051620151808152f35b346102955761158d61151536610d9d565b61153661152760208395949501612bbf565b61153081611d68565b15612bc9565b61154c6001610e01855f525f60205260405f2090565b9061157161156460208401516001600160a01b031690565b608084015190838761470f565b60a08161098861158661096a60808401612c98565b5487614776565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49283156104a9577f567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc9361047193610c97925f926115f0575b506115e93685612fc4565b90876148c2565b61160a91925060a03d60a011610b4757610b39818361275f565b905f6115de565b346102955761161f3661083a565b906116316006610bb160208501612bbf565b61163a8161460e565b6116476108d33683612c0e565b91608061165660208401612c98565b920135916116668382848761470f565b611672610c0183613110565b9261167c856149d5565b156116b2575050610a9981610c977f587faad1bcd589ce902468251883e1976a645af8563c773eed7356d78433210c9386614a31565b90916116ee60a0826116d46116cd61096a6101608401612c98565b5488614cce565b60405162ea54e760e01b815293849283926004840161373f565b038173__$b69fb814c294bfc16f92e50d7aeced4bde$__5af49283156104a9577f17eb0a6bd5a0de45d1029ce3444941070e149df35b22176fc439f930f73c09f794610a9994610c97935f91611751575b5061174a3686612fc4565b8989614d2b565b61176a915060a03d60a01161137d57611369818361275f565b5f61173f565b6080366003190112610295576004356024356001600160401b0381116102955761179e90369060040161082b565b6044356001600160401b038111610295576117bd9036906004016102d5565b90916117c7610302565b926117d9855f525f60205260405f2090565b6117e5600182016131f8565b936117f1825460ff1690565b906117fb826123bb565b6001821495868015611adb575b61181190612c82565b61181d600585016128e3565b9261185b61182a88613110565b6001600160401b0361185261184688516001600160401b031690565b6001600160401b031690565b91161015613aa3565b60208201516001600160a01b0316978a6080840151956001600160401b036118966118466118888d613110565b93516001600160401b031690565b91161115611a8d57506118eb61192d9493926004926118d660208c01926118d160016118c186612bbf565b6118ca81611d68565b1415612bc9565b6123bb565b80611a6d575b6118e69015612bc9565b612bbf565b6118f481611d68565b1480611a3a575b61190590156131e2565b6119118489898d61470f565b60a08761098861192661096a60808401612c98565b548d614776565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49182156104a9577f07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a996014996119bb8d8b6119af6119ee9a6119c197611a0c9e6119aa6119d69c6119df9e5f91611a1b575b506119a33688612fc4565b8d896152c4565b613eb2565b93919490923690612fc4565b90613fe3565b845460ff191660021785555163ffffffff1690565b63ffffffff1690565b6001600160401b034216613ad9565b9301805467ffffffffffffffff19166001600160401b038516179055565b61047160405192839283613af9565b611a34915060a03d60a011610b4757610b39818361275f565b5f611998565b50337f00000000000000000000000000000000000000000000000000000000000000006001600160a01b031614156118fb565b506118e66009611a7c83612bbf565b611a8581611d68565b1490506118dc565b6119d69392506119c19150996014996119bb7f07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a9c8b6119af6119ee9a6119df9a611a0c9e6119aa3415612bdf565b50611ae5836123bb565b60048314611808565b604036600319011261029557600435611b0681610284565b6001600160a01b0360243591611b1d831515613b19565b611b25615604565b611b30838233615498565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f00551690815f52600660205260405f205490808201809211611bd8575f516020615d695f395f51905f5291837f2da466a7b24304f47e87fa2e1e5a81b9831ce54fec19055ce277ca2f39ba42c4611bc561047194835f5260066020528460405f2055604051918291829190602083019252565b0390a26040519081529081906020820190565b6132a7565b34610295575f36600319011261029557602060405162093a808152f35b3461029557611c1f611c0b36610d9d565b6115366003610bb160208496959601612bbf565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49283156104a9577f188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf9869361047193610c97925f926115f057506115e93685612fc4565b34610295575f36600319011261029557600354600454905f805b82841015611d3c577fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b8401545f90815260026020526040902091611cd783615559565b611d2a57611ce483615589565b15611d1357611d0a916004611cfb611d04936139b6565b9401549061331f565b936139b6565b915b9192611c94565b92509250505b604080519182526020820192909252f35b915092611d36906139b6565b91611d0c565b92509050611d19565b634e487b7160e01b5f52602160045260245ffd5b60041115611d6357565b611d45565b600a1115611d6357565b90600a821015611d635752565b805180835260209291819084018484015e5f828201840152601f01601f1916010190565b6104f2916001600160401b038251168152611dc660208301516020830190611d72565b60408201516040820152611e336060830151606083019060c080916001600160401b0381511684526001600160a01b03602082015116602085015260ff6040820151166040850152606081015160608501526080810151608085015260a081015160a08501520151910152565b60808281015180516001600160401b031661014084015260208101516001600160a01b0316610160840152604081015160ff1661018084015260608101516101a0840152908101516101c083015260a08101516101e083015260c0015161020082015260c0611eb460a0840151610260610220850152610260840190611d7f565b92015190610240818403910152611d7f565b92936001600160401b0360c0956104f298979482948752611ee681611d59565b602087015216604085015216606083015260808201528160a08201520190611da3565b3461029557602036600319011261029557600435611f25613b65565b505f52600260205260405f2060405190611f3e826126f2565b80548252610561600182015491611f89611f79611f5b8560ff1690565b94611f6a602088019687613ba9565b60081c6001600160a01b031690565b6001600160a01b03166040860152565b6002810154606085015260038101546001600160401b0380821660808701908152959160401c166001600160401b031660a0820190815291612001611888611fdf600560048501549460c08701958652016128e3565b9360e0810194855251965197611ff489611d59565b516001600160401b031690565b905191519260405196879687611ec6565b346102955760603660031901126102955760043561202f81610284565b5f516020615d695f395f51905f526104716024359261204d84610284565b604435936120656001600160a01b03831615156139eb565b612070851515613b19565b6120a46001600160a01b037f00000000000000000000000000000000000000000000000000000000000000001633146131e2565b7f7084f5476618d8e60b11ef0d7d3f06914655adb8793e28ff7f018d4c76d505d5611bc58661211e6001600160a01b038516988995865f5260066020526120fb8260405f20546120f682821015613bb5565b61332c565b9788612118836001600160a01b03165f52600660205260405f2090565b556155b8565b6040519081529081906020820190565b346102955761213c3661083a565b61214d6008610bb160208401612bbf565b61215a6108d33684612c0e565b916121bb61216a60208301612c98565b9161217b608082013584868861470f565b6121853685612fc4565b61218e866149d5565b93868515612284575b505060a081610e87610e8061096a60206060850151016001600160a01b0390511690565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49182156104a9576121f8935f9361225f575b506121f2903690612c0e565b866148c2565b1561222e576104717f3142fb397e715d80415dff7b527bf1c451def4675da6e1199ee1b4588e3f630a9160405191829182613070565b6104717f26afbcb9eb52c21f42eb9cfe8f263718ffb65afbf84abe8ad8cce2acfb2242b89160405191829182613070565b6121f291935061227d9060a03d60a011610b4757610b39818361275f565b92906121e6565b610a376122a2926122948661460e565b610f6e366101408b01612ee5565b505f86612197565b9160a0936001600160401b03916104f297969385526122c881611d59565b602085015216604083015260608201528160808201520190611da3565b3461029557602036600319011261029557600435612301613b65565b505f52600560205260405f206040519061231a8261270e565b80548252610561600182015491612351611f7960ff851694602087019561234081611d59565b865260081c6001600160a01b031690565b6002810154606085015260038101546001600160401b03166001600160401b031660808501908152936123aa612395600560048501549460a08501958652016128e3565b9160c0810192835251945195611ff487611d59565b9151905191604051958695866122aa565b60061115611d6357565b906006821015611d635752565b919260a0610120946123eb85612454959a99989a6123c5565b63ffffffff81511660208601526001600160a01b0360208201511660408601526001600160a01b0360408201511660608601526001600160401b036060820151166080860152608081015182860152015160c084015261014060e0840152610140830190611da3565b946101008201520152565b34610295576020366003190112610295576004355f60a060405161248281612729565b82815282602082015282604082015282606082015282608082015201526124a7613b65565b505f525f6020526124ba60405f20613bd7565b80516124c5816123bb565b61056160208301519260408101519060606124ed61184660808401516001600160401b031690565b91015191604051958695866123d2565b61251d61250936610d9d565b6115366002610bb160208496959601612bbf565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49283156104a9577f6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f417786962066969361047193610c97925f926115f057506115e93685612fc4565b3461029557612586366111b5565b61258e615604565b6001600160a01b038116916125a48315156139eb565b6001600160a01b036125e1826125cb336001600160a01b03165f52600860205260405f2090565b906001600160a01b03165f5260205260405f2090565b54916125ee831515613b19565b5f61260e826125cb336001600160a01b03165f52600860205260405f2090565b551691818361268857612631915f808080858a5af161262b613c34565b50613c63565b60405190815233907f7b8d70738154be94a9a068a6d2f5dd8cfc65c52855859dc8f47de1ff185f8b5590602090a461104860017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055565b6126929184615662565b612631565b34610295576110486126a83661104a565b91613c8b565b34610295575f36600319011261029557602060405160018152f35b156126d057565b6287a33760e41b5f5260045ffd5b634e487b7160e01b5f52604160045260245ffd5b61010081019081106001600160401b0382111761082657604052565b60e081019081106001600160401b0382111761082657604052565b60c081019081106001600160401b0382111761082657604052565b60a081019081106001600160401b0382111761082657604052565b90601f801991011681019081106001600160401b0382111761082657604052565b6040519061278f60408361275f565b565b6040519061278f60e08361275f565b906040516127ad8161270e565b60c0600482946127ea60ff82546001600160401b03811687526001600160a01b03808260401c1616602088015260e01c16604086019060ff169052565b6001810154606085015260028101546080850152600381015460a08501520154910152565b90600182811c9216801561283d575b602083101461282957565b634e487b7160e01b5f52602260045260245ffd5b91607f169161281e565b5f92918154916128568361280f565b80835292600181169081156128ab575060011461287257505050565b5f9081526020812093945091925b838310612891575060209250010190565b600181602092949394548385870101520191019190612880565b915050602093945060ff929192191683830152151560051b010190565b9061278f6128dc9260405193848092612847565b038361275f565b906040516128f08161270e565b809260ff81546001600160401b038116845260401c1690600a821015611d6357600d6129619160c093602086015260018101546040860152612934600282016127a0565b6060860152612945600782016127a0565b6080860152612956600c82016128c8565b60a0860152016128c8565b910152565b5190600482101561029557565b6001600160401b0381160361029557565b5190811515820361029557565b908160c0910312610295576129f960a0604051926129ae84612729565b80518452602081015160208501526129c860408201612966565b604085015260608101516129db81612973565b606085015260808101516129ee81612973565b608085015201612984565b60a082015290565b908151612a0d81611d59565b815260806001600160401b0381612a33602086015160a0602087015260a0860190611da3565b946040810151604086015282606082015116606086015201511691015290565b9060206104f2928181520190612a01565b6040513d5f823e3d90fd5b90600d6104f292612a9781546001600160401b038116855260ff602086019160401c16611d72565b60018101546040840152612b036060840160028301600460c09160ff8082546001600160401b03811687526001600160a01b038160401c16602088015260e01c161660408501526001810154606085015260028101546080850152600381015460a08501520154910152565b60078101546001600160401b038116610140850152604081901c6001600160a01b031661016085015260e01c60ff1661018084015260088101546101a084015260098101546101c0840152600a8101546101e0840152600b810154610200840152610260610220840152612b7e6102608401600c8301612847565b9261024081850391015201612847565b906001600160401b03612bae602092959495604085526040850190612a6f565b9416910152565b600a111561029557565b356104f281612bb5565b15612bd057565b633226144f60e21b5f5260045ffd5b15612be657565b636956f2ab60e11b5f5260045ffd5b63ffffffff81160361029557565b359061278f82612973565b91908260c091031261029557604051612c2681612729565b60a08082948035612c3681612bf5565b84526020810135612c4681610284565b60208501526040810135612c5981610284565b60408501526060810135612c6c81612973565b6060850152608081013560808501520135910152565b15612c8957565b631e40ad6360e31b5f5260045ffd5b356104f281610284565b908160a09103126102955760405190612cba82612744565b80518252602081015160208301526040810151600681101561029557612cfb9160809160408501526060810151612cf081612973565b606085015201612984565b608082015290565b90612d0f8183516123c5565b60806001600160401b0381612d33602086015160a0602087015260a0860190611da3565b94604081015160408601526060810151606086015201511691015290565b359061278f82612bb5565b60c080916001600160401b038135612d7381612973565b1684526001600160a01b036020820135612d8c81610284565b16602085015260ff612da06040830161079f565b166040850152606081013560608501526080810135608085015260a081013560a08501520135910152565b9035601e19823603018112156102955701602081359101916001600160401b03821161029557813603831361029557565b908060209392818452848401375f828201840152601f01601f1916010190565b6104f2916001600160401b038235612e3381612973565b168152612e516020830135612e4781612bb5565b6020830190611d72565b60408201356040820152612e6b6060820160608401612d5c565b612e7d61014082016101408401612d5c565b612eb1612ea5612e91610220850185612dcb565b610260610220860152610260850191612dfc565b92610240810190612dcb565b91610240818503910152612dfc565b9091612ed76104f293604084526040840190612d03565b916020818403910152612e1c565b91908260e091031261029557604051612efd8161270e565b60c08082948035612f0d81612973565b84526020810135612f1d81610284565b6020850152612f2e6040820161079f565b6040850152606081013560608501526080810135608085015260a081013560a08501520135910152565b6001600160401b03811161082657601f01601f191660200190565b929192612f7f82612f58565b91612f8d604051938461275f565b829481845281830111610295578281602093845f960137010152565b9080601f83011215610295578160206104f293359101612f73565b9190916102608184031261029557612fda612791565b92612fe482612c03565b8452612ff260208301612d51565b60208501526040820135604085015261300e8160608401612ee5565b6060850152613021816101408401612ee5565b60808501526102208201356001600160401b0381116102955781613046918401612fa9565b60a08501526102408201356001600160401b038111610295576130699201612fa9565b60c0830152565b9060206104f2928181520190612e1c565b60e09060a06104f2949363ffffffff813561309b81612bf5565b1683526001600160a01b0360208201356130b481610284565b1660208401526001600160a01b0360408201356130d081610284565b1660408401526001600160401b0360608201356130ec81612973565b16606084015260808101356080840152013560a08201528160c08201520190612e1c565b356104f281612973565b9091612ed76104f293604084526040840190612a01565b634e487b7160e01b5f52603260045260245ffd5b60035481101561315d5760035f5260205f2001905f90565b613131565b805482101561315d575f5260205f2001905f90565b916131909183549060031b91821b915f19901b19161790565b9055565b60035468010000000000000000811015610826576001810160035560035481101561315d5760035f527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b0155565b156131e957565b6370a8bfcd60e11b5f5260045ffd5b9060405161320581612729565b60a0600382946001600160a01b03815463ffffffff8116865260201c16602085015261325d6001600160401b0360018301546001600160a01b03808216166040880152851c1660608601906001600160401b03169052565b600281015460808501520154910152565b6040519061327d60208361275f565b5f8252565b90916132996104f293604084526040840190612d03565b916020818403910152611da3565b634e487b7160e01b5f52601160045260245ffd5b6001600160401b0381116108265760051b60200190565b604051906132e160208361275f565b5f808352366020840137565b906132f7826132bb565b613304604051918261275f565b8281528092613315601f19916132bb565b0190602036910137565b91908201809211611bd857565b91908203918211611bd857565b805182101561315d5760209160051b010190565b91906003549080840293808504821490151715611bd857818410156133d157830190818411611bd8578082116133c9575b5061339161338c848361332c565b6132ed565b92805b8281106133a057505050565b806133af6105de600193613145565b6133c26133bc858461332c565b88613339565b5201613394565b90505f61337e565b505090506104f26132d2565b906006811015611d635760ff80198354169116179055565b9060206104f2928181520190611da3565b90613418825f525f60205260405f2090565b613424600182016131f8565b91613430825460ff1690565b918461343e600583016128e3565b91600261345560208801516001600160a01b031690565b9561345f816123bb565b148061364e575b6135755750505061347e6001610bb160208401612bbf565b61348e608084015183838761470f565b6134c160a0826134a661098161096a60808401612c98565b604051632a2d120f60e21b8152938492839260048401612ec0565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af480156104a957610fb661354f9461352b88937f04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a898613542965f92613554575b506135243689612fc4565b90866148c2565b6001600160a01b03165f52600160205260405f2090565b5060405191829182613070565b0390a2565b61356e91925060a03d60a011610b4757610b39818361275f565b905f613519565b7f04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a8955061364192935061354f946135d46014836135bc610fb695600360ff19825416179055565b5f601382015501805467ffffffffffffffff19169055565b61352b60608601613600815160606135f660208301516001600160a01b031690565b9101519085614ae5565b5160a061361760208301516001600160a01b031690565b910151907f0000000000000000000000000000000000000000000000000000000000000000614ae5565b50604051918291826133f5565b506014810154426001600160401b0390911610613466565b1561366d57565b6336c7a86b60e21b5f5260045ffd5b9061368681611d59565b60ff80198354169116179055565b9060206104f2928181520190612a6f565b908160a091031261029557612cfb6080604051926136c284612744565b80518452602081015160208501526136dc60408201612966565b60408501526060810151612cf081612973565b9081516136fb81611d59565b8152608080613719602085015160a0602086015260a0850190611da3565b93604081015160408501526001600160401b036060820151166060850152015191015290565b9091612ed76104f2936040845260408401906136ef565b916137618284614c4f565b61394d57613777825f52600560205260405f2090565b9061378484835414613666565b600182018054929060026137a7600886901c6001600160a01b03165b9560ff1690565b6137b081611d59565b1480613935575b61384e57506002906137d06007610bb160208601612bbf565b0154906137df8284838861470f565b6137ee60a0826116d487614c71565b038173__$b69fb814c294bfc16f92e50d7aeced4bde$__5af49283156104a9577f2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d19461384994610c97935f91611751575061174a3686612fc4565b0390a3565b805460ff191660031790557f2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d1925060059150600481015f815491556138a0600383016001600160401b03198154169055565b5f516020615d695f395f51905f526001600160a01b036138f36138d1600c8601546001600160a01b039060401c1690565b936138ed856001600160a01b03165f52600660205260405f2090565b5461331f565b9283613910826001600160a01b03165f52600660205260405f2090565b556040519384521691602090a261392561447e565b6138496040519283920182613694565b506003820154426001600160401b03909116106137b7565b613849816139836007610bb160207f6d0cf3d243d63f08f50db493a8af34b27d4e3bc9ec4098e82700abfeffe2d4989601612bbf565b610c8d613997865f525f60205260405f2090565b600181015460039060201c6001600160a01b031691015490838861470f565b5f198114611bd85760010190565b9060206104f29281815201906136ef565b156139dc57565b6306ee4dcd60e01b5f5260045ffd5b156139f257565b63e6c4247b60e01b5f5260045ffd5b15613a095750565b60ff906357470ffd60e01b5f521660045260245ffd5b15613a2657565b63c1606c2f60e01b5f5260045ffd5b6001600160401b03602061278f93613a7a6001600160a01b0382511685906001600160a01b031673ffffffffffffffffffffffffffffffffffffffff19825416179055565b0151825467ffffffffffffffff60a01b1916911660a01b67ffffffffffffffff60a01b16179055565b15613aaa57565b637d95736160e01b5f5260045ffd5b6001600160401b0362015180911601906001600160401b038211611bd857565b906001600160401b03809116911601906001600160401b038211611bd857565b906001600160401b03612bae602092959495604085526040850190612e1c565b15613b2057565b6334b2073960e11b5f5260045ffd5b60405190613b3c8261270e565b5f60c0838281528260208201528260408201528260608201528260808201528260a08201520152565b60405190613b728261270e565b606060c0835f81525f60208201525f6040820152613b8e613b2f565b83820152613b9a613b2f565b60808201528260a08201520152565b613bb282611d59565b52565b15613bbc57565b631e9acf1760e31b5f5260045ffd5b6006821015611d635752565b90604051613be481612744565b60806001600160401b0360148395613c0060ff82541686613bcb565b613c0c600182016131f8565b6020860152613c1d600582016128e3565b604086015260138101546060860152015416910152565b3d15613c5e573d90613c4582612f58565b91613c53604051938461275f565b82523d5f602084013e565b606090565b15613c6c575050565b6001600160a01b039063296c17bb60e21b5f521660045260245260445ffd5b91613c9682846156bb565b613e1c57613cac825f52600260205260405f2090565b90613cb984835414613666565b60018201805492906002613cd9600886901c6001600160a01b03166137a0565b613ce281611d59565b1480613df9575b613d7b5750600290613d026005610bb160208601612bbf565b015490613d118284838861470f565b613d2060c082610cb4876140d5565b038173__$682d6198b4eca5bc7e038b912a26498e7e$__5af49283156104a9577f1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e9461384994610c97935f91610d3b5750610d2b3686612fc4565b805460ff191660031790557f1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e9260059250613df19060048301905f82549255613dda600385016fffffffffffffffff0000000000000000198154169055565b600c84015460401c6001600160a01b031690614ae5565b61392561447e565b50600382015460401c6001600160401b03166001600160401b0342911610613ce9565b613849816139836005610bb160207f32e24720f56fd5a7f4cb219d7ff3278ae95196e79c85b5801395894a6f53466c9601612bbf565b15613e5957565b6306a41ced60e21b5f5260045ffd5b15613e705750565b60ff9063399eb60560e01b5f521660045260245ffd5b15613e8f575050565b9060ff6001600160401b039263975133f360e01b5f52166004521660245260445ffd5b9291908015613f8c57801561315d57613f0191843560f81c9081613f0557507f000000000000000000000000000000000000000000000000000000000000000094600101925f19909201919050565b9091565b600180613f1884613f1f949060ff161c90565b1614613e52565b613f7f613f378260ff165f52600760205260405f2090565b546001600160a01b0381169290613f6c90613f6790613f5884871515613e68565b60a01c6001600160401b031690565b613ab9565b906001600160401b038216421015613e86565b93600101915f1990910190565b63ac241e1160e01b5f5260045ffd5b90816020910312610295575190565b9392606093613fd56001600160a01b0394612bae949998998852608060208901526080880190611d7f565b918683036040880152612dfc565b9193929590613ff1906156d3565b916002821015611d63576020956001600160a01b039261407a5761402d905b604051635850a09b60e11b81529889978896879560048701613faa565b0392165afa80156104a95761278f915f9161404b575b501515613a1f565b61406d915060203d602011614073575b614065818361275f565b810190613f9b565b5f614043565b503d61405b565b5061402d7f0000000000000000000000000000000000000000000000000000000000000000614010565b604051906140b182612744565b5f6080838281526140c0613b65565b60208201528260408201528260608201520152565b6140dd6140a4565b905f5260026020526001600160401b0380600360405f2060ff60018201541661410581611d59565b8552614113600582016128e3565b6020860152600481015460408601520154818116606085015260401c1616608082015290565b600160ff1b8114611bd8575f0390565b936141b694602094939682614166835f52600260205260405f2090565b9860a08701956141768751151590565b156144495760808201518901516001600160a01b0316998a975b60408a018d81516141a081611d59565b6141a981611d59565b61442b575b505051151590565b614418575b50505050506141d460608401516001600160401b031690565b6001600160401b0381166143ef575b5060038601805460808501516001600160401b039081169160401c168190036143b8575b50505f8351135f1461436b576142299061422184516158e5565b92839161548a565b6142386004860191825461331f565b90555b0180515f8113156142d057505f516020615d695f395f51905f52916142686001600160a01b0392516158e5565b6142b960046142928361428c866001600160a01b03165f52600660205260405f2090565b5461332c565b96876142af866001600160a01b03165f52600660205260405f2090565b550191825461331f565b90556040519384521691602090a25b61278f61447e565b90505f81126142e2575b5050506142c8565b5f516020615d695f395f51905f529161430a6143056001600160a01b0393614139565b6158e5565b614355600461432e836138ed866001600160a01b03165f52600660205260405f2090565b968761434b866001600160a01b03165f52600660205260405f2090565b550191825461332c565b90556040519384521691602090a25f80806142da565b6143753415612bdf565b8251905f8212614388575b50505061423b565b61439761430561439f93614139565b928391614ae5565b6143ae6004860191825461332c565b9055825f80614380565b81546fffffffffffffffff0000000000000000191660409190911b6fffffffffffffffff0000000000000000161790555f80614207565b6144129060038801906001600160401b03166001600160401b0319825416179055565b5f6141e3565b614421946157eb565b5f808281806141bb565b600161444292519161443c83611d59565b0161367c565b5f8d6141ae565b600c8b015460401c6001600160a01b0316998a97614190565b9291906144796020916040865260408601906104ae565b930152565b6003546004545f928390828411156145e85761449a838561332c565b806040105f146145da57506144b4604095949392956132ed565b925b808310806145d0575b156145c2576144d06105de84613145565b6144e56105fd825f52600260205260405f2090565b956144ef81615559565b6145ad576144fc81615589565b1561455b576001600160a01b036145436105fd600198999a6106ab955f866106ba610661600c5f516020615d695f395f51905f529a01546001600160a01b039060401c1690565b604051938452961691602090a25b94939291946144b6565b5050509391925061456b90600455565b80614574575050565b81817f8fac6141d748dc9c9bc16cc25f636385597618190a44c03d33be5656e01b364293526145a860405192839283614462565b0390a1565b5050929394916145bc906139b6565b92614551565b509391925061456b90600455565b50604085106144bf565b6144b49095949392956132ed565b5f61449a565b356104f281612bf5565b156145ff57565b630596b15b60e01b5f5260045ffd5b6001600160a01b03602082013561462481610284565b166146308115156139eb565b6001600160a01b03604083013561464681610284565b817f00000000000000000000000000000000000000000000000000000000000000001691829116036146ce5781146146bc5750806201518063ffffffff61468f61278f946145ee565b161015908161469f575b506145f8565b62093a8091506146b363ffffffff916145ee565b1611155f614699565b63abfa558d60e01b5f5260045260245ffd5b6308ad910960e21b5f5260045ffd5b903590601e198136030182121561029557018035906001600160401b0382116102955760200191813603831361029557565b909161278f9361473f61474d926147348361472e6102208901896146dd565b90613eb2565b908888949394615949565b61472e6102408501856146dd565b91937f000000000000000000000000000000000000000000000000000000000000000093615949565b9060146001600160401b039161478a6140a4565b935f525f60205260405f20906147a460ff83541686613bcb565b6147b0600583016128e3565b6020860152601382015460408601526060850152015416608082015290565b9060a060039163ffffffff81511663ffffffff198554161784556001600160a01b036020820151167fffffffffffffffff0000000000000000000000000000000000000000ffffffff77ffffffffffffffffffffffffffffffffffffffff0000000086549260201b1691161784556148b16001850161488461485b60408501516001600160a01b031690565b82906001600160a01b031673ffffffffffffffffffffffffffffffffffffffff19825416179055565b6060830151815467ffffffffffffffff60a01b191660a09190911b67ffffffffffffffff60a01b16179055565b608081015160028501550151910155565b926148fe8161494d946080946148df885f525f60205260405f2090565b976148eb895460ff1690565b6148f4816123bb565b156149c3576152c4565b60408101805161490d816123bb565b614916816123bb565b151580614998575b61497e575b5060148401805460608301516001600160401b03908116911681900361495c575b50500151151590565b6149545750565b60135f910155565b815467ffffffffffffffff19166001600160401b039091161790555f80614944565b614992905161498c816123bb565b856133dd565b5f614923565b50845460ff168151906149aa826123bb565b6149b3826123bb565b6149bc816123bb565b141561491e565b6149d08260018b016147cf565b6152c4565b805f525f60205260ff60405f2054166006811015611d63578015908115614a1d575b50614a18575f525f6020526001600160401b03600760405f20015416461490565b505f90565b60059150614a2a816123bb565b145f6149f7565b90614a8391805f525f602052614a4c600160405f20016131f8565b60a083614a68614a6161096a60808401612c98565b5485614776565b604051632a2d120f60e21b8152968792839260048401612ec0565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af49283156104a95761278f945f94614ac0575b50614aba903690612fc4565b916148c2565b614aba919450614ade9060a03d60a011610b4757610b39818361275f565b9390614aae565b90614af89291614af3615604565b614b1e565b60017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055565b9190918115614c4a576001600160a01b0383169283614bc2576001600160a01b038216925f8080808488620186a0f1614b55613c34565b5015614b62575050505050565b614ba5613849926125cb7fbf182be802245e8ed88e4b8d3e4344c0863dd2a70334f089fd07265389306fcf956001600160a01b03165f52600860205260405f2090565b614bb082825461331f565b90556040519081529081906020820190565b614bd4614bd0848484615add565b1590565b614bdf575b50505050565b81614c286001600160a01b03926125cb7fbf182be802245e8ed88e4b8d3e4344c0863dd2a70334f089fd07265389306fcf956001600160a01b03165f52600860205260405f2090565b614c3385825461331f565b90556040519384521691602090a35f808080614bd9565b505050565b905f52600560205260405f2054159081614c67575090565b6104f291506149d5565b614c796140a4565b905f5260056020526001600160401b03600360405f2060ff600182015416614ca081611d59565b8452614cae600582016128e3565b60208501526004810154604085015201541660608201525f608082015290565b90614cd76140a4565b915f5260056020526001600160401b03600360405f2060ff600182015416614cfe81611d59565b8552614d0c600582016128e3565b6020860152600481015460408601520154166060830152608082015290565b6020939291614db691614d46815f52600560205260405f2090565b97604086018051614d5681611d59565b614d5f81611d59565b614e45575b5087856080880194614d768651151590565b614e32575b505050505060038701614d9581546001600160401b031690565b60608601516001600160401b039081169116819003614e1057505051151590565b15614df757608001518201516001600160a01b031680935b8251905f821315614de857614229915061422184516158e5565b5f82126143885750505061423b565b50600c84015460401c6001600160a01b03168093614dce565b815467ffffffffffffffff19166001600160401b039091161790555f806141ae565b614e3b94615b4a565b5f80878582614d7b565b614e5c9051614e5381611d59565b60018b0161367c565b5f614d64565b9160ff6001600160a01b03928360405195466020880152166040860152166060840152166080820152608081526104f260a08261275f565b805192835f947a184f03e93ff9f4daa797ed6e38ed64bf6a1f010000000000000000821015615036575b806d04ee2d6d415b85acef8100000000600a92101561501a575b662386f26fc10000811015615005575b6305f5e100811015614ff3575b612710811015614fe3575b6064811015614fd4575b1015614fc9575b614f606021614f2860018801615c08565b968701015b5f1901917f3031323334353637383961626364656600000000000000000000000000000000600a82061a8353600a900490565b908115614f7057614f6090614f2d565b50506001600160a01b03614f9584614f89858498615b9c565b60208151910120615bf2565b911693168314614fc157614fb39181602061142d9351910120615bf2565b14614fbc575f90565b600190565b505050600190565b600190940193614f17565b60029060649004960195614f10565b6004906127109004960195614f06565b6008906305f5e1009004960195614efb565b601090662386f26fc100009004960195614eee565b6020906d04ee2d6d415b85acef81000000009004960195614ede565b50604094507a184f03e93ff9f4daa797ed6e38ed64bf6a1f0100000000000000008104614ec4565b90600a811015611d635768ff000000000000000082549160401b169068ff00000000000000001916179055565b8151815460208401516040808601516001600160401b039094167fffffff000000000000000000000000000000000000000000000000000000000090931692909217911b7bffffffffffffffffffffffffffffffffffffffff0000000000000000161760e09190911b60ff60e01b16178155606082015160018201556080820151600282015560a0820151600382015560c090910151600490910155565b601f821161513657505050565b5f5260205f20906020601f840160051c8301931061516e575b601f0160051c01905b818110615163575050565b5f8155600101615158565b909150819061514f565b91909182516001600160401b0381116108265761519f81615199845461280f565b84615129565b6020601f82116001146151da5781906131909394955f926151cf575b50508160011b915f199060031b1c19161790565b015190505f806151bb565b601f198216906151ed845f5260205f2090565b915f5b8181106152275750958360019596971061520f575b505050811b019055565b01515f1960f88460031b161c191690555f8080615205565b9192602060018192868b0151815501940192016151f0565b8151815467ffffffffffffffff19166001600160401b0391909116178155602082015191600a831015611d635760c0600d9161527e61278f958561505e565b6040810151600185015561529960608201516002860161508b565b6152aa60808201516007860161508b565b6152bb60a0820151600c8601615178565b01519101615178565b9161531360206152e1615305959694965f525f60205260405f2090565b956152f982606086015101516001600160a01b031690565b9586946005890161523f565b01516001600160a01b031690565b5f8351135f1461547b5761532783516158e5565b61533281848461548a565b6153416013870191825461331f565b90555b602083019283515f81136153fa575b5051905f82126153d2575b505050515f8112615375575b50505061278f61447e565b5f516020615d695f395f51905f52916153986143056001600160a01b0393614139565b6153bc601361432e836138ed866001600160a01b03165f52600660205260405f2090565b90556040519384521691602090a25f808061536a565b6143976143056153e193614139565b6153f06013850191825461332c565b9055815f8061535e565b615403906158e5565b6154228161428c866001600160a01b03165f52600660205260405f2090565b908161543f866001600160a01b03165f52600660205260405f2090565b5561544f6013890191825461331f565b90556040519081526001600160a01b038416905f516020615d695f395f51905f5290602090a25f615353565b6154853415612bdf565b615344565b90614af89291615498615604565b908215614c4a576001600160a01b0316918215801561554a576154bc823414612bdf565b156154c657505050565b6001600160a01b03604051926323b872dd60e01b5f52166004523060245260445260205f60648180865af160015f511481161561552b575b6040919091525f606052156155105750565b635274afe760e01b5f526001600160a01b031660045260245ffd5b6001811516615541573d15833b151516166154fe565b503d5f823e3d90fd5b6155543415612bdf565b6154bc565b6001015460ff1661556981611d59565b60038114908115615578575090565b6002915061558581611d59565b1490565b6001600160401b0360038201541642101590816155a4575090565b600180925060ff9101541661558581611d59565b90614af892916155c6615604565b91908115614c4a576001600160a01b031691826155fb5761278f92505f808080856001600160a01b0386165af161262b613c34565b61278f92615662565b60027f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0054146156535760027f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055565b633ee5aeb560e01b5f5260045ffd5b916001600160a01b036040519263a9059cbb60e01b5f521660045260245260205f60448180865af160015f51148116156156a5575b604091909152156155105750565b6001811516615541573d15833b15151616615697565b905f52600260205260405f2054159081614c67575090565b6001600160401b03815116906020810151600a811015611d635761577a8260406157da94015161571a60806060840151930151946040519760208901526040880190611d72565b6060860152608085019060c080916001600160401b0381511684526001600160a01b03602082015116602085015260ff6040820151166040850152606081015160608501526080810151608085015260a081015160a08501520151910152565b80516001600160401b031661016084015260208101516001600160a01b0316610180840152604081015160ff166101a084015260608101516101c084015260808101516101e084015260a081015161020084015260c00151610220830152565b61022081526104f26102408261275f565b9190915f52600260205260405f20918255600582019261582b6001600160401b0383511685906001600160401b03166001600160401b0319825416179055565b602082015193600a851015611d635760c06158e19361584f6002976158979461505e565b6040810151600687015561586a60608201516007880161508b565b61587b6080820151600c880161508b565b61588c60a082015160118801615178565b015160128501615178565b60018301907fffffffffffffffffffffff0000000000000000000000000000000000000000ff74ffffffffffffffffffffffffffffffffffffffff0083549260081b169116179055565b0155565b5f81126158ef5790565b635467221960e11b5f5260045260245ffd5b90604051918281549182825260208201905f5260205f20925f5b81811061593057505061278f9250038361275f565b845483526001948501948794506020909301920161591b565b6001600160a01b039061402d61596f61596a60209895999697993690612fc4565b6156d3565b936040519889978896879563600109bb60e01b875260048701613faa565b6001810190825f528160205260405f2054155f146159f557805468010000000000000000811015610826576159e26159cc826001879401855584613162565b819391549060031b91821b915f19901b19161790565b905554915f5260205260405f2055600190565b5050505f90565b80548015615a23575f190190615a128282613162565b8154905f199060031b1b1916905555565b634e487b7160e01b5f52603160045260245ffd5b6001810191805f528260205260405f2054928315155f14615ad5575f198401848111611bd85783545f19810194908511611bd8575f958583615a9297615a859503615a98575b5050506159fc565b905f5260205260405f2090565b55600190565b615abe615ab891615aaf6105de615acc9588613162565b92839187613162565b90613177565b85905f5260205260405f2090565b555f8080615a7d565b505050505f90565b60405163a9059cbb60e01b60208281019182526001600160a01b03909416602483015260448083019590955293815290925f91615b1b60648261275f565b51908285620186a0f15f51913d91156159f5578115615b415750602011614a1857151590565b9150503b151590565b9190915f52600560205260405f20918255600582019261582b6001600160401b0383511685906001600160401b03166001600160401b0319825416179055565b805191908290602001825e015f815290565b61278f90615be4615bde94936040519586937f19457468657265756d205369676e6564204d6573736167653a0a0000000000006020860152603a850190615b8a565b90615b8a565b03601f19810184528361275f565b6104f291615bff91615c30565b90929192615c6a565b90615c1282612f58565b615c1f604051918261275f565b8281528092613315601f1991612f58565b8151919060418303615c6057615c599250602082015190606060408401519301515f1a90615ce6565b9192909190565b50505f9160029190565b615c7381611d59565b80615c7c575050565b615c8581611d59565b60018103615c9c5763f645eedf60e01b5f5260045ffd5b615ca581611d59565b60028103615cc0575063fce698f760e01b5f5260045260245ffd5b80615ccc600392611d59565b14615cd45750565b6335e2f38360e21b5f5260045260245ffd5b91907f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a08411615d5d579160209360809260ff5f9560405194855216868401526040830152606082015282805260015afa156104a9575f516001600160a01b03811615615d5357905f905f90565b505f906001905f90565b5050505f916003919056fe05f47829691a1f710b0620aedd52749bb09d8abe4bb530d306db920a71b0d7cea26469706673582212209ca9c46147c1489a82a869c349b42f4d493d4ac3ebf3c0ac4fa5a93e0113024c64736f6c634300081e0033",
}

// ChannelHubABI is the input ABI used to generate the binding from.
// Deprecated: Use ChannelHubMetaData.ABI instead.
var ChannelHubABI = ChannelHubMetaData.ABI

// ChannelHubBin is the compiled bytecode used for deploying new contracts.
// Deprecated: Use ChannelHubMetaData.Bin instead.
var ChannelHubBin = ChannelHubMetaData.Bin

// DeployChannelHub deploys a new Ethereum contract, binding an instance of ChannelHub to it.
func DeployChannelHub(auth *bind.TransactOpts, backend bind.ContractBackend, _defaultSigValidator common.Address, _node common.Address) (common.Address, *types.Transaction, *ChannelHub, error) {
	parsed, err := ChannelHubMetaData.GetAbi()
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	if parsed == nil {
		return common.Address{}, nil, nil, errors.New("GetABI returned nil")
	}

	address, tx, contract, err := bind.DeployContract(auth, *parsed, common.FromHex(ChannelHubBin), backend, _defaultSigValidator, _node)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &ChannelHub{ChannelHubCaller: ChannelHubCaller{contract: contract}, ChannelHubTransactor: ChannelHubTransactor{contract: contract}, ChannelHubFilterer: ChannelHubFilterer{contract: contract}}, nil
}

// ChannelHub is an auto generated Go binding around an Ethereum contract.
type ChannelHub struct {
	ChannelHubCaller     // Read-only binding to the contract
	ChannelHubTransactor // Write-only binding to the contract
	ChannelHubFilterer   // Log filterer for contract events
}

// ChannelHubCaller is an auto generated read-only Go binding around an Ethereum contract.
type ChannelHubCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChannelHubTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ChannelHubTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChannelHubFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ChannelHubFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ChannelHubSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ChannelHubSession struct {
	Contract     *ChannelHub       // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// ChannelHubCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ChannelHubCallerSession struct {
	Contract *ChannelHubCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts     // Call options to use throughout this session
}

// ChannelHubTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ChannelHubTransactorSession struct {
	Contract     *ChannelHubTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts     // Transaction auth options to use throughout this session
}

// ChannelHubRaw is an auto generated low-level Go binding around an Ethereum contract.
type ChannelHubRaw struct {
	Contract *ChannelHub // Generic contract binding to access the raw methods on
}

// ChannelHubCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ChannelHubCallerRaw struct {
	Contract *ChannelHubCaller // Generic read-only contract binding to access the raw methods on
}

// ChannelHubTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ChannelHubTransactorRaw struct {
	Contract *ChannelHubTransactor // Generic write-only contract binding to access the raw methods on
}

// NewChannelHub creates a new instance of ChannelHub, bound to a specific deployed contract.
func NewChannelHub(address common.Address, backend bind.ContractBackend) (*ChannelHub, error) {
	contract, err := bindChannelHub(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ChannelHub{ChannelHubCaller: ChannelHubCaller{contract: contract}, ChannelHubTransactor: ChannelHubTransactor{contract: contract}, ChannelHubFilterer: ChannelHubFilterer{contract: contract}}, nil
}

// NewChannelHubCaller creates a new read-only instance of ChannelHub, bound to a specific deployed contract.
func NewChannelHubCaller(address common.Address, caller bind.ContractCaller) (*ChannelHubCaller, error) {
	contract, err := bindChannelHub(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ChannelHubCaller{contract: contract}, nil
}

// NewChannelHubTransactor creates a new write-only instance of ChannelHub, bound to a specific deployed contract.
func NewChannelHubTransactor(address common.Address, transactor bind.ContractTransactor) (*ChannelHubTransactor, error) {
	contract, err := bindChannelHub(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ChannelHubTransactor{contract: contract}, nil
}

// NewChannelHubFilterer creates a new log filterer instance of ChannelHub, bound to a specific deployed contract.
func NewChannelHubFilterer(address common.Address, filterer bind.ContractFilterer) (*ChannelHubFilterer, error) {
	contract, err := bindChannelHub(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ChannelHubFilterer{contract: contract}, nil
}

// bindChannelHub binds a generic wrapper to an already deployed contract.
func bindChannelHub(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ChannelHubABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ChannelHub *ChannelHubRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ChannelHub.Contract.ChannelHubCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ChannelHub *ChannelHubRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChannelHubTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ChannelHub *ChannelHubRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChannelHubTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ChannelHub *ChannelHubCallerRaw) Call(opts *bind.CallOpts, result *[]interface{}, method string, params ...interface{}) error {
	return _ChannelHub.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ChannelHub *ChannelHubTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ChannelHub.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ChannelHub *ChannelHubTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ChannelHub.Contract.contract.Transact(opts, method, params...)
}

// DEFAULTSIGVALIDATOR is a free data retrieval call binding the contract method 0x71a47141.
//
// Solidity: function DEFAULT_SIG_VALIDATOR() view returns(address)
func (_ChannelHub *ChannelHubCaller) DEFAULTSIGVALIDATOR(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "DEFAULT_SIG_VALIDATOR")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// DEFAULTSIGVALIDATOR is a free data retrieval call binding the contract method 0x71a47141.
//
// Solidity: function DEFAULT_SIG_VALIDATOR() view returns(address)
func (_ChannelHub *ChannelHubSession) DEFAULTSIGVALIDATOR() (common.Address, error) {
	return _ChannelHub.Contract.DEFAULTSIGVALIDATOR(&_ChannelHub.CallOpts)
}

// DEFAULTSIGVALIDATOR is a free data retrieval call binding the contract method 0x71a47141.
//
// Solidity: function DEFAULT_SIG_VALIDATOR() view returns(address)
func (_ChannelHub *ChannelHubCallerSession) DEFAULTSIGVALIDATOR() (common.Address, error) {
	return _ChannelHub.Contract.DEFAULTSIGVALIDATOR(&_ChannelHub.CallOpts)
}

// ESCROWDEPOSITUNLOCKDELAY is a free data retrieval call binding the contract method 0x5a0745b4.
//
// Solidity: function ESCROW_DEPOSIT_UNLOCK_DELAY() view returns(uint32)
func (_ChannelHub *ChannelHubCaller) ESCROWDEPOSITUNLOCKDELAY(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "ESCROW_DEPOSIT_UNLOCK_DELAY")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// ESCROWDEPOSITUNLOCKDELAY is a free data retrieval call binding the contract method 0x5a0745b4.
//
// Solidity: function ESCROW_DEPOSIT_UNLOCK_DELAY() view returns(uint32)
func (_ChannelHub *ChannelHubSession) ESCROWDEPOSITUNLOCKDELAY() (uint32, error) {
	return _ChannelHub.Contract.ESCROWDEPOSITUNLOCKDELAY(&_ChannelHub.CallOpts)
}

// ESCROWDEPOSITUNLOCKDELAY is a free data retrieval call binding the contract method 0x5a0745b4.
//
// Solidity: function ESCROW_DEPOSIT_UNLOCK_DELAY() view returns(uint32)
func (_ChannelHub *ChannelHubCallerSession) ESCROWDEPOSITUNLOCKDELAY() (uint32, error) {
	return _ChannelHub.Contract.ESCROWDEPOSITUNLOCKDELAY(&_ChannelHub.CallOpts)
}

// MAXCHALLENGEDURATION is a free data retrieval call binding the contract method 0xb9f4420d.
//
// Solidity: function MAX_CHALLENGE_DURATION() view returns(uint32)
func (_ChannelHub *ChannelHubCaller) MAXCHALLENGEDURATION(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "MAX_CHALLENGE_DURATION")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// MAXCHALLENGEDURATION is a free data retrieval call binding the contract method 0xb9f4420d.
//
// Solidity: function MAX_CHALLENGE_DURATION() view returns(uint32)
func (_ChannelHub *ChannelHubSession) MAXCHALLENGEDURATION() (uint32, error) {
	return _ChannelHub.Contract.MAXCHALLENGEDURATION(&_ChannelHub.CallOpts)
}

// MAXCHALLENGEDURATION is a free data retrieval call binding the contract method 0xb9f4420d.
//
// Solidity: function MAX_CHALLENGE_DURATION() view returns(uint32)
func (_ChannelHub *ChannelHubCallerSession) MAXCHALLENGEDURATION() (uint32, error) {
	return _ChannelHub.Contract.MAXCHALLENGEDURATION(&_ChannelHub.CallOpts)
}

// MAXDEPOSITESCROWSTEPS is a free data retrieval call binding the contract method 0x5ae2accc.
//
// Solidity: function MAX_DEPOSIT_ESCROW_STEPS() view returns(uint32)
func (_ChannelHub *ChannelHubCaller) MAXDEPOSITESCROWSTEPS(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "MAX_DEPOSIT_ESCROW_STEPS")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// MAXDEPOSITESCROWSTEPS is a free data retrieval call binding the contract method 0x5ae2accc.
//
// Solidity: function MAX_DEPOSIT_ESCROW_STEPS() view returns(uint32)
func (_ChannelHub *ChannelHubSession) MAXDEPOSITESCROWSTEPS() (uint32, error) {
	return _ChannelHub.Contract.MAXDEPOSITESCROWSTEPS(&_ChannelHub.CallOpts)
}

// MAXDEPOSITESCROWSTEPS is a free data retrieval call binding the contract method 0x5ae2accc.
//
// Solidity: function MAX_DEPOSIT_ESCROW_STEPS() view returns(uint32)
func (_ChannelHub *ChannelHubCallerSession) MAXDEPOSITESCROWSTEPS() (uint32, error) {
	return _ChannelHub.Contract.MAXDEPOSITESCROWSTEPS(&_ChannelHub.CallOpts)
}

// MINCHALLENGEDURATION is a free data retrieval call binding the contract method 0x94191051.
//
// Solidity: function MIN_CHALLENGE_DURATION() view returns(uint32)
func (_ChannelHub *ChannelHubCaller) MINCHALLENGEDURATION(opts *bind.CallOpts) (uint32, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "MIN_CHALLENGE_DURATION")

	if err != nil {
		return *new(uint32), err
	}

	out0 := *abi.ConvertType(out[0], new(uint32)).(*uint32)

	return out0, err

}

// MINCHALLENGEDURATION is a free data retrieval call binding the contract method 0x94191051.
//
// Solidity: function MIN_CHALLENGE_DURATION() view returns(uint32)
func (_ChannelHub *ChannelHubSession) MINCHALLENGEDURATION() (uint32, error) {
	return _ChannelHub.Contract.MINCHALLENGEDURATION(&_ChannelHub.CallOpts)
}

// MINCHALLENGEDURATION is a free data retrieval call binding the contract method 0x94191051.
//
// Solidity: function MIN_CHALLENGE_DURATION() view returns(uint32)
func (_ChannelHub *ChannelHubCallerSession) MINCHALLENGEDURATION() (uint32, error) {
	return _ChannelHub.Contract.MINCHALLENGEDURATION(&_ChannelHub.CallOpts)
}

// NODE is a free data retrieval call binding the contract method 0x51bfcdbd.
//
// Solidity: function NODE() view returns(address)
func (_ChannelHub *ChannelHubCaller) NODE(opts *bind.CallOpts) (common.Address, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "NODE")

	if err != nil {
		return *new(common.Address), err
	}

	out0 := *abi.ConvertType(out[0], new(common.Address)).(*common.Address)

	return out0, err

}

// NODE is a free data retrieval call binding the contract method 0x51bfcdbd.
//
// Solidity: function NODE() view returns(address)
func (_ChannelHub *ChannelHubSession) NODE() (common.Address, error) {
	return _ChannelHub.Contract.NODE(&_ChannelHub.CallOpts)
}

// NODE is a free data retrieval call binding the contract method 0x51bfcdbd.
//
// Solidity: function NODE() view returns(address)
func (_ChannelHub *ChannelHubCallerSession) NODE() (common.Address, error) {
	return _ChannelHub.Contract.NODE(&_ChannelHub.CallOpts)
}

// TRANSFERGASLIMIT is a free data retrieval call binding the contract method 0x38a66be2.
//
// Solidity: function TRANSFER_GAS_LIMIT() view returns(uint256)
func (_ChannelHub *ChannelHubCaller) TRANSFERGASLIMIT(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "TRANSFER_GAS_LIMIT")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// TRANSFERGASLIMIT is a free data retrieval call binding the contract method 0x38a66be2.
//
// Solidity: function TRANSFER_GAS_LIMIT() view returns(uint256)
func (_ChannelHub *ChannelHubSession) TRANSFERGASLIMIT() (*big.Int, error) {
	return _ChannelHub.Contract.TRANSFERGASLIMIT(&_ChannelHub.CallOpts)
}

// TRANSFERGASLIMIT is a free data retrieval call binding the contract method 0x38a66be2.
//
// Solidity: function TRANSFER_GAS_LIMIT() view returns(uint256)
func (_ChannelHub *ChannelHubCallerSession) TRANSFERGASLIMIT() (*big.Int, error) {
	return _ChannelHub.Contract.TRANSFERGASLIMIT(&_ChannelHub.CallOpts)
}

// VALIDATORACTIVATIONDELAY is a free data retrieval call binding the contract method 0xa4594631.
//
// Solidity: function VALIDATOR_ACTIVATION_DELAY() view returns(uint64)
func (_ChannelHub *ChannelHubCaller) VALIDATORACTIVATIONDELAY(opts *bind.CallOpts) (uint64, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "VALIDATOR_ACTIVATION_DELAY")

	if err != nil {
		return *new(uint64), err
	}

	out0 := *abi.ConvertType(out[0], new(uint64)).(*uint64)

	return out0, err

}

// VALIDATORACTIVATIONDELAY is a free data retrieval call binding the contract method 0xa4594631.
//
// Solidity: function VALIDATOR_ACTIVATION_DELAY() view returns(uint64)
func (_ChannelHub *ChannelHubSession) VALIDATORACTIVATIONDELAY() (uint64, error) {
	return _ChannelHub.Contract.VALIDATORACTIVATIONDELAY(&_ChannelHub.CallOpts)
}

// VALIDATORACTIVATIONDELAY is a free data retrieval call binding the contract method 0xa4594631.
//
// Solidity: function VALIDATOR_ACTIVATION_DELAY() view returns(uint64)
func (_ChannelHub *ChannelHubCallerSession) VALIDATORACTIVATIONDELAY() (uint64, error) {
	return _ChannelHub.Contract.VALIDATORACTIVATIONDELAY(&_ChannelHub.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint8)
func (_ChannelHub *ChannelHubCaller) VERSION(opts *bind.CallOpts) (uint8, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "VERSION")

	if err != nil {
		return *new(uint8), err
	}

	out0 := *abi.ConvertType(out[0], new(uint8)).(*uint8)

	return out0, err

}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint8)
func (_ChannelHub *ChannelHubSession) VERSION() (uint8, error) {
	return _ChannelHub.Contract.VERSION(&_ChannelHub.CallOpts)
}

// VERSION is a free data retrieval call binding the contract method 0xffa1ad74.
//
// Solidity: function VERSION() view returns(uint8)
func (_ChannelHub *ChannelHubCallerSession) VERSION() (uint8, error) {
	return _ChannelHub.Contract.VERSION(&_ChannelHub.CallOpts)
}

// EscrowHead is a free data retrieval call binding the contract method 0x82d3e15d.
//
// Solidity: function escrowHead() view returns(uint256)
func (_ChannelHub *ChannelHubCaller) EscrowHead(opts *bind.CallOpts) (*big.Int, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "escrowHead")

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// EscrowHead is a free data retrieval call binding the contract method 0x82d3e15d.
//
// Solidity: function escrowHead() view returns(uint256)
func (_ChannelHub *ChannelHubSession) EscrowHead() (*big.Int, error) {
	return _ChannelHub.Contract.EscrowHead(&_ChannelHub.CallOpts)
}

// EscrowHead is a free data retrieval call binding the contract method 0x82d3e15d.
//
// Solidity: function escrowHead() view returns(uint256)
func (_ChannelHub *ChannelHubCallerSession) EscrowHead() (*big.Int, error) {
	return _ChannelHub.Contract.EscrowHead(&_ChannelHub.CallOpts)
}

// GetChannelData is a free data retrieval call binding the contract method 0xe617208c.
//
// Solidity: function getChannelData(bytes32 channelId) view returns(uint8 status, (uint32,address,address,uint64,uint256,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) lastState, uint256 challengeExpiry, uint256 lockedFunds)
func (_ChannelHub *ChannelHubCaller) GetChannelData(opts *bind.CallOpts, channelId [32]byte) (struct {
	Status          uint8
	Definition      ChannelDefinition
	LastState       State
	ChallengeExpiry *big.Int
	LockedFunds     *big.Int
}, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getChannelData", channelId)

	outstruct := new(struct {
		Status          uint8
		Definition      ChannelDefinition
		LastState       State
		ChallengeExpiry *big.Int
		LockedFunds     *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Status = *abi.ConvertType(out[0], new(uint8)).(*uint8)
	outstruct.Definition = *abi.ConvertType(out[1], new(ChannelDefinition)).(*ChannelDefinition)
	outstruct.LastState = *abi.ConvertType(out[2], new(State)).(*State)
	outstruct.ChallengeExpiry = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.LockedFunds = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetChannelData is a free data retrieval call binding the contract method 0xe617208c.
//
// Solidity: function getChannelData(bytes32 channelId) view returns(uint8 status, (uint32,address,address,uint64,uint256,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) lastState, uint256 challengeExpiry, uint256 lockedFunds)
func (_ChannelHub *ChannelHubSession) GetChannelData(channelId [32]byte) (struct {
	Status          uint8
	Definition      ChannelDefinition
	LastState       State
	ChallengeExpiry *big.Int
	LockedFunds     *big.Int
}, error) {
	return _ChannelHub.Contract.GetChannelData(&_ChannelHub.CallOpts, channelId)
}

// GetChannelData is a free data retrieval call binding the contract method 0xe617208c.
//
// Solidity: function getChannelData(bytes32 channelId) view returns(uint8 status, (uint32,address,address,uint64,uint256,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) lastState, uint256 challengeExpiry, uint256 lockedFunds)
func (_ChannelHub *ChannelHubCallerSession) GetChannelData(channelId [32]byte) (struct {
	Status          uint8
	Definition      ChannelDefinition
	LastState       State
	ChallengeExpiry *big.Int
	LockedFunds     *big.Int
}, error) {
	return _ChannelHub.Contract.GetChannelData(&_ChannelHub.CallOpts, channelId)
}

// GetChannelIds is a free data retrieval call binding the contract method 0x187576d8.
//
// Solidity: function getChannelIds(address user) view returns(bytes32[])
func (_ChannelHub *ChannelHubCaller) GetChannelIds(opts *bind.CallOpts, user common.Address) ([][32]byte, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getChannelIds", user)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetChannelIds is a free data retrieval call binding the contract method 0x187576d8.
//
// Solidity: function getChannelIds(address user) view returns(bytes32[])
func (_ChannelHub *ChannelHubSession) GetChannelIds(user common.Address) ([][32]byte, error) {
	return _ChannelHub.Contract.GetChannelIds(&_ChannelHub.CallOpts, user)
}

// GetChannelIds is a free data retrieval call binding the contract method 0x187576d8.
//
// Solidity: function getChannelIds(address user) view returns(bytes32[])
func (_ChannelHub *ChannelHubCallerSession) GetChannelIds(user common.Address) ([][32]byte, error) {
	return _ChannelHub.Contract.GetChannelIds(&_ChannelHub.CallOpts, user)
}

// GetEscrowDepositData is a free data retrieval call binding the contract method 0xd888ccae.
//
// Solidity: function getEscrowDepositData(bytes32 escrowId) view returns(bytes32 channelId, uint8 status, uint64 unlockAt, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubCaller) GetEscrowDepositData(opts *bind.CallOpts, escrowId [32]byte) (struct {
	ChannelId       [32]byte
	Status          uint8
	UnlockAt        uint64
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getEscrowDepositData", escrowId)

	outstruct := new(struct {
		ChannelId       [32]byte
		Status          uint8
		UnlockAt        uint64
		ChallengeExpiry uint64
		LockedAmount    *big.Int
		InitState       State
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ChannelId = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Status = *abi.ConvertType(out[1], new(uint8)).(*uint8)
	outstruct.UnlockAt = *abi.ConvertType(out[2], new(uint64)).(*uint64)
	outstruct.ChallengeExpiry = *abi.ConvertType(out[3], new(uint64)).(*uint64)
	outstruct.LockedAmount = *abi.ConvertType(out[4], new(*big.Int)).(**big.Int)
	outstruct.InitState = *abi.ConvertType(out[5], new(State)).(*State)

	return *outstruct, err

}

// GetEscrowDepositData is a free data retrieval call binding the contract method 0xd888ccae.
//
// Solidity: function getEscrowDepositData(bytes32 escrowId) view returns(bytes32 channelId, uint8 status, uint64 unlockAt, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubSession) GetEscrowDepositData(escrowId [32]byte) (struct {
	ChannelId       [32]byte
	Status          uint8
	UnlockAt        uint64
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	return _ChannelHub.Contract.GetEscrowDepositData(&_ChannelHub.CallOpts, escrowId)
}

// GetEscrowDepositData is a free data retrieval call binding the contract method 0xd888ccae.
//
// Solidity: function getEscrowDepositData(bytes32 escrowId) view returns(bytes32 channelId, uint8 status, uint64 unlockAt, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubCallerSession) GetEscrowDepositData(escrowId [32]byte) (struct {
	ChannelId       [32]byte
	Status          uint8
	UnlockAt        uint64
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	return _ChannelHub.Contract.GetEscrowDepositData(&_ChannelHub.CallOpts, escrowId)
}

// GetEscrowDepositIds is a free data retrieval call binding the contract method 0x5b9acbf9.
//
// Solidity: function getEscrowDepositIds(uint256 page, uint256 pageSize) view returns(bytes32[] ids)
func (_ChannelHub *ChannelHubCaller) GetEscrowDepositIds(opts *bind.CallOpts, page *big.Int, pageSize *big.Int) ([][32]byte, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getEscrowDepositIds", page, pageSize)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetEscrowDepositIds is a free data retrieval call binding the contract method 0x5b9acbf9.
//
// Solidity: function getEscrowDepositIds(uint256 page, uint256 pageSize) view returns(bytes32[] ids)
func (_ChannelHub *ChannelHubSession) GetEscrowDepositIds(page *big.Int, pageSize *big.Int) ([][32]byte, error) {
	return _ChannelHub.Contract.GetEscrowDepositIds(&_ChannelHub.CallOpts, page, pageSize)
}

// GetEscrowDepositIds is a free data retrieval call binding the contract method 0x5b9acbf9.
//
// Solidity: function getEscrowDepositIds(uint256 page, uint256 pageSize) view returns(bytes32[] ids)
func (_ChannelHub *ChannelHubCallerSession) GetEscrowDepositIds(page *big.Int, pageSize *big.Int) ([][32]byte, error) {
	return _ChannelHub.Contract.GetEscrowDepositIds(&_ChannelHub.CallOpts, page, pageSize)
}

// GetEscrowWithdrawalData is a free data retrieval call binding the contract method 0xdd73d494.
//
// Solidity: function getEscrowWithdrawalData(bytes32 escrowId) view returns(bytes32 channelId, uint8 status, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubCaller) GetEscrowWithdrawalData(opts *bind.CallOpts, escrowId [32]byte) (struct {
	ChannelId       [32]byte
	Status          uint8
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getEscrowWithdrawalData", escrowId)

	outstruct := new(struct {
		ChannelId       [32]byte
		Status          uint8
		ChallengeExpiry uint64
		LockedAmount    *big.Int
		InitState       State
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.ChannelId = *abi.ConvertType(out[0], new([32]byte)).(*[32]byte)
	outstruct.Status = *abi.ConvertType(out[1], new(uint8)).(*uint8)
	outstruct.ChallengeExpiry = *abi.ConvertType(out[2], new(uint64)).(*uint64)
	outstruct.LockedAmount = *abi.ConvertType(out[3], new(*big.Int)).(**big.Int)
	outstruct.InitState = *abi.ConvertType(out[4], new(State)).(*State)

	return *outstruct, err

}

// GetEscrowWithdrawalData is a free data retrieval call binding the contract method 0xdd73d494.
//
// Solidity: function getEscrowWithdrawalData(bytes32 escrowId) view returns(bytes32 channelId, uint8 status, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubSession) GetEscrowWithdrawalData(escrowId [32]byte) (struct {
	ChannelId       [32]byte
	Status          uint8
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	return _ChannelHub.Contract.GetEscrowWithdrawalData(&_ChannelHub.CallOpts, escrowId)
}

// GetEscrowWithdrawalData is a free data retrieval call binding the contract method 0xdd73d494.
//
// Solidity: function getEscrowWithdrawalData(bytes32 escrowId) view returns(bytes32 channelId, uint8 status, uint64 challengeExpiry, uint256 lockedAmount, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState)
func (_ChannelHub *ChannelHubCallerSession) GetEscrowWithdrawalData(escrowId [32]byte) (struct {
	ChannelId       [32]byte
	Status          uint8
	ChallengeExpiry uint64
	LockedAmount    *big.Int
	InitState       State
}, error) {
	return _ChannelHub.Contract.GetEscrowWithdrawalData(&_ChannelHub.CallOpts, escrowId)
}

// GetNodeBalance is a free data retrieval call binding the contract method 0x07f241ce.
//
// Solidity: function getNodeBalance(address token) view returns(uint256)
func (_ChannelHub *ChannelHubCaller) GetNodeBalance(opts *bind.CallOpts, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getNodeBalance", token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetNodeBalance is a free data retrieval call binding the contract method 0x07f241ce.
//
// Solidity: function getNodeBalance(address token) view returns(uint256)
func (_ChannelHub *ChannelHubSession) GetNodeBalance(token common.Address) (*big.Int, error) {
	return _ChannelHub.Contract.GetNodeBalance(&_ChannelHub.CallOpts, token)
}

// GetNodeBalance is a free data retrieval call binding the contract method 0x07f241ce.
//
// Solidity: function getNodeBalance(address token) view returns(uint256)
func (_ChannelHub *ChannelHubCallerSession) GetNodeBalance(token common.Address) (*big.Int, error) {
	return _ChannelHub.Contract.GetNodeBalance(&_ChannelHub.CallOpts, token)
}

// GetNodeValidator is a free data retrieval call binding the contract method 0x3c684f92.
//
// Solidity: function getNodeValidator(uint8 validatorId) view returns(address validator, uint64 registeredAt)
func (_ChannelHub *ChannelHubCaller) GetNodeValidator(opts *bind.CallOpts, validatorId uint8) (struct {
	Validator    common.Address
	RegisteredAt uint64
}, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getNodeValidator", validatorId)

	outstruct := new(struct {
		Validator    common.Address
		RegisteredAt uint64
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Validator = *abi.ConvertType(out[0], new(common.Address)).(*common.Address)
	outstruct.RegisteredAt = *abi.ConvertType(out[1], new(uint64)).(*uint64)

	return *outstruct, err

}

// GetNodeValidator is a free data retrieval call binding the contract method 0x3c684f92.
//
// Solidity: function getNodeValidator(uint8 validatorId) view returns(address validator, uint64 registeredAt)
func (_ChannelHub *ChannelHubSession) GetNodeValidator(validatorId uint8) (struct {
	Validator    common.Address
	RegisteredAt uint64
}, error) {
	return _ChannelHub.Contract.GetNodeValidator(&_ChannelHub.CallOpts, validatorId)
}

// GetNodeValidator is a free data retrieval call binding the contract method 0x3c684f92.
//
// Solidity: function getNodeValidator(uint8 validatorId) view returns(address validator, uint64 registeredAt)
func (_ChannelHub *ChannelHubCallerSession) GetNodeValidator(validatorId uint8) (struct {
	Validator    common.Address
	RegisteredAt uint64
}, error) {
	return _ChannelHub.Contract.GetNodeValidator(&_ChannelHub.CallOpts, validatorId)
}

// GetOpenChannels is a free data retrieval call binding the contract method 0x6898234b.
//
// Solidity: function getOpenChannels(address user) view returns(bytes32[] openChannels)
func (_ChannelHub *ChannelHubCaller) GetOpenChannels(opts *bind.CallOpts, user common.Address) ([][32]byte, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getOpenChannels", user)

	if err != nil {
		return *new([][32]byte), err
	}

	out0 := *abi.ConvertType(out[0], new([][32]byte)).(*[][32]byte)

	return out0, err

}

// GetOpenChannels is a free data retrieval call binding the contract method 0x6898234b.
//
// Solidity: function getOpenChannels(address user) view returns(bytes32[] openChannels)
func (_ChannelHub *ChannelHubSession) GetOpenChannels(user common.Address) ([][32]byte, error) {
	return _ChannelHub.Contract.GetOpenChannels(&_ChannelHub.CallOpts, user)
}

// GetOpenChannels is a free data retrieval call binding the contract method 0x6898234b.
//
// Solidity: function getOpenChannels(address user) view returns(bytes32[] openChannels)
func (_ChannelHub *ChannelHubCallerSession) GetOpenChannels(user common.Address) ([][32]byte, error) {
	return _ChannelHub.Contract.GetOpenChannels(&_ChannelHub.CallOpts, user)
}

// GetReclaimBalance is a free data retrieval call binding the contract method 0x735181f0.
//
// Solidity: function getReclaimBalance(address account, address token) view returns(uint256)
func (_ChannelHub *ChannelHubCaller) GetReclaimBalance(opts *bind.CallOpts, account common.Address, token common.Address) (*big.Int, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getReclaimBalance", account, token)

	if err != nil {
		return *new(*big.Int), err
	}

	out0 := *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)

	return out0, err

}

// GetReclaimBalance is a free data retrieval call binding the contract method 0x735181f0.
//
// Solidity: function getReclaimBalance(address account, address token) view returns(uint256)
func (_ChannelHub *ChannelHubSession) GetReclaimBalance(account common.Address, token common.Address) (*big.Int, error) {
	return _ChannelHub.Contract.GetReclaimBalance(&_ChannelHub.CallOpts, account, token)
}

// GetReclaimBalance is a free data retrieval call binding the contract method 0x735181f0.
//
// Solidity: function getReclaimBalance(address account, address token) view returns(uint256)
func (_ChannelHub *ChannelHubCallerSession) GetReclaimBalance(account common.Address, token common.Address) (*big.Int, error) {
	return _ChannelHub.Contract.GetReclaimBalance(&_ChannelHub.CallOpts, account, token)
}

// GetUnlockableEscrowDepositStats is a free data retrieval call binding the contract method 0xc9408398.
//
// Solidity: function getUnlockableEscrowDepositStats() view returns(uint256 count, uint256 totalAmount)
func (_ChannelHub *ChannelHubCaller) GetUnlockableEscrowDepositStats(opts *bind.CallOpts) (struct {
	Count       *big.Int
	TotalAmount *big.Int
}, error) {
	var out []interface{}
	err := _ChannelHub.contract.Call(opts, &out, "getUnlockableEscrowDepositStats")

	outstruct := new(struct {
		Count       *big.Int
		TotalAmount *big.Int
	})
	if err != nil {
		return *outstruct, err
	}

	outstruct.Count = *abi.ConvertType(out[0], new(*big.Int)).(**big.Int)
	outstruct.TotalAmount = *abi.ConvertType(out[1], new(*big.Int)).(**big.Int)

	return *outstruct, err

}

// GetUnlockableEscrowDepositStats is a free data retrieval call binding the contract method 0xc9408398.
//
// Solidity: function getUnlockableEscrowDepositStats() view returns(uint256 count, uint256 totalAmount)
func (_ChannelHub *ChannelHubSession) GetUnlockableEscrowDepositStats() (struct {
	Count       *big.Int
	TotalAmount *big.Int
}, error) {
	return _ChannelHub.Contract.GetUnlockableEscrowDepositStats(&_ChannelHub.CallOpts)
}

// GetUnlockableEscrowDepositStats is a free data retrieval call binding the contract method 0xc9408398.
//
// Solidity: function getUnlockableEscrowDepositStats() view returns(uint256 count, uint256 totalAmount)
func (_ChannelHub *ChannelHubCallerSession) GetUnlockableEscrowDepositStats() (struct {
	Count       *big.Int
	TotalAmount *big.Int
}, error) {
	return _ChannelHub.Contract.GetUnlockableEscrowDepositStats(&_ChannelHub.CallOpts)
}

// ChallengeChannel is a paid mutator transaction binding the contract method 0xb25a1d38.
//
// Solidity: function challengeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, bytes challengerSig, uint8 challengerIdx) payable returns()
func (_ChannelHub *ChannelHubTransactor) ChallengeChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "challengeChannel", channelId, candidate, challengerSig, challengerIdx)
}

// ChallengeChannel is a paid mutator transaction binding the contract method 0xb25a1d38.
//
// Solidity: function challengeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, bytes challengerSig, uint8 challengerIdx) payable returns()
func (_ChannelHub *ChannelHubSession) ChallengeChannel(channelId [32]byte, candidate State, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeChannel(&_ChannelHub.TransactOpts, channelId, candidate, challengerSig, challengerIdx)
}

// ChallengeChannel is a paid mutator transaction binding the contract method 0xb25a1d38.
//
// Solidity: function challengeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, bytes challengerSig, uint8 challengerIdx) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) ChallengeChannel(channelId [32]byte, candidate State, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeChannel(&_ChannelHub.TransactOpts, channelId, candidate, challengerSig, challengerIdx)
}

// ChallengeEscrowDeposit is a paid mutator transaction binding the contract method 0x16b390b1.
//
// Solidity: function challengeEscrowDeposit(bytes32 escrowId, bytes challengerSig, uint8 challengerIdx) returns()
func (_ChannelHub *ChannelHubTransactor) ChallengeEscrowDeposit(opts *bind.TransactOpts, escrowId [32]byte, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "challengeEscrowDeposit", escrowId, challengerSig, challengerIdx)
}

// ChallengeEscrowDeposit is a paid mutator transaction binding the contract method 0x16b390b1.
//
// Solidity: function challengeEscrowDeposit(bytes32 escrowId, bytes challengerSig, uint8 challengerIdx) returns()
func (_ChannelHub *ChannelHubSession) ChallengeEscrowDeposit(escrowId [32]byte, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeEscrowDeposit(&_ChannelHub.TransactOpts, escrowId, challengerSig, challengerIdx)
}

// ChallengeEscrowDeposit is a paid mutator transaction binding the contract method 0x16b390b1.
//
// Solidity: function challengeEscrowDeposit(bytes32 escrowId, bytes challengerSig, uint8 challengerIdx) returns()
func (_ChannelHub *ChannelHubTransactorSession) ChallengeEscrowDeposit(escrowId [32]byte, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeEscrowDeposit(&_ChannelHub.TransactOpts, escrowId, challengerSig, challengerIdx)
}

// ChallengeEscrowWithdrawal is a paid mutator transaction binding the contract method 0x8d0b12a5.
//
// Solidity: function challengeEscrowWithdrawal(bytes32 escrowId, bytes challengerSig, uint8 challengerIdx) returns()
func (_ChannelHub *ChannelHubTransactor) ChallengeEscrowWithdrawal(opts *bind.TransactOpts, escrowId [32]byte, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "challengeEscrowWithdrawal", escrowId, challengerSig, challengerIdx)
}

// ChallengeEscrowWithdrawal is a paid mutator transaction binding the contract method 0x8d0b12a5.
//
// Solidity: function challengeEscrowWithdrawal(bytes32 escrowId, bytes challengerSig, uint8 challengerIdx) returns()
func (_ChannelHub *ChannelHubSession) ChallengeEscrowWithdrawal(escrowId [32]byte, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeEscrowWithdrawal(&_ChannelHub.TransactOpts, escrowId, challengerSig, challengerIdx)
}

// ChallengeEscrowWithdrawal is a paid mutator transaction binding the contract method 0x8d0b12a5.
//
// Solidity: function challengeEscrowWithdrawal(bytes32 escrowId, bytes challengerSig, uint8 challengerIdx) returns()
func (_ChannelHub *ChannelHubTransactorSession) ChallengeEscrowWithdrawal(escrowId [32]byte, challengerSig []byte, challengerIdx uint8) (*types.Transaction, error) {
	return _ChannelHub.Contract.ChallengeEscrowWithdrawal(&_ChannelHub.TransactOpts, escrowId, challengerSig, challengerIdx)
}

// CheckpointChannel is a paid mutator transaction binding the contract method 0x9691b468.
//
// Solidity: function checkpointChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) CheckpointChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "checkpointChannel", channelId, candidate)
}

// CheckpointChannel is a paid mutator transaction binding the contract method 0x9691b468.
//
// Solidity: function checkpointChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) CheckpointChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CheckpointChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// CheckpointChannel is a paid mutator transaction binding the contract method 0x9691b468.
//
// Solidity: function checkpointChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) CheckpointChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CheckpointChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// ClaimFunds is a paid mutator transaction binding the contract method 0xf766f8d6.
//
// Solidity: function claimFunds(address token, address destination) returns()
func (_ChannelHub *ChannelHubTransactor) ClaimFunds(opts *bind.TransactOpts, token common.Address, destination common.Address) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "claimFunds", token, destination)
}

// ClaimFunds is a paid mutator transaction binding the contract method 0xf766f8d6.
//
// Solidity: function claimFunds(address token, address destination) returns()
func (_ChannelHub *ChannelHubSession) ClaimFunds(token common.Address, destination common.Address) (*types.Transaction, error) {
	return _ChannelHub.Contract.ClaimFunds(&_ChannelHub.TransactOpts, token, destination)
}

// ClaimFunds is a paid mutator transaction binding the contract method 0xf766f8d6.
//
// Solidity: function claimFunds(address token, address destination) returns()
func (_ChannelHub *ChannelHubTransactorSession) ClaimFunds(token common.Address, destination common.Address) (*types.Transaction, error) {
	return _ChannelHub.Contract.ClaimFunds(&_ChannelHub.TransactOpts, token, destination)
}

// CloseChannel is a paid mutator transaction binding the contract method 0x5dc46a74.
//
// Solidity: function closeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) CloseChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "closeChannel", channelId, candidate)
}

// CloseChannel is a paid mutator transaction binding the contract method 0x5dc46a74.
//
// Solidity: function closeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) CloseChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CloseChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// CloseChannel is a paid mutator transaction binding the contract method 0x5dc46a74.
//
// Solidity: function closeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) CloseChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CloseChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x41b660ef.
//
// Solidity: function createChannel((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState) payable returns()
func (_ChannelHub *ChannelHubTransactor) CreateChannel(opts *bind.TransactOpts, def ChannelDefinition, initState State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "createChannel", def, initState)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x41b660ef.
//
// Solidity: function createChannel((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState) payable returns()
func (_ChannelHub *ChannelHubSession) CreateChannel(def ChannelDefinition, initState State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CreateChannel(&_ChannelHub.TransactOpts, def, initState)
}

// CreateChannel is a paid mutator transaction binding the contract method 0x41b660ef.
//
// Solidity: function createChannel((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initState) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) CreateChannel(def ChannelDefinition, initState State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CreateChannel(&_ChannelHub.TransactOpts, def, initState)
}

// DepositToChannel is a paid mutator transaction binding the contract method 0xf4ac51f5.
//
// Solidity: function depositToChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactor) DepositToChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "depositToChannel", channelId, candidate)
}

// DepositToChannel is a paid mutator transaction binding the contract method 0xf4ac51f5.
//
// Solidity: function depositToChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubSession) DepositToChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.DepositToChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// DepositToChannel is a paid mutator transaction binding the contract method 0xf4ac51f5.
//
// Solidity: function depositToChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) DepositToChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.DepositToChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// DepositToNode is a paid mutator transaction binding the contract method 0xb65b78d1.
//
// Solidity: function depositToNode(address token, uint256 amount) payable returns()
func (_ChannelHub *ChannelHubTransactor) DepositToNode(opts *bind.TransactOpts, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "depositToNode", token, amount)
}

// DepositToNode is a paid mutator transaction binding the contract method 0xb65b78d1.
//
// Solidity: function depositToNode(address token, uint256 amount) payable returns()
func (_ChannelHub *ChannelHubSession) DepositToNode(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ChannelHub.Contract.DepositToNode(&_ChannelHub.TransactOpts, token, amount)
}

// DepositToNode is a paid mutator transaction binding the contract method 0xb65b78d1.
//
// Solidity: function depositToNode(address token, uint256 amount) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) DepositToNode(token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ChannelHub.Contract.DepositToNode(&_ChannelHub.TransactOpts, token, amount)
}

// FinalizeEscrowDeposit is a paid mutator transaction binding the contract method 0xff5bc09e.
//
// Solidity: function finalizeEscrowDeposit(bytes32 channelId, bytes32 escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) FinalizeEscrowDeposit(opts *bind.TransactOpts, channelId [32]byte, escrowId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "finalizeEscrowDeposit", channelId, escrowId, candidate)
}

// FinalizeEscrowDeposit is a paid mutator transaction binding the contract method 0xff5bc09e.
//
// Solidity: function finalizeEscrowDeposit(bytes32 channelId, bytes32 escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) FinalizeEscrowDeposit(channelId [32]byte, escrowId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.FinalizeEscrowDeposit(&_ChannelHub.TransactOpts, channelId, escrowId, candidate)
}

// FinalizeEscrowDeposit is a paid mutator transaction binding the contract method 0xff5bc09e.
//
// Solidity: function finalizeEscrowDeposit(bytes32 channelId, bytes32 escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) FinalizeEscrowDeposit(channelId [32]byte, escrowId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.FinalizeEscrowDeposit(&_ChannelHub.TransactOpts, channelId, escrowId, candidate)
}

// FinalizeEscrowWithdrawal is a paid mutator transaction binding the contract method 0x6840dbd2.
//
// Solidity: function finalizeEscrowWithdrawal(bytes32 channelId, bytes32 escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) FinalizeEscrowWithdrawal(opts *bind.TransactOpts, channelId [32]byte, escrowId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "finalizeEscrowWithdrawal", channelId, escrowId, candidate)
}

// FinalizeEscrowWithdrawal is a paid mutator transaction binding the contract method 0x6840dbd2.
//
// Solidity: function finalizeEscrowWithdrawal(bytes32 channelId, bytes32 escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) FinalizeEscrowWithdrawal(channelId [32]byte, escrowId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.FinalizeEscrowWithdrawal(&_ChannelHub.TransactOpts, channelId, escrowId, candidate)
}

// FinalizeEscrowWithdrawal is a paid mutator transaction binding the contract method 0x6840dbd2.
//
// Solidity: function finalizeEscrowWithdrawal(bytes32 channelId, bytes32 escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) FinalizeEscrowWithdrawal(channelId [32]byte, escrowId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.FinalizeEscrowWithdrawal(&_ChannelHub.TransactOpts, channelId, escrowId, candidate)
}

// FinalizeMigration is a paid mutator transaction binding the contract method 0x53269198.
//
// Solidity: function finalizeMigration(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) FinalizeMigration(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "finalizeMigration", channelId, candidate)
}

// FinalizeMigration is a paid mutator transaction binding the contract method 0x53269198.
//
// Solidity: function finalizeMigration(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) FinalizeMigration(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.FinalizeMigration(&_ChannelHub.TransactOpts, channelId, candidate)
}

// FinalizeMigration is a paid mutator transaction binding the contract method 0x53269198.
//
// Solidity: function finalizeMigration(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) FinalizeMigration(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.FinalizeMigration(&_ChannelHub.TransactOpts, channelId, candidate)
}

// InitiateEscrowDeposit is a paid mutator transaction binding the contract method 0x47de477a.
//
// Solidity: function initiateEscrowDeposit((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactor) InitiateEscrowDeposit(opts *bind.TransactOpts, def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "initiateEscrowDeposit", def, candidate)
}

// InitiateEscrowDeposit is a paid mutator transaction binding the contract method 0x47de477a.
//
// Solidity: function initiateEscrowDeposit((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubSession) InitiateEscrowDeposit(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateEscrowDeposit(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateEscrowDeposit is a paid mutator transaction binding the contract method 0x47de477a.
//
// Solidity: function initiateEscrowDeposit((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactorSession) InitiateEscrowDeposit(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateEscrowDeposit(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateEscrowWithdrawal is a paid mutator transaction binding the contract method 0xa5c82680.
//
// Solidity: function initiateEscrowWithdrawal((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) InitiateEscrowWithdrawal(opts *bind.TransactOpts, def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "initiateEscrowWithdrawal", def, candidate)
}

// InitiateEscrowWithdrawal is a paid mutator transaction binding the contract method 0xa5c82680.
//
// Solidity: function initiateEscrowWithdrawal((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) InitiateEscrowWithdrawal(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateEscrowWithdrawal(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateEscrowWithdrawal is a paid mutator transaction binding the contract method 0xa5c82680.
//
// Solidity: function initiateEscrowWithdrawal((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) InitiateEscrowWithdrawal(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateEscrowWithdrawal(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateMigration is a paid mutator transaction binding the contract method 0xdc23f29e.
//
// Solidity: function initiateMigration((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) InitiateMigration(opts *bind.TransactOpts, def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "initiateMigration", def, candidate)
}

// InitiateMigration is a paid mutator transaction binding the contract method 0xdc23f29e.
//
// Solidity: function initiateMigration((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) InitiateMigration(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateMigration(&_ChannelHub.TransactOpts, def, candidate)
}

// InitiateMigration is a paid mutator transaction binding the contract method 0xdc23f29e.
//
// Solidity: function initiateMigration((uint32,address,address,uint64,uint256,bytes32) def, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) InitiateMigration(def ChannelDefinition, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.InitiateMigration(&_ChannelHub.TransactOpts, def, candidate)
}

// PurgeEscrowDeposits is a paid mutator transaction binding the contract method 0x3115f630.
//
// Solidity: function purgeEscrowDeposits(uint256 maxSteps) returns()
func (_ChannelHub *ChannelHubTransactor) PurgeEscrowDeposits(opts *bind.TransactOpts, maxSteps *big.Int) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "purgeEscrowDeposits", maxSteps)
}

// PurgeEscrowDeposits is a paid mutator transaction binding the contract method 0x3115f630.
//
// Solidity: function purgeEscrowDeposits(uint256 maxSteps) returns()
func (_ChannelHub *ChannelHubSession) PurgeEscrowDeposits(maxSteps *big.Int) (*types.Transaction, error) {
	return _ChannelHub.Contract.PurgeEscrowDeposits(&_ChannelHub.TransactOpts, maxSteps)
}

// PurgeEscrowDeposits is a paid mutator transaction binding the contract method 0x3115f630.
//
// Solidity: function purgeEscrowDeposits(uint256 maxSteps) returns()
func (_ChannelHub *ChannelHubTransactorSession) PurgeEscrowDeposits(maxSteps *big.Int) (*types.Transaction, error) {
	return _ChannelHub.Contract.PurgeEscrowDeposits(&_ChannelHub.TransactOpts, maxSteps)
}

// RegisterNodeValidator is a paid mutator transaction binding the contract method 0x8e31c735.
//
// Solidity: function registerNodeValidator(uint8 validatorId, address validator, bytes signature) returns()
func (_ChannelHub *ChannelHubTransactor) RegisterNodeValidator(opts *bind.TransactOpts, validatorId uint8, validator common.Address, signature []byte) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "registerNodeValidator", validatorId, validator, signature)
}

// RegisterNodeValidator is a paid mutator transaction binding the contract method 0x8e31c735.
//
// Solidity: function registerNodeValidator(uint8 validatorId, address validator, bytes signature) returns()
func (_ChannelHub *ChannelHubSession) RegisterNodeValidator(validatorId uint8, validator common.Address, signature []byte) (*types.Transaction, error) {
	return _ChannelHub.Contract.RegisterNodeValidator(&_ChannelHub.TransactOpts, validatorId, validator, signature)
}

// RegisterNodeValidator is a paid mutator transaction binding the contract method 0x8e31c735.
//
// Solidity: function registerNodeValidator(uint8 validatorId, address validator, bytes signature) returns()
func (_ChannelHub *ChannelHubTransactorSession) RegisterNodeValidator(validatorId uint8, validator common.Address, signature []byte) (*types.Transaction, error) {
	return _ChannelHub.Contract.RegisterNodeValidator(&_ChannelHub.TransactOpts, validatorId, validator, signature)
}

// WithdrawFromChannel is a paid mutator transaction binding the contract method 0xc74a2d10.
//
// Solidity: function withdrawFromChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactor) WithdrawFromChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "withdrawFromChannel", channelId, candidate)
}

// WithdrawFromChannel is a paid mutator transaction binding the contract method 0xc74a2d10.
//
// Solidity: function withdrawFromChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubSession) WithdrawFromChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.WithdrawFromChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// WithdrawFromChannel is a paid mutator transaction binding the contract method 0xc74a2d10.
//
// Solidity: function withdrawFromChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) returns()
func (_ChannelHub *ChannelHubTransactorSession) WithdrawFromChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.WithdrawFromChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// WithdrawFromNode is a paid mutator transaction binding the contract method 0xd91a1283.
//
// Solidity: function withdrawFromNode(address to, address token, uint256 amount) returns()
func (_ChannelHub *ChannelHubTransactor) WithdrawFromNode(opts *bind.TransactOpts, to common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "withdrawFromNode", to, token, amount)
}

// WithdrawFromNode is a paid mutator transaction binding the contract method 0xd91a1283.
//
// Solidity: function withdrawFromNode(address to, address token, uint256 amount) returns()
func (_ChannelHub *ChannelHubSession) WithdrawFromNode(to common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ChannelHub.Contract.WithdrawFromNode(&_ChannelHub.TransactOpts, to, token, amount)
}

// WithdrawFromNode is a paid mutator transaction binding the contract method 0xd91a1283.
//
// Solidity: function withdrawFromNode(address to, address token, uint256 amount) returns()
func (_ChannelHub *ChannelHubTransactorSession) WithdrawFromNode(to common.Address, token common.Address, amount *big.Int) (*types.Transaction, error) {
	return _ChannelHub.Contract.WithdrawFromNode(&_ChannelHub.TransactOpts, to, token, amount)
}

// ChannelHubChannelChallengedIterator is returned from FilterChannelChallenged and is used to iterate over the raw logs and unpacked data for ChannelChallenged events raised by the ChannelHub contract.
type ChannelHubChannelChallengedIterator struct {
	Event *ChannelHubChannelChallenged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubChannelChallengedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubChannelChallenged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubChannelChallenged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubChannelChallengedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubChannelChallengedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubChannelChallenged represents a ChannelChallenged event raised by the ChannelHub contract.
type ChannelHubChannelChallenged struct {
	ChannelId         [32]byte
	Candidate         State
	ChallengeExpireAt uint64
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterChannelChallenged is a free log retrieval operation binding the contract event 0x07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a.
//
// Solidity: event ChannelChallenged(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) FilterChannelChallenged(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubChannelChallengedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ChannelChallenged", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubChannelChallengedIterator{contract: _ChannelHub.contract, event: "ChannelChallenged", logs: logs, sub: sub}, nil
}

// WatchChannelChallenged is a free log subscription operation binding the contract event 0x07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a.
//
// Solidity: event ChannelChallenged(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) WatchChannelChallenged(opts *bind.WatchOpts, sink chan<- *ChannelHubChannelChallenged, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ChannelChallenged", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubChannelChallenged)
				if err := _ChannelHub.contract.UnpackLog(event, "ChannelChallenged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChannelChallenged is a log parse operation binding the contract event 0x07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a.
//
// Solidity: event ChannelChallenged(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) ParseChannelChallenged(log types.Log) (*ChannelHubChannelChallenged, error) {
	event := new(ChannelHubChannelChallenged)
	if err := _ChannelHub.contract.UnpackLog(event, "ChannelChallenged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubChannelCheckpointedIterator is returned from FilterChannelCheckpointed and is used to iterate over the raw logs and unpacked data for ChannelCheckpointed events raised by the ChannelHub contract.
type ChannelHubChannelCheckpointedIterator struct {
	Event *ChannelHubChannelCheckpointed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubChannelCheckpointedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubChannelCheckpointed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubChannelCheckpointed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubChannelCheckpointedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubChannelCheckpointedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubChannelCheckpointed represents a ChannelCheckpointed event raised by the ChannelHub contract.
type ChannelHubChannelCheckpointed struct {
	ChannelId [32]byte
	Candidate State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterChannelCheckpointed is a free log retrieval operation binding the contract event 0x567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc.
//
// Solidity: event ChannelCheckpointed(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) FilterChannelCheckpointed(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubChannelCheckpointedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ChannelCheckpointed", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubChannelCheckpointedIterator{contract: _ChannelHub.contract, event: "ChannelCheckpointed", logs: logs, sub: sub}, nil
}

// WatchChannelCheckpointed is a free log subscription operation binding the contract event 0x567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc.
//
// Solidity: event ChannelCheckpointed(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) WatchChannelCheckpointed(opts *bind.WatchOpts, sink chan<- *ChannelHubChannelCheckpointed, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ChannelCheckpointed", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubChannelCheckpointed)
				if err := _ChannelHub.contract.UnpackLog(event, "ChannelCheckpointed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChannelCheckpointed is a log parse operation binding the contract event 0x567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc.
//
// Solidity: event ChannelCheckpointed(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) ParseChannelCheckpointed(log types.Log) (*ChannelHubChannelCheckpointed, error) {
	event := new(ChannelHubChannelCheckpointed)
	if err := _ChannelHub.contract.UnpackLog(event, "ChannelCheckpointed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubChannelClosedIterator is returned from FilterChannelClosed and is used to iterate over the raw logs and unpacked data for ChannelClosed events raised by the ChannelHub contract.
type ChannelHubChannelClosedIterator struct {
	Event *ChannelHubChannelClosed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubChannelClosedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubChannelClosed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubChannelClosed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubChannelClosedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubChannelClosedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubChannelClosed represents a ChannelClosed event raised by the ChannelHub contract.
type ChannelHubChannelClosed struct {
	ChannelId  [32]byte
	FinalState State
	Raw        types.Log // Blockchain specific contextual infos
}

// FilterChannelClosed is a free log retrieval operation binding the contract event 0x04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a8.
//
// Solidity: event ChannelClosed(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) finalState)
func (_ChannelHub *ChannelHubFilterer) FilterChannelClosed(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubChannelClosedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ChannelClosed", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubChannelClosedIterator{contract: _ChannelHub.contract, event: "ChannelClosed", logs: logs, sub: sub}, nil
}

// WatchChannelClosed is a free log subscription operation binding the contract event 0x04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a8.
//
// Solidity: event ChannelClosed(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) finalState)
func (_ChannelHub *ChannelHubFilterer) WatchChannelClosed(opts *bind.WatchOpts, sink chan<- *ChannelHubChannelClosed, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ChannelClosed", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubChannelClosed)
				if err := _ChannelHub.contract.UnpackLog(event, "ChannelClosed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChannelClosed is a log parse operation binding the contract event 0x04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a8.
//
// Solidity: event ChannelClosed(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) finalState)
func (_ChannelHub *ChannelHubFilterer) ParseChannelClosed(log types.Log) (*ChannelHubChannelClosed, error) {
	event := new(ChannelHubChannelClosed)
	if err := _ChannelHub.contract.UnpackLog(event, "ChannelClosed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubChannelCreatedIterator is returned from FilterChannelCreated and is used to iterate over the raw logs and unpacked data for ChannelCreated events raised by the ChannelHub contract.
type ChannelHubChannelCreatedIterator struct {
	Event *ChannelHubChannelCreated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubChannelCreatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubChannelCreated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubChannelCreated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubChannelCreatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubChannelCreatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubChannelCreated represents a ChannelCreated event raised by the ChannelHub contract.
type ChannelHubChannelCreated struct {
	ChannelId    [32]byte
	User         common.Address
	Definition   ChannelDefinition
	InitialState State
	Raw          types.Log // Blockchain specific contextual infos
}

// FilterChannelCreated is a free log retrieval operation binding the contract event 0xb0d099feaab5034d04a1c610e86b8832343f2127b3c667b705834dafdf96e9e4.
//
// Solidity: event ChannelCreated(bytes32 indexed channelId, address indexed user, (uint32,address,address,uint64,uint256,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initialState)
func (_ChannelHub *ChannelHubFilterer) FilterChannelCreated(opts *bind.FilterOpts, channelId [][32]byte, user []common.Address) (*ChannelHubChannelCreatedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ChannelCreated", channelIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubChannelCreatedIterator{contract: _ChannelHub.contract, event: "ChannelCreated", logs: logs, sub: sub}, nil
}

// WatchChannelCreated is a free log subscription operation binding the contract event 0xb0d099feaab5034d04a1c610e86b8832343f2127b3c667b705834dafdf96e9e4.
//
// Solidity: event ChannelCreated(bytes32 indexed channelId, address indexed user, (uint32,address,address,uint64,uint256,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initialState)
func (_ChannelHub *ChannelHubFilterer) WatchChannelCreated(opts *bind.WatchOpts, sink chan<- *ChannelHubChannelCreated, channelId [][32]byte, user []common.Address) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}
	var userRule []interface{}
	for _, userItem := range user {
		userRule = append(userRule, userItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ChannelCreated", channelIdRule, userRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubChannelCreated)
				if err := _ChannelHub.contract.UnpackLog(event, "ChannelCreated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChannelCreated is a log parse operation binding the contract event 0xb0d099feaab5034d04a1c610e86b8832343f2127b3c667b705834dafdf96e9e4.
//
// Solidity: event ChannelCreated(bytes32 indexed channelId, address indexed user, (uint32,address,address,uint64,uint256,bytes32) definition, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) initialState)
func (_ChannelHub *ChannelHubFilterer) ParseChannelCreated(log types.Log) (*ChannelHubChannelCreated, error) {
	event := new(ChannelHubChannelCreated)
	if err := _ChannelHub.contract.UnpackLog(event, "ChannelCreated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubChannelDepositedIterator is returned from FilterChannelDeposited and is used to iterate over the raw logs and unpacked data for ChannelDeposited events raised by the ChannelHub contract.
type ChannelHubChannelDepositedIterator struct {
	Event *ChannelHubChannelDeposited // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubChannelDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubChannelDeposited)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubChannelDeposited)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubChannelDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubChannelDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubChannelDeposited represents a ChannelDeposited event raised by the ChannelHub contract.
type ChannelHubChannelDeposited struct {
	ChannelId [32]byte
	Candidate State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterChannelDeposited is a free log retrieval operation binding the contract event 0x6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f41778696206696.
//
// Solidity: event ChannelDeposited(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) FilterChannelDeposited(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubChannelDepositedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ChannelDeposited", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubChannelDepositedIterator{contract: _ChannelHub.contract, event: "ChannelDeposited", logs: logs, sub: sub}, nil
}

// WatchChannelDeposited is a free log subscription operation binding the contract event 0x6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f41778696206696.
//
// Solidity: event ChannelDeposited(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) WatchChannelDeposited(opts *bind.WatchOpts, sink chan<- *ChannelHubChannelDeposited, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ChannelDeposited", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubChannelDeposited)
				if err := _ChannelHub.contract.UnpackLog(event, "ChannelDeposited", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChannelDeposited is a log parse operation binding the contract event 0x6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f41778696206696.
//
// Solidity: event ChannelDeposited(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) ParseChannelDeposited(log types.Log) (*ChannelHubChannelDeposited, error) {
	event := new(ChannelHubChannelDeposited)
	if err := _ChannelHub.contract.UnpackLog(event, "ChannelDeposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubChannelWithdrawnIterator is returned from FilterChannelWithdrawn and is used to iterate over the raw logs and unpacked data for ChannelWithdrawn events raised by the ChannelHub contract.
type ChannelHubChannelWithdrawnIterator struct {
	Event *ChannelHubChannelWithdrawn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubChannelWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubChannelWithdrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubChannelWithdrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubChannelWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubChannelWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubChannelWithdrawn represents a ChannelWithdrawn event raised by the ChannelHub contract.
type ChannelHubChannelWithdrawn struct {
	ChannelId [32]byte
	Candidate State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterChannelWithdrawn is a free log retrieval operation binding the contract event 0x188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf986.
//
// Solidity: event ChannelWithdrawn(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) FilterChannelWithdrawn(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubChannelWithdrawnIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ChannelWithdrawn", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubChannelWithdrawnIterator{contract: _ChannelHub.contract, event: "ChannelWithdrawn", logs: logs, sub: sub}, nil
}

// WatchChannelWithdrawn is a free log subscription operation binding the contract event 0x188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf986.
//
// Solidity: event ChannelWithdrawn(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) WatchChannelWithdrawn(opts *bind.WatchOpts, sink chan<- *ChannelHubChannelWithdrawn, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ChannelWithdrawn", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubChannelWithdrawn)
				if err := _ChannelHub.contract.UnpackLog(event, "ChannelWithdrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseChannelWithdrawn is a log parse operation binding the contract event 0x188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf986.
//
// Solidity: event ChannelWithdrawn(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate)
func (_ChannelHub *ChannelHubFilterer) ParseChannelWithdrawn(log types.Log) (*ChannelHubChannelWithdrawn, error) {
	event := new(ChannelHubChannelWithdrawn)
	if err := _ChannelHub.contract.UnpackLog(event, "ChannelWithdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubDepositedIterator is returned from FilterDeposited and is used to iterate over the raw logs and unpacked data for Deposited events raised by the ChannelHub contract.
type ChannelHubDepositedIterator struct {
	Event *ChannelHubDeposited // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubDepositedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubDeposited)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubDeposited)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubDepositedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubDepositedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubDeposited represents a Deposited event raised by the ChannelHub contract.
type ChannelHubDeposited struct {
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterDeposited is a free log retrieval operation binding the contract event 0x2da466a7b24304f47e87fa2e1e5a81b9831ce54fec19055ce277ca2f39ba42c4.
//
// Solidity: event Deposited(address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) FilterDeposited(opts *bind.FilterOpts, token []common.Address) (*ChannelHubDepositedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "Deposited", tokenRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubDepositedIterator{contract: _ChannelHub.contract, event: "Deposited", logs: logs, sub: sub}, nil
}

// WatchDeposited is a free log subscription operation binding the contract event 0x2da466a7b24304f47e87fa2e1e5a81b9831ce54fec19055ce277ca2f39ba42c4.
//
// Solidity: event Deposited(address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) WatchDeposited(opts *bind.WatchOpts, sink chan<- *ChannelHubDeposited, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "Deposited", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubDeposited)
				if err := _ChannelHub.contract.UnpackLog(event, "Deposited", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseDeposited is a log parse operation binding the contract event 0x2da466a7b24304f47e87fa2e1e5a81b9831ce54fec19055ce277ca2f39ba42c4.
//
// Solidity: event Deposited(address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) ParseDeposited(log types.Log) (*ChannelHubDeposited, error) {
	event := new(ChannelHubDeposited)
	if err := _ChannelHub.contract.UnpackLog(event, "Deposited", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowDepositChallengedIterator is returned from FilterEscrowDepositChallenged and is used to iterate over the raw logs and unpacked data for EscrowDepositChallenged events raised by the ChannelHub contract.
type ChannelHubEscrowDepositChallengedIterator struct {
	Event *ChannelHubEscrowDepositChallenged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowDepositChallengedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowDepositChallenged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowDepositChallenged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowDepositChallengedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowDepositChallengedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowDepositChallenged represents a EscrowDepositChallenged event raised by the ChannelHub contract.
type ChannelHubEscrowDepositChallenged struct {
	EscrowId          [32]byte
	State             State
	ChallengeExpireAt uint64
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositChallenged is a free log retrieval operation binding the contract event 0xba075bd445233f7cad862c72f0343b3503aad9c8e704a2295f122b82abf8e801.
//
// Solidity: event EscrowDepositChallenged(bytes32 indexed escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositChallenged(opts *bind.FilterOpts, escrowId [][32]byte) (*ChannelHubEscrowDepositChallengedIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositChallenged", escrowIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositChallengedIterator{contract: _ChannelHub.contract, event: "EscrowDepositChallenged", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositChallenged is a free log subscription operation binding the contract event 0xba075bd445233f7cad862c72f0343b3503aad9c8e704a2295f122b82abf8e801.
//
// Solidity: event EscrowDepositChallenged(bytes32 indexed escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowDepositChallenged(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowDepositChallenged, escrowId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowDepositChallenged", escrowIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowDepositChallenged)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositChallenged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowDepositChallenged is a log parse operation binding the contract event 0xba075bd445233f7cad862c72f0343b3503aad9c8e704a2295f122b82abf8e801.
//
// Solidity: event EscrowDepositChallenged(bytes32 indexed escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowDepositChallenged(log types.Log) (*ChannelHubEscrowDepositChallenged, error) {
	event := new(ChannelHubEscrowDepositChallenged)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositChallenged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowDepositFinalizedIterator is returned from FilterEscrowDepositFinalized and is used to iterate over the raw logs and unpacked data for EscrowDepositFinalized events raised by the ChannelHub contract.
type ChannelHubEscrowDepositFinalizedIterator struct {
	Event *ChannelHubEscrowDepositFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowDepositFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowDepositFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowDepositFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowDepositFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowDepositFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowDepositFinalized represents a EscrowDepositFinalized event raised by the ChannelHub contract.
type ChannelHubEscrowDepositFinalized struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositFinalized is a free log retrieval operation binding the contract event 0x1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e.
//
// Solidity: event EscrowDepositFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositFinalized(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowDepositFinalizedIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositFinalized", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositFinalizedIterator{contract: _ChannelHub.contract, event: "EscrowDepositFinalized", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositFinalized is a free log subscription operation binding the contract event 0x1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e.
//
// Solidity: event EscrowDepositFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowDepositFinalized(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowDepositFinalized, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowDepositFinalized", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowDepositFinalized)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowDepositFinalized is a log parse operation binding the contract event 0x1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e.
//
// Solidity: event EscrowDepositFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowDepositFinalized(log types.Log) (*ChannelHubEscrowDepositFinalized, error) {
	event := new(ChannelHubEscrowDepositFinalized)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowDepositFinalizedOnHomeIterator is returned from FilterEscrowDepositFinalizedOnHome and is used to iterate over the raw logs and unpacked data for EscrowDepositFinalizedOnHome events raised by the ChannelHub contract.
type ChannelHubEscrowDepositFinalizedOnHomeIterator struct {
	Event *ChannelHubEscrowDepositFinalizedOnHome // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowDepositFinalizedOnHomeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowDepositFinalizedOnHome)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowDepositFinalizedOnHome)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowDepositFinalizedOnHomeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowDepositFinalizedOnHomeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowDepositFinalizedOnHome represents a EscrowDepositFinalizedOnHome event raised by the ChannelHub contract.
type ChannelHubEscrowDepositFinalizedOnHome struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositFinalizedOnHome is a free log retrieval operation binding the contract event 0x32e24720f56fd5a7f4cb219d7ff3278ae95196e79c85b5801395894a6f53466c.
//
// Solidity: event EscrowDepositFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositFinalizedOnHome(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowDepositFinalizedOnHomeIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositFinalizedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositFinalizedOnHomeIterator{contract: _ChannelHub.contract, event: "EscrowDepositFinalizedOnHome", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositFinalizedOnHome is a free log subscription operation binding the contract event 0x32e24720f56fd5a7f4cb219d7ff3278ae95196e79c85b5801395894a6f53466c.
//
// Solidity: event EscrowDepositFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowDepositFinalizedOnHome(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowDepositFinalizedOnHome, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowDepositFinalizedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowDepositFinalizedOnHome)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositFinalizedOnHome", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowDepositFinalizedOnHome is a log parse operation binding the contract event 0x32e24720f56fd5a7f4cb219d7ff3278ae95196e79c85b5801395894a6f53466c.
//
// Solidity: event EscrowDepositFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowDepositFinalizedOnHome(log types.Log) (*ChannelHubEscrowDepositFinalizedOnHome, error) {
	event := new(ChannelHubEscrowDepositFinalizedOnHome)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositFinalizedOnHome", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowDepositInitiatedIterator is returned from FilterEscrowDepositInitiated and is used to iterate over the raw logs and unpacked data for EscrowDepositInitiated events raised by the ChannelHub contract.
type ChannelHubEscrowDepositInitiatedIterator struct {
	Event *ChannelHubEscrowDepositInitiated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowDepositInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowDepositInitiated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowDepositInitiated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowDepositInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowDepositInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowDepositInitiated represents a EscrowDepositInitiated event raised by the ChannelHub contract.
type ChannelHubEscrowDepositInitiated struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositInitiated is a free log retrieval operation binding the contract event 0xede7867afa7cdb9c443667efd8244d98bf9df1dce68e60dc94dca6605125ca76.
//
// Solidity: event EscrowDepositInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositInitiated(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowDepositInitiatedIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositInitiated", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositInitiatedIterator{contract: _ChannelHub.contract, event: "EscrowDepositInitiated", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositInitiated is a free log subscription operation binding the contract event 0xede7867afa7cdb9c443667efd8244d98bf9df1dce68e60dc94dca6605125ca76.
//
// Solidity: event EscrowDepositInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowDepositInitiated(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowDepositInitiated, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowDepositInitiated", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowDepositInitiated)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositInitiated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowDepositInitiated is a log parse operation binding the contract event 0xede7867afa7cdb9c443667efd8244d98bf9df1dce68e60dc94dca6605125ca76.
//
// Solidity: event EscrowDepositInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowDepositInitiated(log types.Log) (*ChannelHubEscrowDepositInitiated, error) {
	event := new(ChannelHubEscrowDepositInitiated)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowDepositInitiatedOnHomeIterator is returned from FilterEscrowDepositInitiatedOnHome and is used to iterate over the raw logs and unpacked data for EscrowDepositInitiatedOnHome events raised by the ChannelHub contract.
type ChannelHubEscrowDepositInitiatedOnHomeIterator struct {
	Event *ChannelHubEscrowDepositInitiatedOnHome // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowDepositInitiatedOnHomeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowDepositInitiatedOnHome)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowDepositInitiatedOnHome)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowDepositInitiatedOnHomeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowDepositInitiatedOnHomeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowDepositInitiatedOnHome represents a EscrowDepositInitiatedOnHome event raised by the ChannelHub contract.
type ChannelHubEscrowDepositInitiatedOnHome struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositInitiatedOnHome is a free log retrieval operation binding the contract event 0x471c4ebe4e57d25ef7117e141caac31c6b98f067b8098a7a7bbd38f637c2f980.
//
// Solidity: event EscrowDepositInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositInitiatedOnHome(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowDepositInitiatedOnHomeIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositInitiatedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositInitiatedOnHomeIterator{contract: _ChannelHub.contract, event: "EscrowDepositInitiatedOnHome", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositInitiatedOnHome is a free log subscription operation binding the contract event 0x471c4ebe4e57d25ef7117e141caac31c6b98f067b8098a7a7bbd38f637c2f980.
//
// Solidity: event EscrowDepositInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowDepositInitiatedOnHome(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowDepositInitiatedOnHome, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowDepositInitiatedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowDepositInitiatedOnHome)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositInitiatedOnHome", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowDepositInitiatedOnHome is a log parse operation binding the contract event 0x471c4ebe4e57d25ef7117e141caac31c6b98f067b8098a7a7bbd38f637c2f980.
//
// Solidity: event EscrowDepositInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowDepositInitiatedOnHome(log types.Log) (*ChannelHubEscrowDepositInitiatedOnHome, error) {
	event := new(ChannelHubEscrowDepositInitiatedOnHome)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositInitiatedOnHome", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowDepositsPurgedIterator is returned from FilterEscrowDepositsPurged and is used to iterate over the raw logs and unpacked data for EscrowDepositsPurged events raised by the ChannelHub contract.
type ChannelHubEscrowDepositsPurgedIterator struct {
	Event *ChannelHubEscrowDepositsPurged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowDepositsPurgedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowDepositsPurged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowDepositsPurged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowDepositsPurgedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowDepositsPurgedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowDepositsPurged represents a EscrowDepositsPurged event raised by the ChannelHub contract.
type ChannelHubEscrowDepositsPurged struct {
	EscrowIds   [][32]byte
	PurgedCount *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositsPurged is a free log retrieval operation binding the contract event 0x8fac6141d748dc9c9bc16cc25f636385597618190a44c03d33be5656e01b3642.
//
// Solidity: event EscrowDepositsPurged(bytes32[] escrowIds, uint256 purgedCount)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositsPurged(opts *bind.FilterOpts) (*ChannelHubEscrowDepositsPurgedIterator, error) {

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositsPurged")
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositsPurgedIterator{contract: _ChannelHub.contract, event: "EscrowDepositsPurged", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositsPurged is a free log subscription operation binding the contract event 0x8fac6141d748dc9c9bc16cc25f636385597618190a44c03d33be5656e01b3642.
//
// Solidity: event EscrowDepositsPurged(bytes32[] escrowIds, uint256 purgedCount)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowDepositsPurged(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowDepositsPurged) (event.Subscription, error) {

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowDepositsPurged")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowDepositsPurged)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositsPurged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowDepositsPurged is a log parse operation binding the contract event 0x8fac6141d748dc9c9bc16cc25f636385597618190a44c03d33be5656e01b3642.
//
// Solidity: event EscrowDepositsPurged(bytes32[] escrowIds, uint256 purgedCount)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowDepositsPurged(log types.Log) (*ChannelHubEscrowDepositsPurged, error) {
	event := new(ChannelHubEscrowDepositsPurged)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowDepositsPurged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowWithdrawalChallengedIterator is returned from FilterEscrowWithdrawalChallenged and is used to iterate over the raw logs and unpacked data for EscrowWithdrawalChallenged events raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalChallengedIterator struct {
	Event *ChannelHubEscrowWithdrawalChallenged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowWithdrawalChallengedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowWithdrawalChallenged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowWithdrawalChallenged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowWithdrawalChallengedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowWithdrawalChallengedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowWithdrawalChallenged represents a EscrowWithdrawalChallenged event raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalChallenged struct {
	EscrowId          [32]byte
	State             State
	ChallengeExpireAt uint64
	Raw               types.Log // Blockchain specific contextual infos
}

// FilterEscrowWithdrawalChallenged is a free log retrieval operation binding the contract event 0xb8568a1f475f3c76759a620e08a653d28348c5c09e2e0bc91d533339801fefd8.
//
// Solidity: event EscrowWithdrawalChallenged(bytes32 indexed escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowWithdrawalChallenged(opts *bind.FilterOpts, escrowId [][32]byte) (*ChannelHubEscrowWithdrawalChallengedIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowWithdrawalChallenged", escrowIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowWithdrawalChallengedIterator{contract: _ChannelHub.contract, event: "EscrowWithdrawalChallenged", logs: logs, sub: sub}, nil
}

// WatchEscrowWithdrawalChallenged is a free log subscription operation binding the contract event 0xb8568a1f475f3c76759a620e08a653d28348c5c09e2e0bc91d533339801fefd8.
//
// Solidity: event EscrowWithdrawalChallenged(bytes32 indexed escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowWithdrawalChallenged(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowWithdrawalChallenged, escrowId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowWithdrawalChallenged", escrowIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowWithdrawalChallenged)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalChallenged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowWithdrawalChallenged is a log parse operation binding the contract event 0xb8568a1f475f3c76759a620e08a653d28348c5c09e2e0bc91d533339801fefd8.
//
// Solidity: event EscrowWithdrawalChallenged(bytes32 indexed escrowId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state, uint64 challengeExpireAt)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowWithdrawalChallenged(log types.Log) (*ChannelHubEscrowWithdrawalChallenged, error) {
	event := new(ChannelHubEscrowWithdrawalChallenged)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalChallenged", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowWithdrawalFinalizedIterator is returned from FilterEscrowWithdrawalFinalized and is used to iterate over the raw logs and unpacked data for EscrowWithdrawalFinalized events raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalFinalizedIterator struct {
	Event *ChannelHubEscrowWithdrawalFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowWithdrawalFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowWithdrawalFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowWithdrawalFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowWithdrawalFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowWithdrawalFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowWithdrawalFinalized represents a EscrowWithdrawalFinalized event raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalFinalized struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowWithdrawalFinalized is a free log retrieval operation binding the contract event 0x2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d1.
//
// Solidity: event EscrowWithdrawalFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowWithdrawalFinalized(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowWithdrawalFinalizedIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowWithdrawalFinalized", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowWithdrawalFinalizedIterator{contract: _ChannelHub.contract, event: "EscrowWithdrawalFinalized", logs: logs, sub: sub}, nil
}

// WatchEscrowWithdrawalFinalized is a free log subscription operation binding the contract event 0x2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d1.
//
// Solidity: event EscrowWithdrawalFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowWithdrawalFinalized(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowWithdrawalFinalized, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowWithdrawalFinalized", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowWithdrawalFinalized)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowWithdrawalFinalized is a log parse operation binding the contract event 0x2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d1.
//
// Solidity: event EscrowWithdrawalFinalized(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowWithdrawalFinalized(log types.Log) (*ChannelHubEscrowWithdrawalFinalized, error) {
	event := new(ChannelHubEscrowWithdrawalFinalized)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowWithdrawalFinalizedOnHomeIterator is returned from FilterEscrowWithdrawalFinalizedOnHome and is used to iterate over the raw logs and unpacked data for EscrowWithdrawalFinalizedOnHome events raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalFinalizedOnHomeIterator struct {
	Event *ChannelHubEscrowWithdrawalFinalizedOnHome // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowWithdrawalFinalizedOnHomeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowWithdrawalFinalizedOnHome)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowWithdrawalFinalizedOnHome)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowWithdrawalFinalizedOnHomeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowWithdrawalFinalizedOnHomeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowWithdrawalFinalizedOnHome represents a EscrowWithdrawalFinalizedOnHome event raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalFinalizedOnHome struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowWithdrawalFinalizedOnHome is a free log retrieval operation binding the contract event 0x6d0cf3d243d63f08f50db493a8af34b27d4e3bc9ec4098e82700abfeffe2d498.
//
// Solidity: event EscrowWithdrawalFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowWithdrawalFinalizedOnHome(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowWithdrawalFinalizedOnHomeIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowWithdrawalFinalizedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowWithdrawalFinalizedOnHomeIterator{contract: _ChannelHub.contract, event: "EscrowWithdrawalFinalizedOnHome", logs: logs, sub: sub}, nil
}

// WatchEscrowWithdrawalFinalizedOnHome is a free log subscription operation binding the contract event 0x6d0cf3d243d63f08f50db493a8af34b27d4e3bc9ec4098e82700abfeffe2d498.
//
// Solidity: event EscrowWithdrawalFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowWithdrawalFinalizedOnHome(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowWithdrawalFinalizedOnHome, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowWithdrawalFinalizedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowWithdrawalFinalizedOnHome)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalFinalizedOnHome", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowWithdrawalFinalizedOnHome is a log parse operation binding the contract event 0x6d0cf3d243d63f08f50db493a8af34b27d4e3bc9ec4098e82700abfeffe2d498.
//
// Solidity: event EscrowWithdrawalFinalizedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowWithdrawalFinalizedOnHome(log types.Log) (*ChannelHubEscrowWithdrawalFinalizedOnHome, error) {
	event := new(ChannelHubEscrowWithdrawalFinalizedOnHome)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalFinalizedOnHome", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowWithdrawalInitiatedIterator is returned from FilterEscrowWithdrawalInitiated and is used to iterate over the raw logs and unpacked data for EscrowWithdrawalInitiated events raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalInitiatedIterator struct {
	Event *ChannelHubEscrowWithdrawalInitiated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowWithdrawalInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowWithdrawalInitiated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowWithdrawalInitiated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowWithdrawalInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowWithdrawalInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowWithdrawalInitiated represents a EscrowWithdrawalInitiated event raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalInitiated struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowWithdrawalInitiated is a free log retrieval operation binding the contract event 0x17eb0a6bd5a0de45d1029ce3444941070e149df35b22176fc439f930f73c09f7.
//
// Solidity: event EscrowWithdrawalInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowWithdrawalInitiated(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowWithdrawalInitiatedIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowWithdrawalInitiated", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowWithdrawalInitiatedIterator{contract: _ChannelHub.contract, event: "EscrowWithdrawalInitiated", logs: logs, sub: sub}, nil
}

// WatchEscrowWithdrawalInitiated is a free log subscription operation binding the contract event 0x17eb0a6bd5a0de45d1029ce3444941070e149df35b22176fc439f930f73c09f7.
//
// Solidity: event EscrowWithdrawalInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowWithdrawalInitiated(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowWithdrawalInitiated, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowWithdrawalInitiated", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowWithdrawalInitiated)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalInitiated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowWithdrawalInitiated is a log parse operation binding the contract event 0x17eb0a6bd5a0de45d1029ce3444941070e149df35b22176fc439f930f73c09f7.
//
// Solidity: event EscrowWithdrawalInitiated(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowWithdrawalInitiated(log types.Log) (*ChannelHubEscrowWithdrawalInitiated, error) {
	event := new(ChannelHubEscrowWithdrawalInitiated)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubEscrowWithdrawalInitiatedOnHomeIterator is returned from FilterEscrowWithdrawalInitiatedOnHome and is used to iterate over the raw logs and unpacked data for EscrowWithdrawalInitiatedOnHome events raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalInitiatedOnHomeIterator struct {
	Event *ChannelHubEscrowWithdrawalInitiatedOnHome // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubEscrowWithdrawalInitiatedOnHomeIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubEscrowWithdrawalInitiatedOnHome)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubEscrowWithdrawalInitiatedOnHome)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubEscrowWithdrawalInitiatedOnHomeIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubEscrowWithdrawalInitiatedOnHomeIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubEscrowWithdrawalInitiatedOnHome represents a EscrowWithdrawalInitiatedOnHome event raised by the ChannelHub contract.
type ChannelHubEscrowWithdrawalInitiatedOnHome struct {
	EscrowId  [32]byte
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterEscrowWithdrawalInitiatedOnHome is a free log retrieval operation binding the contract event 0x587faad1bcd589ce902468251883e1976a645af8563c773eed7356d78433210c.
//
// Solidity: event EscrowWithdrawalInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowWithdrawalInitiatedOnHome(opts *bind.FilterOpts, escrowId [][32]byte, channelId [][32]byte) (*ChannelHubEscrowWithdrawalInitiatedOnHomeIterator, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowWithdrawalInitiatedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowWithdrawalInitiatedOnHomeIterator{contract: _ChannelHub.contract, event: "EscrowWithdrawalInitiatedOnHome", logs: logs, sub: sub}, nil
}

// WatchEscrowWithdrawalInitiatedOnHome is a free log subscription operation binding the contract event 0x587faad1bcd589ce902468251883e1976a645af8563c773eed7356d78433210c.
//
// Solidity: event EscrowWithdrawalInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchEscrowWithdrawalInitiatedOnHome(opts *bind.WatchOpts, sink chan<- *ChannelHubEscrowWithdrawalInitiatedOnHome, escrowId [][32]byte, channelId [][32]byte) (event.Subscription, error) {

	var escrowIdRule []interface{}
	for _, escrowIdItem := range escrowId {
		escrowIdRule = append(escrowIdRule, escrowIdItem)
	}
	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "EscrowWithdrawalInitiatedOnHome", escrowIdRule, channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubEscrowWithdrawalInitiatedOnHome)
				if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalInitiatedOnHome", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseEscrowWithdrawalInitiatedOnHome is a log parse operation binding the contract event 0x587faad1bcd589ce902468251883e1976a645af8563c773eed7356d78433210c.
//
// Solidity: event EscrowWithdrawalInitiatedOnHome(bytes32 indexed escrowId, bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseEscrowWithdrawalInitiatedOnHome(log types.Log) (*ChannelHubEscrowWithdrawalInitiatedOnHome, error) {
	event := new(ChannelHubEscrowWithdrawalInitiatedOnHome)
	if err := _ChannelHub.contract.UnpackLog(event, "EscrowWithdrawalInitiatedOnHome", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubFundsClaimedIterator is returned from FilterFundsClaimed and is used to iterate over the raw logs and unpacked data for FundsClaimed events raised by the ChannelHub contract.
type ChannelHubFundsClaimedIterator struct {
	Event *ChannelHubFundsClaimed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubFundsClaimedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubFundsClaimed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubFundsClaimed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubFundsClaimedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubFundsClaimedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubFundsClaimed represents a FundsClaimed event raised by the ChannelHub contract.
type ChannelHubFundsClaimed struct {
	Account     common.Address
	Token       common.Address
	Destination common.Address
	Amount      *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterFundsClaimed is a free log retrieval operation binding the contract event 0x7b8d70738154be94a9a068a6d2f5dd8cfc65c52855859dc8f47de1ff185f8b55.
//
// Solidity: event FundsClaimed(address indexed account, address indexed token, address indexed destination, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) FilterFundsClaimed(opts *bind.FilterOpts, account []common.Address, token []common.Address, destination []common.Address) (*ChannelHubFundsClaimedIterator, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var destinationRule []interface{}
	for _, destinationItem := range destination {
		destinationRule = append(destinationRule, destinationItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "FundsClaimed", accountRule, tokenRule, destinationRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubFundsClaimedIterator{contract: _ChannelHub.contract, event: "FundsClaimed", logs: logs, sub: sub}, nil
}

// WatchFundsClaimed is a free log subscription operation binding the contract event 0x7b8d70738154be94a9a068a6d2f5dd8cfc65c52855859dc8f47de1ff185f8b55.
//
// Solidity: event FundsClaimed(address indexed account, address indexed token, address indexed destination, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) WatchFundsClaimed(opts *bind.WatchOpts, sink chan<- *ChannelHubFundsClaimed, account []common.Address, token []common.Address, destination []common.Address) (event.Subscription, error) {

	var accountRule []interface{}
	for _, accountItem := range account {
		accountRule = append(accountRule, accountItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}
	var destinationRule []interface{}
	for _, destinationItem := range destination {
		destinationRule = append(destinationRule, destinationItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "FundsClaimed", accountRule, tokenRule, destinationRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubFundsClaimed)
				if err := _ChannelHub.contract.UnpackLog(event, "FundsClaimed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseFundsClaimed is a log parse operation binding the contract event 0x7b8d70738154be94a9a068a6d2f5dd8cfc65c52855859dc8f47de1ff185f8b55.
//
// Solidity: event FundsClaimed(address indexed account, address indexed token, address indexed destination, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) ParseFundsClaimed(log types.Log) (*ChannelHubFundsClaimed, error) {
	event := new(ChannelHubFundsClaimed)
	if err := _ChannelHub.contract.UnpackLog(event, "FundsClaimed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubMigrationInFinalizedIterator is returned from FilterMigrationInFinalized and is used to iterate over the raw logs and unpacked data for MigrationInFinalized events raised by the ChannelHub contract.
type ChannelHubMigrationInFinalizedIterator struct {
	Event *ChannelHubMigrationInFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubMigrationInFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubMigrationInFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubMigrationInFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubMigrationInFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubMigrationInFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubMigrationInFinalized represents a MigrationInFinalized event raised by the ChannelHub contract.
type ChannelHubMigrationInFinalized struct {
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterMigrationInFinalized is a free log retrieval operation binding the contract event 0x7b20773c41402791c5f18914dbbeacad38b1ebcc4c55d8eb3bfe0a4cde26c826.
//
// Solidity: event MigrationInFinalized(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterMigrationInFinalized(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubMigrationInFinalizedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "MigrationInFinalized", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubMigrationInFinalizedIterator{contract: _ChannelHub.contract, event: "MigrationInFinalized", logs: logs, sub: sub}, nil
}

// WatchMigrationInFinalized is a free log subscription operation binding the contract event 0x7b20773c41402791c5f18914dbbeacad38b1ebcc4c55d8eb3bfe0a4cde26c826.
//
// Solidity: event MigrationInFinalized(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchMigrationInFinalized(opts *bind.WatchOpts, sink chan<- *ChannelHubMigrationInFinalized, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "MigrationInFinalized", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubMigrationInFinalized)
				if err := _ChannelHub.contract.UnpackLog(event, "MigrationInFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMigrationInFinalized is a log parse operation binding the contract event 0x7b20773c41402791c5f18914dbbeacad38b1ebcc4c55d8eb3bfe0a4cde26c826.
//
// Solidity: event MigrationInFinalized(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseMigrationInFinalized(log types.Log) (*ChannelHubMigrationInFinalized, error) {
	event := new(ChannelHubMigrationInFinalized)
	if err := _ChannelHub.contract.UnpackLog(event, "MigrationInFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubMigrationInInitiatedIterator is returned from FilterMigrationInInitiated and is used to iterate over the raw logs and unpacked data for MigrationInInitiated events raised by the ChannelHub contract.
type ChannelHubMigrationInInitiatedIterator struct {
	Event *ChannelHubMigrationInInitiated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubMigrationInInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubMigrationInInitiated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubMigrationInInitiated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubMigrationInInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubMigrationInInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubMigrationInInitiated represents a MigrationInInitiated event raised by the ChannelHub contract.
type ChannelHubMigrationInInitiated struct {
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterMigrationInInitiated is a free log retrieval operation binding the contract event 0x26afbcb9eb52c21f42eb9cfe8f263718ffb65afbf84abe8ad8cce2acfb2242b8.
//
// Solidity: event MigrationInInitiated(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterMigrationInInitiated(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubMigrationInInitiatedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "MigrationInInitiated", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubMigrationInInitiatedIterator{contract: _ChannelHub.contract, event: "MigrationInInitiated", logs: logs, sub: sub}, nil
}

// WatchMigrationInInitiated is a free log subscription operation binding the contract event 0x26afbcb9eb52c21f42eb9cfe8f263718ffb65afbf84abe8ad8cce2acfb2242b8.
//
// Solidity: event MigrationInInitiated(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchMigrationInInitiated(opts *bind.WatchOpts, sink chan<- *ChannelHubMigrationInInitiated, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "MigrationInInitiated", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubMigrationInInitiated)
				if err := _ChannelHub.contract.UnpackLog(event, "MigrationInInitiated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMigrationInInitiated is a log parse operation binding the contract event 0x26afbcb9eb52c21f42eb9cfe8f263718ffb65afbf84abe8ad8cce2acfb2242b8.
//
// Solidity: event MigrationInInitiated(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseMigrationInInitiated(log types.Log) (*ChannelHubMigrationInInitiated, error) {
	event := new(ChannelHubMigrationInInitiated)
	if err := _ChannelHub.contract.UnpackLog(event, "MigrationInInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubMigrationOutFinalizedIterator is returned from FilterMigrationOutFinalized and is used to iterate over the raw logs and unpacked data for MigrationOutFinalized events raised by the ChannelHub contract.
type ChannelHubMigrationOutFinalizedIterator struct {
	Event *ChannelHubMigrationOutFinalized // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubMigrationOutFinalizedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubMigrationOutFinalized)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubMigrationOutFinalized)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubMigrationOutFinalizedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubMigrationOutFinalizedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubMigrationOutFinalized represents a MigrationOutFinalized event raised by the ChannelHub contract.
type ChannelHubMigrationOutFinalized struct {
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterMigrationOutFinalized is a free log retrieval operation binding the contract event 0x9a6f675cc94b83b55f1ecc0876affd4332a30c92e6faa2aca0199b1b6df922c3.
//
// Solidity: event MigrationOutFinalized(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterMigrationOutFinalized(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubMigrationOutFinalizedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "MigrationOutFinalized", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubMigrationOutFinalizedIterator{contract: _ChannelHub.contract, event: "MigrationOutFinalized", logs: logs, sub: sub}, nil
}

// WatchMigrationOutFinalized is a free log subscription operation binding the contract event 0x9a6f675cc94b83b55f1ecc0876affd4332a30c92e6faa2aca0199b1b6df922c3.
//
// Solidity: event MigrationOutFinalized(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchMigrationOutFinalized(opts *bind.WatchOpts, sink chan<- *ChannelHubMigrationOutFinalized, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "MigrationOutFinalized", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubMigrationOutFinalized)
				if err := _ChannelHub.contract.UnpackLog(event, "MigrationOutFinalized", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMigrationOutFinalized is a log parse operation binding the contract event 0x9a6f675cc94b83b55f1ecc0876affd4332a30c92e6faa2aca0199b1b6df922c3.
//
// Solidity: event MigrationOutFinalized(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseMigrationOutFinalized(log types.Log) (*ChannelHubMigrationOutFinalized, error) {
	event := new(ChannelHubMigrationOutFinalized)
	if err := _ChannelHub.contract.UnpackLog(event, "MigrationOutFinalized", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubMigrationOutInitiatedIterator is returned from FilterMigrationOutInitiated and is used to iterate over the raw logs and unpacked data for MigrationOutInitiated events raised by the ChannelHub contract.
type ChannelHubMigrationOutInitiatedIterator struct {
	Event *ChannelHubMigrationOutInitiated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubMigrationOutInitiatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubMigrationOutInitiated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubMigrationOutInitiated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubMigrationOutInitiatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubMigrationOutInitiatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubMigrationOutInitiated represents a MigrationOutInitiated event raised by the ChannelHub contract.
type ChannelHubMigrationOutInitiated struct {
	ChannelId [32]byte
	State     State
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterMigrationOutInitiated is a free log retrieval operation binding the contract event 0x3142fb397e715d80415dff7b527bf1c451def4675da6e1199ee1b4588e3f630a.
//
// Solidity: event MigrationOutInitiated(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) FilterMigrationOutInitiated(opts *bind.FilterOpts, channelId [][32]byte) (*ChannelHubMigrationOutInitiatedIterator, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "MigrationOutInitiated", channelIdRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubMigrationOutInitiatedIterator{contract: _ChannelHub.contract, event: "MigrationOutInitiated", logs: logs, sub: sub}, nil
}

// WatchMigrationOutInitiated is a free log subscription operation binding the contract event 0x3142fb397e715d80415dff7b527bf1c451def4675da6e1199ee1b4588e3f630a.
//
// Solidity: event MigrationOutInitiated(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) WatchMigrationOutInitiated(opts *bind.WatchOpts, sink chan<- *ChannelHubMigrationOutInitiated, channelId [][32]byte) (event.Subscription, error) {

	var channelIdRule []interface{}
	for _, channelIdItem := range channelId {
		channelIdRule = append(channelIdRule, channelIdItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "MigrationOutInitiated", channelIdRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubMigrationOutInitiated)
				if err := _ChannelHub.contract.UnpackLog(event, "MigrationOutInitiated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseMigrationOutInitiated is a log parse operation binding the contract event 0x3142fb397e715d80415dff7b527bf1c451def4675da6e1199ee1b4588e3f630a.
//
// Solidity: event MigrationOutInitiated(bytes32 indexed channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) state)
func (_ChannelHub *ChannelHubFilterer) ParseMigrationOutInitiated(log types.Log) (*ChannelHubMigrationOutInitiated, error) {
	event := new(ChannelHubMigrationOutInitiated)
	if err := _ChannelHub.contract.UnpackLog(event, "MigrationOutInitiated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubNodeBalanceUpdatedIterator is returned from FilterNodeBalanceUpdated and is used to iterate over the raw logs and unpacked data for NodeBalanceUpdated events raised by the ChannelHub contract.
type ChannelHubNodeBalanceUpdatedIterator struct {
	Event *ChannelHubNodeBalanceUpdated // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubNodeBalanceUpdatedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubNodeBalanceUpdated)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubNodeBalanceUpdated)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubNodeBalanceUpdatedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubNodeBalanceUpdatedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubNodeBalanceUpdated represents a NodeBalanceUpdated event raised by the ChannelHub contract.
type ChannelHubNodeBalanceUpdated struct {
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterNodeBalanceUpdated is a free log retrieval operation binding the contract event 0x05f47829691a1f710b0620aedd52749bb09d8abe4bb530d306db920a71b0d7ce.
//
// Solidity: event NodeBalanceUpdated(address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) FilterNodeBalanceUpdated(opts *bind.FilterOpts, token []common.Address) (*ChannelHubNodeBalanceUpdatedIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "NodeBalanceUpdated", tokenRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubNodeBalanceUpdatedIterator{contract: _ChannelHub.contract, event: "NodeBalanceUpdated", logs: logs, sub: sub}, nil
}

// WatchNodeBalanceUpdated is a free log subscription operation binding the contract event 0x05f47829691a1f710b0620aedd52749bb09d8abe4bb530d306db920a71b0d7ce.
//
// Solidity: event NodeBalanceUpdated(address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) WatchNodeBalanceUpdated(opts *bind.WatchOpts, sink chan<- *ChannelHubNodeBalanceUpdated, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "NodeBalanceUpdated", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubNodeBalanceUpdated)
				if err := _ChannelHub.contract.UnpackLog(event, "NodeBalanceUpdated", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseNodeBalanceUpdated is a log parse operation binding the contract event 0x05f47829691a1f710b0620aedd52749bb09d8abe4bb530d306db920a71b0d7ce.
//
// Solidity: event NodeBalanceUpdated(address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) ParseNodeBalanceUpdated(log types.Log) (*ChannelHubNodeBalanceUpdated, error) {
	event := new(ChannelHubNodeBalanceUpdated)
	if err := _ChannelHub.contract.UnpackLog(event, "NodeBalanceUpdated", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubTransferFailedIterator is returned from FilterTransferFailed and is used to iterate over the raw logs and unpacked data for TransferFailed events raised by the ChannelHub contract.
type ChannelHubTransferFailedIterator struct {
	Event *ChannelHubTransferFailed // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubTransferFailedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubTransferFailed)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubTransferFailed)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubTransferFailedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubTransferFailedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubTransferFailed represents a TransferFailed event raised by the ChannelHub contract.
type ChannelHubTransferFailed struct {
	Recipient common.Address
	Token     common.Address
	Amount    *big.Int
	Raw       types.Log // Blockchain specific contextual infos
}

// FilterTransferFailed is a free log retrieval operation binding the contract event 0xbf182be802245e8ed88e4b8d3e4344c0863dd2a70334f089fd07265389306fcf.
//
// Solidity: event TransferFailed(address indexed recipient, address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) FilterTransferFailed(opts *bind.FilterOpts, recipient []common.Address, token []common.Address) (*ChannelHubTransferFailedIterator, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "TransferFailed", recipientRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubTransferFailedIterator{contract: _ChannelHub.contract, event: "TransferFailed", logs: logs, sub: sub}, nil
}

// WatchTransferFailed is a free log subscription operation binding the contract event 0xbf182be802245e8ed88e4b8d3e4344c0863dd2a70334f089fd07265389306fcf.
//
// Solidity: event TransferFailed(address indexed recipient, address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) WatchTransferFailed(opts *bind.WatchOpts, sink chan<- *ChannelHubTransferFailed, recipient []common.Address, token []common.Address) (event.Subscription, error) {

	var recipientRule []interface{}
	for _, recipientItem := range recipient {
		recipientRule = append(recipientRule, recipientItem)
	}
	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "TransferFailed", recipientRule, tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubTransferFailed)
				if err := _ChannelHub.contract.UnpackLog(event, "TransferFailed", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseTransferFailed is a log parse operation binding the contract event 0xbf182be802245e8ed88e4b8d3e4344c0863dd2a70334f089fd07265389306fcf.
//
// Solidity: event TransferFailed(address indexed recipient, address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) ParseTransferFailed(log types.Log) (*ChannelHubTransferFailed, error) {
	event := new(ChannelHubTransferFailed)
	if err := _ChannelHub.contract.UnpackLog(event, "TransferFailed", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubValidatorRegisteredIterator is returned from FilterValidatorRegistered and is used to iterate over the raw logs and unpacked data for ValidatorRegistered events raised by the ChannelHub contract.
type ChannelHubValidatorRegisteredIterator struct {
	Event *ChannelHubValidatorRegistered // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubValidatorRegisteredIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubValidatorRegistered)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubValidatorRegistered)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubValidatorRegisteredIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubValidatorRegisteredIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubValidatorRegistered represents a ValidatorRegistered event raised by the ChannelHub contract.
type ChannelHubValidatorRegistered struct {
	ValidatorId uint8
	Validator   common.Address
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterValidatorRegistered is a free log retrieval operation binding the contract event 0x9ee792368f12db92ad66335fa19df35feaec025c86445fea202ab5412a180e05.
//
// Solidity: event ValidatorRegistered(uint8 indexed validatorId, address indexed validator)
func (_ChannelHub *ChannelHubFilterer) FilterValidatorRegistered(opts *bind.FilterOpts, validatorId []uint8, validator []common.Address) (*ChannelHubValidatorRegisteredIterator, error) {

	var validatorIdRule []interface{}
	for _, validatorIdItem := range validatorId {
		validatorIdRule = append(validatorIdRule, validatorIdItem)
	}
	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "ValidatorRegistered", validatorIdRule, validatorRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubValidatorRegisteredIterator{contract: _ChannelHub.contract, event: "ValidatorRegistered", logs: logs, sub: sub}, nil
}

// WatchValidatorRegistered is a free log subscription operation binding the contract event 0x9ee792368f12db92ad66335fa19df35feaec025c86445fea202ab5412a180e05.
//
// Solidity: event ValidatorRegistered(uint8 indexed validatorId, address indexed validator)
func (_ChannelHub *ChannelHubFilterer) WatchValidatorRegistered(opts *bind.WatchOpts, sink chan<- *ChannelHubValidatorRegistered, validatorId []uint8, validator []common.Address) (event.Subscription, error) {

	var validatorIdRule []interface{}
	for _, validatorIdItem := range validatorId {
		validatorIdRule = append(validatorIdRule, validatorIdItem)
	}
	var validatorRule []interface{}
	for _, validatorItem := range validator {
		validatorRule = append(validatorRule, validatorItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "ValidatorRegistered", validatorIdRule, validatorRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubValidatorRegistered)
				if err := _ChannelHub.contract.UnpackLog(event, "ValidatorRegistered", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseValidatorRegistered is a log parse operation binding the contract event 0x9ee792368f12db92ad66335fa19df35feaec025c86445fea202ab5412a180e05.
//
// Solidity: event ValidatorRegistered(uint8 indexed validatorId, address indexed validator)
func (_ChannelHub *ChannelHubFilterer) ParseValidatorRegistered(log types.Log) (*ChannelHubValidatorRegistered, error) {
	event := new(ChannelHubValidatorRegistered)
	if err := _ChannelHub.contract.UnpackLog(event, "ValidatorRegistered", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}

// ChannelHubWithdrawnIterator is returned from FilterWithdrawn and is used to iterate over the raw logs and unpacked data for Withdrawn events raised by the ChannelHub contract.
type ChannelHubWithdrawnIterator struct {
	Event *ChannelHubWithdrawn // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ChannelHubWithdrawnIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ChannelHubWithdrawn)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ChannelHubWithdrawn)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ChannelHubWithdrawnIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ChannelHubWithdrawnIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ChannelHubWithdrawn represents a Withdrawn event raised by the ChannelHub contract.
type ChannelHubWithdrawn struct {
	Token  common.Address
	Amount *big.Int
	Raw    types.Log // Blockchain specific contextual infos
}

// FilterWithdrawn is a free log retrieval operation binding the contract event 0x7084f5476618d8e60b11ef0d7d3f06914655adb8793e28ff7f018d4c76d505d5.
//
// Solidity: event Withdrawn(address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) FilterWithdrawn(opts *bind.FilterOpts, token []common.Address) (*ChannelHubWithdrawnIterator, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "Withdrawn", tokenRule)
	if err != nil {
		return nil, err
	}
	return &ChannelHubWithdrawnIterator{contract: _ChannelHub.contract, event: "Withdrawn", logs: logs, sub: sub}, nil
}

// WatchWithdrawn is a free log subscription operation binding the contract event 0x7084f5476618d8e60b11ef0d7d3f06914655adb8793e28ff7f018d4c76d505d5.
//
// Solidity: event Withdrawn(address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) WatchWithdrawn(opts *bind.WatchOpts, sink chan<- *ChannelHubWithdrawn, token []common.Address) (event.Subscription, error) {

	var tokenRule []interface{}
	for _, tokenItem := range token {
		tokenRule = append(tokenRule, tokenItem)
	}

	logs, sub, err := _ChannelHub.contract.WatchLogs(opts, "Withdrawn", tokenRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ChannelHubWithdrawn)
				if err := _ChannelHub.contract.UnpackLog(event, "Withdrawn", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ParseWithdrawn is a log parse operation binding the contract event 0x7084f5476618d8e60b11ef0d7d3f06914655adb8793e28ff7f018d4c76d505d5.
//
// Solidity: event Withdrawn(address indexed token, uint256 amount)
func (_ChannelHub *ChannelHubFilterer) ParseWithdrawn(log types.Log) (*ChannelHubWithdrawn, error) {
	event := new(ChannelHubWithdrawn)
	if err := _ChannelHub.contract.UnpackLog(event, "Withdrawn", log); err != nil {
		return nil, err
	}
	event.Raw = log
	return event, nil
}
