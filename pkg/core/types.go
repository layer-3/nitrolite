package core

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/shopspring/decimal"
)

type ChannelType uint8

var (
	ChannelTypeHome   ChannelType = 1
	ChannelTypeEscrow ChannelType = 2
)

type ChannelParticipant uint8

var (
	ChannelParticipantUser ChannelParticipant = 0
	ChannelParticipantNode ChannelParticipant = 1
)

type ChannelStatus uint8

var (
	ChannelStatusVoid       ChannelStatus = 0
	ChannelStatusOpen       ChannelStatus = 1
	ChannelStatusChallenged ChannelStatus = 2
	ChannelStatusClosed     ChannelStatus = 3
)

func (s ChannelStatus) String() string {
	switch s {
	case ChannelStatusVoid:
		return "void"
	case ChannelStatusOpen:
		return "open"
	case ChannelStatusChallenged:
		return "challenged"
	case ChannelStatusClosed:
		return "closed"
	default:
		return "unknown"
	}
}

func (s *ChannelStatus) Scan(src any) error {
	switch v := src.(type) {
	case int64:
		*s = ChannelStatus(uint8(v))
		return nil
	case int32:
		*s = ChannelStatus(uint8(v))
		return nil
	case int:
		*s = ChannelStatus(uint8(v))
		return nil
	case string:
		return s.scanString(v)
	default:
		return fmt.Errorf("unsupported ChannelStatus scan type %T", src)
	}
}

func (s *ChannelStatus) scanString(v string) error {
	v = strings.TrimSpace(v)
	// if numeric
	if n, err := strconv.Atoi(v); err == nil {
		*s = ChannelStatus(uint8(n))
		return nil
	}
	// else map names
	switch strings.ToLower(v) {
	case ChannelStatusVoid.String():
		*s = ChannelStatusVoid
	case ChannelStatusOpen.String():
		*s = ChannelStatusOpen
	case ChannelStatusChallenged.String():
		*s = ChannelStatusChallenged
	case ChannelStatusClosed.String():
		*s = ChannelStatusClosed
	default:
		return fmt.Errorf("unknown ChannelStatus %q", v)
	}
	return nil
}

const (
	INTENT_OPERATE                    = 0
	INTENT_CLOSE                      = 1
	INTENT_DEPOSIT                    = 2
	INTENT_WITHDRAW                   = 3
	INTENT_INITIATE_ESCROW_DEPOSIT    = 4
	INTENT_FINALIZE_ESCROW_DEPOSIT    = 5
	INTENT_INITIATE_ESCROW_WITHDRAWAL = 6
	INTENT_FINALIZE_ESCROW_WITHDRAWAL = 7
	INTENT_INITIATE_MIGRATION         = 8
	INTENT_FINALIZE_MIGRATION         = 9
)

// Channel represents an on-chain channel
type Channel struct {
	ChannelID             string        `json:"channel_id"`                     // Unique identifier for the channel
	UserWallet            string        `json:"user_wallet"`                    // User wallet address
	Asset                 string        `json:"asset"`                          // Asset symbol (e.g. USDC, ETH)
	Type                  ChannelType   `json:"type"`                           // Type of the channel (home, escrow)
	BlockchainID          uint64        `json:"blockchain_id"`                  // Unique identifier for the blockchain
	TokenAddress          string        `json:"token_address"`                  // Address of the token used in the channel
	ChallengeDuration     uint32        `json:"challenge_duration"`             // Challenge period for the channel in seconds
	ChallengeExpiresAt    *time.Time    `json:"challenge_expires_at,omitempty"` // Timestamp when the challenge period elapses
	Nonce                 uint64        `json:"nonce"`                          // Nonce for the channel
	ApprovedSigValidators string        `json:"approved_sig_validators"`        // Bitmask representing approved signature validators for the channel
	Status                ChannelStatus `json:"status"`                         // Current status of the channel (void, open, challenged, closed)
	StateVersion          uint64        `json:"state_version"`                  // On-chain state version of the channel
}

func NewChannel(channelID, userWallet, asset string, ChType ChannelType, blockchainID uint64, tokenAddress string, nonce uint64, challenge uint32, approvedSigValidators string) *Channel {
	return &Channel{
		ChannelID:             channelID,
		UserWallet:            userWallet,
		Asset:                 asset,
		Type:                  ChType,
		BlockchainID:          blockchainID,
		TokenAddress:          tokenAddress,
		Nonce:                 nonce,
		ChallengeDuration:     challenge,
		ApprovedSigValidators: approvedSigValidators,
		Status:                ChannelStatusVoid,
		StateVersion:          0,
	}
}

