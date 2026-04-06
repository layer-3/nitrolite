package sdk

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/rpc"
	"github.com/layer-3/nitrolite/pkg/sign"
	"github.com/shopspring/decimal"
)

const (
	// DefaultChallengePeriod is the default challenge period for channels (1 day in seconds)
	DefaultChallengePeriod = 86400
)

// Deposit prepares a deposit state for the user's channel.
// This method handles two scenarios automatically:
//  1. If no channel exists: Creates a new channel with the initial deposit
//  2. If channel exists: Advances the state with a deposit transition
//
// The returned state is signed by both the user and the node, but has not yet been
// submitted to the blockchain. Use Checkpoint to execute the on-chain transaction.
//
// Parameters:
//   - ctx: Context for the operation
//   - blockchainID: The blockchain network ID (e.g., 80002 for Polygon Amoy)
//   - asset: The asset symbol to deposit (e.g., "usdc")
//   - amount: The amount to deposit
//
// Returns:
//   - The co-signed state ready for on-chain checkpoint
//   - Error if the operation fails
//
// Example:
//
//	state, err := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
//	txHash, err := client.Checkpoint(ctx, "usdc")
//	fmt.Printf("Deposit transaction: %s\n", txHash)
func (c *Client) Deposit(ctx context.Context, blockchainID uint64, asset string, amount decimal.Decimal) (*core.State, error) {
	userWallet := c.GetUserAddress()

	// Get node address
	nodeAddress, err := c.getNodeAddress(ctx)
	if err != nil {
		return nil, err
	}

	// Get token address for this asset on this blockchain
	tokenAddress, err := c.getTokenAddress(ctx, blockchainID, asset)
	if err != nil {
		return nil, err
	}

	// Try to get latest state to determine if channel exists
	state, err := c.GetLatestState(ctx, userWallet, asset, false)

	// Scenario A: Channel doesn't exist - create it
	if err != nil || state.HomeChannelID == nil {
		// Get supported sig validators bitmap from node config
		bitmap, err := c.getSupportedSigValidatorsBitmap(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get sig validators bitmap: %w", err)
		}

		// Create channel definition
		channelDef := core.ChannelDefinition{
			Nonce:                 generateNonce(),
			Challenge:             DefaultChallengePeriod,
			ApprovedSigValidators: bitmap,
		}

		// HomeChannelID is intentionally not checked here: a non-nil state with a
		// nil HomeChannelID is valid — it represents a user who received funds but
		// has not yet opened a channel. Only replace with a void state when there
		// is truly no prior state at all.
		if state == nil {
			state = core.NewVoidState(asset, userWallet)
		}
		newState := state.NextState()

		_, err = newState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to apply channel creation: %w", err)
		}

		// Apply deposit transition
		_, err = newState.ApplyHomeDepositTransition(amount)
		if err != nil {
			return nil, fmt.Errorf("failed to apply deposit transition: %w", err)
		}

		// Sign state
		sig, err := c.ValidateAndSignState(state, newState)
		if err != nil {
			return nil, fmt.Errorf("failed to sign state: %w", err)
		}
		newState.UserSig = &sig

		// Request channel creation from node
		nodeSig, err := c.requestChannelCreation(ctx, *newState, channelDef)
		if err != nil {
			return nil, fmt.Errorf("failed to request channel creation: %w", err)
		}
		newState.NodeSig = &nodeSig

		return newState, nil
	}

	// Scenario B: Channel exists - checkpoint deposit
	// Create next state
	nextState := state.NextState()

	// Apply deposit transition
	_, err = nextState.ApplyHomeDepositTransition(amount)
	if err != nil {
		return nil, fmt.Errorf("failed to apply deposit transition: %w", err)
	}

	// Sign and submit state to node
	_, err = c.signAndSubmitState(ctx, state, nextState)
	if err != nil {
		return nextState, err
	}

	return nextState, nil
}

