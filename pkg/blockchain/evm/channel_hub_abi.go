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
	ABI: "[{\"type\":\"constructor\",\"inputs\":[{\"name\":\"_defaultSigValidator\",\"type\":\"address\",\"internalType\":\"contractISignatureValidator\"},{\"name\":\"_node\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"DEFAULT_SIG_VALIDATOR\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"contractISignatureValidator\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"ESCROW_DEPOSIT_UNLOCK_DELAY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MAX_DEPOSIT_ESCROW_STEPS\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"MIN_CHALLENGE_DURATION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint32\",\"internalType\":\"uint32\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"NODE\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"address\",\"internalType\":\"address\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"TRANSFER_GAS_LIMIT\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VALIDATOR_ACTIVATION_DELAY\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"VERSION\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"challengeChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"challengerIdx\",\"type\":\"uint8\",\"internalType\":\"enumParticipantIndex\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"challengeEscrowDeposit\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"challengerIdx\",\"type\":\"uint8\",\"internalType\":\"enumParticipantIndex\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"challengeEscrowWithdrawal\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"challengerSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"challengerIdx\",\"type\":\"uint8\",\"internalType\":\"enumParticipantIndex\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"checkpointChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"claimFunds\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"destination\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"closeChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"createChannel\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"depositToChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"depositToNode\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"escrowHead\",\"inputs\":[],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"finalizeEscrowDeposit\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"finalizeEscrowWithdrawal\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"finalizeMigration\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"getChannelData\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumChannelStatus\"},{\"name\":\"definition\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"lastState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpiry\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"lockedFunds\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getChannelIds\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowDepositData\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumEscrowStatus\"},{\"name\":\"unlockAt\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"challengeExpiry\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"lockedAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowDepositIds\",\"inputs\":[{\"name\":\"page\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"pageSize\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[{\"name\":\"ids\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getEscrowWithdrawalData\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}],\"outputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"status\",\"type\":\"uint8\",\"internalType\":\"enumEscrowStatus\"},{\"name\":\"challengeExpiry\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"lockedAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"initState\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNodeBalance\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getNodeValidator\",\"inputs\":[{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}],\"outputs\":[{\"name\":\"validator\",\"type\":\"address\",\"internalType\":\"contractISignatureValidator\"},{\"name\":\"registeredAt\",\"type\":\"uint64\",\"internalType\":\"uint64\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getOpenChannels\",\"inputs\":[{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"openChannels\",\"type\":\"bytes32[]\",\"internalType\":\"bytes32[]\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getReclaimBalance\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}],\"outputs\":[{\"name\":\"\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"getUnlockableEscrowDepositStats\",\"inputs\":[],\"outputs\":[{\"name\":\"count\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"totalAmount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"stateMutability\":\"view\"},{\"type\":\"function\",\"name\":\"initiateEscrowDeposit\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"initiateEscrowWithdrawal\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"initiateMigration\",\"inputs\":[{\"name\":\"def\",\"type\":\"tuple\",\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"purgeEscrowDeposits\",\"inputs\":[{\"name\":\"maxSteps\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"registerNodeValidator\",\"inputs\":[{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"validator\",\"type\":\"address\",\"internalType\":\"contractISignatureValidator\"},{\"name\":\"signature\",\"type\":\"bytes\",\"internalType\":\"bytes\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"function\",\"name\":\"withdrawFromChannel\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"outputs\":[],\"stateMutability\":\"payable\"},{\"type\":\"function\",\"name\":\"withdrawFromNode\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}],\"outputs\":[],\"stateMutability\":\"nonpayable\"},{\"type\":\"event\",\"name\":\"ChannelChallenged\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelCheckpointed\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelClosed\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"finalState\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelCreated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"user\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"definition\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structChannelDefinition\",\"components\":[{\"name\":\"challengeDuration\",\"type\":\"uint32\",\"internalType\":\"uint32\"},{\"name\":\"user\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"node\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"nonce\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"approvedSignatureValidators\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"name\":\"initialState\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelDeposited\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ChannelWithdrawn\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"candidate\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Deposited\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositChallenged\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositFinalized\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositFinalizedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositInitiated\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositInitiatedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowDepositsPurged\",\"inputs\":[{\"name\":\"purgedCount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalChallenged\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]},{\"name\":\"challengeExpireAt\",\"type\":\"uint64\",\"indexed\":false,\"internalType\":\"uint64\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalFinalized\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalFinalizedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalInitiated\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"EscrowWithdrawalInitiatedOnHome\",\"inputs\":[{\"name\":\"escrowId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"FundsClaimed\",\"inputs\":[{\"name\":\"account\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"destination\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationInFinalized\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationInInitiated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationOutFinalized\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"MigrationOutInitiated\",\"inputs\":[{\"name\":\"channelId\",\"type\":\"bytes32\",\"indexed\":true,\"internalType\":\"bytes32\"},{\"name\":\"state\",\"type\":\"tuple\",\"indexed\":false,\"internalType\":\"structState\",\"components\":[{\"name\":\"version\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"intent\",\"type\":\"uint8\",\"internalType\":\"enumStateIntent\"},{\"name\":\"metadata\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"},{\"name\":\"homeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"nonHomeLedger\",\"type\":\"tuple\",\"internalType\":\"structLedger\",\"components\":[{\"name\":\"chainId\",\"type\":\"uint64\",\"internalType\":\"uint64\"},{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"decimals\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"userAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"userNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"},{\"name\":\"nodeAllocation\",\"type\":\"uint256\",\"internalType\":\"uint256\"},{\"name\":\"nodeNetFlow\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"name\":\"userSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"},{\"name\":\"nodeSig\",\"type\":\"bytes\",\"internalType\":\"bytes\"}]}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"NodeBalanceUpdated\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"TransferFailed\",\"inputs\":[{\"name\":\"recipient\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"ValidatorRegistered\",\"inputs\":[{\"name\":\"validatorId\",\"type\":\"uint8\",\"indexed\":true,\"internalType\":\"uint8\"},{\"name\":\"validator\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"contractISignatureValidator\"}],\"anonymous\":false},{\"type\":\"event\",\"name\":\"Withdrawn\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"indexed\":true,\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"indexed\":false,\"internalType\":\"uint256\"}],\"anonymous\":false},{\"type\":\"error\",\"name\":\"AddressCollision\",\"inputs\":[{\"name\":\"collision\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ChallengerVersionTooLow\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureLength\",\"inputs\":[{\"name\":\"length\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"ECDSAInvalidSignatureS\",\"inputs\":[{\"name\":\"s\",\"type\":\"bytes32\",\"internalType\":\"bytes32\"}]},{\"type\":\"error\",\"name\":\"EmptySignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectAmount\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectChallengeDuration\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectChannelId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectChannelStatus\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectNode\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectSignature\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectStateIntent\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"IncorrectValue\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InsufficientBalance\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidAddress\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"InvalidValidatorId\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"NativeTransferFailed\",\"inputs\":[{\"name\":\"to\",\"type\":\"address\",\"internalType\":\"address\"},{\"name\":\"amount\",\"type\":\"uint256\",\"internalType\":\"uint256\"}]},{\"type\":\"error\",\"name\":\"NoChannelIdFoundForEscrow\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ReentrancyGuardReentrantCall\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"SafeCastOverflowedIntToUint\",\"inputs\":[{\"name\":\"value\",\"type\":\"int256\",\"internalType\":\"int256\"}]},{\"type\":\"error\",\"name\":\"SafeERC20FailedOperation\",\"inputs\":[{\"name\":\"token\",\"type\":\"address\",\"internalType\":\"address\"}]},{\"type\":\"error\",\"name\":\"ValidatorAlreadyRegistered\",\"inputs\":[{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}]},{\"type\":\"error\",\"name\":\"ValidatorNotActive\",\"inputs\":[{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"},{\"name\":\"activatesAt\",\"type\":\"uint64\",\"internalType\":\"uint64\"}]},{\"type\":\"error\",\"name\":\"ValidatorNotApproved\",\"inputs\":[]},{\"type\":\"error\",\"name\":\"ValidatorNotRegistered\",\"inputs\":[{\"name\":\"validatorId\",\"type\":\"uint8\",\"internalType\":\"uint8\"}]}]",
	Bin: "0x60c0346100fd57601f615a4738819003918201601f19168301916001600160401b038311848410176101015780849260409485528339810103126100fd5780516001600160a01b038116918282036100fd5760200151916001600160a01b038316908184036100fd5760017f9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f0055156100ee57156100ee5760805260a052604051615931908161011682396080518181816110400152613bba015260a051818181610c34015281816112ff01528181611e760152818161336401528181613dbb015281816142f401526143d60152f35b63e6c4247b60e01b5f5260045ffd5b5f80fd5b634e487b7160e01b5f52604160045260245ffdfe60806040526004361015610011575f80fd5b5f3560e01c806307f241ce1461026f57806316b390b11461026a578063187576d8146102655780633115f6301461026057806338a66be21461025b5780633c684f921461025657806341b660ef1461025157806347de477a1461024c57806351bfcdbd1461024757806353269198146102425780635a0745b41461023d5780635ae2accc146102385780635b9acbf9146102335780635dc46a741461022e5780636840dbd2146102295780636898234b1461022457806371a471411461021f578063735181f01461021a57806382d3e15d146102155780638d0b12a5146102105780638e31c7351461020b57806394191051146102015780639691b46814610206578063a459463114610201578063a5c82680146101fc578063b25a1d38146101f7578063b65b78d1146101f2578063c74a2d10146101ed578063c9408398146101e8578063d888ccae146101e3578063d91a1283146101de578063dc23f29e146101d9578063dd73d494146101d4578063e617208c146101cf578063f4ac51f5146101ca578063f766f8d6146101c5578063ff5bc09e146101c05763ffa1ad74146101bb575f80fd5b612491565b61247a565b61236e565b6122f3565b612255565b6120ce565b611f17565b611e0b565b611d02565b611ac4565b611a49565b611988565b611652565b6114f3565b6113ce565b6113eb565b611233565b6110ec565b6110cf565b611089565b611021565b610f35565b610f1e565b610ed8565b610eb6565b610e9b565b610e7f565b610c87565b610c15565b610aa1565b6107fa565b610734565b6106f9565b61055d565b6104d7565b610341565b610289565b6001600160a01b0381160361028557565b5f80fd5b34610285576020366003190112610285576001600160a01b036004356102ae81610274565b165f526006602052602060405f2054604051908152f35b9181601f84011215610285578235916001600160401b038311610285576020838186019501011161028557565b60643590600282101561028557565b9060606003198301126102855760043591602435906001600160401b03821161028557610330916004016102c5565b909160443560028110156102855790565b34610285576103a36103dd61035536610301565b9294916103b8610370879693965f52600260205260405f2090565b9485549261037f8415156124ac565b600187015460059060081c6001600160a01b031696879260028a01549a8b91613b99565b9192909901986103b28a6126c6565b87613ce6565b60c06103c387613e10565b604051809481926301999b9360e61b835260048301612836565b038173__$682d6198b4eca5bc7e038b912a26498e7e$__5af48015610499577fba075bd445233f7cad862c72f0343b3503aad9c8e704a2295f122b82abf8e80195610451946080945f93610466575b5082610443939461043c896126c6565b908b613e84565b01516001600160401b031690565b90610461604051928392836128a9565b0390a2005b610443935061048c9060c03d60c011610492575b6104848183612542565b810190612774565b9261042c565b503d61047a565b612847565b60206040818301928281528451809452019201905f5b8181106104c15750505090565b82518452602093840193909201916001016104b4565b34610285576020366003190112610285576001600160a01b036004356104fc81610274565b165f52600160205260405f206040519081602082549182815201915f5260205f20905f5b818110610547576105438561053781870382612542565b6040519182918261049e565b0390f35b8254845260209093019260019283019201610520565b34610285576020366003190112610285576004355f905f60035491600454925b808410806106f0575b156106e5576105bb6105b56105a761059d87612ea7565b90549060031b1c90565b5f52600260205260405f2090565b936136de565b946105c58461518c565b6106d3576105d2846151bc565b15610690575f5160206158bc5f395f51905f526001600160a01b0361067961067361065e945f610610600c8b01546001600160a01b039060401c1690565b9961066d60016106318d6001600160a01b03165f52600660205260405f2090565b54928d6106446004830195865490613066565b9b8c916001600160a01b03165f52600660205260405f2090565b5501805460ff19166003179055565b556136de565b976136de565b604051938452951691602090a25b9391929361057d565b945050505061069e90600455565b806106a557005b6040519081527f61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd14590602090a1005b936106df9193506136de565b91610687565b50505060045561069e565b50818310610586565b34610285575f366003190112610285576020604051620186a08152f35b6004359060ff8216820361028557565b359060ff8216820361028557565b346102855760203660031901126102855760ff61074f610716565b165f52600760205260405f2060405160408101918183106001600160401b038411176107ad576040928352546001600160a01b03811680835260a09190911c6001600160401b03166020928301819052835191825291810191909152f35b6124c1565b90816102609103126102855790565b90600319820160e081126102855760c0136102855760049160c435906001600160401b038211610285576107f7916004016107b2565b90565b610803366107c1565b6020810160026108128261294f565b61081b81611bac565b148015610a86575b8015610a68575b61083390612959565b6108bd6108486108433686612988565b614282565b92610855602086016129fc565b9061085f866142b7565b61086f6080870135838388614394565b60c0816108a261089b610884608084016129fc565b6001600160a01b03165f52600660205260405f2090565b54886143fb565b604051632a2d120f60e21b8152958692839260048401612c22565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af4908115610499577fb0d099feaab5034d04a1c610e86b8832343f2127b3c667b705834dafdf96e9e4946109326109b3936001600160a01b03965f91610a39575b50610921368b612988565b61092b3686612d26565b908a61452f565b61095687610951866001600160a01b03165f52600160205260405f2090565b6154d0565b5060026109628261294f565b61096b81611bac565b036109b85750857f6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f41778696206696604051806109a18582612dd2565b0390a25b604051938493169683612de3565b0390a3005b6109c360039161294f565b6109cc81611bac565b03610a0957857f188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf98660405180610a018582612dd2565b0390a26109a5565b857f567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc60405180610a018582612dd2565b610a5b915060c03d60c011610a61575b610a538183612542565b810190612a06565b5f610916565b503d610a49565b50610833610a758261294f565b610a7e81611bac565b15905061082a565b506003610a928261294f565b610a9b81611bac565b14610823565b610aaa366107c1565b90610acb6004610abc6020850161294f565b610ac581611bac565b14612959565b610ad4816142b7565b610ae16108433683612988565b916080610af0602084016129fc565b92013591610b0083828487614394565b610b12610b0c83612e72565b85614642565b92610b1c85614671565b15610b5e5750506109b381610b527f471c4ebe4e57d25ef7117e141caac31c6b98f067b8098a7a7bbd38f637c2f98093866146cd565b60405191829182612dd2565b9091610b8a60c082610b6f87613e10565b604051632ef10bcd60e21b8152938492839260048401612e7c565b038173__$682d6198b4eca5bc7e038b912a26498e7e$__5af4928315610499577fede7867afa7cdb9c443667efd8244d98bf9df1dce68e60dc94dca6605125ca76946109b394610bed935f91610bf6575b50610be63686612d26565b8989613e84565b610b5284612ef6565b610c0f915060c03d60c011610492576104848183612542565b5f610bdb565b34610285575f3660031901126102855760206040516001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000168152f35b9060406003198301126102855760043591602435906001600160401b038211610285576107f7916004016107b2565b3461028557610c9536610c58565b610ca66009610abc6020840161294f565b610cc26001610cbc845f525f60205260405f2090565b01612f3f565b610d5d610cd960208301516001600160a01b031690565b91610cea6080820151848688614394565b610cf43685612d26565b61014085019386610d0486612e72565b6001600160401b031646149586610e17575b50505060c081610d42610d3b61088460206060850151016001600160a01b0390511690565b54896143fb565b604051632a2d120f60e21b8152958692839260048401612fc9565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af491821561049957610d8f935f93610df6575b508661452f565b15610dc5576104617f9a6f675cc94b83b55f1ecc0876affd4332a30c92e6faa2aca0199b1b6df922c39160405191829182612dd2565b6104617f7b20773c41402791c5f18914dbbeacad38b1ebcc4c55d8eb3bfe0a4cde26c8269160405191829182612dd2565b610e1091935060c03d60c011610a6157610a538183612542565b915f610d88565b610e7692610e29610e71923690612c47565b6060860152610e3b3660608b01612c47565b6080860152610e48612fb5565b60a0860152610e55612fb5565b60c08601526001600160a01b03165f52600160205260405f2090565b615575565b505f8681610d16565b34610285575f366003190112610285576020604051612a308152f35b34610285575f36600319011261028557602060405160408152f35b3461028557604036600319011261028557610543610537602435600435613094565b610eea610ee436610c58565b9061314d565b005b6060600319820112610285576004359160243591604435906001600160401b038211610285576107f7916004016107b2565b3461028557610eea610f2f36610eec565b9161349f565b34610285576020366003190112610285576001600160a01b03600435610f5a81610274565b165f526001602052610f6e60405f20615444565b5f905f5b815181101561100e57610fa0610f99610f8b8385613080565b515f525f60205260405f2090565b5460ff1690565b610fa9816121a4565b60038114159081610ff9575b50610fc3575b600101610f72565b91610fd6818460019310610fde576136de565b929050610fbb565b610fe88585613080565b51610ff38286613080565b526136de565b60059150611006816121a4565b14155f610fb5565b506105439181526040519182918261049e565b34610285575f3660031901126102855760206040516001600160a01b037f0000000000000000000000000000000000000000000000000000000000000000168152f35b60409060031901126102855760043561107c81610274565b906024356107f781610274565b346102855760206110c66001600160a01b036110a436611064565b91165f526008835260405f20906001600160a01b03165f5260205260405f2090565b54604051908152f35b34610285575f366003190112610285576020600454604051908152f35b34610285576110fa36610301565b611146611112859493945f52600560205260405f2090565b918254946111218615156124ac565b60a061112c88614925565b604051809581926312031f5d60e11b8352600483016136ec565b038173__$b69fb814c294bfc16f92e50d7aeced4bde$__5af4908115610499577fb8568a1f475f3c76759a620e08a653d28348c5c09e2e0bc91d533339801fefd8966103b296610451966060965f956111f0575b50916111e0859661044396938560056111c460016111d49901546001600160a01b039060081c1690565b97889360028401549a8b91613b99565b92909193019e8f6126c6565b6111e9896126c6565b908b6149df565b61044395506111d4939192966112206111e09260a03d60a01161122c575b6112188183612542565b8101906133db565b9650969291935061119a565b503d61120e565b346102855760603660031901126102855761124c610716565b60243561125881610274565b6044356001600160401b0381116102855761136b9161127e6113a89236906004016102c5565b93909461133161132c60ff8316966112978815156136fd565b6001600160a01b038616986112ad8a1515613713565b6112ee856112e86112dc6112dc6112cf8460ff165f52600760205260405f2090565b546001600160a01b031690565b6001600160a01b031690565b15613729565b6113266112fc8b8730614aa7565b917f0000000000000000000000000000000000000000000000000000000000000000933691612cd5565b90614adf565b613747565b61134b61133c612563565b6001600160a01b039094168452565b426001600160401b0316602084015260ff165f52600760205260405f2090565b8151815460209093015167ffffffffffffffff60a01b60a09190911b166001600160e01b03199093166001600160a01b0390911617919091179055565b7f9ee792368f12db92ad66335fa19df35feaec025c86445fea202ab5412a180e055f80a3005b34610285575f366003190112610285576020604051620151808152f35b61146f6113f736610c58565b6114186114096020839594950161294f565b61141281611bac565b15612959565b61142e6001610cbc855f525f60205260405f2090565b9061145361144660208401516001600160a01b031690565b6080840151908387614394565b60c0816108a2611468610884608084016129fc565b54876143fb565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af4928315610499577f567044ba1cdd4671ac3979c114241e1e3b56c9e9051f63f2f234f7a2795019cc9361046193610b52925f926114d2575b506114cb3685612d26565b908761452f565b6114ec91925060c03d60c011610a6157610a538183612542565b905f6114c0565b3461028557611501366107c1565b906115136006610abc6020850161294f565b61151c816142b7565b6115296108433683612988565b916080611538602084016129fc565b9201359161154883828487614394565b611554610b0c83612e72565b9261155e85614671565b156115945750506109b381610b527f587faad1bcd589ce902468251883e1976a645af8563c773eed7356d78433210c93866146cd565b90916115d060a0826115b66115af61088461016084016129fc565b5488614982565b60405162ea54e760e01b8152938492839260048401613488565b038173__$b69fb814c294bfc16f92e50d7aeced4bde$__5af4928315610499577f17eb0a6bd5a0de45d1029ce3444941070e149df35b22176fc439f930f73c09f7946109b394610b52935f91611633575b5061162c3686612d26565b89896149df565b61164c915060a03d60a01161122c576112188183612542565b5f611621565b6080366003190112610285576004356024356001600160401b038111610285576116809036906004016107b2565b6044356001600160401b0381116102855761169f9036906004016102c5565b91906116a96102f2565b926116bb855f525f60205260405f2090565b9185846116ca60018601612f3f565b936116d6865460ff1690565b936116e0856121a4565b60018514808015611975575b6116f59061375d565b611701600589016126c6565b9561173f61170e86612e72565b6001600160401b0361173661172a8b516001600160401b031690565b6001600160401b031690565b91161015613773565b60208801516001600160a01b0316966001600160401b0361177861172a61176a60808d015199612e72565b93516001600160401b031690565b911611611835575b5050946117db7f07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a998998966117d56118089760149c6117c96117f9996117f0996118269e613b99565b93919490923690612d26565b90613ce6565b845460ff191660021785555163ffffffff1690565b63ffffffff1690565b6001600160401b0342166137a9565b9301805467ffffffffffffffff19166001600160401b038516179055565b610461604051928392836137c9565b6118a69495509061187e9161187160208c9b999c9a959a019161186c600161185c8561294f565b61186581611bac565b1415612959565b6121a4565b81611958575b5015612959565b61188a8489898d614394565b60c0876108a261189f610884608084016129fc565b548d6143fb565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af4918215610499577f07b9206d5a6026d3bd2a8f9a9b79f6fa4bfbd6a016975829fbaf07488019f28a996014996117d58d8b6117c96118089a6117db976118269e6119236117f09c6117f99e5f91611939575b5061191c3688612d26565b8d89614ed0565b999e509950995050509750509698995099611780565b611952915060c03d60c011610a6157610a538183612542565b5f611911565b600991506119659061294f565b61196e81611bac565b145f611877565b5061197f866121a4565b600486146116ec565b6040366003190112610285576004356119a081610274565b602435906119af8215156137f0565b6001600160a01b03811691825f52600660205260405f2054818101809111611a4457837f2da466a7b24304f47e87fa2e1e5a81b9831ce54fec19055ce277ca2f39ba42c4611a3184611a21610461965f5160206158bc5f395f51905f5298865f5260066020528760405f2055336150a2565b6040519081529081906020820190565b0390a26040519081529081906020820190565b612fee565b611a69611a5536610c58565b6114186003610abc6020849695960161294f565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af4928315610499577f188e0ade7d115cc397426774adb960ae3e8c83e72f0a6cad4b7085e1d60bf9869361046193610b52925f926114d257506114cb3685612d26565b34610285575f36600319011261028557600354600454905f805b82841015611b80577fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b8401545f90815260026020526040902091611b218361518c565b611b6e57611b2e836151bc565b15611b5757611b4e916004611b456105b5936136de565b94015490613066565b915b9192611ade565b92509250505b604080519182526020820192909252f35b915092611b7a906136de565b91611b50565b92509050611b5d565b634e487b7160e01b5f52602160045260245ffd5b60041115611ba757565b611b89565b600a1115611ba757565b90600a821015611ba75752565b60c080916001600160401b0381511684526001600160a01b03602082015116602085015260ff6040820151166040850152606081015160608501526080810151608085015260a081015160a08501520151910152565b805180835260209291819084018484015e5f828201840152601f01601f1916010190565b6107f7916001600160401b038251168152611c6060208301516020830190611bb6565b60408201516040820152611c7c60608301516060830190611bc3565b611c8f6080830151610140830190611bc3565b60c0611cad60a0840151610260610220850152610260840190611c19565b92015190610240818403910152611c19565b92936001600160401b0360c0956107f798979482948752611cdf81611b9d565b602087015216604085015216606083015260808201528160a08201520190611c3d565b3461028557602036600319011261028557600435611d1e61383c565b505f52600260205260405f2060405190611d37826124d5565b80548252610543600182015491611d82611d72611d548560ff1690565b94611d63602088019687613880565b60081c6001600160a01b031690565b6001600160a01b03166040860152565b6002810154606085015260038101546001600160401b0380821660808701908152959160401c166001600160401b031660a0820190815291611dfa61176a611dd8600560048501549460c08701958652016126c6565b9360e0810194855251965197611ded89611b9d565b516001600160401b031690565b905191519260405196879687611cbf565b3461028557606036600319011261028557600435611e2881610274565b5f5160206158bc5f395f51905f5261046160243592611e4684610274565b60443593611e5e6001600160a01b0383161515613713565b611e698515156137f0565b611e9d6001600160a01b037f000000000000000000000000000000000000000000000000000000000000000016331461388c565b7f7084f5476618d8e60b11ef0d7d3f06914655adb8793e28ff7f018d4c76d505d5611a3186611a216001600160a01b038516988995865f526006602052611ef48260405f2054611eef828210156138a2565b613073565b9788611f11836001600160a01b03165f52600660205260405f2090565b556151eb565b3461028557611f25366107c1565b611f366008610abc6020840161294f565b611f436108433684612988565b91611fa4611f53602083016129fc565b91611f646080820135848688614394565b611f6e3685612d26565b611f7786614671565b9386851561206d575b505060c081610d42610d3b61088460206060850151016001600160a01b0390511690565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af491821561049957611fe1935f93612048575b50611fdb903690612988565b8661452f565b15612017576104617f3142fb397e715d80415dff7b527bf1c451def4675da6e1199ee1b4588e3f630a9160405191829182612dd2565b6104617f26afbcb9eb52c21f42eb9cfe8f263718ffb65afbf84abe8ad8cce2acfb2242b89160405191829182612dd2565b611fdb9193506120669060c03d60c011610a6157610a538183612542565b9290611fcf565b61095161208b9261207d866142b7565b610e29366101408b01612c47565b505f86611f80565b9160a0936001600160401b03916107f797969385526120b181611b9d565b602085015216604083015260608201528160808201520190611c3d565b34610285576020366003190112610285576004356120ea61383c565b505f52600560205260405f2060405190612103826124f1565b8054825261054360018201549161213a611d7260ff851694602087019561212981611b9d565b865260081c6001600160a01b031690565b6002810154606085015260038101546001600160401b03166001600160401b0316608085019081529361219361217e600560048501549460a08501958652016126c6565b9160c0810192835251945195611ded87611b9d565b915190519160405195869586612093565b60061115611ba757565b906006821015611ba75752565b60a0809163ffffffff81511684526001600160a01b0360208201511660208501526001600160a01b0360408201511660408501526001600160401b036060820151166060850152608081015160808501520151910152565b91926122376101209461222d8561224a959a99989a6121ae565b60208501906121bb565b61014060e0840152610140830190611c3d565b946101008201520152565b34610285576020366003190112610285576004355f60a06040516122788161250c565b828152826020820152826040820152826060820152826080820152015261229d61383c565b505f525f6020526122b060405f206138c4565b80516122bb816121a4565b61054360208301519260408101519060606122e361172a60808401516001600160401b031690565b9101519160405195869586612213565b6123136122ff36610c58565b6114186002610abc6020849695960161294f565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af4928315610499577f6085f5128b19e0d3cc37524413de47259383f0f75265d5d66f417786962066969361046193610b52925f926114d257506114cb3685612d26565b346102855761237c36611064565b612384615237565b6001600160a01b0381169161239a831515613713565b6001600160a01b036123d7826123c1336001600160a01b03165f52600860205260405f2090565b906001600160a01b03165f5260205260405f2090565b54916123e48315156137f0565b5f612404826123c1336001600160a01b03165f52600860205260405f2090565b551691818361246b57612427915f808080858a5af1612421613921565b50613950565b60405190815233907f7b8d70738154be94a9a068a6d2f5dd8cfc65c52855859dc8f47de1ff185f8b5590602090a4610eea60015f5160206158dc5f395f51905f5255565b612475918461526f565b612427565b3461028557610eea61248b36610eec565b91613978565b34610285575f36600319011261028557602060405160018152f35b156124b357565b6287a33760e41b5f5260045ffd5b634e487b7160e01b5f52604160045260245ffd5b61010081019081106001600160401b038211176107ad57604052565b60e081019081106001600160401b038211176107ad57604052565b60c081019081106001600160401b038211176107ad57604052565b60a081019081106001600160401b038211176107ad57604052565b90601f801991011681019081106001600160401b038211176107ad57604052565b60405190612572604083612542565b565b6040519061257260e083612542565b90604051612590816124f1565b60c0600482946125cd60ff82546001600160401b03811687526001600160a01b03808260401c1616602088015260e01c16604086019060ff169052565b6001810154606085015260028101546080850152600381015460a08501520154910152565b90600182811c92168015612620575b602083101461260c57565b634e487b7160e01b5f52602260045260245ffd5b91607f1691612601565b5f9291815491612639836125f2565b808352926001811690811561268e575060011461265557505050565b5f9081526020812093945091925b838310612674575060209250010190565b600181602092949394548385870101520191019190612663565b915050602093945060ff929192191683830152151560051b010190565b906125726126bf926040519384809261262a565b0383612542565b906040516126d3816124f1565b809260ff81546001600160401b038116845260401c1690600a821015611ba757600d6127449160c09360208601526001810154604086015261271760028201612583565b606086015261272860078201612583565b6080860152612739600c82016126ab565b60a0860152016126ab565b910152565b5190600482101561028557565b6001600160401b0381160361028557565b5190811515820361028557565b908160c0910312610285576127dc60a0604051926127918461250c565b80518452602081015160208501526127ab60408201612749565b604085015260608101516127be81612756565b606085015260808101516127d181612756565b608085015201612767565b60a082015290565b9081516127f081611b9d565b815260806001600160401b0381612816602086015160a0602087015260a0860190611c3d565b946040810151604086015282606082015116606086015201511691015290565b9060206107f79281815201906127e4565b6040513d5f823e3d90fd5b600460c09160ff8082546001600160401b03811687526001600160a01b038160401c16602088015260e01c161660408501526001810154606085015260028101546080850152600381015460a08501520154910152565b9291602061293561257293604087526128dc81546001600160401b03811660408a015260ff60608a019160401c16611bb6565b600181015460808801526128f660a0880160028301612852565b612907610180880160078301612852565b61026080880152600d6129216102a08901600c840161262a565b888103603f19016102808a0152910161262a565b9401906001600160401b03169052565b600a111561028557565b356107f781612945565b1561296057565b633226144f60e21b5f5260045ffd5b63ffffffff81160361028557565b359061257282612756565b91908260c0910312610285576040516129a08161250c565b60a080829480356129b08161296f565b845260208101356129c081610274565b602085015260408101356129d381610274565b604085015260608101356129e681612756565b6060850152608081013560808501520135910152565b356107f781610274565b908160c09103126102855760405190612a1e8261250c565b805182526020810151602083015260408101516006811015610285576127dc9160a09160408501526060810151612a5481612756565b60608501526127d160808201612767565b90612a718183516121ae565b60806001600160401b0381612a95602086015160a0602087015260a0860190611c3d565b94604081015160408601526060810151606086015201511691015290565b359061257282612945565b60c080916001600160401b038135612ad581612756565b1684526001600160a01b036020820135612aee81610274565b16602085015260ff612b0260408301610726565b166040850152606081013560608501526080810135608085015260a081013560a08501520135910152565b9035601e19823603018112156102855701602081359101916001600160401b03821161028557813603831361028557565b908060209392818452848401375f828201840152601f01601f1916010190565b6107f7916001600160401b038235612b9581612756565b168152612bb36020830135612ba981612945565b6020830190611bb6565b60408201356040820152612bcd6060820160608401612abe565b612bdf61014082016101408401612abe565b612c13612c07612bf3610220850185612b2d565b610260610220860152610260850191612b5e565b92610240810190612b2d565b91610240818503910152612b5e565b9091612c396107f793604084526040840190612a65565b916020818403910152612b7e565b91908260e091031261028557604051612c5f816124f1565b60c08082948035612c6f81612756565b84526020810135612c7f81610274565b6020850152612c9060408201610726565b6040850152606081013560608501526080810135608085015260a081013560a08501520135910152565b6001600160401b0381116107ad57601f01601f191660200190565b929192612ce182612cba565b91612cef6040519384612542565b829481845281830111610285578281602093845f960137010152565b9080601f83011215610285578160206107f793359101612cd5565b9190916102608184031261028557612d3c612574565b92612d468261297d565b8452612d5460208301612ab3565b602085015260408201356040850152612d708160608401612c47565b6060850152612d83816101408401612c47565b60808501526102208201356001600160401b0381116102855781612da8918401612d0b565b60a08501526102408201356001600160401b03811161028557612dcb9201612d0b565b60c0830152565b9060206107f7928181520190612b7e565b60e09060a06107f7949363ffffffff8135612dfd8161296f565b1683526001600160a01b036020820135612e1681610274565b1660208401526001600160a01b036040820135612e3281610274565b1660408401526001600160401b036060820135612e4e81612756565b16606084015260808101356080840152013560a08201528160c08201520190612b7e565b356107f781612756565b9091612c396107f7936040845260408401906127e4565b634e487b7160e01b5f52603260045260245ffd5b600354811015612ebf5760035f5260205f2001905f90565b612e93565b8054821015612ebf575f5260205f2001905f90565b91612ef29183549060031b91821b915f19901b19161790565b9055565b600354600160401b8110156107ad5760018101600355600354811015612ebf5760035f527fc2575a0e9e593c00f959f8c92f12db2869c3395a3b0502d05e2516446f71f85b0155565b90604051612f4c8161250c565b60a0600382946001600160a01b03815463ffffffff8116865260201c166020850152612fa46001600160401b0360018301546001600160a01b03808216166040880152851c1660608601906001600160401b03169052565b600281015460808501520154910152565b60405190612fc4602083612542565b5f8252565b9091612fe06107f793604084526040840190612a65565b916020818403910152611c3d565b634e487b7160e01b5f52601160045260245ffd5b6001600160401b0381116107ad5760051b60200190565b60405190613028602083612542565b5f808352366020840137565b9061303e82613002565b61304b6040519182612542565b828152809261305c601f1991613002565b0190602036910137565b91908201809211611a4457565b91908203918211611a4457565b8051821015612ebf5760209160051b010190565b91906003549080840293808504821490151715611a44578184101561311857830190818411611a4457808211613110575b506130d86130d38483613073565b613034565b92805b8281106130e757505050565b806130f661059d600193612ea7565b6131096131038584613073565b88613080565b52016130db565b90505f6130c5565b505090506107f7613019565b906006811015611ba75760ff80198354169116179055565b9060206107f7928181520190611c3d565b9061315f825f525f60205260405f2090565b61316b60018201612f3f565b91613177825460ff1690565b9184613185600583016126c6565b91600261319c60208801516001600160a01b031690565b956131a6816121a4565b1480613395575b6132bc575050506131c56001610abc6020840161294f565b6131d56080840151838387614394565b61320860c0826131ed61089b610884608084016129fc565b604051632a2d120f60e21b8152938492839260048401612c22565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af4801561049957610e716132969461327288937f04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a898613289965f9261329b575b5061326b3689612d26565b908661452f565b6001600160a01b03165f52600160205260405f2090565b5060405191829182612dd2565b0390a2565b6132b591925060c03d60c011610a6157610a538183612542565b905f613260565b7f04cd8c68bf83e7bc531ca5a5d75c34e36513c2acf81e07e6470ba79e29da13a895506133889293506132969461331b601483613303610e7195600360ff19825416179055565b5f601382015501805467ffffffffffffffff19169055565b613272606086016133478151606061333d60208301516001600160a01b031690565b9101519085614781565b5160a061335e60208301516001600160a01b031690565b910151907f0000000000000000000000000000000000000000000000000000000000000000614781565b506040519182918261313c565b506014810154426001600160401b03909116106131ad565b156133b457565b6336c7a86b60e21b5f5260045ffd5b906133cd81611b9d565b60ff80198354169116179055565b908160a0910312610285576134306080604051926133f884612527565b805184526020810151602085015261341260408201612749565b6040850152606081015161342581612756565b606085015201612767565b608082015290565b90815161344481611b9d565b8152608080613462602085015160a0602086015260a0850190611c3d565b93604081015160408501526001600160401b036060820151166060850152015191015290565b9091612c396107f793604084526040840190613438565b916134b2825f52600560205260405f2090565b906134bd83856148d8565b613686576134cd848354146133ad565b600182018054929060026134f0600886901c6001600160a01b03165b9560ff1690565b6134f981611b9d565b148061366e575b61359757506002906135196007610abc6020860161294f565b01549061352882848388614394565b61353760a0826115b687614925565b038173__$b69fb814c294bfc16f92e50d7aeced4bde$__5af4928315610499577f2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d19461359294610b52935f91611633575061162c3686612d26565b0390a3565b805460ff191660031790557f2fdac1380dbe23ae259b6871582b7f33e34461547f400bdd20d74991250317d1925061359291905f5160206158bc5f395f51905f526001600160a01b0361363c61361a600c60048601955f87549755613609600382016001600160401b03198154169055565b015460401c6001600160a01b031690565b93613636856001600160a01b03165f52600660205260405f2090565b54613066565b9283613659826001600160a01b03165f52600660205260405f2090565b556040519384521691602090a2610b5261417c565b506003820154426001600160401b0390911610613500565b7f6d0cf3d243d63f08f50db493a8af34b27d4e3bc9ec4098e82700abfeffe2d498915061359290610b526136d760016136c6885f525f60205260405f2090565b015460201c6001600160a01b031690565b82876148fa565b5f198114611a445760010190565b9060206107f7928181520190613438565b1561370457565b6306ee4dcd60e01b5f5260045ffd5b1561371a57565b63e6c4247b60e01b5f5260045ffd5b156137315750565b60ff906357470ffd60e01b5f521660045260245ffd5b1561374e57565b63c1606c2f60e01b5f5260045ffd5b1561376457565b631e40ad6360e31b5f5260045ffd5b1561377a57565b637d95736160e01b5f5260045ffd5b6001600160401b0362015180911601906001600160401b038211611a4457565b906001600160401b03809116911601906001600160401b038211611a4457565b906001600160401b036137e9602092959495604085526040850190612b7e565b9416910152565b156137f757565b6334b2073960e11b5f5260045ffd5b60405190613813826124f1565b5f60c0838281528260208201528260408201528260608201528260808201528260a08201520152565b60405190613849826124f1565b606060c0835f81525f60208201525f6040820152613865613806565b83820152613871613806565b60808201528260a08201520152565b61388982611b9d565b52565b1561389357565b6308ad910960e21b5f5260045ffd5b156138a957565b631e9acf1760e31b5f5260045ffd5b6006821015611ba75752565b906040516138d181612527565b60806001600160401b03601483956138ed60ff825416866138b8565b6138f960018201612f3f565b602086015261390a600582016126c6565b604086015260138101546060860152015416910152565b3d1561394b573d9061393282612cba565b916139406040519384612542565b82523d5f602084013e565b606090565b15613959575050565b6001600160a01b039063296c17bb60e21b5f521660045260245260445ffd5b9161398b825f52600260205260405f2090565b9061399683856152c8565b613af9576139a6848354146133ad565b600182018054929060026139c6600886901c6001600160a01b03166134e9565b6139cf81611b9d565b1480613ad6575b613a6857506002906139ef6005610abc6020860161294f565b0154906139fe82848388614394565b613a0d60c082610b6f87613e10565b038173__$682d6198b4eca5bc7e038b912a26498e7e$__5af4928315610499577f1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e9461359294610b52935f91610bf65750610be63686612d26565b805460ff191660031790557f1b92e8ef67d8a7c0d29c99efcd180a5e0d98d60ac41d52abbbb5950882c78e4e926135929291613ace91613ac8600c60048401935f855495556136096003820167ffffffffffffffff60401b198154169055565b90614781565b610b5261417c565b50600382015460401c6001600160401b03166001600160401b03429116106139d6565b7f32e24720f56fd5a7f4cb219d7ff3278ae95196e79c85b5801395894a6f53466c915061359290610b526136d760016136c6885f525f60205260405f2090565b15613b4057565b6306a41ced60e21b5f5260045ffd5b15613b575750565b60ff9063399eb60560e01b5f521660045260245ffd5b15613b76575050565b9060ff6001600160401b039263975133f360e01b5f52166004521660245260445ffd5b9291908015613c73578015612ebf57613be891843560f81c9081613bec57507f000000000000000000000000000000000000000000000000000000000000000094600101925f19909201919050565b9091565b600180613bff84613c06949060ff161c90565b1614613b39565b613c66613c1e8260ff165f52600760205260405f2090565b546001600160a01b0381169290613c5390613c4e90613c3f84871515613b4f565b60a01c6001600160401b031690565b613789565b906001600160401b038216421015613b6d565b93600101915f1990910190565b63ac241e1160e01b5f5260045ffd5b805191908290602001825e015f815290565b60021115611ba757565b90816020910312610285575190565b9392606093613cd86001600160a01b03946137e9949998998852608060208901526080880190611c19565b918683036040880152612b5e565b602095613d1693959497613d39613d046001600160a01b03956152e0565b613d2b6040519788928c840190613c82565b686368616c6c656e676560b81b815260090190565b03601f198101875286612542565b613d4281613c94565b613db557613d68905b60405163600109bb60e01b81529889978896879560048701613cad565b0392165afa801561049957612572915f91613d86575b501515613747565b613da8915060203d602011613dae575b613da08183612542565b810190613c9e565b5f613d7e565b503d613d96565b50613d687f0000000000000000000000000000000000000000000000000000000000000000613d4b565b60405190613dec82612527565b5f608083828152613dfb61383c565b60208201528260408201528260608201520152565b613e18613ddf565b905f5260026020526001600160401b0380600360405f2060ff600182015416613e4081611b9d565b8552613e4e600582016126c6565b6020860152600481015460408601520154818116606085015260401c1616608082015290565b600160ff1b8114611a44575f0390565b6020939291613f1891613e9f815f52600260205260405f2090565b97604086018051613eaf81611b9d565b613eb881611b9d565b61415f575b50878560a0880194613ecf8651151590565b61414c575b5050505050613eed60608501516001600160401b031690565b6001600160401b038116614123575b5060808401516001600160401b0316806140ed575b5051151590565b156140d457608001518201516001600160a01b031680935b8251905f82131561409457613f529150613f4a8451615428565b9283916150a2565b613f6160048601918254613066565b90555b0180515f811315613ff957505f5160206158bc5f395f51905f5291613f916001600160a01b039251615428565b613fe26004613fbb83613fb5866001600160a01b03165f52600660205260405f2090565b54613073565b9687613fd8866001600160a01b03165f52600660205260405f2090565b5501918254613066565b90556040519384521691602090a25b61257261417c565b90505f811261400b575b505050613ff1565b5f5160206158bc5f395f51905f529161403361402e6001600160a01b0393613e74565b615428565b61407e600461405783613636866001600160a01b03165f52600660205260405f2090565b9687614074866001600160a01b03165f52600660205260405f2090565b5501918254613073565b90556040519384521691602090a25f8080614003565b5f82126140a4575b505050613f64565b6140b361402e6140bb93613e74565b928391614781565b6140ca60048601918254613073565b9055825f8061409c565b50600c84015460401c6001600160a01b03168093613f30565b61411d90600389019067ffffffffffffffff60401b82549160401b169067ffffffffffffffff60401b1916179055565b5f613f11565b6141469060038901906001600160401b03166001600160401b0319825416179055565b5f613efc565b61415594615352565b5f80878582613ed4565b614176905161416d81611b9d565b60018b016133c3565b5f613ebd565b5f905f60035491600454925b80841080614278575b1561426b576141a86105b56105a761059d87612ea7565b946141b28461518c565b614259576141bf846151bc565b15614214575f5160206158bc5f395f51905f526001600160a01b036141fd61067361065e945f610610600c8b01546001600160a01b039060401c1690565b604051938452951691602090a25b93919293614188565b93919450506142239150600455565b8061422b5750565b6040519081527f61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd14590602090a1565b936142659193506136de565b9161420b565b5092916142239150600455565b5060408310614191565b6040516142936020820180936121bb565b60c081526142a260e082612542565b5190206001600160f81b0316600160f81b1790565b6001600160a01b0360208201356142cd81610274565b166142d9811515613713565b6001600160a01b0361431e60408401356142f281610274565b7f000000000000000000000000000000000000000000000000000000000000000083169216821461388c565b8114614350575063ffffffff6201518091356143398161296f565b161061434157565b630596b15b60e01b5f5260045ffd5b63abfa558d60e01b5f5260045260245ffd5b903590601e198136030182121561028557018035906001600160401b0382116102855760200191813603831361028557565b9091612572936143c46143d2926143b9836143b3610220890189614362565b90613b99565b90888894939461548c565b6143b3610240850185614362565b91937f00000000000000000000000000000000000000000000000000000000000000009361548c565b9060146001600160401b039161440f613ddf565b935f525f60205260405f209061442960ff835416866138b8565b614435600583016126c6565b6020860152601382015460408601526060850152015416608082015290565b9060a060039163ffffffff81511663ffffffff198554161784556001600160a01b03602082015116640100000000600160c01b0385549160201b1690640100000000600160c01b03191617845561451e600185016144e86144bf60408501516001600160a01b031690565b825473ffffffffffffffffffffffffffffffffffffffff19166001600160a01b03909116178255565b60608301516001600160401b0316815467ffffffffffffffff60a01b191660a09190911b67ffffffffffffffff60a01b16179055565b608081015160028501550151910155565b9261456b816145ba9460a09461454c885f525f60205260405f2090565b97614558895460ff1690565b614561816121a4565b1561463057614ed0565b60408101805161457a816121a4565b614583816121a4565b151580614605575b6145eb575b5060148401805460608301516001600160401b0390811691168190036145c9575b50500151151590565b6145c15750565b60135f910155565b815467ffffffffffffffff19166001600160401b039091161790555f806145b1565b6145ff90516145f9816121a4565b85613124565b5f614590565b50845460ff16815190614617826121a4565b614620826121a4565b614629816121a4565b141561458b565b61463d8260018b01614454565b614ed0565b906001600160401b0360405191602083019384521660408201526040815261466b606082612542565b51902090565b805f525f60205260ff60405f2054166006811015611ba75780159081156146b9575b506146b4575f525f6020526001600160401b03600760405f20015416461490565b505f90565b600591506146c6816121a4565b145f614693565b9061471f91805f525f6020526146e8600160405f2001612f3f565b60c0836147046146fd610884608084016129fc565b54856143fb565b604051632a2d120f60e21b8152968792839260048401612c22565b038173__$c00a153e45d4e7ce60e0acf48b0547b51a$__5af492831561049957612572945f9461475c575b50614756903690612d26565b9161452f565b61475691945061477a9060c03d60c011610a6157610a538183612542565b939061474a565b90614794929161478f615237565b6147a7565b60015f5160206158dc5f395f51905f5255565b91909181156148d3576001600160a01b038316928361484b576001600160a01b038216925f8080808488620186a0f16147de613921565b50156147eb575050505050565b61482e613592926123c17fbf182be802245e8ed88e4b8d3e4344c0863dd2a70334f089fd07265389306fcf956001600160a01b03165f52600860205260405f2090565b614839828254613066565b90556040519081529081906020820190565b61485d61485984848461561b565b1590565b614868575b50505050565b816148b16001600160a01b03926123c17fbf182be802245e8ed88e4b8d3e4344c0863dd2a70334f089fd07265389306fcf956001600160a01b03165f52600860205260405f2090565b6148bc858254613066565b90556040519384521691602090a35f808080614862565b505050565b905f52600560205260405f20541590816148f0575090565b6107f79150614671565b61471f926146e86149176001610cbc855f525f60205260405f2090565b916080830151908585614394565b61492d613ddf565b905f5260056020526001600160401b03600360405f2060ff60018201541661495481611b9d565b8452614962600582016126c6565b60208501526004810154604085015201541660608201525f608082015290565b9061498b613ddf565b915f5260056020526001600160401b03600360405f2060ff6001820154166149b281611b9d565b85526149c0600582016126c6565b6020860152600481015460408601520154166060830152608082015290565b6020939291613f18916149fa815f52600560205260405f2090565b97604086018051614a0a81611b9d565b614a1381611b9d565b614a93575b5087856080880194614a2a8651151590565b614a80575b5050505050614a4860608501516001600160401b031690565b6001600160401b038116614a5d575051151590565b61411d9060038901906001600160401b03166001600160401b0319825416179055565b614a89946156af565b5f80878582614a2f565b614aa1905161416d81611b9d565b5f614a18565b9160ff6001600160a01b03928360405195466020880152166040860152166060840152166080820152608081526107f760a082612542565b805192835f9472184f03e93ff9f4daa797ed6e38ed64bf6a1f0160401b821015614c69575b806d04ee2d6d415b85acef8100000000600a921015614c4d575b662386f26fc10000811015614c38575b6305f5e100811015614c26575b612710811015614c16575b6064811015614c07575b1015614bfc575b614b936021614b686001880161575b565b968701015b5f1901916f181899199a1a9b1b9c1cb0b131b232b360811b600a82061a8353600a900490565b908115614ba357614b9390614b6d565b50506001600160a01b03614bc884614bbc8584986156ef565b60208151910120615745565b911693168314614bf457614be6918160206112dc9351910120615745565b14614bef575f90565b600190565b505050600190565b600190940193614b57565b60029060649004960195614b50565b6004906127109004960195614b46565b6008906305f5e1009004960195614b3b565b601090662386f26fc100009004960195614b2e565b6020906d04ee2d6d415b85acef81000000009004960195614b1e565b506040945072184f03e93ff9f4daa797ed6e38ed64bf6a1f0160401b8104614b04565b90600a811015611ba75760ff60401b82549160401b169060ff60401b1916179055565b8151815460208401516040808601516001600160401b039094166001600160e81b031990931692909217911b7bffffffffffffffffffffffffffffffffffffffff0000000000000000161760e09190911b60ff60e01b16178155606082015160018201556080820151600282015560a0820151600382015560c090910151600490910155565b601f8211614d4257505050565b5f5260205f20906020601f840160051c83019310614d7a575b601f0160051c01905b818110614d6f575050565b5f8155600101614d64565b9091508190614d5b565b91909182516001600160401b0381116107ad57614dab81614da584546125f2565b84614d35565b6020601f8211600114614de6578190612ef29394955f92614ddb575b50508160011b915f199060031b1c19161790565b015190505f80614dc7565b601f19821690614df9845f5260205f2090565b915f5b818110614e3357509583600195969710614e1b575b505050811b019055565b01515f1960f88460031b161c191690555f8080614e11565b9192602060018192868b015181550194019201614dfc565b8151815467ffffffffffffffff19166001600160401b0391909116178155602082015191600a831015611ba75760c0600d91614e8a6125729585614c8c565b60408101516001850155614ea5606082015160028601614caf565b614eb6608082015160078601614caf565b614ec760a0820151600c8601614d84565b01519101614d84565b60206060614f1082614eef614f21959896985f525f60205260405f2090565b97614efd6080880151151590565b615090575b01516001600160a01b031690565b94015101516001600160a01b031690565b809282515f8113615065575b50602083019283515f8113614fe4575b5051905f8212614fbc575b505050515f8112614f5f575b50505061257261417c565b5f5160206158bc5f395f51905f5291614f8261402e6001600160a01b0393613e74565b614fa6601361405783613636866001600160a01b03165f52600660205260405f2090565b90556040519384521691602090a25f8080614f54565b6140b361402e614fcb93613e74565b614fda60138501918254613073565b9055815f80614f48565b614fed90615428565b61500c81613fb5866001600160a01b03165f52600660205260405f2090565b9081615029866001600160a01b03165f52600660205260405f2090565b5561503960138901918254613066565b90556040519081526001600160a01b038416905f5160206158bc5f395f51905f5290602090a25f614f3d565b61506e90615428565b6150798184846150a2565b61508860138701918254613066565b90555f614f2d565b61509d8860058b01614e4b565b614f02565b9061479492916150b0615237565b6150cb565b156150bc57565b636956f2ab60e11b5f5260045ffd5b9082156148d3576001600160a01b0316918215801561517d576150ef8234146150b5565b156150f957505050565b6001600160a01b03604051926323b872dd60e01b5f52166004523060245260445260205f60648180865af160015f511481161561515e575b6040919091525f606052156151435750565b635274afe760e01b5f526001600160a01b031660045260245ffd5b6001811516615174573d15833b15151616615131565b503d5f823e3d90fd5b61518734156150b5565b6150ef565b6001015460ff1661519c81611b9d565b600381149081156151ab575090565b600291506151b881611b9d565b1490565b6001600160401b0360038201541642101590816151d7575090565b600180925060ff910154166151b881611b9d565b9061479492916151f9615237565b919081156148d3576001600160a01b0316918261522e5761257292505f808080856001600160a01b0386165af1612421613921565b6125729261526f565b60025f5160206158dc5f395f51905f5254146152605760025f5160206158dc5f395f51905f5255565b633ee5aeb560e01b5f5260045ffd5b916001600160a01b036040519263a9059cbb60e01b5f521660045260245260205f60448180865af160015f51148116156152b2575b604091909152156151435750565b6001811516615174573d15833b151516166152a4565b905f52600260205260405f20541590816148f0575090565b6001600160401b03815116906020810151600a811015611ba75761533682604061534194015161532760806060840151930151946040519760208901526040880190611bb6565b60608601526080850190611bc3565b610160830190611bc3565b61022081526107f761024082612542565b9190915f52600260205260405f2091825560058201926153926001600160401b0383511685906001600160401b03166001600160401b0319825416179055565b602082015193600a851015611ba75760c0615424936153b66002976153fe94614c8c565b604081015160068701556153d1606082015160078801614caf565b6153e26080820151600c8801614caf565b6153f360a082015160118801614d84565b015160128501614d84565b6001830190610100600160a81b0382549160081b1690610100600160a81b031916179055565b0155565b5f81126154325790565b635467221960e11b5f5260045260245ffd5b90604051918281549182825260208201905f5260205f20925f5b81811061547357505061257292500383612542565b845483526001948501948794506020909301920161545e565b6001600160a01b0390613d686154b26154ad60209895999697993690612d26565b6152e0565b936040519889978896879563600109bb60e01b875260048701613cad565b6001810190825f528160205260405f2054155f14615533578054600160401b8110156107ad5761552061550a826001879401855584612ec4565b819391549060031b91821b915f19901b19161790565b905554915f5260205260405f2055600190565b5050505f90565b80548015615561575f1901906155508282612ec4565b8154905f199060031b1b1916905555565b634e487b7160e01b5f52603160045260245ffd5b6001810191805f528260205260405f2054928315155f14615613575f198401848111611a445783545f19810194908511611a44575f9585836155d0976155c395036155d6575b50505061553a565b905f5260205260405f2090565b55600190565b6155fc6155f6916155ed61059d61560a9588612ec4565b92839187612ec4565b90612ed9565b85905f5260205260405f2090565b555f80806155bb565b505050505f90565b60405163a9059cbb60e01b602082019081526001600160a01b03939093166024820152604480820194909452928352915f91829161565a606482612542565b51908285620186a0f19061566c613921565b91156156a95781519081156156a05750602081101561568b5750505f90565b81602091810103126102855760200151151590565b9150503b151590565b50505f90565b9190915f52600560205260405f2091825560058201926153926001600160401b0383511685906001600160401b03166001600160401b0319825416179055565b6125729061573761573194936040519586937f19457468657265756d205369676e6564204d6573736167653a0a0000000000006020860152603a850190613c82565b90613c82565b03601f198101845283612542565b6107f79161575291615783565b909291926157bd565b9061576582612cba565b6157726040519182612542565b828152809261305c601f1991612cba565b81519190604183036157b3576157ac9250602082015190606060408401519301515f1a90615839565b9192909190565b50505f9160029190565b6157c681611b9d565b806157cf575050565b6157d881611b9d565b600181036157ef5763f645eedf60e01b5f5260045ffd5b6157f881611b9d565b60028103615813575063fce698f760e01b5f5260045260245ffd5b8061581f600392611b9d565b146158275750565b6335e2f38360e21b5f5260045260245ffd5b91907f7fffffffffffffffffffffffffffffff5d576e7357a4501ddfe92f46681b20a084116158b0579160209360809260ff5f9560405194855216868401526040830152606082015282805260015afa15610499575f516001600160a01b038116156158a657905f905f90565b505f906001905f90565b5050505f916003919056fe05f47829691a1f710b0620aedd52749bb09d8abe4bb530d306db920a71b0d7ce9b779b17422d0df92223018b32b4d1fa46e071723d6817e2486d003becc55f00a2646970667358221220a570aef58c98bd9abf8f1bffbdf7d77776f6b91090e2ebdca3a8bced81d9fd7564736f6c634300081e0033",
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
// Solidity: function checkpointChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactor) CheckpointChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "checkpointChannel", channelId, candidate)
}

// CheckpointChannel is a paid mutator transaction binding the contract method 0x9691b468.
//
// Solidity: function checkpointChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubSession) CheckpointChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CheckpointChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// CheckpointChannel is a paid mutator transaction binding the contract method 0x9691b468.
//
// Solidity: function checkpointChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
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
// Solidity: function closeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactor) CloseChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "closeChannel", channelId, candidate)
}

