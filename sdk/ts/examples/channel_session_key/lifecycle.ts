/**
 * Example: Channel Session Key Lifecycle
 *
 * Requirements to run this example:
 *
 *  1. A reachable nitronode WebSocket endpoint (set via wsURL below).
 *     The default points at the stress environment.
 *
 *  2. One EVM wallet with a hex private key (replace the placeholder below).
 *
 *  3. Minimum off-chain (channel) balances on the node:
 *       - 0.00005 YUSD   (one deposit + one withdraw via session key)
 *       - 0.00005 YELLOW (one deposit + one withdraw via session key)
 *
 *     An open channel is NOT a hard prerequisite. If the wallet already has
 *     funds on the node but no acknowledged channel yet, acknowledge() is run
 *     first to open one.
 *
 *  4. chainId below must match the asset's home blockchain for the target
 *     nitronode deployment, and rpcURL must point at a JSON-RPC endpoint for
 *     that chain. Both deposit and withdraw are followed by an on-chain
 *     checkpoint; the example then polls getHomeChannel until the node has
 *     observed the checkpoint event before moving on. Without a working RPC
 *     these calls fail.
 *
 * This example demonstrates:
 *  1. Open YUSD and YELLOW channels for the wallet (acknowledge)
 *  2. Generate a fresh session key
 *  3. Register session key v1 with both assets allowed
 *  4. Deposit YUSD and YELLOW via a session-key-backed client (success)
 *  5. Update session key v2 -> [YELLOW] only
 *  6. Withdraw YELLOW (success); attempt YUSD withdraw via session key (expected fail)
 *  7. Update session key v3 -> [YUSD] only
 *  8. Withdraw YUSD (success); attempt YELLOW deposit via session key (expected fail)
 *  9. Revoke session key v4 -> []
 * 10. Attempt YUSD deposit, YELLOW deposit, and channel closure via session key
 *     (all expected to fail)
 */

// Node <22 does not expose a stable global WebSocket. The SDK dialer uses
// `new WebSocket(url)` directly, so we polyfill the global from the `ws`
// package before any client is constructed.
import WebSocket from 'ws';
if (typeof (globalThis as { WebSocket?: unknown }).WebSocket === 'undefined') {
  (globalThis as { WebSocket: unknown }).WebSocket = WebSocket;
}

import Decimal from 'decimal.js';
import { Address, Hex } from 'viem';
import { generatePrivateKey, privateKeyToAccount } from 'viem/accounts';
import { Client } from '../../src/client';
import {
  createSigners,
  EthereumMsgSigner,
  EthereumRawSigner,
  ChannelDefaultSigner,
  ChannelSessionKeyStateSigner,
} from '../../src/signers';
import { withBlockchainRPC } from '../../src/config';
import { ChannelSessionKeyStateV1 } from '../../src/rpc/types';
import { getChannelSessionKeyAuthMetadataHashV1 } from '../../src/core/utils';
import { isFinal } from '../../src/core/state';