// Withdraw prepares a withdrawal state to remove funds from the user's channel.
// This operation handles two scenarios automatically:
//  1. If no channel exists: Creates a new channel with the withdrawal transition
//  2. If channel exists: Advances the state with a withdrawal transition
//
// The returned state is signed by both the user and the node, but has not yet been
// submitted to the blockchain. Use Checkpoint to execute the on-chain transaction.
//
// Parameters:
//   - ctx: Context for the operation
//   - blockchainID: The blockchain network ID (e.g., 80002 for Polygon Amoy)
//   - asset: The asset symbol to withdraw (e.g., "usdc")
//   - amount: The amount to withdraw
//
// Returns:
//   - The co-signed state ready for on-chain checkpoint
//   - Error if the operation fails
//
// Example:
//
//	state, err := client.Withdraw(ctx, 80002, "usdc", decimal.NewFromInt(25))
//	txHash, err := client.Checkpoint(ctx, "usdc")
//	fmt.Printf("Withdrawal transaction: %s\n", txHash)
func (c *Client) Withdraw(ctx context.Context, blockchainID uint64, asset string, amount decimal.Decimal) (*core.State, error) {
	userWallet := c.GetUserAddress()

	// Get node address (Required for channel creation flow)
	nodeAddress, err := c.getNodeAddress(ctx)
	if err != nil {
		return nil, err
	}

	// Get token address for this asset on this blockchain (Required for channel creation flow)
	tokenAddress, err := c.getTokenAddress(ctx, blockchainID, asset)
	if err != nil {
		return nil, err
	}

	// Try to get latest state to determine if channel exists
	state, err := c.GetLatestState(ctx, userWallet, asset, false)

	// Channel doesn't exist - create it and withdraw
	if err != nil || state.HomeChannelID == nil {
		// Get supported sig validators bitmap from node config
		bitmap, err := c.getSupportedSigValidatorsBitmap(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get sig validators bitmap: %w", err)
		}

		// Create channel definition
		channelDef := core.ChannelDefinition{
			Nonce:                 generateNonce(),
			Challenge:             DefaultChallengePeriod,
			ApprovedSigValidators: bitmap,
		}

		// HomeChannelID is intentionally not checked here: a non-nil state with a
		// nil HomeChannelID is valid — it represents a user who received funds but
		// has not yet opened a channel. Only replace with a void state when there
		// is truly no prior state at all.
		if state == nil {
			state = core.NewVoidState(asset, userWallet)
		}
		newState := state.NextState()

		_, err = newState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to apply channel creation: %w", err)
		}

		// Apply withdrawal transition
		// Note: Ensure your core logic allows withdrawal on a fresh state
		// (assuming the smart contract handles the net balance check)
		_, err = newState.ApplyHomeWithdrawalTransition(amount)
		if err != nil {
			return nil, fmt.Errorf("failed to apply withdrawal transition: %w", err)
		}

		// Sign state
		sig, err := c.ValidateAndSignState(state, newState)
		if err != nil {
			return nil, fmt.Errorf("failed to sign state: %w", err)
		}
		newState.UserSig = &sig

		// Request channel creation from node
		nodeSig, err := c.requestChannelCreation(ctx, *newState, channelDef)
		if err != nil {
			return nil, fmt.Errorf("failed to request channel creation: %w", err)
		}
		newState.NodeSig = &nodeSig

		return newState, nil
	}

	// Create next state
	nextState := state.NextState()

	// Apply withdrawal transition
	_, err = nextState.ApplyHomeWithdrawalTransition(amount)
	if err != nil {
		return nil, fmt.Errorf("failed to apply withdrawal transition: %w", err)
	}

	// Sign and submit state to node
	_, err = c.signAndSubmitState(ctx, state, nextState)
	if err != nil {
		return nil, err
	}

	return nextState, nil
}