// ChannelDefinition represents configuration for creating a channel
type ChannelDefinition struct {
	Nonce                 uint64 `json:"nonce"`                   // A unique number to prevent replay attacks
	Challenge             uint32 `json:"challenge"`               // Challenge period for the channel in seconds
	ApprovedSigValidators string `json:"approved_sig_validators"` // Bitmask representing approved signature validators for the channel
}

// State represents the current state of the user stored on Node
type State struct {
	ID              string     `json:"id"`                          // Deterministic ID (hash) of the state
	Transition      Transition `json:"transition"`                  // The transition that led to this state
	Asset           string     `json:"asset"`                       // Asset type of the state
	UserWallet      string     `json:"user_wallet"`                 // User wallet address
	Epoch           uint64     `json:"epoch"`                       // User Epoch Index
	Version         uint64     `json:"version"`                     // Version of the state
	HomeChannelID   *string    `json:"home_channel_id,omitempty"`   // Identifier for the home Channel ID
	EscrowChannelID *string    `json:"escrow_channel_id,omitempty"` // Identifier for the escrow Channel ID
	HomeLedger      Ledger     `json:"home_ledger"`                 // User and node balances for the home channel
	EscrowLedger    *Ledger    `json:"escrow_ledger,omitempty"`     // User and node balances for the escrow channel
	UserSig         *string    `json:"user_sig,omitempty"`          // User signature for the state
	NodeSig         *string    `json:"node_sig,omitempty"`          // Node signature for the state
}

func NewVoidState(asset, userWallet string) *State {
	return &State{
		Asset:      asset,
		UserWallet: userWallet,
		HomeLedger: Ledger{
			UserBalance: decimal.Zero,
			UserNetFlow: decimal.Zero,
			NodeBalance: decimal.Zero,
			NodeNetFlow: decimal.Zero,
		},
	}
}

func (state State) NextState() *State {
	var nextState *State
	if state.IsFinal() {
		nextState = &State{
			Asset:           state.Asset,
			UserWallet:      state.UserWallet,
			Epoch:           state.Epoch + 1,
			Version:         0,
			HomeChannelID:   nil,
			EscrowChannelID: nil,
			HomeLedger: Ledger{
				UserBalance: decimal.Zero,
				UserNetFlow: decimal.Zero,
				NodeBalance: decimal.Zero,
				NodeNetFlow: decimal.Zero,
			},
			EscrowLedger: nil,
		}
	} else {
		nextState = &State{
			Asset:           state.Asset,
			UserWallet:      state.UserWallet,
			Epoch:           state.Epoch,
			Version:         state.Version + 1,
			HomeChannelID:   state.HomeChannelID,
			EscrowChannelID: state.EscrowChannelID,
			HomeLedger:      state.HomeLedger,
			EscrowLedger:    nil,
		}
	}
	if state.EscrowLedger != nil {
		nextState.EscrowLedger = &Ledger{
			TokenAddress: state.EscrowLedger.TokenAddress,
			BlockchainID: state.EscrowLedger.BlockchainID,
			UserBalance:  state.EscrowLedger.UserBalance,
			UserNetFlow:  state.EscrowLedger.UserNetFlow,
			NodeBalance:  state.EscrowLedger.NodeBalance,
			NodeNetFlow:  state.EscrowLedger.NodeNetFlow,
		}

		if transitionType := state.Transition.Type; transitionType == TransitionTypeEscrowDeposit || transitionType == TransitionTypeEscrowWithdraw {
			// escrowChannelID, escrowLedger: not-nil -> nil
			nextState.EscrowChannelID = nil
			nextState.EscrowLedger = nil
		}
	}
	nextState.ID = GetStateID(nextState.UserWallet, nextState.Asset, nextState.Epoch, nextState.Version)

	return nextState
}

// ApplyChannelCreation applies channel creation parameters to the state and returns the calculated home channel ID.
func (state *State) ApplyChannelCreation(channelDef ChannelDefinition, blockchainID uint64, tokenAddress, nodeAddress string) (string, error) {
	// Set home ledger
	state.HomeLedger.TokenAddress = tokenAddress
	state.HomeLedger.BlockchainID = blockchainID

	// Calculate home channel ID
	homeChannelID, err := GetHomeChannelID(
		nodeAddress,
		state.UserWallet,
		state.Asset,
		channelDef.Nonce,
		channelDef.Challenge,
		channelDef.ApprovedSigValidators,
	)
	if err != nil {
		return "", fmt.Errorf("failed to calculate home channel ID: %w", err)
	}
	state.HomeChannelID = &homeChannelID

	return homeChannelID, nil
}

func (state *State) IsFinal() bool {
	return state.Transition.Type == TransitionTypeFinalize
}