async function main() {
  const wsURL = 'wss://nitronode-stress.yellow.org/v1/ws';

  // Replace with your hex private key. The wallet must have minimum off-chain
  // balance for YUSD and YELLOW; channels are auto-opened below if missing.
  const walletPrivateKey = '0x7d6071201765d2630ca9eb83cbe3e2e2e76f9b56ea3ed13a49a00208ebcdf843' as Hex;

  // chainId is the home blockchain ID used for deposit / withdraw calls. Set
  // it to the asset's home chain on the target nitronode deployment. 11155111
  // is Ethereum Sepolia (the stress environment).
  const chainId = 11155111n;

  // rpcURL is a JSON-RPC endpoint for chainId. Replace with your own provider
  // if the public endpoint is rate-limited.
  const rpcURL = 'https://sepolia.drpc.org';

  // --- Setup wallet signers + wallet-backed SDK client ---
  const walletSigners = createSigners(walletPrivateKey);
  const walletAddress = walletSigners.stateSigner.getAddress();
  console.log(`Wallet: ${walletAddress}\n`);

  const walletClient = await Client.create(
    wsURL,
    walletSigners.stateSigner,
    walletSigners.txSigner,
    withBlockchainRPC(chainId, rpcURL)
  );

  try {
    // --- Step 1: ensure YUSD and YELLOW channels are open ---
    console.log('=== Step 1: Ensuring channels are open ===');
    await ensureChannelOpen(walletClient, 'yusd');
    await ensureChannelOpen(walletClient, 'yellow');
    console.log();

    // --- Step 2: generate a fresh session key ---
    console.log('=== Step 2: Generating session key ===');
    const sessionKeyPrivateKey = generatePrivateKey();
    const sessionKeyAccount = privateKeyToAccount(sessionKeyPrivateKey);
    const sessionKeyAddress = sessionKeyAccount.address;
    const sessionKeyMsgSigner = new EthereumMsgSigner(sessionKeyPrivateKey);
    console.log(`Session key: ${sessionKeyAddress}\n`);

    // --- Step 3: register session key v1 with both assets allowed ---
    console.log('=== Step 3: Registering session key v1 ([yusd, yellow]) ===');
    const stateV1 = await submitSessionKey(walletClient, walletAddress, sessionKeyAddress, sessionKeyMsgSigner, 1n, ['yusd', 'yellow']);
    console.log('✓ v1 registered\n');

    // --- Step 4: deposit YUSD and YELLOW via session-key client ---
    console.log('=== Step 4: Depositing via session-key client (v1) ===');
    const skClient1 = await newSessionKeyClient(wsURL, walletPrivateKey, sessionKeyPrivateKey, stateV1, chainId, rpcURL);
    try {
      const yusdDeposit = await skClient1.deposit(chainId, 'yusd', new Decimal(0.00001));
      console.log('✓ YUSD deposited via session key');
      await checkpointAndWait(skClient1, 'yusd', yusdDeposit.version);

      const yellowDeposit = await skClient1.deposit(chainId, 'yellow', new Decimal(0.00001));
      console.log('✓ YELLOW deposited via session key');
      await checkpointAndWait(skClient1, 'yellow', yellowDeposit.version);
    } finally {
      await skClient1.close();
    }
    console.log();

    // --- Step 5: update session key v2 -> [yellow] ---
    console.log('=== Step 5: Updating session key v2 ([yellow]) ===');
    const stateV2 = await submitSessionKey(walletClient, walletAddress, sessionKeyAddress, sessionKeyMsgSigner, 2n, ['yellow']);
    const skClient2 = await newSessionKeyClient(wsURL, walletPrivateKey, sessionKeyPrivateKey, stateV2, chainId, rpcURL);
    console.log('✓ v2 registered\n');

    try {
      // --- Step 6: withdraw YELLOW (ok); attempt YUSD withdraw (fail) ---
      console.log('=== Step 6: Withdraw via v2 (yellow only) ===');
      const yellowWithdraw = await skClient2.withdraw(chainId, 'yellow', new Decimal(0.000005));
      console.log('✓ YELLOW withdrawn via session key');
      await checkpointAndWait(skClient2, 'yellow', yellowWithdraw.version);

      try {
        await skClient2.withdraw(chainId, 'yusd', new Decimal(0.000005));
        console.log('✗ Unexpected: YUSD withdraw succeeded under v2');
      } catch (err) {
        console.log(`✓ Expected: YUSD withdraw rejected by node: ${err}`);
      }
    } finally {
      await skClient2.close();
    }
    console.log();

    // --- Step 7: update session key v3 -> [yusd] ---
    console.log('=== Step 7: Updating session key v3 ([yusd]) ===');
    const stateV3 = await submitSessionKey(walletClient, walletAddress, sessionKeyAddress, sessionKeyMsgSigner, 3n, ['yusd']);
    const skClient3 = await newSessionKeyClient(wsURL, walletPrivateKey, sessionKeyPrivateKey, stateV3, chainId, rpcURL);
    console.log('✓ v3 registered\n');

    try {
      // --- Step 8: withdraw YUSD (ok); attempt YELLOW deposit (fail) ---
      console.log('=== Step 8: Withdraw via v3 (yusd only) ===');
      const yusdWithdraw = await skClient3.withdraw(chainId, 'yusd', new Decimal(0.000005));
      console.log('✓ YUSD withdrawn via session key');
      await checkpointAndWait(skClient3, 'yusd', yusdWithdraw.version);

      try {
        await skClient3.deposit(chainId, 'yellow', new Decimal(0.000005));
        console.log('✗ Unexpected: YELLOW deposit succeeded under v3');
      } catch (err) {
        console.log(`✓ Expected: YELLOW deposit rejected by node: ${err}`);
      }
    } finally {
      await skClient3.close();
    }
    console.log();

    // --- Step 9: revoke session key v4 -> [] ---
    // Empty assets disables every per-asset check on the node, so the next
    // version of the key cannot authorize any channel operation.
    console.log('=== Step 9: Revoking session key v4 (empty assets) ===');
    const stateV4 = await submitSessionKey(walletClient, walletAddress, sessionKeyAddress, sessionKeyMsgSigner, 4n, []);
    const skClient4 = await newSessionKeyClient(wsURL, walletPrivateKey, sessionKeyPrivateKey, stateV4, chainId, rpcURL);
    console.log('✓ v4 registered (revoked)\n');

    try {
      // --- Step 10: every session-key operation must fail ---
      console.log('=== Step 10: Verifying revoked session key cannot operate ===');
      try {
        await skClient4.deposit(chainId, 'yusd', new Decimal(0.000005));
        console.log('✗ Unexpected: YUSD deposit succeeded under v4');
      } catch (err) {
        console.log(`✓ Expected: YUSD deposit rejected by node: ${err}`);
      }
      try {
        await skClient4.deposit(chainId, 'yellow', new Decimal(0.000005));
        console.log('✗ Unexpected: YELLOW deposit succeeded under v4');
      } catch (err) {
        console.log(`✓ Expected: YELLOW deposit rejected by node: ${err}`);
      }
      try {
        await skClient4.closeHomeChannel('yusd');
        console.log('✗ Unexpected: YUSD channel close succeeded under v4');
      } catch (err) {
        console.log(`✓ Expected: YUSD channel close rejected by node: ${err}`);
      }
    } finally {
      await skClient4.close();
    }

    console.log('\n=== Example Complete ===');
  } finally {
    await walletClient.close();
  }
}