// Transfer prepares a transfer state to send funds to another wallet address.
// This method handles two scenarios automatically:
//  1. If no channel exists: Creates a new channel with the transfer transition
//  2. If channel exists: Advances the state with a transfer send transition
//
// The returned state is signed by both the user and the node. For existing channels,
// no blockchain interaction is needed. For new channels, use Checkpoint to create
// the channel on-chain.
//
// Parameters:
//   - ctx: Context for the operation
//   - recipientWallet: The recipient's wallet address (e.g., "0x1234...")
//   - asset: The asset symbol to transfer (e.g., "usdc")
//   - amount: The amount to transfer
//
// Returns:
//   - The co-signed state with the transfer transition applied
//   - Error if the operation fails
//
// Example:
//
//	state, err := client.Transfer(ctx, "0xRecipient...", "usdc", decimal.NewFromInt(50))
//	fmt.Printf("Transfer tx ID: %s\n", state.Transition.TxID)
func (c *Client) Transfer(ctx context.Context, recipientWallet string, asset string, amount decimal.Decimal) (*core.State, error) {
	// Get sender's latest state
	senderWallet := c.GetUserAddress()
	state, err := c.GetLatestState(ctx, senderWallet, asset, false)
	if err != nil || state.HomeChannelID == nil {
		// Get supported sig validators bitmap from node config
		bitmap, err := c.getSupportedSigValidatorsBitmap(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get sig validators bitmap: %w", err)
		}

		// Create channel definition
		channelDef := core.ChannelDefinition{
			Nonce:                 generateNonce(),
			Challenge:             DefaultChallengePeriod,
			ApprovedSigValidators: bitmap,
		}

		// HomeChannelID is intentionally not checked here: a non-nil state with a
		// nil HomeChannelID is valid — it represents a user who received funds but
		// has not yet opened a channel. Only replace with a void state when there
		// is truly no prior state at all.
		if state == nil {
			state = core.NewVoidState(asset, senderWallet)
		}
		newState := state.NextState()

		blockchainID, ok := c.homeBlockchains[asset]
		if !ok {
			if state.HomeLedger.BlockchainID != 0 {
				blockchainID = state.HomeLedger.BlockchainID
			} else {
				blockchainID, err = c.assetStore.GetSuggestedBlockchainID(asset)
				if err != nil {
					return nil, err
				}
			}
		}

		// Get node address (Required for channel creation flow)
		nodeAddress, err := c.getNodeAddress(ctx)
		if err != nil {
			return nil, err
		}

		// Get token address for this asset on this blockchain
		tokenAddress, err := c.getTokenAddress(ctx, blockchainID, asset)
		if err != nil {
			return nil, err
		}

		_, err = newState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to apply channel creation: %w", err)
		}

		// Apply transfer send transition
		_, err = newState.ApplyTransferSendTransition(recipientWallet, amount)
		if err != nil {
			return nil, fmt.Errorf("failed to apply transfer transition: %w", err)
		}

		sig, err := c.ValidateAndSignState(state, newState)
		if err != nil {
			return nil, fmt.Errorf("failed to sign state: %w", err)
		}
		newState.UserSig = &sig

		// Request channel creation from node
		nodeSig, err := c.requestChannelCreation(ctx, *newState, channelDef)
		if err != nil {
			return nil, fmt.Errorf("failed to request channel creation: %w", err)
		}
		newState.NodeSig = &nodeSig

		return newState, nil
	}

	// Create next state
	nextState := state.NextState()

	// Apply transfer send transition
	_, err = nextState.ApplyTransferSendTransition(recipientWallet, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to apply transfer transition: %w", err)
	}

	// Sign and submit state
	_, err = c.signAndSubmitState(ctx, state, nextState)
	if err != nil {
		return nil, err
	}

	return nextState, nil
}

// CloseHomeChannel prepares a finalize state to close the user's channel for a specific asset.
// This creates a final state with zero user balance and submits it to the node.
//
// The returned state is signed by both the user and the node, but has not yet been
// submitted to the blockchain. Use Checkpoint to execute the on-chain close.
//
// Parameters:
//   - ctx: Context for the operation
//   - asset: The asset symbol to close (e.g., "usdc")
//
// Returns:
//   - The co-signed finalize state ready for on-chain close
//   - Error if the operation fails
//
// Errors:
//   - Returns error if no channel exists for the asset
//   - Returns error if state signing or submission fails
//
// Example:
//
//	state, err := client.CloseHomeChannel(ctx, "usdc")
//	txHash, err := client.Checkpoint(ctx, "usdc")
//	fmt.Printf("Close transaction: %s\n", txHash)
func (c *Client) CloseHomeChannel(ctx context.Context, asset string) (*core.State, error) {
	// Get sender's latest state
	senderWallet := c.GetUserAddress()

	state, err := c.GetLatestState(ctx, senderWallet, asset, false)
	if err != nil {
		return nil, err
	}

	if state.HomeChannelID == nil {
		return nil, fmt.Errorf("no channel exists for asset %s", asset)
	}

	// Create next state
	nextState := state.NextState()

	// Apply finalize transition
	_, err = nextState.ApplyFinalizeTransition()
	if err != nil {
		return nil, fmt.Errorf("failed to apply finalize transition: %w", err)
	}

	// Sign and submit state
	_, err = c.signAndSubmitState(ctx, state, nextState)
	if err != nil {
		return nil, err
	}

	return nextState, nil
}