// CloseChannel is a paid mutator transaction binding the contract method 0x5dc46a74.
//
// Solidity: function closeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubSession) CloseChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.CloseChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// CloseChannel is a paid mutator transaction binding the contract method 0x5dc46a74.
//
// Solidity: function closeChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
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
// Solidity: function withdrawFromChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubTransactor) WithdrawFromChannel(opts *bind.TransactOpts, channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.contract.Transact(opts, "withdrawFromChannel", channelId, candidate)
}

// WithdrawFromChannel is a paid mutator transaction binding the contract method 0xc74a2d10.
//
// Solidity: function withdrawFromChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
func (_ChannelHub *ChannelHubSession) WithdrawFromChannel(channelId [32]byte, candidate State) (*types.Transaction, error) {
	return _ChannelHub.Contract.WithdrawFromChannel(&_ChannelHub.TransactOpts, channelId, candidate)
}

// WithdrawFromChannel is a paid mutator transaction binding the contract method 0xc74a2d10.
//
// Solidity: function withdrawFromChannel(bytes32 channelId, (uint64,uint8,bytes32,(uint64,address,uint8,uint256,int256,uint256,int256),(uint64,address,uint8,uint256,int256,uint256,int256),bytes,bytes) candidate) payable returns()
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
	PurgedCount *big.Int
	Raw         types.Log // Blockchain specific contextual infos
}

// FilterEscrowDepositsPurged is a free log retrieval operation binding the contract event 0x61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd145.
//
// Solidity: event EscrowDepositsPurged(uint256 purgedCount)
func (_ChannelHub *ChannelHubFilterer) FilterEscrowDepositsPurged(opts *bind.FilterOpts) (*ChannelHubEscrowDepositsPurgedIterator, error) {

	logs, sub, err := _ChannelHub.contract.FilterLogs(opts, "EscrowDepositsPurged")
	if err != nil {
		return nil, err
	}
	return &ChannelHubEscrowDepositsPurgedIterator{contract: _ChannelHub.contract, event: "EscrowDepositsPurged", logs: logs, sub: sub}, nil
}

// WatchEscrowDepositsPurged is a free log subscription operation binding the contract event 0x61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd145.
//
// Solidity: event EscrowDepositsPurged(uint256 purgedCount)
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

// ParseEscrowDepositsPurged is a log parse operation binding the contract event 0x61815f4b11c6ea4e14a2e448a010bed8efdc3e53a15efbf183d16a31085cd145.
//
// Solidity: event EscrowDepositsPurged(uint256 purgedCount)
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
