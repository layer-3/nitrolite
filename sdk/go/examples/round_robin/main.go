package main

// Example: Round Robin Test
//
// Exercises every basic channel action (deposit, transfer, close, withdraw)
// for one asset across every chain on which that asset is supported.
//
// ----------------------------------------------------------------------------
// Flow
// ----------------------------------------------------------------------------
//
//  1. Preparation
//     a. Set up wallet signers for private keys A and B.
//     b. Build tokenSet: every supported token for the configured asset,
//        paired with its chain's JSON-RPC endpoint. tokenSet[i].BlockchainID
//        must have an entry in chainRPCs.
//     c. Ensure signer A holds >= minNativeBalances[chainID] of the native
//        gas token on every chain in tokenSet. Each chain pays three
//        checkpoint transactions per run (deposit + close + withdraw).
//     d. Ensure signer A holds >= transferAmount of the configured asset on
//        tokenSet[0].BlockchainID. Subsequent iterations are seeded by the
//        previous iteration's withdrawal — no other chain needs to be
//        pre-funded with the asset.
//     e. If sessionKeyPriv is non-empty, register it as a channel session
//        key for signer A and use the session-key-backed client for every
//        channel operation in the loop.
//
//  2. For i := range tokenSet (let next = (i + 1) % len(tokenSet)):
//     a. Signer A deposits transferAmount of tokenSet[i].
//     b. Signer A transfers transferAmount to signer B (off-chain).
//     c. Signer A closes the home channel for the asset.
//     d. Signer B transfers transferAmount back to signer A (lands on a
//        void state, no chain attached).
//     e. Signer A withdraws transferAmount on tokenSet[next].BlockchainID,
//        which auto-creates a new channel on that chain. On the last
//        iteration this wraps to tokenSet[0], closing the loop.
//
// Together (2.a–2.e) exercise deposit / transfer-send / close /
// transfer-receive / withdraw on every chain in tokenSet.
//
// ----------------------------------------------------------------------------
// Operational notes
// ----------------------------------------------------------------------------
//
//   - wsURL must point at a reachable nitronode WebSocket endpoint.
//   - The nitronode must maintain reserves of the asset on every chain in
//     tokenSet. 2.e withdraws tokens signer A never deposited on that chain,
//     which only works if the node holds liquidity there.
//   - tokenSet is derived at runtime from GetAssets(asset).Tokens; the
//     iteration order matches the order returned by the node.
//   - The example exits non-zero on the first failure. It does not reset
//     state, so re-running on the same wallets resumes from whatever state
//     the node holds. Fresh wallets give the cleanest preflight numbers.

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/shopspring/decimal"

	"github.com/layer-3/nitrolite/pkg/core"
	"github.com/layer-3/nitrolite/pkg/sign"
	sdk "github.com/layer-3/nitrolite/sdk/go"
)

// ============================================================================
// Configuration
// ============================================================================

const (
	wsURL = "wss://nitronode-sandbox.yellow.org/v1/ws"

	// Replace with your hex private keys. privA performs every channel
	// operation; privB only receives transfers and sends them back.
	privA = "0x7d6071..."
	privB = "0xf63695..."

	// Empty string disables the session-key path; the wallet client performs
	// channel operations directly. Otherwise this key is registered as a
	// channel session key for privA and used for the loop.
	sessionKeyPriv = ""

	// Asset symbol to test. tokenSet is built from this asset's tokens as
	// returned by the node.
	asset = "yusd"
)

// transferAmount is the value used for every deposit / transfer / withdraw
// in the loop. Keep this small for testnets.
var transferAmount = decimal.NewFromFloat(0.00001)

// chainRPCs maps blockchain ID -> JSON-RPC endpoint. Must cover every chain
// in tokenSet (derived at runtime). Missing entries fail preflight.
var chainRPCs = map[uint64]string{
	11155111: "https://sepolia.drpc.org",            // Ethereum Sepolia
	84532:    "https://sepolia.base.org",            // Base Sepolia
	80002:    "https://rpc-amoy.polygon.technology", // Polygon Amoy
	59141:    "https://rpc.sepolia.linea.build",     // Linea Sepolia
	1449000:  "https://rpc.testnet.xrplevm.org",     // XRP LVM Testnet

	1:       "https://0xrpc.io/eth",                   // Ethereum Mainnet
	14:      "https://rpc.ankr.com/flare",             // Flare Mainnet
	56:      "https://bsc.api.pocket.network",         // BNB Smart Chain Mainnet
	137:     "https://polygon-bor-rpc.publicnode.com", // Polygon Mainnet
	8453:    "https://base-rpc.publicnode.com",        // Base Mainnet
	59144:   "https://linea.drpc.org",                 // Linea Mainnet
	1440000: "https://xrpl.drpc.org",                  // XRP EVM Mainnet
}