// Acknowledge prepares an acknowledgement state for the given asset.
// This is used when a user receives a transfer but hasn't yet acknowledged the state,
// or to acknowledge channel creation without a deposit.
//
// This method handles two scenarios automatically:
//  1. If no channel exists: Creates a new channel with the acknowledgement transition
//  2. If channel exists: Advances the state with an acknowledgement transition
//
// The returned state is signed by both the user and the node.
//
// Parameters:
//   - ctx: Context for the operation
//   - asset: The asset symbol to acknowledge (e.g., "usdc")
//
// Returns:
//   - The co-signed state with the acknowledgement transition applied
//   - Error if the operation fails
//
// Requirements:
//   - Home blockchain must be set for the asset (use SetHomeBlockchain) when no channel exists
//
// Example:
//
//	state, err := client.Acknowledge(ctx, "usdc")
func (c *Client) Acknowledge(ctx context.Context, asset string) (*core.State, error) {
	userWallet := c.GetUserAddress()

	// Try to get latest state to determine if channel exists
	state, err := c.GetLatestState(ctx, userWallet, asset, false)

	// No channel path - create channel with acknowledgement
	if err != nil || state.HomeChannelID == nil {
		// Get supported sig validators bitmap from node config
		bitmap, err := c.getSupportedSigValidatorsBitmap(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get sig validators bitmap: %w", err)
		}

		channelDef := core.ChannelDefinition{
			Nonce:                 generateNonce(),
			Challenge:             DefaultChallengePeriod,
			ApprovedSigValidators: bitmap,
		}

		// HomeChannelID is intentionally not checked here: a non-nil state with a
		// nil HomeChannelID is valid — it represents a user who received funds but
		// has not yet opened a channel. Only replace with a void state when there
		// is truly no prior state at all.
		if state == nil {
			state = core.NewVoidState(asset, userWallet)
		}
		newState := state.NextState()

		blockchainID, ok := c.homeBlockchains[asset]
		if !ok {
			if state.HomeLedger.BlockchainID != 0 {
				blockchainID = state.HomeLedger.BlockchainID
			} else {
				blockchainID, err = c.assetStore.GetSuggestedBlockchainID(asset)
				if err != nil {
					return nil, err
				}
			}
		}

		nodeAddress, err := c.getNodeAddress(ctx)
		if err != nil {
			return nil, err
		}

		tokenAddress, err := c.getTokenAddress(ctx, blockchainID, asset)
		if err != nil {
			return nil, err
		}

		_, err = newState.ApplyChannelCreation(channelDef, blockchainID, tokenAddress, nodeAddress)
		if err != nil {
			return nil, fmt.Errorf("failed to apply channel creation: %w", err)
		}

		_, err = newState.ApplyAcknowledgementTransition()
		if err != nil {
			return nil, fmt.Errorf("failed to apply acknowledgement transition: %w", err)
		}

		sig, err := c.ValidateAndSignState(state, newState)
		if err != nil {
			return nil, fmt.Errorf("failed to sign state: %w", err)
		}
		newState.UserSig = &sig

		nodeSig, err := c.requestChannelCreation(ctx, *newState, channelDef)
		if err != nil {
			return nil, fmt.Errorf("failed to request channel creation: %w", err)
		}
		newState.NodeSig = &nodeSig

		return newState, nil
	}

	if state.UserSig != nil {
		return nil, fmt.Errorf("state already acknowledged by user")
	}

	// Has channel path - submit acknowledgement
	nextState := state.NextState()

	_, err = nextState.ApplyAcknowledgementTransition()
	if err != nil {
		return nil, fmt.Errorf("failed to apply acknowledgement transition: %w", err)
	}

	_, err = c.signAndSubmitState(ctx, state, nextState)
	if err != nil {
		return nil, err
	}

	return nextState, nil
}