func (state *State) ApplyAcknowledgementTransition() (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}

	txID := common.Hash{}.Hex()      // Placeholder txID for acknowledgement transition
	accountID := common.Hash{}.Hex() // Placeholder accountID for acknowledgement transition
	newTransition := NewTransition(TransitionTypeAcknowledgement, txID, accountID, decimal.Zero)
	state.Transition = *newTransition

	return *newTransition, nil
}

func (state *State) ApplyHomeDepositTransition(amount decimal.Decimal) (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}
	if state.HomeChannelID == nil {
		return Transition{}, fmt.Errorf("missing home channel ID")
	}

	accountID := *state.HomeChannelID
	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeHomeDeposit, txID, accountID, amount)
	state.Transition = *newTransition
	state.HomeLedger.UserNetFlow = state.HomeLedger.UserNetFlow.Add(newTransition.Amount)
	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Add(newTransition.Amount)

	return *newTransition, nil
}

func (state *State) ApplyHomeWithdrawalTransition(amount decimal.Decimal) (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}
	if state.HomeChannelID == nil {
		return Transition{}, fmt.Errorf("missing home channel ID")
	}

	accountID := *state.HomeChannelID
	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeHomeWithdrawal, txID, accountID, amount)
	state.Transition = *newTransition
	state.HomeLedger.UserNetFlow = state.HomeLedger.UserNetFlow.Sub(newTransition.Amount)
	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Sub(newTransition.Amount)

	return *newTransition, nil
}

func (state *State) ApplyTransferSendTransition(recipient string, amount decimal.Decimal) (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}
	// TODO: maybe validate that recipient is a correct UserWallet format
	accountID := recipient
	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeTransferSend, txID, accountID, amount)
	state.Transition = *newTransition
	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Sub(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Sub(newTransition.Amount)

	return *newTransition, nil
}

func (state *State) ApplyTransferReceiveTransition(sender string, amount decimal.Decimal, txID string) (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}
	// TODO: maybe validate that recipient is a correct UserWallet format
	accountID := sender

	newTransition := NewTransition(TransitionTypeTransferReceive, txID, accountID, amount)
	state.Transition = *newTransition
	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Add(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Add(newTransition.Amount)
	return *newTransition, nil
}

func (state *State) ApplyCommitTransition(accountID string, amount decimal.Decimal) (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}
	// TODO: maybe validate that AccountID has correct AppSessionID format
	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeCommit, txID, accountID, amount)
	state.Transition = *newTransition
	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Sub(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Sub(newTransition.Amount)

	return *newTransition, nil
}

func (state *State) ApplyReleaseTransition(accountID string, amount decimal.Decimal) (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}
	// TODO: maybe validate that recipient is a correct UserWallet format
	txID, err := GetReceiverTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeRelease, txID, accountID, amount)
	state.Transition = *newTransition
	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Add(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Add(newTransition.Amount)
	return *newTransition, nil
}

func (state *State) ApplyMutualLockTransition(blockchainID uint64, tokenAddress string, amount decimal.Decimal) (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}
	if state.HomeChannelID == nil {
		return Transition{}, fmt.Errorf("missing home channel ID")
	}
	if blockchainID == 0 {
		return Transition{}, fmt.Errorf("invalid blockchain ID")
	}
	if tokenAddress == "" {
		return Transition{}, fmt.Errorf("invalid token address")
	}

	escrowChannelID, err := GetEscrowChannelID(*state.HomeChannelID, state.Version)
	if err != nil {
		return Transition{}, err
	}
	state.EscrowChannelID = &escrowChannelID
	accountID := escrowChannelID

	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeMutualLock, txID, accountID, amount)
	state.Transition = *newTransition

	state.HomeLedger.NodeBalance = state.HomeLedger.NodeBalance.Add(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Add(newTransition.Amount)

	state.EscrowLedger = &Ledger{
		BlockchainID: blockchainID,
		TokenAddress: tokenAddress,
		UserBalance:  decimal.Zero.Add(newTransition.Amount),
		UserNetFlow:  decimal.Zero.Add(newTransition.Amount),
		NodeBalance:  decimal.Zero,
		NodeNetFlow:  decimal.Zero,
	}

	return *newTransition, nil
}

func (state *State) ApplyEscrowDepositTransition(amount decimal.Decimal) (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}
	if state.EscrowChannelID == nil {
		return Transition{}, fmt.Errorf("internal error: escrow channel ID is nil")
	}
	if state.EscrowLedger == nil {
		return Transition{}, fmt.Errorf("escrow ledger is nil")
	}
	accountID := *state.EscrowChannelID

	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeEscrowDeposit, txID, accountID, amount)
	state.Transition = *newTransition

	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Add(newTransition.Amount)
	state.HomeLedger.NodeBalance = decimal.Zero

	state.EscrowLedger.UserBalance = state.EscrowLedger.UserBalance.Sub(newTransition.Amount)
	state.EscrowLedger.NodeNetFlow = state.EscrowLedger.NodeNetFlow.Sub(newTransition.Amount)

	return *newTransition, nil
}

