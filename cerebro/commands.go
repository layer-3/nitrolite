package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/layer-3/nitrolite/pkg/app"
	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
)

// ============================================================================
// Help & Config
// ============================================================================

func (o *Operator) showHelp() {
	fmt.Println(`
Clearnode CLI - SDK Development Tool
=====================================

SETUP COMMANDS
  help                          Display this help message
  config                        Display current configuration
  wallet                        Display wallet address
  import wallet                 Configure wallet (import or generate)
  import rpc <chain_id> <url>   Configure blockchain RPC endpoint

HIGH-LEVEL OPERATIONS (Smart Client)
  token-balance <chain_id> <asset>             Check on-chain token balance for your wallet
  approve <chain_id> <asset> <amount>          Approve token spending for deposits
  deposit <chain_id> <asset> <amount>          Deposit to channel (auto-create if needed)
  withdraw <chain_id> <asset> <amount>         Withdraw from channel
  transfer <recipient> <asset> <amount>        Transfer to another wallet
  acknowledge <asset>                          Acknowledge transfer or channel creation
  close-channel <asset>                        Close home channel on-chain
  checkpoint <asset>                           Submit latest state on-chain

NODE INFORMATION (Base Client)
  ping                          Test node connection
  node info                     Get node configuration
  chains                        List supported blockchains
  assets [chain_id]             List supported assets (optionally filter by chain)

USER QUERIES (Base Client)
  balances [wallet]             Get user balances (defaults to configured wallet)
  transactions [wallet]         Get transaction history (defaults to configured wallet)
  action-allowances [wallet]    Get action allowances (defaults to configured wallet)

LOW-LEVEL STATE MANAGEMENT (Base Client)
  state [wallet] <asset>        Get latest state (wallet defaults to configured)
  home-channel [wallet] <asset> Get home channel (wallet defaults to configured)
  escrow-channel <channel_id>   Get escrow channel by ID

APP REGISTRY
  app-info <app_id>                    Show application details
  my-apps                              List your registered applications
  register-app <app_id> [no-approval]  Register a new application

LOW-LEVEL APP SESSIONS (Base Client)
  app-sessions                  List app sessions

SESSION KEY MANAGEMENT
  generate-session-key                                                Generate or import session key (stores locally)
  session-key                                                         Show current session key info
  clear-session-key                                                   Clear session key, revert to default wallet signer
  create-channel-session-key <session_key> <expires_hours> <assets>   Register channel session key (auto-activates if stored)
  channel-session-keys                                                List active channel session keys
  create-app-session-key <session_key> <expires_hours> [app_ids] [session_ids]  Register app session key (IDs: comma-separated)
  app-session-keys                                                    List active app session keys

OTHER
  exit                          Exit the CLI

EXAMPLES
  import wallet
  import rpc 80002 https://polygon-amoy.g.alchemy.com/v2/KEY
  approve 80002 usdc 1000000
  deposit 80002 usdc 100
  transfer 0x1234... usdc 50
  balances              # Uses configured wallet
  balances 0x1234...    # Query specific wallet
  state usdc            # Get state for USDC
  chains
  generate-session-key                               # Step 1: generate/import
  create-channel-session-key 0xabcd... 24 usdc,weth  # Step 2: register + activate
  create-app-session-key 0xabcd... 24 app1,app2`)
}

func (o *Operator) showConfig(ctx context.Context) {
	fmt.Println("Current Configuration")
	fmt.Println("=====================")

	// Private key status
	_, err := o.store.GetPrivateKey()
	if err != nil {
		fmt.Println("Wallet:     Not configured")
	} else {
		// Get signer to show address
		privateKey, _ := o.store.GetPrivateKey()
		signer, err := sign.NewEthereumRawSigner(privateKey)
		if err == nil {
			fmt.Printf("Wallet:     Configured (%s)\n", signer.PublicKey().Address().String())
		} else {
			fmt.Println("Wallet:     Configured")
		}
	}

	// Session key status
	skPrivateKey, _, _, skErr := o.store.GetSessionKey()
	if skErr != nil {
		fmt.Println("Session Key: Not configured (using default wallet signer)")
	} else {
		skSigner, skSignerErr := sign.NewEthereumRawSigner(skPrivateKey)
		if skSignerErr == nil {
			fmt.Printf("Session Key: Configured (%s)\n", skSigner.PublicKey().Address().String())
		} else {
			fmt.Println("Session Key: Configured (invalid key)")
		}
	}

	// RPC status
	rpcs, err := o.store.GetAllRPCs()
	if err != nil || len(rpcs) == 0 {
		fmt.Println("RPCs:       None configured")
	} else {
		fmt.Printf("RPCs:       %d configured\n", len(rpcs))
		for chainID, rpcURL := range rpcs {
			// Truncate URL for display
			displayURL := rpcURL
			if len(displayURL) > 50 {
				displayURL = displayURL[:47] + "..."
			}
			fmt.Printf("   - Chain %d: %s\n", chainID, displayURL)
		}
	}

	// Node info
	nodeConfig, err := o.client.GetConfig(ctx)
	if err == nil {
		fmt.Printf("\nNode Info\n")
		fmt.Printf("   Address:   %s\n", nodeConfig.NodeAddress)
		fmt.Printf("   Version:   %s\n", nodeConfig.NodeVersion)
		fmt.Printf("   Chains:    %d\n", len(nodeConfig.Blockchains))
	}
}

