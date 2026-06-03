/**
 * Example: Complete App Session Lifecycle
 *
 * Requirements to run this example:
 *
 *  1. A reachable nitronode WebSocket endpoint (set via wsURL below).
 *     The default points at the public sandbox.
 *
 *  2. Three EVM wallets with hex private keys (replace the placeholders below).
 *     Wallet 3 may be a fresh key — it only receives funds via redistribution.
 *
 *  3. Minimum off-chain (channel) balances on the node:
 *       - Wallet 1: 0.0001 YUSD     (deposited into Session 1)
 *       - Wallet 2: 0.00015 YELLOW  (deposited into Session 2)
 *       - Wallet 3: none required   (receives funds via redistribution)
 *
 *     An open channel is NOT a hard prerequisite. If a wallet already has
 *     funds on the node but no acknowledged channel for the asset yet, the
 *     example calls acknowledge() first to open one. Wallet 3 also needs no
 *     pre-existing channel; the withdraw step will open/credit its ledger
 *     automatically.
 *
 * This example demonstrates:
 *  1. Create first app session for wallet 1
 *  2. Deposit YUSD into first app session by wallet 1
 *     (auto-opens wallet 1's YUSD channel via acknowledge() if missing)
 *  3. Create second app session for wallet 2 with wallet 3 as a participant
 *  4. Deposit YELLOW into second app session by wallet 2
 *     (auto-opens wallet 2's YELLOW channel via acknowledge() if missing)
 *  5. Redistribute app state within app session so that participant with wallet 3 also has some allocation
 *  6. Wallet 3 withdraws from his app session
 *  7. Close both app sessions
 */

// Node <22 does not expose a stable global WebSocket. The SDK dialer uses
// `new WebSocket(url)` directly, so we polyfill the global from the `ws`
// package before any client is constructed.
import WebSocket from 'ws';
if (typeof (globalThis as { WebSocket?: unknown }).WebSocket === 'undefined') {
  (globalThis as { WebSocket: unknown }).WebSocket = WebSocket;
}

import Decimal from 'decimal.js';
import { Hex } from 'viem';
import { generatePrivateKey, privateKeyToAccount } from 'viem/accounts';
import { Client } from '../../src/client';
import {
  createSigners,
  EthereumMsgSigner,
  AppSessionWalletSignerV1,
  AppSessionKeySignerV1,
} from '../../src/signers';
import {
  AppDefinitionV1,
  AppStateUpdateV1,
  AppStateUpdateIntent,
  AppSessionKeyStateV1,
} from '../../src/app/types';
import { packCreateAppSessionRequestV1, packAppStateUpdateV1, packAppSessionKeyStateV1 } from '../../src/app/packing';
import { isFinal } from '../../src/core/state';

/**
 * ensureChannelOpen guarantees that the given wallet has an acknowledged
 * channel open for asset. If the node holds no state for the wallet/asset
 * pair, or the latest state is still awaiting the user's signature (or has
 * been finalized), acknowledge() is invoked to create or progress the channel.
 * Already-acknowledged channels are left untouched.
 */
async function ensureChannelOpen(label: string, client: Client, asset: string): Promise<void> {
    const wallet = client.getUserAddress();
    const state = await client.getLatestState(wallet, asset, false);

    const hasOpenChannel =
        state !== null &&
        state.homeChannelId !== undefined &&
        !isFinal(state) &&
        !!state.userSig;
    if (hasOpenChannel) {
        console.log(`✓ ${label} already has an open ${asset} channel`);
        return;
    }

    await client.acknowledge(asset);
    console.log(`✓ ${label} acknowledged ${asset} channel`);
}