// minNativeBalances maps blockchain ID -> minimum native gas balance for
// privA. Sized to cover three checkpoint transactions (deposit + close +
// withdraw). Tune per chain based on observed gas costs.
var minNativeBalances = map[uint64]decimal.Decimal{
	11155111: decimal.NewFromFloat(0.01),
	84532:    decimal.NewFromFloat(0.005),
	80002:    decimal.NewFromFloat(0.05),
	59141:    decimal.NewFromFloat(0.005),
	1449000:  decimal.NewFromFloat(0.001),

	1:       decimal.NewFromFloat(0.001),
	14:      decimal.NewFromFloat(0.001),
	56:      decimal.NewFromFloat(0.001),
	137:     decimal.NewFromFloat(3),
	8453:    decimal.NewFromFloat(0.001),
	59144:   decimal.NewFromFloat(0.001),
	1440000: decimal.NewFromFloat(0.001),
}

// ============================================================================
// Main
// ============================================================================

func main() {
	ctx := context.Background()

	// --- Build signers for A and B ---
	signersA := buildSigners(privA, "A")
	signersB := buildSigners(privB, "B")
	fmt.Printf("Wallet A: %s\nWallet B: %s\n\n", signersA.address, signersB.address)

	// --- Build wallet clients with every RPC pre-registered ---
	walletA := newClient(signersA.channelSigner, signersA.rawSigner)
	defer walletA.Close()
	walletB := newClient(signersB.channelSigner, signersB.rawSigner)
	defer walletB.Close()

	// --- 1.b: discover tokenSet for the configured asset ---
	fmt.Println("=== 1.b: Discovering tokenSet ===")
	tokenSet := discoverTokenSet(ctx, walletA, asset)
	for i, t := range tokenSet {
		fmt.Printf("  [%d] %s on chain %d (%s)\n", i, t.Symbol, t.BlockchainID, t.Address)
	}
	fmt.Println()

	// --- 1.c + 1.d: preflight native and asset balances ---
	fmt.Println("=== 1.c + 1.d: Preflight ===")
	preflight(ctx, walletA, signersA.address, tokenSet)
	fmt.Println("✓ preflight passed")
	fmt.Println()

	// --- 1.e: optional session-key registration ---
	opsClient := walletA
	if sessionKeyPriv != "" {
		fmt.Println("=== 1.e: Registering channel session key ===")
		opsClient = setupSessionKeyClient(ctx, walletA, signersA)
		defer opsClient.Close()
		fmt.Println("✓ session-key client ready")
		fmt.Println()
	}

	// --- 2: round-robin loop ---
	fmt.Println("=== 2: Round robin ===")
	for i := range tokenSet {
		next := (i + 1) % len(tokenSet)
		runIteration(ctx, i, opsClient, walletB, signersA.address, signersB.address, tokenSet[i], tokenSet[next])
	}

	fmt.Println("\n=== Example Complete ===")
}

// ============================================================================
// Iteration
// ============================================================================