// ============================================================================
// Wallet Commands
// ============================================================================

func (o *Operator) showWallet(_ context.Context) {
	// Get private key
	privateKey, err := o.store.GetPrivateKey()
	if err != nil {
		fmt.Println("ERROR: No wallet configured")
		fmt.Println("INFO: Use 'import wallet' to configure wallet")
		return
	}

	// Create signer to get address
	signer, err := sign.NewEthereumRawSigner(privateKey)
	if err != nil {
		fmt.Printf("ERROR: Failed to get wallet address: %v\n", err)
		return
	}

	address := signer.PublicKey().Address().String()

	fmt.Println("Wallet Configuration")
	fmt.Println("====================")
	fmt.Printf("Address: %s\n", address)
}

// ============================================================================
// Import Commands
// ============================================================================

func (o *Operator) importWallet(_ context.Context) {
	fmt.Println("Wallet Configuration")
	fmt.Println("====================")
	fmt.Println()
	fmt.Println("Choose an option:")
	fmt.Println("  1. Import existing private key")
	fmt.Println("  2. Generate new wallet")
	fmt.Println()
	fmt.Print("Enter choice (1 or 2): ")

	var choice string
	fmt.Scanln(&choice)
	choice = strings.TrimSpace(choice)

	var privateKey string
	var signer sign.Signer
	var err error

	switch choice {
	case "1":
		// Import existing key
		fmt.Println()
		fmt.Println("Import Existing Wallet")
		fmt.Print("Enter private key (with or without 0x prefix): ")
		fmt.Scanln(&privateKey)

		privateKey = strings.TrimSpace(privateKey)
		if privateKey == "" {
			fmt.Println("ERROR: Private key cannot be empty")
			return
		}

		// Validate by creating signer
		signer, err = sign.NewEthereumRawSigner(privateKey)
		if err != nil {
			fmt.Printf("ERROR: Invalid private key: %v\n", err)
			return
		}

	case "2":
		// Generate new wallet
		fmt.Println()
		fmt.Println("Generate New Wallet")
		privateKey, err = generatePrivateKey()
		if err != nil {
			fmt.Printf("ERROR: Failed to generate private key: %v\n", err)
			return
		}

		signer, err = sign.NewEthereumRawSigner(privateKey)
		if err != nil {
			fmt.Printf("ERROR: Failed to create signer: %v\n", err)
			return
		}

		fmt.Println()
		fmt.Println("WARNING: Save your private key securely!")
		fmt.Println("=========================================")
		fmt.Printf("Private Key: %s\n", privateKey)
		fmt.Println("=========================================")
		fmt.Println()
		fmt.Print("Type 'I have saved my private key' to continue: ")

		var confirmation string
		fmt.Scanln(&confirmation)
		// Read the full line
		if confirmation == "" {
			fmt.Println("ERROR: You must confirm that you saved the private key")
			return
		}

	default:
		fmt.Println("ERROR: Invalid choice")
		return
	}

	// Save to storage
	if err := o.store.SetPrivateKey(privateKey); err != nil {
		fmt.Printf("ERROR: Failed to save private key: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Wallet configured successfully\n")
	fmt.Printf("Address: %s\n", signer.PublicKey().Address().String())

	if choice == "2" {
		fmt.Println()
		fmt.Println("Security Recommendations:")
		fmt.Println("   - Store your private key in a secure location")
		fmt.Println("   - Never share your private key with anyone")
		fmt.Println("   - Consider using a hardware wallet for large amounts")
	}
}

func (o *Operator) importRPC(_ context.Context, chainIDStr, rpcURL string) {
	chainID, err := o.parseChainID(chainIDStr)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	if err := o.store.SetRPC(chainID, rpcURL); err != nil {
		fmt.Printf("ERROR: Failed to save RPC: %v\n", err)
		return
	}
	// TODO: add to SDK Client dynamically

	fmt.Printf("SUCCESS: RPC configured for chain %d\n", chainID)
}

func (o *Operator) setHomeBlockchain(_ context.Context, asset, chainIDStr string) {
	chainID, err := o.parseChainID(chainIDStr)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	if err := o.client.SetHomeBlockchain(asset, chainID); err != nil {
		fmt.Printf("ERROR: Failed to set home blockchain: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Home blockchain for asset %s is set to %d\n", asset, chainID)
}

// ============================================================================
// High-Level Operations (Smart Client)
// ============================================================================

func (o *Operator) deposit(ctx context.Context, chainIDStr, asset, amountStr string) {
	chainID, err := o.parseChainID(chainIDStr)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	amount, err := o.parseAmount(amountStr)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Printf("Depositing %s %s on chain %d...\n", amount.String(), asset, chainID)

	_, err = o.client.Deposit(ctx, chainID, asset, amount)
	if err != nil {
		fmt.Printf("ERROR: Deposit failed: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Deposit state prepared. Run 'checkpoint %s' to submit to the blockchain.\n", asset)
}

func (o *Operator) tokenBalance(ctx context.Context, chainIDStr, asset string) {
	chainID, err := o.parseChainID(chainIDStr)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	wallet := o.getImportedWalletAddress()
	if wallet == "" {
		fmt.Println("ERROR: No wallet configured. Use 'import wallet' first.")
		return
	}

	fmt.Printf("Querying on-chain %s balance on chain %d for %s...\n", asset, chainID, wallet)

	balance, err := o.client.GetOnChainBalance(ctx, chainID, asset, wallet)
	if err != nil {
		fmt.Printf("ERROR: Failed to get on-chain balance: %v\n", err)
		return
	}

	fmt.Printf("On-chain %s balance: %s\n", asset, balance.String())
}

func (o *Operator) withdraw(ctx context.Context, chainIDStr, asset, amountStr string) {
	chainID, err := o.parseChainID(chainIDStr)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	amount, err := o.parseAmount(amountStr)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Printf("Withdrawing %s %s from chain %d...\n", amount.String(), asset, chainID)

	_, err = o.client.Withdraw(ctx, chainID, asset, amount)
	if err != nil {
		fmt.Printf("ERROR: Withdrawal failed: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Withdrawal state prepared. Run 'checkpoint %s' to submit to the blockchain.\n", asset)
}

func (o *Operator) transfer(ctx context.Context, recipient, asset, amountStr string) {
	amount, err := o.parseAmount(amountStr)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Printf("Transferring %s %s to %s...\n", amount.String(), asset, recipient)

	_, err = o.client.Transfer(ctx, recipient, asset, amount)
	if err != nil {
		fmt.Printf("ERROR: Transfer failed: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Transfer completed\n")
}

func (o *Operator) closeChannel(ctx context.Context, asset string) {
	fmt.Printf("Initiating channel closure for asset: %s...\n", asset)
	fmt.Println("INFO: This involves signing a final state and submitting a transaction to the blockchain.")

	_, err := o.client.CloseHomeChannel(ctx, asset)
	if err != nil {
		fmt.Printf("ERROR: Failed to close channel: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Channel close state prepared. Run 'checkpoint %s' to submit to the blockchain.\n", asset)
}

func (o *Operator) acknowledge(ctx context.Context, asset string) {
	fmt.Printf("Acknowledging state for asset: %s...\n", asset)

	_, err := o.client.Acknowledge(ctx, asset)
	if err != nil {
		fmt.Printf("ERROR: Acknowledgement failed: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Acknowledgement completed\n")
}

func (o *Operator) approveToken(ctx context.Context, chainIDStr, asset, amountStr string) {
	chainID, err := o.parseChainID(chainIDStr)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	amount, err := o.parseAmount(amountStr)
	if err != nil {
		fmt.Printf("ERROR: %v\n", err)
		return
	}

	fmt.Printf("Approving %s %s on chain %d...\n", amount.String(), asset, chainID)

	txHash, err := o.client.ApproveToken(ctx, chainID, asset, amount)
	if err != nil {
		fmt.Printf("ERROR: Approve failed: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Token spending approved\n")
	fmt.Printf("Transaction Hash: %s\n", txHash)
}

func (o *Operator) checkpoint(ctx context.Context, asset string) {
	fmt.Printf("Submitting checkpoint for asset: %s...\n", asset)
	fmt.Println("INFO: This submits the latest co-signed state to the blockchain.")

	txHash, err := o.client.Checkpoint(ctx, asset)
	if err != nil {
		fmt.Printf("ERROR: Checkpoint failed: %v\n", err)
		return
	}

	fmt.Printf("SUCCESS: Checkpoint completed\n")
	fmt.Printf("Transaction Hash: %s\n", txHash)
}

// ============================================================================
// Node Information (Base Client)
// ============================================================================

func (o *Operator) ping(ctx context.Context) {
	fmt.Print("Pinging node... ")
	err := o.client.Ping(ctx)
	if err != nil {
		fmt.Printf("ERROR: Failed: %v\n", err)
		return
	}
	fmt.Println("Success")
}

func (o *Operator) nodeInfo(ctx context.Context) {
	config, err := o.client.GetConfig(ctx)
	if err != nil {
		fmt.Printf("ERROR: Failed to get node info: %v\n", err)
		return
	}

	fmt.Println("Node Information")
	fmt.Println("================")
	fmt.Printf("Address:   %s\n", config.NodeAddress)
	fmt.Printf("Version:   %s\n", config.NodeVersion)
	fmt.Printf("Chains:    %d\n", len(config.Blockchains))

	if len(config.SupportedSigValidators) > 0 {
		fmt.Printf("\nSupported Signature Validators:\n")
		for _, v := range config.SupportedSigValidators {
			switch v {
			case core.ChannelSignerType_Default:
				fmt.Printf("  - Default Wallet (0x%02x)\n", uint8(v))
			case core.ChannelSignerType_SessionKey:
				fmt.Printf("  - Session Key (0x%02x)\n", uint8(v))
			default:
				fmt.Printf("  - Unknown (0x%02x)\n", uint8(v))
			}
		}
	}

	fmt.Println("\nSupported Blockchains:")
	for _, bc := range config.Blockchains {
		fmt.Printf("  - %s (ID: %d)\n", bc.Name, bc.ID)
		fmt.Printf("    Contract:   %s\n", bc.ChannelHubAddress)
	}
}

func (o *Operator) listChains(ctx context.Context) {
	chains, err := o.client.GetBlockchains(ctx)
	if err != nil {
		fmt.Printf("ERROR: Failed to list chains: %v\n", err)
		return
	}

	fmt.Printf("Supported Blockchains (%d)\n", len(chains))
	fmt.Println("==========================")
	for _, chain := range chains {
		fmt.Printf("- %s\n", chain.Name)
		fmt.Printf("  Chain ID:  %d\n", chain.ID)
		fmt.Printf("  Contract:  %s\n", chain.ChannelHubAddress)

		// Check if RPC is configured
		_, err := o.store.GetRPC(chain.ID)
		if err == nil {
			fmt.Printf("  RPC:       Configured\n")
		} else {
			fmt.Printf("  RPC:       Not configured\n")
		}
		fmt.Println()
	}
}

func (o *Operator) listAssets(ctx context.Context, chainIDStr string) {
	var chainID *uint64
	if chainIDStr != "" {
		parsed, err := o.parseChainID(chainIDStr)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			return
		}
		chainID = &parsed
	}

	assets, err := o.client.GetAssets(ctx, chainID)
	if err != nil {
		fmt.Printf("ERROR: Failed to list assets: %v\n", err)
		return
	}

	if chainID != nil {
		fmt.Printf("Assets on Chain %d (%d)\n", *chainID, len(assets))
	} else {
		fmt.Printf("All Supported Assets (%d)\n", len(assets))
	}
	fmt.Println("==========================")

	for _, asset := range assets {
		fmt.Printf("- %s (%s)\n", asset.Name, asset.Symbol)
		fmt.Printf("  Decimals:  %d\n", asset.Decimals)
		fmt.Printf("  Tokens:    %d connected\n", len(asset.Tokens))

		// Show token details
		if len(asset.Tokens) > 0 {
			if chainID != nil {
				// When filtering by chain, show detailed info for each token
				for _, token := range asset.Tokens {
					fmt.Printf("    - Chain %d: %s\n", token.BlockchainID, token.Address)
					fmt.Printf("      Decimals: %d\n", token.Decimals)
				}
			} else {
				// When showing all assets, list chains with their token details
				for _, token := range asset.Tokens {
					fmt.Printf("    - Chain %d: %s (decimals: %d)\n", token.BlockchainID, token.Address, token.Decimals)
				}
			}
		}
		fmt.Println()
	}
}

// ============================================================================
// User Queries (Base Client)
// ============================================================================

func (o *Operator) getBalances(ctx context.Context, wallet string) {
	balances, err := o.client.GetBalances(ctx, wallet)
	if err != nil {
		fmt.Printf("ERROR: Failed to get balances: %v\n", err)
		return
	}

	fmt.Printf("Balances for %s\n", wallet)
	fmt.Println("========================================")
	if len(balances) == 0 {
		fmt.Println("No balances found")
		return
	}

	for _, balance := range balances {
		fmt.Printf("- %s: %s\n", balance.Asset, balance.Balance.String())
	}
}

func (o *Operator) getHomeChannel(ctx context.Context, wallet, asset string) {
	channel, err := o.client.GetHomeChannel(ctx, wallet, asset)
	if err != nil {
		fmt.Printf("ERROR: Failed to get home channel: %v\n", err)
		return
	}

	typeStr := "unknown"
	switch channel.Type {
	case core.ChannelTypeHome:
		typeStr = "Home"
	case core.ChannelTypeEscrow:
		typeStr = "Escrow"
	}

	statusStr := "unknown"
	switch channel.Status {
	case core.ChannelStatusVoid:
		statusStr = "Void"
	case core.ChannelStatusOpen:
		statusStr = "Open"
	case core.ChannelStatusChallenged:
		statusStr = "Challenged"
	case core.ChannelStatusClosed:
		statusStr = "Closed"
	}

	fmt.Printf("Home Channel for %s (%s)\n", wallet, asset)
	fmt.Println("=========================================")
	fmt.Printf("Channel ID:  %s\n", channel.ChannelID)
	fmt.Printf("Type:        %s\n", typeStr)
	fmt.Printf("Status:      %s\n", statusStr)
	fmt.Printf("Version:     %d\n", channel.StateVersion)
	fmt.Printf("Nonce:       %d\n", channel.Nonce)
	fmt.Printf("Chain ID:    %d\n", channel.BlockchainID)
	fmt.Printf("Token:       %s\n", channel.TokenAddress)
	fmt.Printf("Challenge:   %d seconds\n", channel.ChallengeDuration)
}

func (o *Operator) getEscrowChannel(ctx context.Context, escrowChannelID string) {
	channel, err := o.client.GetEscrowChannel(ctx, escrowChannelID)
	if err != nil {
		fmt.Printf("ERROR: Failed to get escrow channel: %v\n", err)
		return
	}

	typeStr := "unknown"
	switch channel.Type {
	case core.ChannelTypeHome:
		typeStr = "Home"
	case core.ChannelTypeEscrow:
		typeStr = "Escrow"
	}

	statusStr := "unknown"
	switch channel.Status {
	case core.ChannelStatusVoid:
		statusStr = "Void"
	case core.ChannelStatusOpen:
		statusStr = "Open"
	case core.ChannelStatusChallenged:
		statusStr = "Challenged"
	case core.ChannelStatusClosed:
		statusStr = "Closed"
	}

	fmt.Printf("Escrow Channel %s\n", escrowChannelID)
	fmt.Println("=========================================")
	fmt.Printf("Channel ID:  %s\n", channel.ChannelID)
	fmt.Printf("User Wallet: %s\n", channel.UserWallet)
	fmt.Printf("Type:        %s\n", typeStr)
	fmt.Printf("Status:      %s\n", statusStr)
	fmt.Printf("Version:     %d\n", channel.StateVersion)
	fmt.Printf("Nonce:       %d\n", channel.Nonce)
	fmt.Printf("Chain ID:    %d\n", channel.BlockchainID)
	fmt.Printf("Token:       %s\n", channel.TokenAddress)
	fmt.Printf("Challenge:   %d seconds\n", channel.ChallengeDuration)
}

func (o *Operator) listTransactions(ctx context.Context, wallet string) {
	limit := uint32(20)
	opts := &sdk.GetTransactionsOptions{
		Pagination: &core.PaginationParams{
			Limit: &limit,
		},
	}

	txs, meta, err := o.client.GetTransactions(ctx, wallet, opts)
	if err != nil {
		fmt.Printf("ERROR: Failed to list transactions: %v\n", err)
		return
	}

	fmt.Printf("Recent Transactions for %s (Showing %d of %d)\n", wallet, len(txs), meta.TotalCount)
	fmt.Println("=================================================")
	if len(txs) == 0 {
		fmt.Println("No transactions found")
		return
	}

	for _, tx := range txs {
		fmt.Printf("\n- %s\n", tx.TxType.String())
		fmt.Printf("  Hash:      %s\n", tx.ID)
		fmt.Printf("  From:      %s\n", tx.FromAccount)
		fmt.Printf("  To:        %s\n", tx.ToAccount)
		fmt.Printf("  Amount:    %s %s\n", tx.Amount.String(), tx.Asset)
		fmt.Printf("  Created:   %s\n", tx.CreatedAt.Format("2006-01-02 15:04:05"))
	}
}

func (o *Operator) getActionAllowances(ctx context.Context, wallet string) {
	allowances, err := o.client.GetActionAllowances(ctx, wallet)
	if err != nil {
		fmt.Printf("ERROR: Failed to get action allowances: %v\n", err)
		return
	}

	fmt.Printf("Action Allowances for %s\n", wallet)
	fmt.Println("========================================")
	if len(allowances) == 0 {
		fmt.Println("No action allowances found")
		return
	}

	for _, a := range allowances {
		fmt.Printf("- %s\n", a.GatedAction)
		fmt.Printf("  Window:    %s\n", a.TimeWindow)
		fmt.Printf("  Used:      %d / %d\n", a.Used, a.Allowance)
		remaining := uint64(0)
		if a.Allowance > a.Used {
			remaining = a.Allowance - a.Used
		}
		fmt.Printf("  Remaining: %d\n", remaining)
	}
}

// ============================================================================
// App Registry
// ============================================================================

func (o *Operator) getApps(ctx context.Context, appID *string, ownerWallet *string) {
	fmt.Println("Fetching registered applications...")

	apps, _, err := o.client.GetApps(ctx, &sdk.GetAppsOptions{
		AppID:       appID,
		OwnerWallet: ownerWallet,
	})
	if err != nil {
		fmt.Printf("ERROR: Failed to get apps: %v\n", err)
		return
	}

	if len(apps) == 0 {
		fmt.Println("No applications found.")
		return
	}

	fmt.Printf("Found %d application(s):\n\n", len(apps))
	for _, a := range apps {
		fmt.Printf("  App ID:       %s\n", a.App.ID)
		fmt.Printf("  Owner:        %s\n", a.App.OwnerWallet)
		fmt.Printf("  Version:      %d\n", a.App.Version)
		if a.App.CreationApprovalNotRequired {
			fmt.Println("  Approval:     Not required")
		} else {
			fmt.Println("  Approval:     Required")
		}
		if a.App.Metadata != "" {
			fmt.Printf("  Metadata:     %s\n", a.App.Metadata)
		}
		fmt.Printf("  Created:      %s\n", a.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  Updated:      %s\n", a.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Println()
	}
}

func (o *Operator) registerApp(ctx context.Context, appID, metadata string, creationApprovalNotRequired bool) {
	fmt.Printf("Registering application: %s...\n", appID)

	err := o.client.RegisterApp(ctx, appID, metadata, creationApprovalNotRequired)
	if err != nil {
		fmt.Printf("ERROR: Failed to register app: %v\n", err)
		return
	}

	fmt.Println("SUCCESS: Application registered")
	fmt.Printf("  App ID:   %s\n", appID)
	if creationApprovalNotRequired {
		fmt.Println("  Approval: Not required for session creation")
	} else {
		fmt.Println("  Approval: Required for session creation")
	}
}

// ============================================================================
// Low-Level State Management (Base Client)
// ============================================================================

func (o *Operator) getLatestState(ctx context.Context, wallet, asset string) {
	state, err := o.client.GetLatestState(ctx, wallet, asset, false)
	if err != nil {
		fmt.Printf("ERROR: Failed to get state: %v\n", err)
		return
	}

	fmt.Printf("Latest State for %s (%s)\n", wallet, asset)
	fmt.Println("====================================")
	fmt.Printf("Version:    %d\n", state.Version)
	fmt.Printf("Epoch:      %d\n", state.Epoch)
	fmt.Printf("State ID:   %s\n", state.ID)
	if state.HomeChannelID != nil {
		fmt.Printf("Channel:    %s\n", *state.HomeChannelID)
	}
	fmt.Printf("\nHome Ledger:\n")
	fmt.Printf("  Chain:      %d\n", state.HomeLedger.BlockchainID)
	fmt.Printf("  Token:      %s\n", state.HomeLedger.TokenAddress)
	fmt.Printf("  User NetFlow:   %s\n", state.HomeLedger.UserNetFlow.String())
	fmt.Printf("  User Bal:   %s\n", state.HomeLedger.UserBalance.String())
	fmt.Printf("  Node Bal:   %s\n", state.HomeLedger.NodeBalance.String())
	fmt.Printf("  Node NetFlow:   %s\n", state.HomeLedger.NodeNetFlow.String())
	fmt.Printf("\nTransition:\n")
	fmt.Printf("    Type:          %s\n", state.Transition.Type.String())
	fmt.Printf("    TransactionID: %s\n", state.Transition.TxID)
	fmt.Printf("    AccountID:     %s\n", state.Transition.TxID)
	fmt.Printf("    Amount:        %s\n", state.Transition.Amount.String())
}

// ============================================================================
// Low-Level App Sessions (Base Client)
// ============================================================================

func (o *Operator) listAppSessions(ctx context.Context, wallet string) {
	limit := uint32(20)
	sessions, meta, err := o.client.GetAppSessions(ctx, &sdk.GetAppSessionsOptions{
		Participant: &wallet,
		Pagination: &core.PaginationParams{
			Limit: &limit,
		},
	})
	if err != nil {
		fmt.Printf("ERROR: Failed to list app sessions: %v\n", err)
		return
	}

	fmt.Printf("App Sessions (Total: %d)\n", meta.TotalCount)
	fmt.Println("=========================")
	if len(sessions) == 0 {
		fmt.Println("No app sessions found")
		return
	}

	for _, session := range sessions {
		fmt.Printf("\n- Session %s\n", session.AppSessionID)
		fmt.Printf("  Version:      %d\n", session.Version)
		fmt.Printf("  Nonce:        %d\n", session.AppDefinition.Nonce)
		fmt.Printf("  Quorum:       %d\n", session.AppDefinition.Quorum)
		fmt.Printf("  Closed:       %v\n", session.IsClosed)
		fmt.Printf("  Participants: %d\n", len(session.AppDefinition.Participants))
		fmt.Printf("  Allocations:  %d\n", len(session.Allocations))
	}
}

// ============================================================================
// Session Key Management
// ============================================================================

func (o *Operator) generateSessionKey(_ context.Context) {
	fmt.Println("Session Key Setup")
	fmt.Println("=================")
	fmt.Println()
	fmt.Println("Choose an option:")
	fmt.Println("  1. Generate new session key")
	fmt.Println("  2. Import existing private key")
	fmt.Println()
	fmt.Print("Enter choice (1 or 2): ")

	var choice string
	fmt.Scanln(&choice)
	choice = strings.TrimSpace(choice)

	var privateKeyHex string
	var err error

	switch choice {
	case "1":
		privateKeyHex, err = generatePrivateKey()
		if err != nil {
			fmt.Printf("ERROR: Failed to generate session key: %v\n", err)
			return
		}
	case "2":
		fmt.Print("Enter session key private key (hex): ")
		fmt.Scanln(&privateKeyHex)
		privateKeyHex = strings.TrimSpace(privateKeyHex)
		if privateKeyHex == "" {
			fmt.Println("ERROR: Private key cannot be empty")
			return
		}
	default:
		fmt.Println("ERROR: Invalid choice")
		return
	}

	signer, err := sign.NewEthereumRawSigner(privateKeyHex)
	if err != nil {
		fmt.Printf("ERROR: Invalid private key: %v\n", err)
		return
	}

	// Store the session key private key locally (no metadata yet — will be set on registration)
	if err := o.store.SetSessionKeyPrivateKey(privateKeyHex); err != nil {
		fmt.Printf("ERROR: Failed to store session key: %v\n", err)
		return
	}

	address := signer.PublicKey().Address().String()

	fmt.Println()
	fmt.Println("SUCCESS: Session key stored locally")
	fmt.Printf("  Address: %s\n", address)
	if choice == "1" {
		fmt.Printf("  Private Key: %s\n", privateKeyHex)
		fmt.Println()
		fmt.Println("WARNING: Save the private key securely!")
	}
	fmt.Println()
	fmt.Println("Next step: Register it on the clearnode with:")
	fmt.Printf("  create-channel-session-key %s <expires_hours> <assets>\n", address)
}

func (o *Operator) showSessionKey() {
	skPrivateKey, metadataHash, _, err := o.store.GetSessionKey()
	if err != nil {
		// Check if we have just the private key (generated but not yet registered)
		pk, pkErr := o.store.GetSessionKeyPrivateKey()
		if pkErr != nil {
			fmt.Println("No session key configured")
			fmt.Println("INFO: Use 'generate-session-key' to create one.")
			return
		}
		signer, sigErr := sign.NewEthereumRawSigner(pk)
		if sigErr != nil {
			fmt.Printf("ERROR: Invalid stored session key: %v\n", sigErr)
			return
		}
		fmt.Println("Session Key Configuration")
		fmt.Println("=========================")
		fmt.Printf("Address: %s\n", signer.PublicKey().Address().String())
		fmt.Println("Status:  Stored locally (not yet registered on clearnode)")
		fmt.Println()
		fmt.Println("Next step: Register it with:")
		fmt.Printf("  create-channel-session-key %s <expires_hours> <assets>\n", signer.PublicKey().Address().String())
		return
	}

	signer, err := sign.NewEthereumRawSigner(skPrivateKey)
	if err != nil {
		fmt.Printf("ERROR: Invalid stored session key: %v\n", err)
		return
	}

	fmt.Println("Session Key Configuration")
	fmt.Println("=========================")
	fmt.Printf("Address:       %s\n", signer.PublicKey().Address().String())
	fmt.Printf("Metadata Hash: %s\n", metadataHash)
	fmt.Println("Status:        Active (used for state signing)")
}

func (o *Operator) clearSessionKey() {
	if err := o.store.ClearSessionKey(); err != nil {
		fmt.Printf("ERROR: Failed to clear session key: %v\n", err)
		return
	}

	fmt.Println("Reconnecting with default wallet signer...")
	if err := o.reconnect(); err != nil {
		fmt.Printf("ERROR: Failed to reconnect: %v\n", err)
		fmt.Println("INFO: Session key cleared but reconnect failed. Try restarting the CLI.")
		return
	}

	fmt.Println("SUCCESS: Session key cleared. Using default wallet signer.")
}

func (o *Operator) createChannelSessionKey(ctx context.Context, sessionKeyAddr, expiresHoursStr, assetsStr string) {
	expiresHours, err := strconv.ParseUint(expiresHoursStr, 10, 64)
	if err != nil {
		fmt.Printf("ERROR: Invalid expiration hours: %s\n", expiresHoursStr)
		return
	}

	assets := strings.Split(assetsStr, ",")
	for i := range assets {
		assets[i] = strings.TrimSpace(assets[i])
	}

	wallet := o.getImportedWalletAddress()
	if wallet == "" {
		fmt.Println("ERROR: No wallet configured. Use 'import wallet' first.")
		return
	}

	// Determine version by fetching existing keys
	var version uint64 = 1
	existingStates, err := o.client.GetLastChannelKeyStates(ctx, wallet, &sdk.GetLastChannelKeyStatesOptions{
		SessionKey: &sessionKeyAddr,
	})
	if err == nil && len(existingStates) > 0 {
		version = existingStates[0].Version + 1
	}

	expiresAt := time.Now().Add(time.Duration(expiresHours) * time.Hour)

	state := core.ChannelSessionKeyStateV1{
		UserAddress: wallet,
		SessionKey:  sessionKeyAddr,
		Version:     version,
		Assets:      assets,
		ExpiresAt:   expiresAt,
	}

	fmt.Printf("Signing channel session key (version %d)...\n", version)
	sig, err := o.client.SignChannelSessionKeyState(state)
	if err != nil {
		fmt.Printf("ERROR: Failed to sign session key state: %v\n", err)
		return
	}
	state.UserSig = sig

	fmt.Println("Submitting channel session key state...")
	if err := o.client.SubmitChannelSessionKeyState(ctx, state); err != nil {
		fmt.Printf("ERROR: Failed to submit session key state: %v\n", err)
		return
	}

	fmt.Println("SUCCESS: Channel session key registered")
	fmt.Printf("  Session Key: %s\n", sessionKeyAddr)
	fmt.Printf("  Version:     %d\n", version)
	fmt.Printf("  Assets:      %s\n", strings.Join(assets, ", "))
	fmt.Printf("  Expires At:  %s\n", expiresAt.Format("2006-01-02 15:04:05"))

	// If we have a stored session key matching this address, activate it as the state signer
	storedPK, pkErr := o.store.GetSessionKeyPrivateKey()
	if pkErr != nil {
		return
	}
	storedSigner, sigErr := sign.NewEthereumRawSigner(storedPK)
	if sigErr != nil {
		return
	}
	if !strings.EqualFold(storedSigner.PublicKey().Address().String(), sessionKeyAddr) {
		return
	}

	// Compute metadata hash and store full session key data
	metadataHash, err := core.GetChannelSessionKeyAuthMetadataHashV1(version, assets, expiresAt.Unix())
	if err != nil {
		fmt.Printf("WARNING: Failed to compute metadata hash: %v\n", err)
		return
	}

	if err := o.store.SetSessionKey(storedPK, metadataHash.Hex(), sig); err != nil {
		fmt.Printf("WARNING: Failed to store session key data: %v\n", err)
		return
	}

	fmt.Println("Activating session key as state signer...")
	if err := o.reconnect(); err != nil {
		fmt.Printf("WARNING: Failed to reconnect: %v\n", err)
		fmt.Println("INFO: Session key is registered. Restart the CLI to activate it.")
		return
	}

	fmt.Println("SUCCESS: Session key is now used for state signing")
}

func (o *Operator) listChannelSessionKeys(ctx context.Context, wallet string) {
	states, err := o.client.GetLastChannelKeyStates(ctx, wallet, nil)
	if err != nil {
		fmt.Printf("ERROR: Failed to get channel session keys: %v\n", err)
		return
	}

	fmt.Printf("Channel Session Keys for %s (%d)\n", wallet, len(states))
	fmt.Println("===========================================")
	if len(states) == 0 {
		fmt.Println("No active channel session keys found")
		return
	}

	for _, state := range states {
		fmt.Printf("\n- Session Key: %s\n", state.SessionKey)
		fmt.Printf("  Version:    %d\n", state.Version)
		fmt.Printf("  Assets:     %s\n", strings.Join(state.Assets, ", "))
		fmt.Printf("  Expires At: %s\n", state.ExpiresAt.Format("2006-01-02 15:04:05"))
	}
}

func (o *Operator) createAppSessionKey(ctx context.Context, sessionKeyAddr, expiresHoursStr, appIDsStr, sessionIDsStr string) {
	expiresHours, err := strconv.ParseUint(expiresHoursStr, 10, 64)
	if err != nil {
		fmt.Printf("ERROR: Invalid expiration hours: %s\n", expiresHoursStr)
		return
	}

	var applicationIDs []string
	if appIDsStr != "" {
		applicationIDs = strings.Split(appIDsStr, ",")
		for i := range applicationIDs {
			applicationIDs[i] = strings.TrimSpace(applicationIDs[i])
		}
	}

	var appSessionIDs []string
	if sessionIDsStr != "" {
		appSessionIDs = strings.Split(sessionIDsStr, ",")
		for i := range appSessionIDs {
			appSessionIDs[i] = strings.TrimSpace(appSessionIDs[i])
		}
	}

	wallet := o.getImportedWalletAddress()
	if wallet == "" {
		fmt.Println("ERROR: No wallet configured. Use 'import wallet' first.")
		return
	}

	// Determine version by fetching existing keys
	var version uint64 = 1
	existingStates, err := o.client.GetLastAppKeyStates(ctx, wallet, &sdk.GetLastKeyStatesOptions{
		SessionKey: &sessionKeyAddr,
	})
	if err == nil && len(existingStates) > 0 {
		for _, s := range existingStates {
			if s.Version >= version {
				version = s.Version + 1
			}
		}
	}

	state := app.AppSessionKeyStateV1{
		UserAddress:    wallet,
		SessionKey:     sessionKeyAddr,
		Version:        version,
		ApplicationIDs: applicationIDs,
		AppSessionIDs:  appSessionIDs,
		ExpiresAt:      time.Now().Add(time.Duration(expiresHours) * time.Hour),
	}

	fmt.Printf("Signing app session key (version %d)...\n", version)
	sig, err := o.client.SignSessionKeyState(state)
	if err != nil {
		fmt.Printf("ERROR: Failed to sign session key state: %v\n", err)
		return
	}
	state.UserSig = sig

	fmt.Println("Submitting app session key state...")
	if err := o.client.SubmitAppSessionKeyState(ctx, state); err != nil {
		fmt.Printf("ERROR: Failed to submit session key state: %v\n", err)
		return
	}

	fmt.Println("SUCCESS: App session key registered")
	fmt.Printf("  Session Key:     %s\n", sessionKeyAddr)
	fmt.Printf("  Version:         %d\n", version)
	if len(applicationIDs) > 0 {
		fmt.Printf("  Application IDs: %s\n", strings.Join(applicationIDs, ", "))
	}
	if len(appSessionIDs) > 0 {
		fmt.Printf("  Session IDs:     %s\n", strings.Join(appSessionIDs, ", "))
	}
	fmt.Printf("  Expires At:      %s\n", state.ExpiresAt.Format("2006-01-02 15:04:05"))
}

func (o *Operator) listAppSessionKeys(ctx context.Context, wallet string) {
	states, err := o.client.GetLastAppKeyStates(ctx, wallet, nil)
	if err != nil {
		fmt.Printf("ERROR: Failed to get app session keys: %v\n", err)
		return
	}

	fmt.Printf("App Session Keys for %s (%d)\n", wallet, len(states))
	fmt.Println("===========================================")
	if len(states) == 0 {
		fmt.Println("No active app session keys found")
		return
	}

	for _, state := range states {
		fmt.Printf("\n- Session Key: %s\n", state.SessionKey)
		fmt.Printf("  Version:         %d\n", state.Version)
		if len(state.ApplicationIDs) > 0 {
			fmt.Printf("  Application IDs: %s\n", strings.Join(state.ApplicationIDs, ", "))
		}
		if len(state.AppSessionIDs) > 0 {
			fmt.Printf("  Session IDs:     %s\n", strings.Join(state.AppSessionIDs, ", "))
		}
		fmt.Printf("  Expires At:      %s\n", state.ExpiresAt.Format("2006-01-02 15:04:05"))
	}
}

// ============================================================================
// Helper Methods
// ============================================================================

// generatePrivateKey generates a new Ethereum private key
func generatePrivateKey() (string, error) {
	// Generate new ECDSA private key
	privateKey, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}

	// Convert to hex string
	privateKeyBytes := crypto.FromECDSA(privateKey)
	privateKeyHex := hexutil.Encode(privateKeyBytes)

	return privateKeyHex, nil
}