async function main() {
  // Replace with a real deployment url
  const wsURL = 'wss://nitronode-sandbox.yellow.org/v1/ws';

  // --- 0. Setup Wallets ---
  // Replace these strings with your actual hex private keys
  const wallet1PrivateKey = '0x7d607...' as Hex;
  const wallet2PrivateKey = '0x9b652...' as Hex;
  const wallet3PrivateKey = '0xf6369...' as Hex;

  // Create signers from private keys
  const wallet1Signers = createSigners(wallet1PrivateKey);
  const wallet2Signers = createSigners(wallet2PrivateKey);
  const wallet3Signers = createSigners(wallet3PrivateKey);

  // Create app session wallet signers (prepend 0x00 type byte)
  const wallet1MsgSigner = new EthereumMsgSigner(wallet1PrivateKey);
  const wallet2MsgSigner = new EthereumMsgSigner(wallet2PrivateKey);
  const appSession1Signer = new AppSessionWalletSignerV1(wallet1MsgSigner);
  const appSession2Signer = new AppSessionWalletSignerV1(wallet2MsgSigner);

  // Extract wallet addresses
  const wallet1Address = wallet1Signers.stateSigner.getAddress();
  const wallet2Address = wallet2Signers.stateSigner.getAddress();
  const wallet3Address = wallet3Signers.stateSigner.getAddress();

  console.log('--- Wallets Imported ---');
  console.log(`Wallet 1 Address: ${wallet1Address}`);
  console.log(`Wallet 2 Address: ${wallet2Address}`);
  console.log(`Wallet 3 Address: ${wallet3Address}`);
  console.log('------------------------');

  // Create SDK clients (in a real app, these would be separate instances)
  const wallet1Client = await Client.create(
    wsURL,
    wallet1Signers.stateSigner,
    wallet1Signers.txSigner
  );
  const wallet2Client = await Client.create(
    wsURL,
    wallet2Signers.stateSigner,
    wallet2Signers.txSigner
  );
  const wallet3Client = await Client.create(
    wsURL,
    wallet3Signers.stateSigner,
    wallet3Signers.txSigner
  );

  // --- Ensure Required Channels Are Open ---
  // App session deposits require an acknowledged channel for the asset.
  // If the wallet has funds on the node but no channel yet, acknowledge()
  // opens it on the fly so the example only assumes a minimum balance.
  console.log('=== Ensuring Channels Are Open ===');
  await ensureChannelOpen('Wallet 1', wallet1Client, 'yusd');
  await ensureChannelOpen('Wallet 2', wallet2Client, 'yellow');
  console.log();

  // App session IDs — app sessions are created against arbitrary application
  // IDs; no prior registration step is required.
  const suffix = String(Math.floor(Math.random() * 1000000)).padStart(6, '0');
  const app1ID = `test-app-${suffix}`;
  const app2ID = `multi-party-app-${suffix}`;

  // --- 1. Create App Session 1 (Single Participant: Wallet 1) ---
  console.log('=== Step 1: Creating App Session 1 (Wallet 1 only) ===');

  const session1Definition: AppDefinitionV1 = {
    applicationId: app1ID,
    participants: [{ walletAddress: wallet1Address, signatureWeight: 100 }],
    quorum: 100,
    nonce: BigInt(Date.now() * 1000000),
  };

  const session1CreateRequest = packCreateAppSessionRequestV1(session1Definition, '{}');
  const appSession1CreateSig = await appSession1Signer.signMessage(session1CreateRequest);

  const { appSessionId: session1ID } = await wallet1Client.createAppSession(
    session1Definition,
    '{}',
    [appSession1CreateSig]
  );
  console.log(`✓ Created App Session 1: ${session1ID}\n`);

  // --- 3. Deposit YUSD into Session 1 ---
  console.log('=== Step 2: Depositing YUSD into Session 1 ===');

  const session1DepositAmount = new Decimal(0.0001);
  const session1DepositUpdate: AppStateUpdateV1 = {
    appSessionId: session1ID,
    intent: AppStateUpdateIntent.Deposit,
    version: 2n,
    allocations: [
      { participant: wallet1Address, asset: 'yusd', amount: session1DepositAmount },
    ],
    sessionData: '{}',
  };

  const session1DepositRequest = packAppStateUpdateV1(session1DepositUpdate);
  const appSession1DepositSig = await appSession1Signer.signMessage(session1DepositRequest);

  try {
    await wallet1Client.submitAppSessionDeposit(
      session1DepositUpdate,
      [appSession1DepositSig],
      'yusd',
      session1DepositAmount
    );
    console.log(`✓ Deposited ${session1DepositAmount} YUSD into Session 1\n`);
  } catch (err) {
    console.log(`⚠ Deposit warning: ${err}`);
  }

  // --- 3. Create App Session 2 (Multi-Party: Wallet 2 & 3) ---
  console.log('=== Step 3: Creating App Session 2 (Wallet 2 & 3) ===');

  const appID = app2ID;

  // Generate session key for wallet 3
  const sessionKey3PrivateKey = generatePrivateKey();
  const sessionKey3Account = privateKeyToAccount(sessionKey3PrivateKey);
  const sessionKey3MsgSigner = new EthereumMsgSigner(sessionKey3PrivateKey);

  const expiresAt = Math.floor(Date.now() / 1000) + 10 * 60; // 10 minutes from now

  const appSessionKey3State: AppSessionKeyStateV1 = {
    session_key: sessionKey3Account.address,
    user_address: wallet3Address,
    version: '1',
    application_ids: [appID],
    app_session_ids: [],
    expires_at: String(expiresAt),
    user_sig: '',
    session_key_sig: '',
  };

  const packedAppSessionKey3State = packAppSessionKeyStateV1(appSessionKey3State);

  // Wallet's user_sig authorizes the delegation.
  const wallet3MsgSigner = new EthereumMsgSigner(wallet3PrivateKey);
  appSessionKey3State.user_sig = await wallet3MsgSigner.signMessage(packedAppSessionKey3State);

  // Session-key holder's session_key_sig proves possession of the key being registered.
  // Both signatures are required at submit time.
  appSessionKey3State.session_key_sig = await sessionKey3MsgSigner.signMessage(packedAppSessionKey3State);

  await wallet3Client.submitSessionKeyState(appSessionKey3State);

  const appSession3Signer = new AppSessionKeySignerV1(sessionKey3MsgSigner);

  const session2Definition: AppDefinitionV1 = {
    applicationId: appID,
    participants: [
      { walletAddress: wallet2Address, signatureWeight: 50 },
      { walletAddress: wallet3Address, signatureWeight: 50 },
    ],
    quorum: 100,
    nonce: BigInt(Date.now() * 1000000),
  };

  const session2CreateRequest = packCreateAppSessionRequestV1(session2Definition, '{}');
  const appSession2CreateSig = await appSession2Signer.signMessage(session2CreateRequest);
  const appSession3CreateSig = await appSession3Signer.signMessage(session2CreateRequest);

  const { appSessionId: session2ID } = await wallet2Client.createAppSession(
    session2Definition,
    '{}',
    [appSession2CreateSig, appSession3CreateSig]
  );
  console.log(`✓ Created App Session 2: ${session2ID}\n`);

  // --- 4. Deposit YELLOW into Session 2 by Wallet 2 ---
  console.log('=== Step 4: Depositing YELLOW into Session 2 ===');

  const session2DepositAmount = new Decimal(0.00015);
  const session2DepositUpdate: AppStateUpdateV1 = {
    appSessionId: session2ID,
    intent: AppStateUpdateIntent.Deposit,
    version: 2n,
    allocations: [
      { participant: wallet2Address, asset: 'yellow', amount: session2DepositAmount },
    ],
    sessionData: '{}',
  };

  const session2DepositRequest = packAppStateUpdateV1(session2DepositUpdate);
  const appSession2DepositSig = await appSession2Signer.signMessage(session2DepositRequest);
  const appSession3DepositSig = await appSession3Signer.signMessage(session2DepositRequest);

  const nodeSig = await wallet2Client.submitAppSessionDeposit(
    session2DepositUpdate,
    [appSession2DepositSig, appSession3DepositSig],
    'yellow',
    session2DepositAmount
  );
  console.log(`✓ Deposited ${session2DepositAmount} YELLOW into Session 2 (Node Sig: ${nodeSig})\n`);

  // Check Session 2 state before redistribution
  const { sessions: session2InfoBeforeRedist } = await wallet2Client.getAppSessions({
    appSessionId: session2ID,
  });
  if (session2InfoBeforeRedist.length > 0) {
    console.log(
      `Session 2 before redistribution - Version: ${session2InfoBeforeRedist[0].version}, Allocations: ${JSON.stringify(session2InfoBeforeRedist[0].allocations)}\n`
    );
  }

  // --- 5. Redistribute within Session 2 (Wallet 2 -> Wallet 3) ---
  console.log('=== Step 5: Redistributing funds in Session 2 ===');

  const session2RedistributeUpdate: AppStateUpdateV1 = {
    appSessionId: session2ID,
    intent: AppStateUpdateIntent.Operate,
    version: 3n,
    allocations: [
      { participant: wallet2Address, asset: 'yellow', amount: new Decimal(0.0001) },
      { participant: wallet3Address, asset: 'yellow', amount: new Decimal(0.00005) },
    ],
    sessionData: '{}',
  };

  const session2RedistributeRequest = packAppStateUpdateV1(session2RedistributeUpdate);
  const appSession2RedistributeSig = await appSession2Signer.signMessage(
    session2RedistributeRequest
  );
  const appSession3RedistributeSig = await appSession3Signer.signMessage(
    session2RedistributeRequest
  );

  // Multi-sig required for state transition
  try {
    await wallet2Client.submitAppState(session2RedistributeUpdate, [
      appSession2RedistributeSig,
      appSession3RedistributeSig,
    ]);
    console.log('✓ Redistributed YELLOW: Wallet 2 (0.0001) -> Wallet 3 (0.00005)\n');
  } catch (err) {
    console.error(`Redistribution failed: ${err}`);
    throw err;
  }

  // NOTE: Rebalance step is disabled.
  // // --- Rebalance Both App Sessions Atomically ---
  // ... (rebalance code omitted) ...

  // --- 6. Wallet 3 Withdraws from Session 2 ---
  console.log('=== Step 6: Wallet 3 Withdrawing from Session 2 ===');

  const session2WithdrawUpdate: AppStateUpdateV1 = {
    appSessionId: session2ID,
    intent: AppStateUpdateIntent.Withdraw,
    version: 4n,
    allocations: [
      { participant: wallet2Address, asset: 'yellow', amount: new Decimal(0.00005) },
      { participant: wallet3Address, asset: 'yellow', amount: new Decimal(0.00001) },
    ],
    sessionData: '{}',
  };

  const session2WithdrawRequest = packAppStateUpdateV1(session2WithdrawUpdate);
  const appSession2WithdrawSig = await appSession2Signer.signMessage(
    session2WithdrawRequest
  );
  const appSession3WithdrawSig = await appSession3Signer.signMessage(
    session2WithdrawRequest
  );

  try {
    await wallet2Client.submitAppState(session2WithdrawUpdate, [
      appSession2WithdrawSig,
      appSession3WithdrawSig,
    ]);
    console.log('✓ Wallet 3 successfully withdrew YELLOW back to channel\n');
  } catch (err) {
    console.log(`⚠ Withdraw Error: ${err}\n`);
  }

  // --- 7. Close Both App Sessions ---
  console.log('=== Step 7: Closing Both App Sessions ===');

  // Close Session 1
  const session1CloseUpdate: AppStateUpdateV1 = {
    appSessionId: session1ID,
    intent: AppStateUpdateIntent.Close,
    version: 3n,
    allocations: [
      { participant: wallet1Address, asset: 'yusd', amount: new Decimal(0.0001) },
    ],
    sessionData: '{}',
  };

  const session1CloseRequest = packAppStateUpdateV1(session1CloseUpdate);
  const appSession1CloseSig = await appSession1Signer.signMessage(session1CloseRequest);

  try {
    await wallet1Client.submitAppState(session1CloseUpdate, [appSession1CloseSig]);
    console.log('✓ Session 1 successfully closed');
  } catch (err) {
    console.log(`⚠ Close Session 1 Error: ${err}`);
  }

  // Close Session 2
  const session2CloseUpdate: AppStateUpdateV1 = {
    appSessionId: session2ID,
    intent: AppStateUpdateIntent.Close,
    version: 5n,
    allocations: [
      { participant: wallet2Address, asset: 'yellow', amount: new Decimal(0.00005) },
      { participant: wallet3Address, asset: 'yellow', amount: new Decimal(0.00001) },
    ],
    sessionData: '{}',
  };

  const session2CloseRequest = packAppStateUpdateV1(session2CloseUpdate);
  const appSession2CloseSig = await appSession2Signer.signMessage(session2CloseRequest);
  const appSession3CloseSig = await appSession3Signer.signMessage(session2CloseRequest);

  try {
    await wallet2Client.submitAppState(session2CloseUpdate, [
      appSession2CloseSig,
      appSession3CloseSig,
    ]);
    console.log('✓ Session 2 successfully closed');
  } catch (err) {
    console.log(`⚠ Close Session 2 Error: ${err}`);
  }

  console.log('\n=== Example Complete ===');

  // Close clients
  await wallet1Client.close();
  await wallet2Client.close();
  await wallet3Client.close();

  // Exit successfully
  process.exit(0);
}

// Run the example
main().catch((error) => {
  console.error('Fatal error:', error);
  process.exit(1);
});