// runIteration executes one (deposit -> transfer -> close -> transfer-back ->
// withdraw) cycle. cur is the chain on which the deposit/close happen; next
// is the chain on which the withdraw settles.
func runIteration(
	ctx context.Context,
	i int,
	opsClient *sdk.Client,
	walletB *sdk.Client,
	addrA, addrB string,
	cur, next core.Token,
) {
	fmt.Printf("--- iter %d: deposit on chain %d, withdraw on chain %d ---\n", i, cur.BlockchainID, next.BlockchainID)

	// 2.a A deposits transferAmount on cur.BlockchainID. Wait for the approve
	// tx to be mined before issuing Deposit; ApproveToken returns immediately
	// after broadcast, and Deposit will revert if the allowance has not
	// settled on-chain yet.
	approveTx, err := opsClient.ApproveToken(ctx, cur.BlockchainID, asset, transferAmount)
	if err != nil {
		log.Fatalf("iter %d: approve on chain %d failed: %v", i, cur.BlockchainID, err)
	}
	waitForTxReceipt(ctx, chainRPCs[cur.BlockchainID], approveTx)
	depositState, err := opsClient.Deposit(ctx, cur.BlockchainID, asset, transferAmount)
	if err != nil {
		log.Fatalf("iter %d: deposit on chain %d failed: %v", i, cur.BlockchainID, err)
	}
	fmt.Printf("  ✓ A deposited %s %s on chain %d\n", transferAmount, asset, cur.BlockchainID)
	checkpointAndWait(ctx, opsClient, asset, depositState.Version)

	// 2.b A -> B (off-chain).
	if _, err := opsClient.Transfer(ctx, addrB, asset, transferAmount); err != nil {
		log.Fatalf("iter %d: transfer A->B failed: %v", i, err)
	}
	fmt.Printf("  ✓ A transferred %s %s to B (off-chain)\n", transferAmount, asset)

	// 2.c A closes home channel for asset on cur.BlockchainID.
	if _, err := opsClient.CloseHomeChannel(ctx, asset); err != nil {
		log.Fatalf("iter %d: close home channel failed: %v", i, err)
	}
	fmt.Printf("  ✓ A closed home channel for %s\n", asset)
	closeAndWait(ctx, opsClient, asset)

	// 2.d B -> A (off-chain credit, lands on void state, no chain attached).
	if _, err := walletB.Transfer(ctx, addrA, asset, transferAmount); err != nil {
		log.Fatalf("iter %d: transfer B->A failed: %v", i, err)
	}
	fmt.Printf("  ✓ B transferred %s %s back to A (off-chain credit)\n", transferAmount, asset)

	// 2.e A withdraws on next.BlockchainID. Withdraw auto-creates a new
	// channel on next.BlockchainID because A's latest state is void after
	// close + transfer-receive.
	withdrawState, err := opsClient.Withdraw(ctx, next.BlockchainID, asset, transferAmount)
	if err != nil {
		log.Fatalf("iter %d: withdraw on chain %d failed: %v", i, next.BlockchainID, err)
	}
	fmt.Printf("  ✓ A withdrew %s %s on chain %d\n", transferAmount, asset, next.BlockchainID)
	checkpointAndWait(ctx, opsClient, asset, withdrawState.Version)

	// Wait until next chain shows the funds on-chain so the next iteration's
	// deposit doesn't race ERC-20 settlement.
	waitForOnChain(ctx, opsClient, next.BlockchainID, asset, addrA, transferAmount)
	fmt.Println()
}

// ============================================================================
// Preflight
// ============================================================================

func preflight(ctx context.Context, walletA *sdk.Client, addrA string, tokenSet []core.Token) {
	var shortfalls []string

	// Verify chainRPCs / minNativeBalances cover every chain in tokenSet.
	for _, t := range tokenSet {
		if _, ok := chainRPCs[t.BlockchainID]; !ok {
			shortfalls = append(shortfalls, fmt.Sprintf("chainRPCs missing entry for chain %d", t.BlockchainID))
		}
		if _, ok := minNativeBalances[t.BlockchainID]; !ok {
			shortfalls = append(shortfalls, fmt.Sprintf("minNativeBalances missing entry for chain %d", t.BlockchainID))
		}
	}
	if len(shortfalls) > 0 {
		fmt.Println("Configuration shortfalls:")
		for _, s := range shortfalls {
			fmt.Printf("  - %s\n", s)
		}
		os.Exit(1)
	}

	// 1.c: native balance on every chain >= minNativeBalances.
	fmt.Println("Native balance check:")
	for _, t := range tokenSet {
		have, err := nativeBalance(ctx, chainRPCs[t.BlockchainID], addrA)
		if err != nil {
			log.Fatalf("native balance check for chain %d failed: %v", t.BlockchainID, err)
		}
		need := minNativeBalances[t.BlockchainID]
		marker := "✓"
		if have.LessThan(need) {
			marker = "✗"
			shortfalls = append(shortfalls, fmt.Sprintf("chain %d native: need >= %s, have %s", t.BlockchainID, need, have))
		}
		fmt.Printf("  %s chain %d: need >= %s, have %s\n", marker, t.BlockchainID, need, have)
	}

	// 1.d: privA holds >= transferAmount of asset on tokenSet[0].BlockchainID.
	seedChain := tokenSet[0].BlockchainID
	have, err := walletA.GetOnChainBalance(ctx, seedChain, asset, addrA)
	if err != nil {
		log.Fatalf("asset balance check on chain %d failed: %v", seedChain, err)
	}
	marker := "✓"
	if have.LessThan(transferAmount) {
		marker = "✗"
		shortfalls = append(shortfalls, fmt.Sprintf("chain %d %s: need >= %s, have %s", seedChain, asset, transferAmount, have))
	}
	fmt.Printf("Asset balance check:\n  %s chain %d %s: need >= %s, have %s\n", marker, seedChain, asset, transferAmount, have)

	if len(shortfalls) > 0 {
		fmt.Println("\nPreflight failed. Resolve the following before re-running:")
		for _, s := range shortfalls {
			fmt.Printf("  - %s\n", s)
		}
		os.Exit(1)
	}
}