// Checkpoint executes the blockchain transaction for the latest signed state.
// It fetches the latest co-signed state and, based on the transition type and on-chain
// channel status, calls the appropriate blockchain method.
//
// This is the only method that interacts with the blockchain. It should be called after
// any state-building method (Deposit, Withdraw, CloseHomeChannel, etc.) to settle
// the state on-chain. It can also be used as a recovery mechanism if a previous
// blockchain transaction failed (e.g., due to gas issues or network problems).
//
// Blockchain method mapping:
//   - Channel not yet on-chain (status Void): Creates the channel via blockchainClient.Create
//   - HomeDeposit/HomeWithdrawal on existing channel: Checkpoints via blockchainClient.Checkpoint
//   - Finalize: Closes the channel via blockchainClient.Close
//
// Parameters:
//   - ctx: Context for the operation
//   - asset: The asset symbol (e.g., "usdc")
//
// Returns:
//   - Transaction hash of the blockchain transaction
//   - Error if the operation fails or no blockchain operation is needed
//
// Requirements:
//   - Blockchain RPC must be configured for the chain (use WithBlockchainRPC)
//   - A co-signed state must exist (call Deposit, Withdraw, etc. first)
//
// Example:
//
//	state, err := client.Deposit(ctx, 80002, "usdc", decimal.NewFromInt(100))
//	txHash, err := client.Checkpoint(ctx, "usdc")
//	fmt.Printf("On-chain transaction: %s\n", txHash)
func (c *Client) Checkpoint(ctx context.Context, asset string) (string, error) {
	userWallet := c.GetUserAddress()

	// Get latest signed state (both user and node signatures must be present)
	state, err := c.GetLatestState(ctx, userWallet, asset, true)
	if err != nil {
		return "", fmt.Errorf("failed to get latest signed state: %w", err)
	}

	if state.HomeChannelID == nil {
		// NOTE: this should never happen, because signed state MUST have a channel ID
		return "", fmt.Errorf("no channel exists for asset %s", asset)
	}

	blockchainID := state.HomeLedger.BlockchainID

	// Initialize blockchain client if needed
	blockchainClient, err := c.getOrInitBlockchainClient(ctx, blockchainID)
	if err != nil {
		return "", err
	}

	// Get home channel info to determine on-chain status
	channel, err := c.GetHomeChannel(ctx, userWallet, asset)
	if err != nil {
		return "", fmt.Errorf("failed to get home channel: %w", err)
	}

	switch state.Transition.Type {
	case core.TransitionTypeAcknowledgement,
		core.TransitionTypeHomeDeposit,
		core.TransitionTypeHomeWithdrawal,
		core.TransitionTypeTransferSend,
		core.TransitionTypeTransferReceive,
		core.TransitionTypeCommit,
		core.TransitionTypeRelease:
		if channel.Status == core.ChannelStatusVoid {
			// Channel not yet created on-chain, reconstruct definition and call Create
			channelDef := core.ChannelDefinition{
				Nonce:                 channel.Nonce,
				Challenge:             channel.ChallengeDuration,
				ApprovedSigValidators: channel.ApprovedSigValidators,
			}
			txHash, err := blockchainClient.Create(channelDef, *state)
			if err != nil {
				return "", fmt.Errorf("failed to create channel on blockchain: %w", err)
			}
			return txHash, nil
		}

		// Checkpoint existing channel for deposit/withdrawal
		txHash, err := blockchainClient.Checkpoint(*state)
		if err != nil {
			return "", fmt.Errorf("failed to checkpoint on blockchain: %w", err)
		}
		return txHash, nil

	case core.TransitionTypeFinalize:
		txHash, err := blockchainClient.Close(*state)
		if err != nil {
			return "", fmt.Errorf("failed to close channel on blockchain: %w", err)
		}
		return txHash, nil

	default:
		return "", fmt.Errorf("transition type %s does not require a blockchain operation", state.Transition.Type)
	}
}

