package main

// Example: Complete App Session Lifecycle
//
// This example demonstrates:
// 1. Create first app session for wallet 1
// 2. Deposit USDC into first app session by wallet 1
// 3. Create second app session for wallet 2 with wallet 3 as a participant
// 4. Deposit WETH into second app session by wallet 2
// 5. Redistribute app state within app session so that participant with wallet 3 also has some allocation
// 6. Rebalance 2 app sessions atomically
// 7. Wallet 3 withdraws from his app session
// 8. Close both app sessions

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/shopspring/decimal"

	"github.com/erc7824/nitrolite/pkg/app"
	"github.com/erc7824/nitrolite/pkg/core"
	"github.com/erc7824/nitrolite/pkg/sign"
	sdk "github.com/erc7824/nitrolite/sdk/go"
)

func main() {
	ctx := context.Background()
	wsURL := "wss://deployment.yellow.org/ws"

	// --- 0. Setup Wallets ---
	// Replace these strings with your actual hex private keys
	wallet1PrivateKey := "0x7d60..."
	wallet2PrivateKey := "0x9b65..."
	wallet3PrivateKey := "0xf636..."

	// Create signers from private keys
	wallet1Signer, err := sign.NewEthereumMsgSigner(wallet1PrivateKey)
	if err != nil {
		log.Fatalf("Invalid wallet 1 private key: %v", err)
	}
	wallet2Signer, err := sign.NewEthereumMsgSigner(wallet2PrivateKey)
	if err != nil {
		log.Fatalf("Invalid wallet 2 private key: %v", err)
	}
	wallet3Signer, err := sign.NewEthereumMsgSigner(wallet3PrivateKey)
	if err != nil {
		log.Fatalf("Invalid wallet 3 private key: %v", err)
	}

	channel1Signer, err := core.NewChannelDefaultSigner(wallet1Signer)
	if err != nil {
		log.Fatalf("Failed to create channel signer for wallet 1: %v", err)
	}
	channel2Signer, err := core.NewChannelDefaultSigner(wallet2Signer)
	if err != nil {
		log.Fatalf("Failed to create channel signer for wallet 2: %v", err)
	}
	channel3Signer, err := core.NewChannelDefaultSigner(wallet3Signer)
	if err != nil {
		log.Fatalf("Failed to create channel signer for wallet 2: %v", err)
	}

	appSession1Signer, err := app.NewAppSessionWalletSignerV1(wallet1Signer)
	if err != nil {
		log.Fatalf("Failed to create app session signer for wallet 1: %v", err)
	}
	appSession2Signer, err := app.NewAppSessionWalletSignerV1(wallet2Signer)
	if err != nil {
		log.Fatalf("Failed to create app session signer for wallet 2: %v", err)
	}

	// Extract wallet addresses
	wallet1Address := wallet1Signer.PublicKey().Address().String() // 0x053aEAD7d3eebE4359300fDE849bCD9E77384989
	wallet2Address := wallet2Signer.PublicKey().Address().String() // 0x2BfA10aAd64Ae0F7855f54f27117Fcc9C61C6770
	wallet3Address := wallet3Signer.PublicKey().Address().String() // 0xaB5670b44cb4A3B5535BD637cb600DA572148c98

	fmt.Println("--- Wallets Imported ---")
	fmt.Printf("Wallet 1 Address: %s\n", wallet1Address)
	fmt.Printf("Wallet 2 Address: %s\n", wallet2Address)
	fmt.Printf("Wallet 3 Address: %s\n", wallet3Address)
	fmt.Println("------------------------")

	// Create SDK clients (in a real app, these would be separate instances)
	wallet1Client, err := sdk.NewClient(wsURL, channel1Signer, wallet1Signer)
	if err != nil {
		log.Fatal(err)
	}

	// --- 1. Create App Session 1 (Single Participant: Wallet 1) ---
	fmt.Println("=== Step 1: Creating App Session 1 (Wallet 1 only) ===")

	session1Definition := app.AppDefinitionV1{
		ApplicationID: "test-app",
		Participants: []app.AppParticipantV1{
			{WalletAddress: wallet1Address, SignatureWeight: 100},
		},
		Quorum: 100,
		Nonce:  uint64(time.Now().UnixNano()),
	}

	session1CreateRequest, err := app.PackCreateAppSessionRequestV1(session1Definition, "{}")
	if err != nil {
		log.Fatal(err)
	}

	appSession1CreateSession1Sig, _ := appSession1Signer.Sign(session1CreateRequest)
	session1ID, _, _, err := wallet1Client.CreateAppSession(ctx, session1Definition, "{}", []string{appSession1CreateSession1Sig.String()})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Created App Session 1: %s\n\n", session1ID)

	// --- 2. Deposit USDC into Session 1 ---
	fmt.Println("=== Step 2: Depositing USDC into Session 1 ===")

	session1DepositAmount := decimal.NewFromFloat(0.0001)
	session1DepositUpdate := app.AppStateUpdateV1{
		AppSessionID: session1ID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      2,
		Allocations:  []app.AppAllocationV1{{Participant: wallet1Address, Asset: "usdc", Amount: session1DepositAmount}},
	}

	session1DepositRequest, err := app.PackAppStateUpdateV1(session1DepositUpdate)
	if err != nil {
		log.Fatal(err)
	}

	appSession1DepositSig, _ := appSession1Signer.Sign(session1DepositRequest)

	_, err = wallet1Client.SubmitAppSessionDeposit(ctx, session1DepositUpdate, []string{appSession1DepositSig.String()}, "usdc", session1DepositAmount)
	if err != nil {
		log.Printf("⚠ Deposit warning: %v", err)
	}
	fmt.Printf("✓ Deposited %s USDC into Session 1\n\n", session1DepositAmount)

	// --- 3. Create App Session 2 (Multi-Party: Wallet 2 & 3) ---
	fmt.Println("=== Step 3: Creating App Session 2 (Wallet 2 & 3) ===")

	appID := "multi-party-app"
	wallet2Client, err := sdk.NewClient(wsURL, channel2Signer, wallet2Signer)
	if err != nil {
		log.Fatal(err)
	}

	wallet3Client, err := sdk.NewClient(wsURL, channel3Signer, wallet3Signer) // No channel signer needed for wallet 3 since it's not creating channels
	if err != nil {
		log.Fatal(err)
	}

	msgSigner3, err := generateMsgSigner()
	if err != nil {
		log.Fatal(err)
	}

	appSessionKey3State := app.AppSessionKeyStateV1{
		SessionKey:     msgSigner3.PublicKey().Address().String(),
		UserAddress:    wallet3Address,
		Version:        1,
		ApplicationIDs: []string{appID},
		ExpiresAt:      time.Now().Add(10 * time.Minute),
	}
	packedAppSessionKey3State, err := app.PackAppSessionKeyStateV1(appSessionKey3State)
	if err != nil {
		log.Fatal(err)
	}

	appSessionKey3StateSig, err := wallet3Signer.Sign(packedAppSessionKey3State)
	if err != nil {
		log.Fatal(err)
	}
	appSessionKey3State.UserSig = appSessionKey3StateSig.String()

	if err := wallet3Client.SubmitAppSessionKeyState(context.Background(), appSessionKey3State); err != nil {
		log.Fatal(err)
	}

	appSession3Signer, err := app.NewAppSessionKeySignerV1(msgSigner3)
	if err != nil {
		log.Fatalf("Failed to create app session signer for wallet 3: %v", err)
	}

	session2Definition := app.AppDefinitionV1{
		ApplicationID: appID,
		Participants: []app.AppParticipantV1{
			{WalletAddress: wallet2Address, SignatureWeight: 50},
			{WalletAddress: wallet3Address, SignatureWeight: 50},
		},
		Quorum: 100,
		Nonce:  uint64(time.Now().UnixNano()),
	}

	session2CreateRequest, err := app.PackCreateAppSessionRequestV1(session2Definition, "{}")
	if err != nil {
		log.Fatal(err)
	}

	appSession2CreateSession2Sig, _ := appSession2Signer.Sign(session2CreateRequest)
	appSession3CreateSession2Sig, _ := appSession3Signer.Sign(session2CreateRequest)
	session2ID, _, _, err := wallet2Client.CreateAppSession(ctx, session2Definition, "{}", []string{appSession2CreateSession2Sig.String(), appSession3CreateSession2Sig.String()})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Created App Session 2: %s\n\n", session2ID)

	// --- 4. Deposit WETH into Session 2 by Wallet 2 ---
	fmt.Println("=== Step 4: Depositing WETH into Session 2 ===")

	session2DepositAmount := decimal.NewFromFloat(0.015)
	session2DepositUpdate := app.AppStateUpdateV1{
		AppSessionID: session2ID,
		Intent:       app.AppStateUpdateIntentDeposit,
		Version:      2,
		Allocations:  []app.AppAllocationV1{{Participant: wallet2Address, Asset: "weth", Amount: session2DepositAmount}},
	}

	session2DepositRequest, err := app.PackAppStateUpdateV1(session2DepositUpdate)
	if err != nil {
		log.Fatal(err)
	}

	appSession2DepositSig, _ := appSession2Signer.Sign(session2DepositRequest)
	appSession3DepositSig, _ := appSession3Signer.Sign(session2DepositRequest)

	nodeSig, err := wallet2Client.SubmitAppSessionDeposit(ctx, session2DepositUpdate, []string{appSession2DepositSig.String(), appSession3DepositSig.String()}, "weth", session2DepositAmount)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("✓ Deposited %s WETH into Session 2 (Node Sig: %s)\n\n", session2DepositAmount, nodeSig)

	// Check Session 2 state before redistribution
	session2InfoBeforeRedist, _, err := wallet2Client.GetAppSessions(ctx, &sdk.GetAppSessionsOptions{AppSessionID: &session2ID})
	if err != nil {
		log.Fatal(err)
	}
	if len(session2InfoBeforeRedist) > 0 {
		fmt.Printf("Session 2 before redistribution - Version: %d, Allocations: %+v\n\n", session2InfoBeforeRedist[0].Version, session2InfoBeforeRedist[0].Allocations)
	}

	// --- 5. Redistribute within Session 2 (Wallet 2 -> Wallet 3) ---
	fmt.Println("=== Step 5: Redistributing funds in Session 2 ===")

	session2RedistributeUpdate := app.AppStateUpdateV1{
		AppSessionID: session2ID,
		Intent:       app.AppStateUpdateIntentOperate,
		Version:      3,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet2Address, Asset: "weth", Amount: decimal.NewFromFloat(0.01)},
			{Participant: wallet3Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
		},
	}

	session2RedistributeRequest, err := app.PackAppStateUpdateV1(session2RedistributeUpdate)
	if err != nil {
		log.Fatal(err)
	}

	appSession2RedistributeSig, _ := appSession2Signer.Sign(session2RedistributeRequest)
	appSession3RedistributeSig, _ := appSession3Signer.Sign(session2RedistributeRequest)

	// Multi-sig required for state transition
	err = wallet2Client.SubmitAppState(ctx, session2RedistributeUpdate, []string{appSession2RedistributeSig.String(), appSession3RedistributeSig.String()})
	if err != nil {
		log.Fatalf("Redistribution failed: %v", err)
	}
	fmt.Println("✓ Redistributed WETH: Wallet 2 (0.01) -> Wallet 3 (0.005)")

	// --- 6. Rebalance Both App Sessions Atomically ---
	fmt.Println("=== Step 6: Atomic Rebalance Across Sessions ===")

	// Check current allocations before rebalance
	session1InfoBeforeRebalance, _, err := wallet1Client.GetAppSessions(ctx, &sdk.GetAppSessionsOptions{AppSessionID: &session1ID})
	if err != nil {
		log.Fatal(err)
	}
	if len(session1InfoBeforeRebalance) > 0 {
		fmt.Printf("Session 1 before rebalance - Version: %d, Allocations: %+v\n", session1InfoBeforeRebalance[0].Version, session1InfoBeforeRebalance[0].Allocations)
	}

	session2InfoBeforeRebalance, _, err := wallet2Client.GetAppSessions(ctx, &sdk.GetAppSessionsOptions{AppSessionID: &session2ID})
	if err != nil {
		log.Fatal(err)
	}
	if len(session2InfoBeforeRebalance) > 0 {
		fmt.Printf("Session 2 before rebalance - Version: %d, Allocations: %+v\n\n", session2InfoBeforeRebalance[0].Version, session2InfoBeforeRebalance[0].Allocations)
	}

	// Prepare rebalance updates for both sessions
	session1RebalanceUpdate := app.AppStateUpdateV1{
		AppSessionID: session1ID,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      3,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet1Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
			{Participant: wallet1Address, Asset: "usdc", Amount: decimal.NewFromFloat(0.00005)},
		},
	}

	session1RebalanceRequest, err := app.PackAppStateUpdateV1(session1RebalanceUpdate)
	if err != nil {
		log.Fatal(err)
	}

	appSession1RebalanceSig, _ := appSession1Signer.Sign(session1RebalanceRequest)

	session2RebalanceUpdate := app.AppStateUpdateV1{
		AppSessionID: session2ID,
		Intent:       app.AppStateUpdateIntentRebalance,
		Version:      4,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet2Address, Asset: "usdc", Amount: decimal.NewFromFloat(0.00005)},
			{Participant: wallet2Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
			{Participant: wallet3Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
		},
	}

	session2RebalanceRequest, err := app.PackAppStateUpdateV1(session2RebalanceUpdate)
	if err != nil {
		log.Fatal(err)
	}

	appSession2RebalanceSig, _ := appSession2Signer.Sign(session2RebalanceRequest)
	appSession3RebalanceSig, _ := appSession3Signer.Sign(session2RebalanceRequest)

	// Submit atomic rebalance
	signedRebalanceUpdates := []app.SignedAppStateUpdateV1{
		{
			AppStateUpdate: session1RebalanceUpdate,
			QuorumSigs:     []string{appSession1RebalanceSig.String()},
		},
		{
			AppStateUpdate: session2RebalanceUpdate,
			QuorumSigs:     []string{appSession2RebalanceSig.String(), appSession3RebalanceSig.String()},
		},
	}

	rebalanceBatchID, err := wallet2Client.RebalanceAppSessions(ctx, signedRebalanceUpdates)
	if err != nil {
		log.Printf("⚠ Rebalance Error: %v", err)
	} else {
		fmt.Printf("✓ Atomic Rebalance Submitted. BatchID: %s\n\n", rebalanceBatchID)
	}

	// --- 7. Wallet 3 Withdraws from Session 2 ---
	fmt.Println("=== Step 7: Wallet 3 Withdrawing from Session 2 ===")

	session2WithdrawUpdate := app.AppStateUpdateV1{
		AppSessionID: session2ID,
		Intent:       app.AppStateUpdateIntentWithdraw,
		Version:      5,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet2Address, Asset: "usdc", Amount: decimal.NewFromFloat(0.00005)},
			{Participant: wallet2Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
			{Participant: wallet3Address, Asset: "weth", Amount: decimal.NewFromFloat(0.001)},
		},
	}

	session2WithdrawRequest, err := app.PackAppStateUpdateV1(session2WithdrawUpdate)
	if err != nil {
		log.Fatal(err)
	}

	appSession2WithdrawSig, _ := appSession2Signer.Sign(session2WithdrawRequest)
	appSession3WithdrawSig, _ := appSession3Signer.Sign(session2WithdrawRequest)

	err = wallet2Client.SubmitAppState(ctx, session2WithdrawUpdate, []string{appSession2WithdrawSig.String(), appSession3WithdrawSig.String()})
	if err != nil {
		log.Printf("⚠ Withdraw Error: %v", err)
	} else {
		fmt.Println("✓ Wallet 3 successfully withdrew 0.004 WETH back to channel")
	}

	// --- 8. Close Both App Sessions ---
	fmt.Println("=== Step 8: Closing Both App Sessions ===")

	// Close Session 1
	session1CloseUpdate := app.AppStateUpdateV1{
		AppSessionID: session1ID,
		Intent:       app.AppStateUpdateIntentClose,
		Version:      4,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet1Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
			{Participant: wallet1Address, Asset: "usdc", Amount: decimal.NewFromFloat(0.00005)},
		},
	}

	session1CloseRequest, err := app.PackAppStateUpdateV1(session1CloseUpdate)
	if err != nil {
		log.Fatal(err)
	}

	appSession1CloseSig, _ := appSession1Signer.Sign(session1CloseRequest)

	err = wallet1Client.SubmitAppState(ctx, session1CloseUpdate, []string{appSession1CloseSig.String()})
	if err != nil {
		log.Printf("⚠ Close Session 1 Error: %v", err)
	} else {
		fmt.Println("✓ Session 1 successfully closed")
	}

	// Close Session 2
	session2CloseUpdate := app.AppStateUpdateV1{
		AppSessionID: session2ID,
		Intent:       app.AppStateUpdateIntentClose,
		Version:      6,
		Allocations: []app.AppAllocationV1{
			{Participant: wallet2Address, Asset: "usdc", Amount: decimal.NewFromFloat(0.00005)},
			{Participant: wallet2Address, Asset: "weth", Amount: decimal.NewFromFloat(0.005)},
			{Participant: wallet3Address, Asset: "weth", Amount: decimal.NewFromFloat(0.001)},
		},
	}

	session2CloseRequest, err := app.PackAppStateUpdateV1(session2CloseUpdate)
	if err != nil {
		log.Fatal(err)
	}

	appSession2CloseSig, _ := appSession2Signer.Sign(session2CloseRequest)
	appSession3CloseSig, _ := appSession3Signer.Sign(session2CloseRequest)

	err = wallet2Client.SubmitAppState(ctx, session2CloseUpdate, []string{appSession2CloseSig.String(), appSession3CloseSig.String()})
	if err != nil {
		log.Printf("⚠ Close Session 2 Error: %v", err)
	} else {
		fmt.Println("✓ Session 2 successfully closed")
	}

	fmt.Println("\n=== Example Complete ===")
}

func generateMsgSigner() (sign.Signer, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}

	// Convert private key to bytes
	privateKeyBytes := crypto.FromECDSA(privateKey)

	return sign.NewEthereumMsgSigner(hexutil.Encode(privateKeyBytes))
}