/**
 * ensureChannelOpen guarantees that the wallet has an acknowledged channel
 * open for asset. If the node holds no state for the wallet/asset pair, or
 * the latest state is still awaiting the user's signature (or has been
 * finalized), acknowledge() is invoked to create or progress the channel.
 */
async function ensureChannelOpen(client: Client, asset: string): Promise<void> {
  const wallet = client.getUserAddress();
  const state = await client.getLatestState(wallet, asset, false);

  const hasOpenChannel =
    state !== null &&
    state.homeChannelId !== undefined &&
    !isFinal(state) &&
    !!state.userSig;
  if (hasOpenChannel) {
    console.log(`✓ channel already open for ${asset}`);
    return;
  }

  await client.acknowledge(asset);
  console.log(`✓ acknowledged channel for ${asset}`);
}

/**
 * submitSessionKey signs and submits a (version, assets) update for the
 * channel session key using the wallet client. Returns the registered state
 * (including user_sig + session_key_sig) so the caller can derive the matching
 * session-key state signer for subsequent operations.
 */
async function submitSessionKey(
  walletClient: Client,
  walletAddress: Address,
  sessionKeyAddress: Address,
  sessionKeyMsgSigner: EthereumMsgSigner,
  version: bigint,
  assets: string[]
): Promise<ChannelSessionKeyStateV1> {
  const expiresAt = BigInt(Math.floor(Date.now() / 1000) + 24 * 60 * 60);

  const state: ChannelSessionKeyStateV1 = {
    user_address: walletAddress,
    session_key: sessionKeyAddress,
    version: version.toString(),
    assets,
    expires_at: expiresAt.toString(),
    user_sig: '',
    session_key_sig: '',
  };

  state.user_sig = await walletClient.signChannelSessionKeyState(state);
  state.session_key_sig = await walletClient.signChannelSessionKeyOwnership(state, sessionKeyMsgSigner);

  await walletClient.submitChannelSessionKeyState(state);
  return state;
}