// Challenge submits an on-chain challenge for a channel using a co-signed state.
// The state must have both user and node signatures, which are validated before
// the challenge transaction is submitted.
//
// A challenge initiates a dispute period on-chain. If the counterparty does not
// respond with a higher-versioned state before the challenge period expires,
// the channel can be closed with the challenged state.
//
// Parameters:
//   - ctx: Context for the operation
//   - state: A co-signed state (both UserSig and NodeSig must be present)
//
// Returns:
//   - Transaction hash of the on-chain challenge transaction
//   - Error if validation or submission fails
//
// Requirements:
//   - Blockchain RPC must be configured for the chain (use WithBlockchainRPC)
//   - State must have both user and node signatures
//   - State must have a HomeChannelID
//
// Example:
//
//	state, err := client.GetLatestState(ctx, wallet, "usdc", true)
//	txHash, err := client.Challenge(ctx, *state)
//	fmt.Printf("Challenge transaction: %s\n", txHash)
func (c *Client) Challenge(ctx context.Context, state core.State) (string, error) {
	if state.UserSig == nil || state.NodeSig == nil {
		return "", fmt.Errorf("state must have both user and node signatures")
	}

	if state.HomeChannelID == nil {
		return "", fmt.Errorf("state must have a home channel ID")
	}

	// Pack state for signature verification
	packedState, err := core.PackState(state, c.assetStore)
	if err != nil {
		return "", fmt.Errorf("failed to pack state: %w", err)
	}

	sigValidator := core.NewChannelSigValidator(func(wallet, sessionKey, metadataHash string) (bool, error) {
		// Accept signatures from the user's wallet or any valid session key
		return true, nil
	})

	// Verify user signature
	userSigBytes, err := hexutil.Decode(*state.UserSig)
	if err != nil {
		return "", fmt.Errorf("invalid user signature encoding: %w", err)
	}
	if err := sigValidator.Verify(state.UserWallet, packedState, userSigBytes); err != nil {
		return "", fmt.Errorf("invalid user signature: %w", err)
	}

	// Verify node signature
	nodeAddress, err := c.getNodeAddress(ctx)
	if err != nil {
		return "", err
	}
	nodeSigBytes, err := hexutil.Decode(*state.NodeSig)
	if err != nil {
		return "", fmt.Errorf("invalid node signature encoding: %w", err)
	}
	if err := sigValidator.Verify(nodeAddress, packedState, nodeSigBytes); err != nil {
		return "", fmt.Errorf("invalid node signature: %w", err)
	}

	// Create the challenge signature
	challengeData, err := core.PackChallengeState(state, c.assetStore)
	if err != nil {
		return "", fmt.Errorf("failed to pack challenge state: %w", err)
	}
	challengerSig, err := c.stateSigner.Sign(challengeData)
	if err != nil {
		return "", fmt.Errorf("failed to sign challenge: %w", err)
	}

	// Initialize blockchain client and submit
	blockchainID := state.HomeLedger.BlockchainID
	blockchainClient, err := c.getOrInitBlockchainClient(ctx, blockchainID)
	if err != nil {
		return "", err
	}

	txHash, err := blockchainClient.Challenge(state, challengerSig, core.ChannelParticipantUser)
	if err != nil {
		return "", fmt.Errorf("failed to challenge on blockchain: %w", err)
	}

	return txHash, nil
}

// ============================================================================
// Channel Query Methods
// ============================================================================

// GetHomeChannel retrieves home channel information for a user's asset.
//
// Parameters:
//   - wallet: The user's wallet address
//   - asset: The asset symbol
//
// Returns:
//   - Channel information for the home channel
//   - Error if the request fails
//
// Example:
//
//	channel, err := client.GetHomeChannel(ctx, "0x1234...", "usdc")
//	fmt.Printf("Home Channel: %s (Version: %d)\n", channel.ChannelID, channel.StateVersion)
func (c *Client) GetHomeChannel(ctx context.Context, wallet, asset string) (*core.Channel, error) {
	req := rpc.ChannelsV1GetHomeChannelRequest{
		Wallet: wallet,
		Asset:  asset,
	}
	resp, err := c.rpcClient.ChannelsV1GetHomeChannel(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get home channel: %w", err)
	}

	channel, err := transformChannel(resp.Channel)
	if err != nil {
		return nil, fmt.Errorf("failed to transform channel: %w", err)
	}
	return &channel, nil
}

// GetEscrowChannel retrieves escrow channel information for a specific channel ID.
//
// Parameters:
//   - escrowChannelID: The escrow channel ID to query
//
// Returns:
//   - Channel information for the escrow channel
//   - Error if the request fails
//
// Example:
//
//	channel, err := client.GetEscrowChannel(ctx, "0x1234...")
//	fmt.Printf("Escrow Channel: %s (Version: %d)\n", channel.ChannelID, channel.StateVersion)
func (c *Client) GetEscrowChannel(ctx context.Context, escrowChannelID string) (*core.Channel, error) {
	req := rpc.ChannelsV1GetEscrowChannelRequest{
		EscrowChannelID: escrowChannelID,
	}
	resp, err := c.rpcClient.ChannelsV1GetEscrowChannel(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get escrow channel: %w", err)
	}

	channel, err := transformChannel(resp.Channel)
	if err != nil {
		return nil, fmt.Errorf("failed to transform channel: %w", err)
	}
	return &channel, nil
}