func (state *State) ApplyEscrowLockTransition(blockchainID uint64, tokenAddress string, amount decimal.Decimal) (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}
	if state.HomeChannelID == nil {
		return Transition{}, fmt.Errorf("missing home channel ID")
	}
	if blockchainID == 0 {
		return Transition{}, fmt.Errorf("invalid blockchain ID")
	}
	if tokenAddress == "" {
		return Transition{}, fmt.Errorf("invalid token address")
	}

	escrowChannelID, err := GetEscrowChannelID(*state.HomeChannelID, state.Version)
	if err != nil {
		return Transition{}, err
	}
	state.EscrowChannelID = &escrowChannelID
	accountID := escrowChannelID

	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeEscrowLock, txID, accountID, amount)
	state.Transition = *newTransition

	state.EscrowLedger = &Ledger{
		BlockchainID: blockchainID,
		TokenAddress: tokenAddress,
		UserBalance:  decimal.Zero,
		UserNetFlow:  decimal.Zero,
		NodeBalance:  decimal.Zero.Add(newTransition.Amount),
		NodeNetFlow:  decimal.Zero.Add(newTransition.Amount),
	}

	return *newTransition, nil
}

func (state *State) ApplyEscrowWithdrawTransition(amount decimal.Decimal) (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}
	if state.EscrowChannelID == nil {
		return Transition{}, fmt.Errorf("internal error: escrow channel ID is nil")
	}
	if state.EscrowLedger == nil {
		return Transition{}, fmt.Errorf("escrow ledger is nil")
	}
	accountID := *state.EscrowChannelID

	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeEscrowWithdraw, txID, accountID, amount)
	state.Transition = *newTransition

	state.HomeLedger.UserBalance = state.HomeLedger.UserBalance.Sub(newTransition.Amount)
	state.HomeLedger.NodeNetFlow = state.HomeLedger.NodeNetFlow.Sub(newTransition.Amount)

	state.EscrowLedger.UserNetFlow = state.EscrowLedger.UserNetFlow.Sub(newTransition.Amount)
	state.EscrowLedger.NodeBalance = state.EscrowLedger.NodeBalance.Sub(newTransition.Amount)

	return *newTransition, nil
}

func (state *State) ApplyMigrateTransition(amount decimal.Decimal) (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}
	return Transition{}, fmt.Errorf("migrate transition not implemented yet")
}

func (state *State) ApplyFinalizeTransition() (Transition, error) {
	if state.Transition.Type != TransitionTypeVoid {
		return Transition{}, fmt.Errorf("state already has a transition: %s", state.Transition.Type.String())
	}
	if state.HomeChannelID == nil {
		return Transition{}, fmt.Errorf("missing home channel ID")
	}

	accountID := *state.HomeChannelID
	amount := state.HomeLedger.UserBalance
	txID, err := GetSenderTransactionID(accountID, state.ID)
	if err != nil {
		return Transition{}, err
	}

	newTransition := NewTransition(TransitionTypeFinalize, txID, accountID, amount)
	state.Transition = *newTransition

	state.HomeLedger.UserNetFlow = state.HomeLedger.UserNetFlow.Sub(state.HomeLedger.UserBalance)
	state.HomeLedger.UserBalance = decimal.Zero

	return *newTransition, nil
}

// Ledger represents ledger balances
type Ledger struct {
	TokenAddress string          `json:"token_address"` // Address of the token used in this channel
	BlockchainID uint64          `json:"blockchain_id"` // Unique identifier for the blockchain
	UserBalance  decimal.Decimal `json:"user_balance"`  // User balance in the channel
	UserNetFlow  decimal.Decimal `json:"user_net_flow"` // User net flow in the channel
	NodeBalance  decimal.Decimal `json:"node_balance"`  // Node balance in the channel
	NodeNetFlow  decimal.Decimal `json:"node_net_flow"` // Node net flow in the channel
}

func (l1 Ledger) Equal(l2 Ledger) error {
	if l1.TokenAddress != l2.TokenAddress {
		return fmt.Errorf("token address mismatch: expected=%s, proposed=%s", l1.TokenAddress, l2.TokenAddress)
	}
	if l1.BlockchainID != l2.BlockchainID {
		return fmt.Errorf("blockchain ID mismatch: expected=%d, proposed=%d", l1.BlockchainID, l2.BlockchainID)
	}
	if !l1.UserBalance.Equal(l2.UserBalance) {
		return fmt.Errorf("user balance mismatch: expected=%s, proposed=%s", l1.UserBalance.String(), l2.UserBalance.String())
	}
	if !l1.UserNetFlow.Equal(l2.UserNetFlow) {
		return fmt.Errorf("user net flow mismatch: expected=%s, proposed=%s", l1.UserNetFlow.String(), l2.UserNetFlow.String())
	}
	if !l1.NodeBalance.Equal(l2.NodeBalance) {
		return fmt.Errorf("node balance mismatch: expected=%s, proposed=%s", l1.NodeBalance.String(), l2.NodeBalance.String())
	}
	if !l1.NodeNetFlow.Equal(l2.NodeNetFlow) {
		return fmt.Errorf("node net flow mismatch: expected=%s, proposed=%s", l1.NodeNetFlow.String(), l2.NodeNetFlow.String())
	}
	return nil
}