/**
 * newSessionKeyClient builds an SDK client whose state signer is the channel
 * session key derived from the registered state. All channel state operations
 * (deposit, withdraw, closeHomeChannel, …) issued through this client are
 * signed with the session key, and the node validates them against the latest
 * registered (user, session_key, version) tuple — including the asset
 * allow-list and expiry.
 *
 * walletPrivateKey must remain the wallet's key for the txSigner: the SDK
 * uses txSigner to sign on-chain checkpoint transactions, and ChannelHub will
 * only accept calls from the channel's user. Substituting the session key
 * here would either point the client at the wrong on-chain identity or fail
 * tx-level auth.
 */
async function newSessionKeyClient(
  wsURL: string,
  walletPrivateKey: Hex,
  sessionKeyPrivateKey: Hex,
  state: ChannelSessionKeyStateV1,
  chainId: bigint,
  rpcURL: string
): Promise<Client> {
  const walletAddress = privateKeyToAccount(walletPrivateKey).address;

  const metadataHash = getChannelSessionKeyAuthMetadataHashV1(
    walletAddress,
    BigInt(state.version),
    state.assets,
    BigInt(state.expires_at)
  );

  // user_sig is stripped of any ChannelSigner type-byte prefix by
  // Client.signChannelSessionKeyState before submission, so the value we have
  // here is already the raw EIP-191 signature expected by the signer.
  const stateSigner = new ChannelSessionKeyStateSigner(
    sessionKeyPrivateKey,
    walletAddress,
    metadataHash,
    state.user_sig as Hex
  );

  const txSigner = new EthereumRawSigner(walletPrivateKey);

  return await Client.create(wsURL, stateSigner, txSigner, withBlockchainRPC(chainId, rpcURL));
}

/**
 * checkpointAndWait runs checkpoint() for asset and polls getHomeChannel
 * until the node's observed on-chain state_version catches up to
 * expectedVersion. Without this barrier the next deposit/withdraw can race
 * the node's event ingestion and be rejected with "home deposit is still
 * ongoing".
 */
async function checkpointAndWait(client: Client, asset: string, expectedVersion: bigint): Promise<void> {
  const txHash = await client.checkpoint(asset);
  console.log(`  ↳ checkpoint ${asset} tx ${txHash} submitted; waiting for node to observe state_version=${expectedVersion}...`);

  const wallet = client.getUserAddress();
  const deadline = Date.now() + 2 * 60 * 1000;
  while (true) {
    const channel = await client.getHomeChannel(wallet, asset);
    if (channel !== null && channel.stateVersion >= expectedVersion) {
      console.log(`  ↳ node observed state_version=${channel.stateVersion} for ${asset}`);
      return;
    }
    if (Date.now() > deadline) {
      throw new Error(`timed out waiting for ${asset} to reach state_version=${expectedVersion}`);
    }
    await new Promise((resolve) => setTimeout(resolve, 2000));
  }
}

main().catch((error) => {
  console.error('Fatal error:', error);
  process.exit(1);
});