// nativeBalance dials rpcURL and returns the native token balance of addr
// converted to a decimal with 18-decimal precision (standard for EVM native).
func nativeBalance(ctx context.Context, rpcURL, addr string) (decimal.Decimal, error) {
	cl, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		return decimal.Zero, fmt.Errorf("dial %s: %w", rpcURL, err)
	}
	defer cl.Close()
	wei, err := cl.BalanceAt(ctx, common.HexToAddress(addr), nil)
	if err != nil {
		return decimal.Zero, err
	}
	return decimal.NewFromBigInt(wei, 0).Shift(-18), nil
}

// approveConfirmations is the number of additional blocks the example waits
// for after an approve tx receipt before issuing the downstream deposit /
// checkpoint. Public RPC endpoints are often load-balanced across multiple
// nodes, and an eth_call read can hit a node that has not yet indexed the
// approve. Waiting a few blocks past the receipt gives the cluster time to
// converge on the post-tx state.
const approveConfirmations uint64 = 3

// waitForTxReceipt polls rpcURL until the given tx is mined and asserts a
// successful (status=1) receipt, then waits for approveConfirmations more
// blocks on top before returning. Fatals on revert or timeout. Needed because
// the SDK's ApproveToken returns immediately after broadcast, and the
// downstream Deposit / Checkpoint would otherwise race the allowance update
// on load-balanced public RPCs.
func waitForTxReceipt(ctx context.Context, rpcURL, txHash string) {
	cl, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		log.Fatalf("dial %s: %v", rpcURL, err)
	}
	defer cl.Close()

	hash := common.HexToHash(txHash)
	deadline := time.Now().Add(2 * time.Minute)

	var minedBlock uint64
	for {
		receipt, err := cl.TransactionReceipt(ctx, hash)
		if err == nil {
			if receipt.Status != 1 {
				log.Fatalf("tx %s reverted on-chain", txHash)
			}
			minedBlock = receipt.BlockNumber.Uint64()
			break
		}
		if !errors.Is(err, ethereum.NotFound) {
			log.Fatalf("TransactionReceipt %s: %v", txHash, err)
		}
		if time.Now().After(deadline) {
			log.Fatalf("timed out waiting for tx %s", txHash)
		}
		time.Sleep(2 * time.Second)
	}

	target := minedBlock + approveConfirmations
	for {
		head, err := cl.BlockNumber(ctx)
		if err != nil {
			log.Fatalf("BlockNumber on %s: %v", rpcURL, err)
		}
		if head >= target {
			return
		}
		if time.Now().After(deadline) {
			log.Fatalf("timed out waiting for %d confirmations on tx %s (head=%d, target=%d)", approveConfirmations, txHash, head, target)
		}
		time.Sleep(2 * time.Second)
	}
}

// ============================================================================
// Token discovery
// ============================================================================