func (l Ledger) Validate() error {
	if l.TokenAddress == "" {
		return fmt.Errorf("invalid token address")
	}
	if l.BlockchainID == 0 {
		return fmt.Errorf("invalid blockchain ID")
	}
	if l.UserBalance.IsNegative() {
		return fmt.Errorf("user balance cannot be negative")
	}
	if l.NodeBalance.IsNegative() {
		return fmt.Errorf("node balance cannot be negative")
	}
	sumBalances := l.UserBalance.Add(l.NodeBalance)
	sumNetFlows := l.UserNetFlow.Add(l.NodeNetFlow)
	if !sumBalances.Equal(sumNetFlows) {
		return fmt.Errorf("ledger balances do not match net flows: balances=%s, net_flows=%s", sumBalances.String(), sumNetFlows.String())
	}

	return nil
}

// TransactionType represents the type of transaction
type TransactionType uint8

const (
	TransactionTypeHomeDeposit    TransactionType = 10
	TransactionTypeHomeWithdrawal TransactionType = 11

	TransactionTypeEscrowDeposit  TransactionType = 20
	TransactionTypeEscrowWithdraw TransactionType = 21

	TransactionTypeTransfer TransactionType = 30

	TransactionTypeCommit    TransactionType = 40
	TransactionTypeRelease   TransactionType = 41
	TransactionTypeRebalance TransactionType = 42

	TransactionTypeMigrate    TransactionType = 100
	TransactionTypeEscrowLock TransactionType = 110
	TransactionTypeMutualLock TransactionType = 120

	TransactionTypeFinalize = 200
)

// String returns the human-readable name of the transaction type
func (t TransactionType) String() string {
	switch t {
	case TransactionTypeTransfer:
		return "transfer"
	case TransactionTypeRelease:
		return "release"
	case TransactionTypeCommit:
		return "commit"
	case TransactionTypeHomeDeposit:
		return "home_deposit"
	case TransactionTypeHomeWithdrawal:
		return "home_withdrawal"
	case TransactionTypeMutualLock:
		return "mutual_lock"
	case TransactionTypeEscrowDeposit:
		return "escrow_deposit"
	case TransactionTypeEscrowLock:
		return "escrow_lock"
	case TransactionTypeEscrowWithdraw:
		return "escrow_withdraw"
	case TransactionTypeMigrate:
		return "migrate"
	case TransactionTypeRebalance:
		return "rebalance"
	case TransactionTypeFinalize:
		return "finalize"
	default:
		return "unknown"
	}
}

// Transaction represents a transaction record
type Transaction struct {
	ID                 string          `json:"id"`                              // Unique transaction reference
	Asset              string          `json:"asset"`                           // Asset symbol
	TxType             TransactionType `json:"tx_type"`                         // Transaction type
	FromAccount        string          `json:"from_account"`                    // The account that sent the funds
	ToAccount          string          `json:"to_account"`                      // The account that received the funds
	SenderNewStateID   *string         `json:"sender_new_state_id,omitempty"`   // The ID of the new sender's channel state
	ReceiverNewStateID *string         `json:"receiver_new_state_id,omitempty"` // The ID of the new receiver's channel state
	Amount             decimal.Decimal `json:"amount"`                          // Transaction amount
	CreatedAt          time.Time       `json:"created_at"`                      // When the transaction was created
}

// NewTransaction creates a new instance of Transaction
func NewTransaction(id, asset string, txType TransactionType, fromAccount, toAccount string, senderNewStateID, receiverNewStateID *string, amount decimal.Decimal) *Transaction {
	return &Transaction{
		ID:                 id,
		Asset:              asset,
		TxType:             txType,
		FromAccount:        fromAccount,
		ToAccount:          toAccount,
		SenderNewStateID:   senderNewStateID,
		ReceiverNewStateID: receiverNewStateID,
		Amount:             amount,
		CreatedAt:          time.Now().UTC(),
	}
}