// ============================================================================
// State Management Methods
// ============================================================================

// GetLatestState retrieves the latest state for a user's asset.
//
// Parameters:
//   - wallet: The user's wallet address
//   - asset: The asset symbol (e.g., "usdc")
//   - onlySigned: If true, returns only the latest signed state
//
// Returns:
//   - core.State containing all state information
//   - Error if the request fails
//
// Example:
//
//	state, err := client.GetLatestState(ctx, "0x1234...", "usdc", false)
//	fmt.Printf("State Version: %d, Balance: %s\n", state.Version, state.HomeLedger.UserBalance)
func (c *Client) GetLatestState(ctx context.Context, wallet, asset string, onlySigned bool) (*core.State, error) {
	req := rpc.ChannelsV1GetLatestStateRequest{
		Wallet:     wallet,
		Asset:      asset,
		OnlySigned: onlySigned,
	}
	resp, err := c.rpcClient.ChannelsV1GetLatestState(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest state: %w", err)
	}
	state, err := transformState(resp.State)
	if err != nil {
		return nil, fmt.Errorf("failed to transform state: %w", err)
	}
	return &state, nil
}

// submitState submits a signed state update to the node.
// The state must be properly signed by the user before submission.
// This is an internal method used by high-level operations.
func (c *Client) submitState(ctx context.Context, state core.State) (string, error) {
	req := rpc.ChannelsV1SubmitStateRequest{
		State: transformStateToRPC(state),
	}
	resp, err := c.rpcClient.ChannelsV1SubmitState(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to submit state: %w", err)
	}
	return resp.Signature, nil
}

// requestChannelCreation requests the node to co-sign a channel creation state.
// This is an internal method used when creating a new channel (via Deposit, Withdraw,
// Transfer, or Acknowledge).
func (c *Client) requestChannelCreation(ctx context.Context, state core.State, channelDef core.ChannelDefinition) (string, error) {
	req := rpc.ChannelsV1RequestCreationRequest{
		State:             transformStateToRPC(state),
		ChannelDefinition: transformChannelDefinitionToRPC(channelDef),
	}
	resp, err := c.rpcClient.ChannelsV1RequestCreation(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to request channel creation: %w", err)
	}
	return resp.Signature, nil
}

// ============================================================================
// Channel Session Key Methods
// ============================================================================

// GetLastChannelKeyStatesOptions contains optional filters for GetLastChannelKeyStates.
type GetLastChannelKeyStatesOptions struct {
	// SessionKey filters by a specific session key address
	SessionKey *string
}

// SubmitChannelSessionKeyState submits a channel session key state for registration or update.
// The state must be signed by the user's wallet to authorize the session key delegation.
//
// Parameters:
//   - state: The channel session key state containing delegation information
//
// Returns:
//   - Error if the request fails
//
// Example:
//
//	state := core.ChannelSessionKeyStateV1{
//	    UserAddress: "0x1234...",
//	    SessionKey:  "0xabcd...",
//	    Version:     1,
//	    Assets:      []string{"usdc", "weth"},
//	    ExpiresAt:   time.Now().Add(24 * time.Hour),
//	    UserSig:     "0x...",
//	}
//	err := client.SubmitChannelSessionKeyState(ctx, state)
func (c *Client) SubmitChannelSessionKeyState(ctx context.Context, state core.ChannelSessionKeyStateV1) error {
	req := rpc.ChannelsV1SubmitSessionKeyStateRequest{
		State: transformChannelSessionKeyStateToRPC(state),
	}
	_, err := c.rpcClient.ChannelsV1SubmitSessionKeyState(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to submit channel session key state: %w", err)
	}
	return nil
}