func discoverTokenSet(ctx context.Context, client *sdk.Client, symbol string) []core.Token {
	assets, err := client.GetAssets(ctx, nil)
	if err != nil {
		log.Fatalf("GetAssets failed: %v", err)
	}
	for _, a := range assets {
		if strings.EqualFold(a.Symbol, symbol) {
			if len(a.Tokens) == 0 {
				log.Fatalf("asset %s has no supported tokens", symbol)
			}
			return a.Tokens
		}
	}
	log.Fatalf("asset %s not supported by node", symbol)
	return nil
}

// ============================================================================
// Client construction
// ============================================================================

type signers struct {
	address       string
	rawSigner     sign.Signer
	msgSigner     *sign.EthereumMsgSigner
	channelSigner core.ChannelSigner
}

func buildSigners(privateKeyHex, label string) signers {
	raw, err := sign.NewEthereumRawSigner(privateKeyHex)
	if err != nil {
		log.Fatalf("invalid %s private key: %v", label, err)
	}
	msg, err := sign.NewEthereumMsgSignerFromRaw(raw)
	if err != nil {
		log.Fatalf("failed to build %s msg signer: %v", label, err)
	}
	ch, err := core.NewChannelDefaultSigner(msg)
	if err != nil {
		log.Fatalf("failed to build %s channel signer: %v", label, err)
	}
	return signers{
		address:       raw.PublicKey().Address().String(),
		rawSigner:     raw,
		msgSigner:     msg,
		channelSigner: ch,
	}
}

// newClient builds an SDK client with every configured chain RPC attached.
func newClient(channelSigner core.ChannelSigner, rawSigner sign.Signer) *sdk.Client {
	opts := make([]sdk.Option, 0, len(chainRPCs))
	for chainID, url := range chainRPCs {
		opts = append(opts, sdk.WithBlockchainRPC(chainID, url))
	}
	client, err := sdk.NewClient(wsURL, channelSigner, rawSigner, opts...)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	return client
}

// ============================================================================
// Session key
// ============================================================================