// NewTransactionFromTransition maps the transition type to the appropriate transaction type and returns a pointer to a Transaction.
func NewTransactionFromTransition(senderState *State, receiverState *State, transition Transition) (*Transaction, error) {
	var txType TransactionType
	var toAccount, fromAccount string
	// Transition validator is expected to make sure that all the fields are present and valid.

	if transition.Type != TransitionTypeRelease && senderState == nil {
		return nil, fmt.Errorf("sender state must not be nil for non-release transitions")
	}

	var senderStateID *string
	asset := ""
	if senderState != nil {
		senderStateID = &senderState.ID
		asset = senderState.Asset
	} else if receiverState != nil {
		asset = receiverState.Asset
	} else {
		return nil, fmt.Errorf("both sender and receiver states are nil")
	}

	switch transition.Type {
	case TransitionTypeHomeDeposit:
		if senderState.HomeChannelID == nil {
			return nil, fmt.Errorf("sender state has no home channel ID")
		}

		txType = TransactionTypeHomeDeposit
		fromAccount = *senderState.HomeChannelID
		toAccount = senderState.UserWallet

	case TransitionTypeHomeWithdrawal:
		if senderState.HomeChannelID == nil {
			return nil, fmt.Errorf("sender state has no home channel ID")
		}

		txType = TransactionTypeHomeWithdrawal
		fromAccount = senderState.UserWallet
		toAccount = *senderState.HomeChannelID

	case TransitionTypeEscrowDeposit:
		if senderState.EscrowChannelID == nil {
			return nil, fmt.Errorf("sender state has no escrow channel ID")
		}

		txType = TransactionTypeEscrowDeposit
		fromAccount = *senderState.EscrowChannelID
		toAccount = senderState.UserWallet

	case TransitionTypeEscrowWithdraw:
		if senderState.EscrowChannelID == nil || senderState.HomeChannelID == nil {
			return nil, fmt.Errorf("sender state has no escrow or home channel ID")
		}

		txType = TransactionTypeEscrowWithdraw
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	case TransitionTypeTransferSend:
		if receiverState == nil {
			return nil, fmt.Errorf("receiver state must not be nil for 'transfer_send' transition")
		}

		txType = TransactionTypeTransfer
		fromAccount = senderState.UserWallet
		toAccount = transition.AccountID

	case TransitionTypeCommit:
		txType = TransactionTypeCommit
		fromAccount = senderState.UserWallet
		toAccount = transition.AccountID

	case TransitionTypeRelease:
		txType = TransactionTypeRelease
		fromAccount = transition.AccountID
		toAccount = receiverState.UserWallet
		if receiverState == nil {
			return nil, fmt.Errorf("receiver state must not be nil for 'release' transition")
		}

	case TransitionTypeMutualLock:
		if senderState.EscrowChannelID == nil || senderState.HomeChannelID == nil {
			return nil, fmt.Errorf("sender state has no escrow or home channel ID")
		}

		txType = TransactionTypeMutualLock
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	case TransitionTypeEscrowLock:
		if senderState.EscrowChannelID == nil || senderState.HomeChannelID == nil {
			return nil, fmt.Errorf("sender state has no escrow or home channel ID")
		}

		txType = TransactionTypeEscrowLock
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID

	case TransitionTypeMigrate:
		if senderState.EscrowChannelID == nil || senderState.HomeChannelID == nil {
			return nil, fmt.Errorf("sender state has no escrow or home channel ID")
		}

		txType = TransactionTypeMigrate
		fromAccount = *senderState.HomeChannelID
		toAccount = *senderState.EscrowChannelID
	case TransitionTypeFinalize:
		txType = TransactionTypeFinalize
		fromAccount = senderState.UserWallet
		toAccount = *senderState.HomeChannelID
	default:
		return nil, fmt.Errorf("invalid transition type")
	}

	var receiverStateID *string
	var txID string
	var err error
	if receiverState != nil {
		receiverStateID = &receiverState.ID
		txID, err = GetReceiverTransactionID(fromAccount, receiverState.ID)
	} else {
		txID, err = GetSenderTransactionID(toAccount, senderState.ID)
	}
	if err != nil {
		return nil, err
	}

	return NewTransaction(
		txID,
		asset,
		txType,
		fromAccount,
		toAccount,
		senderStateID,
		receiverStateID,
		transition.Amount,
	), nil
}

// TransitionType represents the type of state transition
type TransitionType uint8