// GetLastChannelKeyStates retrieves the latest channel session key states for a user.
//
// Parameters:
//   - userAddress: The user's wallet address
//   - opts: Optional filters (pass nil for no filters)
//
// Returns:
//   - Slice of ChannelSessionKeyStateV1 with the latest non-expired session key states
//   - Error if the request fails
//
// Example:
//
//	states, err := client.GetLastChannelKeyStates(ctx, "0x1234...", nil)
//	for _, state := range states {
//	    fmt.Printf("Session key %s expires at %s\n", state.SessionKey, state.ExpiresAt)
//	}
func (c *Client) GetLastChannelKeyStates(ctx context.Context, userAddress string, opts *GetLastChannelKeyStatesOptions) ([]core.ChannelSessionKeyStateV1, error) {
	req := rpc.ChannelsV1GetLastKeyStatesRequest{
		UserAddress: userAddress,
	}
	if opts != nil {
		req.SessionKey = opts.SessionKey
	}

	resp, err := c.rpcClient.ChannelsV1GetLastKeyStates(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get last channel key states: %w", err)
	}

	states, err := transformChannelSessionKeyStates(resp.States)
	if err != nil {
		return nil, fmt.Errorf("failed to transform channel session key states: %w", err)
	}

	return states, nil
}

// SignChannelSessionKeyState signs a channel session key state using the client's state signer.
// This creates a properly formatted signature that can be set on the state's UserSig field
// before submitting via SubmitChannelSessionKeyState.
//
// Parameters:
//   - state: The channel session key state to sign (UserSig field is excluded from signing)
//
// Returns:
//   - The hex-encoded signature string
//   - Error if signing fails
//
// Example:
//
//	state := core.ChannelSessionKeyStateV1{
//	    UserAddress: client.GetUserAddress(),
//	    SessionKey:  "0xabcd...",
//	    Version:     1,
//	    Assets:      []string{"usdc"},
//	    ExpiresAt:   time.Now().Add(24 * time.Hour),
//	}
//	sig, err := client.SignChannelSessionKeyState(state)
//	state.UserSig = sig
//	err = client.SubmitChannelSessionKeyState(ctx, state)
func (c *Client) SignChannelSessionKeyState(state core.ChannelSessionKeyStateV1) (string, error) {
	metadataHash, err := core.GetChannelSessionKeyAuthMetadataHashV1(state.Version, state.Assets, state.ExpiresAt.Unix())
	if err != nil {
		return "", fmt.Errorf("failed to compute metadata hash: %w", err)
	}

	packed, err := core.PackChannelKeyStateV1(state.SessionKey, metadataHash)
	if err != nil {
		return "", fmt.Errorf("failed to pack channel session key state: %w", err)
	}

	ethMsgSigner, err := sign.NewEthereumMsgSignerFromRaw(c.rawSigner)
	if err != nil {
		return "", fmt.Errorf("failed to create Ethereum message signer: %w", err)
	}

	sig, err := ethMsgSigner.Sign(packed)
	if err != nil {
		return "", fmt.Errorf("failed to sign channel session key state: %w", err)
	}

	return sig.String(), nil
}

// ApproveToken approves the ChannelHub contract to spend tokens on behalf of the user.
// This is required before depositing ERC-20 tokens. Native tokens (e.g., ETH) do not
// require approval and will return an error if attempted.
//
// Parameters:
//   - ctx: Context for the operation
//   - chainID: The blockchain network ID (e.g., 11155111 for Sepolia)
//   - asset: The asset symbol to approve (e.g., "usdc")
//   - amount: The amount to approve for spending
//
// Returns:
//   - Transaction hash of the approval transaction
//   - Error if the operation fails or the asset is a native token
func (c *Client) ApproveToken(ctx context.Context, chainID uint64, asset string, amount decimal.Decimal) (string, error) {
	blockchainClient, err := c.getOrInitBlockchainClient(ctx, chainID)
	if err != nil {
		return "", err
	}

	return blockchainClient.Approve(asset, amount)
}

// GetOnChainBalance queries the on-chain token balance (ERC-20 or native ETH) for a wallet on a specific blockchain.
//
// Parameters:
//   - ctx: Context for the operation
//   - chainID: The blockchain ID to query (e.g., 80002 for Polygon Amoy)
//   - asset: The asset symbol (e.g., "usdc", "weth")
//   - wallet: The Ethereum address to check the balance for
//
// Returns:
//   - The token balance as a decimal (already adjusted for token decimals)
//   - Error if the query fails
//
// Requirements:
//   - Blockchain RPC must be configured for the chain (use WithBlockchainRPC)
func (c *Client) GetOnChainBalance(ctx context.Context, chainID uint64, asset string, wallet string) (decimal.Decimal, error) {
	blockchainClient, err := c.getOrInitBlockchainClient(ctx, chainID)
	if err != nil {
		return decimal.Zero, err
	}

	return blockchainClient.GetTokenBalance(asset, wallet)
}