func setupSessionKeyClient(ctx context.Context, walletA *sdk.Client, sa signers) *sdk.Client {
	skRaw, skMsg := generateSessionKey()
	skAddress := skRaw.PublicKey().Address().String()
	fmt.Printf("Session key: %s\n", skAddress)

	state := core.ChannelSessionKeyStateV1{
		UserAddress: sa.address,
		SessionKey:  skAddress,
		Version:     1,
		Assets:      []string{asset},
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	userSig, err := walletA.SignChannelSessionKeyState(state)
	if err != nil {
		log.Fatalf("sign session key state: %v", err)
	}
	state.UserSig = userSig

	skSig, err := sdk.SignChannelSessionKeyOwnership(state, skMsg)
	if err != nil {
		log.Fatalf("sign session key ownership: %v", err)
	}
	state.SessionKeySig = skSig

	if err := walletA.SubmitChannelSessionKeyState(ctx, state); err != nil {
		log.Fatalf("submit session key state: %v", err)
	}

	metadataHash, err := core.GetChannelSessionKeyAuthMetadataHashV1(state.UserAddress, state.Version, state.Assets, state.ExpiresAt.Unix())
	if err != nil {
		log.Fatalf("compute metadata hash: %v", err)
	}
	skChannelSigner, err := core.NewChannelSessionKeySignerV1(skMsg, metadataHash.Hex(), state.UserSig)
	if err != nil {
		log.Fatalf("build session-key channel signer: %v", err)
	}

	// rawSigner stays as the wallet's key: the SDK uses it to derive the
	// user address and to sign on-chain checkpoint transactions.
	opts := make([]sdk.Option, 0, len(chainRPCs))
	for chainID, url := range chainRPCs {
		opts = append(opts, sdk.WithBlockchainRPC(chainID, url))
	}
	client, err := sdk.NewClient(wsURL, skChannelSigner, sa.rawSigner, opts...)
	if err != nil {
		log.Fatalf("build session-key client: %v", err)
	}
	return client
}

func generateSessionKey() (sign.Signer, *sign.EthereumMsgSigner) {
	priv, err := crypto.GenerateKey()
	if err != nil {
		log.Fatalf("generate session key: %v", err)
	}
	privHex := hexutil.Encode(crypto.FromECDSA(priv))
	raw, err := sign.NewEthereumRawSigner(privHex)
	if err != nil {
		log.Fatalf("session-key raw signer: %v", err)
	}
	msg, err := sign.NewEthereumMsgSigner(privHex)
	if err != nil {
		log.Fatalf("session-key msg signer: %v", err)
	}
	return raw, msg
}

// ============================================================================
// Wait helpers
// ============================================================================

// closeAndWait runs Checkpoint after a Finalize transition and polls
// GetHomeChannel until the node observes the on-chain close. Closure is
// signalled either by the home-channel row dropping out (nil) or by its
// status being reset to Void.
func closeAndWait(ctx context.Context, client *sdk.Client, asset string) {
	txHash, err := client.Checkpoint(ctx, asset)
	if err != nil {
		log.Fatalf("checkpoint %s (close) failed: %v", asset, err)
	}
	fmt.Printf("    ↳ checkpoint %s tx %s; waiting for channel close (nil or status=Void)...\n", asset, txHash)

	wallet := client.GetUserAddress()
	deadline := time.Now().Add(2 * time.Minute)
	for {
		channel, err := client.GetHomeChannel(ctx, wallet, asset)
		if err != nil {
			log.Fatalf("GetHomeChannel %s: %v", asset, err)
		}
		if channel == nil || channel.Status == core.ChannelStatusVoid {
			return
		}
		if time.Now().After(deadline) {
			log.Fatalf("timed out waiting for %s channel to close (last status=%s)", asset, channel.Status)
		}
		time.Sleep(2 * time.Second)
	}
}

// checkpointAndWait runs Checkpoint and polls GetHomeChannel until the node
// observes the expected post-checkpoint state.
//
// When expectedVersion > 0 the helper waits for channel.StateVersion to catch
// up to expectedVersion. When expectedVersion == 0 — which happens for the
// channel-creation transitions issued by Deposit / Withdraw on a void state —
// the state_version stays at 0 even after the checkpoint, so the helper
// instead waits for channel.Status == Open.
func checkpointAndWait(ctx context.Context, client *sdk.Client, asset string, expectedVersion uint64) {
	txHash, err := client.Checkpoint(ctx, asset)
	if err != nil {
		log.Fatalf("checkpoint %s failed: %v", asset, err)
	}
	if expectedVersion == 0 {
		fmt.Printf("    ↳ checkpoint %s tx %s; waiting for channel status=Open...\n", asset, txHash)
	} else {
		fmt.Printf("    ↳ checkpoint %s tx %s; waiting for state_version=%d...\n", asset, txHash, expectedVersion)
	}

	wallet := client.GetUserAddress()
	deadline := time.Now().Add(2 * time.Minute)
	for {
		channel, err := client.GetHomeChannel(ctx, wallet, asset)
		if err != nil {
			log.Fatalf("GetHomeChannel %s: %v", asset, err)
		}
		if channel != nil {
			if expectedVersion == 0 && channel.Status == core.ChannelStatusOpen {
				return
			}
			if expectedVersion > 0 && channel.StateVersion >= expectedVersion {
				return
			}
		}
		if time.Now().After(deadline) {
			if expectedVersion == 0 {
				log.Fatalf("timed out waiting for %s channel to reach status=Open", asset)
			}
			log.Fatalf("timed out waiting for %s to reach state_version=%d", asset, expectedVersion)
		}
		time.Sleep(2 * time.Second)
	}
}

// waitForOnChain polls until addr's ERC-20 balance of asset on chainID is at
// least minAmount. Needed between iterations because Withdraw's on-chain
// settlement can lag the node's state_version bump.
func waitForOnChain(ctx context.Context, client *sdk.Client, chainID uint64, asset, addr string, minAmount decimal.Decimal) {
	deadline := time.Now().Add(2 * time.Minute)
	for {
		have, err := client.GetOnChainBalance(ctx, chainID, asset, addr)
		if err != nil {
			log.Fatalf("GetOnChainBalance chain %d: %v", chainID, err)
		}
		if have.GreaterThanOrEqual(minAmount) {
			fmt.Printf("    ↳ on-chain balance on chain %d settled: %s %s\n", chainID, have, asset)
			return
		}
		if time.Now().After(deadline) {
			log.Fatalf("timed out waiting for chain %d %s balance >= %s (have %s)", chainID, asset, minAmount, have)
		}
		time.Sleep(2 * time.Second)
	}
}