const (
	TransitionTypeVoid                           = 0 // Void transition, used for the initial state with no activity
	TransitionTypeAcknowledgement TransitionType = 1 // Acknowledgement of a received transfer, used for the initial state when a transfer is received without an existing state

	TransitionTypeHomeDeposit    TransitionType = 10 // AccountID: HomeChannelID
	TransitionTypeHomeWithdrawal TransitionType = 11 // AccountID: HomeChannelID

	TransitionTypeEscrowDeposit  TransitionType = 20 // AccountID: EscrowChannelID
	TransitionTypeEscrowWithdraw TransitionType = 21 // AccountID: EscrowChannelID

	TransitionTypeTransferSend    TransitionType = 30 // AccountID: Receiver's UserWallet
	TransitionTypeTransferReceive TransitionType = 31 // AccountID: Sender's UserWallet

	TransitionTypeCommit  TransitionType = 40 // AccountID: AppSessionID
	TransitionTypeRelease TransitionType = 41 // AccountID: AppSessionID

	TransitionTypeMigrate    TransitionType = 100 // AccountID: EscrowChannelID
	TransitionTypeEscrowLock TransitionType = 110 // AccountID: EscrowChannelID
	TransitionTypeMutualLock TransitionType = 120 // AccountID: EscrowChannelID

	TransitionTypeFinalize TransitionType = 200 // AccountID: HomeChannelID
)

// String returns the human-readable name of the transition type
func (t TransitionType) String() string {
	switch t {
	case TransitionTypeVoid:
		return "void"
	case TransitionTypeAcknowledgement:
		return "acknowledgement"
	case TransitionTypeTransferReceive:
		return "transfer_receive"
	case TransitionTypeTransferSend:
		return "transfer_send"
	case TransitionTypeRelease:
		return "release"
	case TransitionTypeCommit:
		return "commit"
	case TransitionTypeHomeDeposit:
		return "home_deposit"
	case TransitionTypeHomeWithdrawal:
		return "home_withdrawal"
	case TransitionTypeMutualLock:
		return "mutual_lock"
	case TransitionTypeEscrowDeposit:
		return "escrow_deposit"
	case TransitionTypeEscrowLock:
		return "escrow_lock"
	case TransitionTypeEscrowWithdraw:
		return "escrow_withdraw"
	case TransitionTypeMigrate:
		return "migrate"
	case TransitionTypeFinalize:
		return "finalize"
	default:
		return "unknown"
	}
}

func (t TransitionType) GatedAction() GatedAction {
	switch t {
	case TransitionTypeTransferSend:
		return GatedActionTransfer
	default:
		return ""
	}
}

// Transition represents a state transition
type Transition struct {
	Type      TransitionType  `json:"type"`       // Type of state transition
	TxID      string          `json:"tx_id"`      // Transaction ID associated with the transition
	AccountID string          `json:"account_id"` // Account identifier (varies based on transition type)
	Amount    decimal.Decimal `json:"amount"`     // Amount involved in the transition
}

// NewTransition creates a new state transition
func NewTransition(transitionType TransitionType, txID, accountID string, amount decimal.Decimal) *Transition {
	return &Transition{
		Type:      transitionType,
		TxID:      txID,
		AccountID: accountID,
		Amount:    amount,
	}
}

// Equal checks if two transitions are equal
func (t1 Transition) Equal(t2 Transition) error {
	if t1.Type != t2.Type {
		return fmt.Errorf("transition type mismatch: expected=%s, proposed=%s", t1.Type.String(), t2.Type.String())
	}
	if t1.TxID != t2.TxID {
		return fmt.Errorf("transaction ID mismatch: expected=%s, proposed=%s", t1.TxID, t2.TxID)
	}
	if t1.AccountID != t2.AccountID {
		return fmt.Errorf("account ID mismatch: expected=%s, proposed=%s", t1.AccountID, t2.AccountID)
	}
	if !t1.Amount.Equal(t2.Amount) {
		return fmt.Errorf("amount mismatch: expected=%s, proposed=%s", t1.Amount.String(), t2.Amount.String())
	}
	return nil
}

// Blockchain represents information about a supported blockchain network
type Blockchain struct {
	Name                   string `json:"name"`                     // Blockchain name
	ID                     uint64 `json:"id"`                       // Blockchain network ID
	ChannelHubAddress      string `json:"channel_hub_address"`      // Address of the ChannelHub contract on this blockchain
	LockingContractAddress string `json:"locking_contract_address"` // Address of the Locking contract on this blockchain
	BlockStep              uint64 `json:"block_step"`               // Number of blocks between each channel update
}

// Asset represents information about a supported asset
type Asset struct {
	Name                  string  `json:"name"`                    // Asset name
	Decimals              uint8   `json:"decimals"`                // Number of decimal places at YN
	Symbol                string  `json:"symbol"`                  // Asset symbol
	SuggestedBlockchainID uint64  `json:"suggested_blockchain_id"` // Suggested blockchain network ID for this asset
	Tokens                []Token `json:"tokens"`                  // Supported tokens for the asset
}

