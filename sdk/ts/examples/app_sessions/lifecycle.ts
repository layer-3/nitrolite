/**
 * Example: Complete App Session Lifecycle
 *
 * Prerequisites (minimum channel balances):
 *   - Wallet 1: 0.0001 USDC
 *   - Wallet 2: 0.00015 WETH
 *   - Wallet 3: no balance required (receives funds via redistribution)
 *
 * This example demonstrates:
 * 1. Register apps in the app registry (required before creating app sessions)
 * 2. Create first app session for wallet 1
 * 3. Deposit USDC into first app session by wallet 1
 * 4. Create second app session for wallet 2 with wallet 3 as a participant
 * 5. Deposit WETH into second app session by wallet 2
 * 6. Redistribute app state within app session so that participant with wallet 3 also has some allocation
 * 7. Wallet 3 withdraws from his app session
 * 8. Close both app sessions
 * 9. Fail case: attempt to create app session for unregistered app (expected to fail)
 */

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

  // --- 1. Register Apps ---
  console.log('=== Step 1: Registering Apps ===');

  const suffix = String(Math.floor(Math.random() * 1000000)).padStart(6, '0');
  const app1ID = `test-app-${suffix}`;
  const app2ID = `multi-party-app-${suffix}`;

  await wallet1Client.registerApp(app1ID, '{}', true);
  console.log(`✓ Registered app: ${app1ID}`);

  await wallet1Client.registerApp(app2ID, '{}', false);
  console.log(`✓ Registered app: ${app2ID} (owner approval required)\n`);

  // --- 2. Create App Session 1 (Single Participant: Wallet 1) ---
  console.log('=== Step 2: Creating App Session 1 (Wallet 1 only) ===');

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

  // --- 3. Deposit USDC into Session 1 ---
  console.log('=== Step 3: Depositing USDC into Session 1 ===');

  const session1DepositAmount = new Decimal(0.0001);
  const session1DepositUpdate: AppStateUpdateV1 = {
    appSessionId: session1ID,
    intent: AppStateUpdateIntent.Deposit,
    version: 2n,
    allocations: [
      { participant: wallet1Address, asset: 'usdc', amount: session1DepositAmount },
    ],
    sessionData: '{}',
  };

  const session1DepositRequest = packAppStateUpdateV1(session1DepositUpdate);
  const appSession1DepositSig = await appSession1Signer.signMessage(session1DepositRequest);

  try {
    await wallet1Client.submitAppSessionDeposit(
      session1DepositUpdate,
      [appSession1DepositSig],
      'usdc',
      session1DepositAmount
    );
    console.log(`✓ Deposited ${session1DepositAmount} USDC into Session 1\n`);
  } catch (err) {
    console.log(`⚠ Deposit warning: ${err}`);
  }

  // --- 4. Create App Session 2 (Multi-Party: Wallet 2 & 3) ---
  console.log('=== Step 4: Creating App Session 2 (Wallet 2 & 3) ===');

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
  };

  const packedAppSessionKey3State = packAppSessionKeyStateV1(appSessionKey3State);
  const wallet3MsgSigner = new EthereumMsgSigner(wallet3PrivateKey);
  const appSessionKey3StateSig = await wallet3MsgSigner.signMessage(packedAppSessionKey3State);
  appSessionKey3State.user_sig = appSessionKey3StateSig;

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

  // Owner approval: wallet1 is the owner of app2, sign the create request using app session signer
  const ownerApprovalSig = await appSession1Signer.signMessage(session2CreateRequest);

  const { appSessionId: session2ID } = await wallet2Client.createAppSession(
    session2Definition,
    '{}',
    [appSession2CreateSig, appSession3CreateSig],
    { ownerSig: ownerApprovalSig }
  );
  console.log(`✓ Created App Session 2: ${session2ID}\n`);

  // --- 5. Deposit WETH into Session 2 by Wallet 2 ---
  console.log('=== Step 5: Depositing WETH into Session 2 ===');

  const session2DepositAmount = new Decimal(0.00015);
  const session2DepositUpdate: AppStateUpdateV1 = {
    appSessionId: session2ID,
    intent: AppStateUpdateIntent.Deposit,
    version: 2n,
    allocations: [
      { participant: wallet2Address, asset: 'weth', amount: session2DepositAmount },
    ],
    sessionData: '{}',
  };

  const session2DepositRequest = packAppStateUpdateV1(session2DepositUpdate);
  const appSession2DepositSig = await appSession2Signer.signMessage(session2DepositRequest);
  const appSession3DepositSig = await appSession3Signer.signMessage(session2DepositRequest);

  const nodeSig = await wallet2Client.submitAppSessionDeposit(
    session2DepositUpdate,
    [appSession2DepositSig, appSession3DepositSig],
    'weth',
    session2DepositAmount
  );
  console.log(`✓ Deposited ${session2DepositAmount} WETH into Session 2 (Node Sig: ${nodeSig})\n`);

  // Check Session 2 state before redistribution
  const { sessions: session2InfoBeforeRedist } = await wallet2Client.getAppSessions({
    appSessionId: session2ID,
  });
  if (session2InfoBeforeRedist.length > 0) {
    console.log(
      `Session 2 before redistribution - Version: ${session2InfoBeforeRedist[0].version}, Allocations: ${JSON.stringify(session2InfoBeforeRedist[0].allocations)}\n`
    );
  }

  // --- 6. Redistribute within Session 2 (Wallet 2 -> Wallet 3) ---
  console.log('=== Step 6: Redistributing funds in Session 2 ===');

  const session2RedistributeUpdate: AppStateUpdateV1 = {
    appSessionId: session2ID,
    intent: AppStateUpdateIntent.Operate,
    version: 3n,
    allocations: [
      { participant: wallet2Address, asset: 'weth', amount: new Decimal(0.0001) },
      { participant: wallet3Address, asset: 'weth', amount: new Decimal(0.00005) },
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
    console.log('✓ Redistributed WETH: Wallet 2 (0.0001) -> Wallet 3 (0.00005)\n');
  } catch (err) {
    console.error(`Redistribution failed: ${err}`);
    throw err;
  }

  // NOTE: Rebalance step is disabled.
  // // --- 7. Rebalance Both App Sessions Atomically ---
  // ... (rebalance code omitted) ...

  // --- 7. Wallet 3 Withdraws from Session 2 ---
  console.log('=== Step 7: Wallet 3 Withdrawing from Session 2 ===');

  const session2WithdrawUpdate: AppStateUpdateV1 = {
    appSessionId: session2ID,
    intent: AppStateUpdateIntent.Withdraw,
    version: 4n,
    allocations: [
      { participant: wallet2Address, asset: 'weth', amount: new Decimal(0.00005) },
      { participant: wallet3Address, asset: 'weth', amount: new Decimal(0.00001) },
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
    console.log('✓ Wallet 3 successfully withdrew WETH back to channel\n');
  } catch (err) {
    console.log(`⚠ Withdraw Error: ${err}\n`);
  }

  // --- 8. Close Both App Sessions ---
  console.log('=== Step 8: Closing Both App Sessions ===');

  // Close Session 1
  const session1CloseUpdate: AppStateUpdateV1 = {
    appSessionId: session1ID,
    intent: AppStateUpdateIntent.Close,
    version: 3n,
    allocations: [
      { participant: wallet1Address, asset: 'usdc', amount: new Decimal(0.0001) },
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
      { participant: wallet2Address, asset: 'weth', amount: new Decimal(0.00005) },
      { participant: wallet3Address, asset: 'weth', amount: new Decimal(0.00001) },
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

  // --- 9. Fail Case: Create App Session for Unregistered App ---
  console.log('\n=== Step 9: Creating App Session for Unregistered App (expected to fail) ===');

  const unregisteredDefinition: AppDefinitionV1 = {
    applicationId: `unregistered-app-${suffix}`,
    participants: [{ walletAddress: wallet1Address, signatureWeight: 100 }],
    quorum: 100,
    nonce: BigInt(Date.now() * 1000000),
  };

  const unregisteredCreateRequest = packCreateAppSessionRequestV1(unregisteredDefinition, '{}');
  const unregisteredSig = await appSession1Signer.signMessage(unregisteredCreateRequest);

  try {
    await wallet1Client.createAppSession(
      unregisteredDefinition,
      '{}',
      [unregisteredSig]
    );
    console.log('✗ Unexpected success: app session was created for unregistered app');
  } catch (err) {
    console.log(`✓ Expected error: ${err}`);
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