// Token represents information about a supported token
type Token struct {
	Name         string `json:"name"`          // Token name
	Symbol       string `json:"symbol"`        // Token symbol
	Address      string `json:"address"`       // Token contract address
	BlockchainID uint64 `json:"blockchain_id"` // Blockchain network ID
	Decimals     uint8  `json:"decimals"`      // Number of decimal places
}

// GatedAction represents an action that can be gated behind certain conditions, such as feature flags or access controls.
type GatedAction string

var (
	GatedActionTransfer GatedAction = "transfer"

	GatedActionAppSessionCreation   GatedAction = "app_session_creation"
	GatedActionAppSessionOperation  GatedAction = "app_session_operation"
	GatedActionAppSessionDeposit    GatedAction = "app_session_deposit"
	GatedActionAppSessionWithdrawal GatedAction = "app_session_withdrawal"
)

// ID returns a unique identifier for the GatedAction, which can be used for efficient storage and retrieval in databases or feature flag systems.
func (g GatedAction) ID() uint8 {
	switch g {
	case GatedActionTransfer:
		return 1
	case GatedActionAppSessionCreation:
		return 10
	case GatedActionAppSessionOperation:
		return 11
	case GatedActionAppSessionDeposit:
		return 12
	case GatedActionAppSessionWithdrawal:
		return 13
	}
	return 0
}

// GatedActionFromID returns the GatedAction corresponding to the given uint8 ID.
// Returns an empty GatedAction and false if the ID is unknown.
func GatedActionFromID(id uint8) (GatedAction, bool) {
	switch id {
	case 1:
		return GatedActionTransfer, true
	case 10:
		return GatedActionAppSessionCreation, true
	case 11:
		return GatedActionAppSessionOperation, true
	case 12:
		return GatedActionAppSessionDeposit, true
	case 13:
		return GatedActionAppSessionWithdrawal, true
	default:
		return "", false
	}
}

// ActionAllowance represents the allowance information for a specific gated action,
// including the time window for which the allowance applies, the total allowance, and the amount used.
type ActionAllowance struct {
	GatedAction GatedAction
	TimeWindow  string
	Allowance   uint64
	Used        uint64
}

// ========= Blockchain CLient Response Types =========

// HomeChannelDataResponse represents the response from getHomeChannelData
type HomeChannelDataResponse struct {
	Definition      ChannelDefinition `json:"definition"`
	Node            string            `json:"node"`
	LastState       State             `json:"last_state"`
	ChallengeExpiry uint64            `json:"challenge_expiry"`
}

// EscrowDepositDataResponse represents the response from getEscrowDepositData
type EscrowDepositDataResponse struct {
	EscrowChannelID string `json:"escrow_channel_id"`
	Node            string `json:"node"`
	LastState       State  `json:"last_state"`
	UnlockExpiry    uint64 `json:"unlock_expiry"`
	ChallengeExpiry uint64 `json:"challenge_expiry"`
}

// EscrowWithdrawalDataResponse represents the response from getEscrowWithdrawalData
type EscrowWithdrawalDataResponse struct {
	EscrowChannelID string `json:"escrow_channel_id"`
	Node            string `json:"node"`
	LastState       State  `json:"last_state"`
}

// ========= Storage Related Types =========

// BalanceEntry represents a balance entry for an asset
type BalanceEntry struct {
	Asset   string          `json:"asset"`   // Asset symbol
	Balance decimal.Decimal `json:"balance"` // Balance amount
}

// PaginationParams provides pagination configuration for getters
type PaginationParams struct {
	Offset *uint32
	Limit  *uint32
	Sort   *string
}

// GetOffsetAndLimit extracts offset and limit from pagination params with defaults and max limit enforcement.
func (p *PaginationParams) GetOffsetAndLimit(defaultLimit, maxLimit uint32) (offset, limit uint32) {
	offset = 0
	limit = defaultLimit

	if p != nil {
		if p.Offset != nil {
			offset = *p.Offset
		}
		if p.Limit != nil {
			limit = min(*p.Limit, maxLimit)
		}
	}

	return offset, limit
}

// PaginationMetadata contains pagination information for list responses.
type PaginationMetadata struct {
	Page       uint32 `json:"page"`        // Current page number
	PerPage    uint32 `json:"per_page"`    // Number of items per page
	TotalCount uint32 `json:"total_count"` // Total number of items
	PageCount  uint32 `json:"page_count"`  // Total number of pages
}

// NodeConfig represents the configuration of a Clearnode instance.
// It includes the node's identity, version, and supported blockchain networks.
type NodeConfig struct {
	// NodeAddress is the Ethereum address of the clearnode operator
	NodeAddress string

	// NodeVersion is the software version of the clearnode instance
	NodeVersion string

	// SupportedSigValidators is the list of supported signature validator types
	SupportedSigValidators []ChannelSignerType

	// Blockchains is the list of supported blockchain networks
	Blockchains []Blockchain
}
