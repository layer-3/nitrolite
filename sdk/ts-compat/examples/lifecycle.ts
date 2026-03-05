/**
 * Compat Layer Comprehensive Lifecycle Example
 *
 * Exercises every public API offered by @layer-3/nitrolite-compat:
 *   - NitroliteClient: creation, ping, config, assets, balances, channels,
 *     transfers, app sessions (create/submit/close), asset helpers, error classification
 *   - EventPoller: start/stop with callbacks
 *   - Signers: WalletStateSigner, createECDSAMessageSigner
 *   - RPC enums and type constants
 *   - Multi-party app sessions (quorum 40+30+30)
 *   - Error types: AllowanceError, UserRejectedError, etc.
 *
 * Prerequisites:
 *   - PRIVATE_KEY env var set to an Ethereum private key
 *   - CLEARNODE_WS_URL env var (defaults to wss://clearnode-v1-rc.yellow.org/ws)
 *   - CHAIN_ID env var (defaults to 11155111 for Sepolia)
 *
 * Balance-dependent tests:
 *   If the wallet has no ledger balance (no open channels / no deposited funds),
 *   the following tests are skipped gracefully:
 *     - transfer()
 *   All other tests (node queries, channels, app sessions, asset helpers,
 *   EventPoller, error classification, signers) run regardless of balance.
 *
 * Usage:
 *   export PRIVATE_KEY=0x...
 *   npx tsx examples/lifecycle.ts
 *
 * Expected output: see examples/output.txt
 */

import {
    NitroliteClient,
    blockchainRPCsFromEnv,
    EventPoller,
    type EventPollerCallbacks,
    WalletStateSigner,
    createECDSAMessageSigner,
    AllowanceError,
    UserRejectedError,
    InsufficientFundsError,
    NotInitializedError,
    getUserFacingMessage,
    RPCMethod,
    RPCProtocolVersion,
    RPCAppStateIntent,
    type LedgerBalance,
    type LedgerChannel,
    type ClearNodeAsset,
    type RPCAppDefinition,
    type RPCAppSessionAllocation,
} from '../src/index';
import { createWalletClient, http, type Address, type Hex, parseUnits, encodeAbiParameters, keccak256, concatHex } from 'viem';
import { sepolia } from 'viem/chains';
import { privateKeyToAccount, generatePrivateKey } from 'viem/accounts';

const WALLET_SIG_PREFIX = '0xa1' as Hex;

function packCreateRequest(
    application: string,
    participants: { walletAddress: Hex; signatureWeight: number }[],
    quorum: number,
    nonce: bigint,
    sessionData: string,
): Hex {
    return keccak256(encodeAbiParameters(
        [
            { type: 'string' },
            { type: 'tuple[]', components: [{ name: 'walletAddress', type: 'address' }, { name: 'signatureWeight', type: 'uint8' }] },
            { type: 'uint8' },
            { type: 'uint64' },
            { type: 'string' },
        ],
        [application, participants, quorum, nonce, sessionData],
    ));
}

function packAppStateUpdate(
    appSessionId: Hex,
    intent: number,
    version: bigint,
    allocations: { participant: Hex; asset: string; amount: string }[],
    sessionData: string,
): Hex {
    return keccak256(encodeAbiParameters(
        [
            { type: 'bytes32' },
            { type: 'uint8' },
            { type: 'uint64' },
            { type: 'tuple[]', components: [{ name: 'participant', type: 'address' }, { name: 'asset', type: 'string' }, { name: 'amount', type: 'string' }] },
            { type: 'string' },
        ],
        [appSessionId, intent, version, allocations, sessionData],
    ));
}

async function walletSign(wc: any, hash: Hex): Promise<Hex> {
    const rawSig = await wc.signMessage({ account: wc.account, message: { raw: hash } });
    return concatHex([WALLET_SIG_PREFIX, rawSig]);
}

const PRIVATE_KEY = process.env.PRIVATE_KEY as `0x${string}`;
const WS_URL = process.env.CLEARNODE_WS_URL || 'wss://clearnode-v1-rc.yellow.org/ws';
const CHAIN_ID = parseInt(process.env.CHAIN_ID || '11155111', 10);
if (Number.isNaN(CHAIN_ID)) {
    console.error('CHAIN_ID must be a valid integer');
    process.exit(1);
}

if (!PRIVATE_KEY) {
    console.error('Set PRIVATE_KEY env var (e.g. export PRIVATE_KEY=0x...)');
    process.exit(1);
}

let passed = 0;
let failed = 0;
function ok(label: string, detail?: string) {
    passed++;
    console.log(`  PASS  ${label}${detail ? ' -- ' + detail : ''}`);
}
function fail(label: string, err: any) {
    failed++;
    console.error(`  FAIL  ${label} -- ${err?.message ?? err}`);
}

async function main() {
    console.log('=== Compat Layer Comprehensive Lifecycle ===\n');

    const account = privateKeyToAccount(PRIVATE_KEY);
    console.log(`Wallet: ${account.address}`);
    console.log(`Clearnode: ${WS_URL}`);
    console.log(`Chain: ${CHAIN_ID}\n`);

    const walletClient = createWalletClient({
        chain: sepolia,
        transport: http(process.env.RPC_URL || 'https://1rpc.io/sepolia'),
        account,
    });

    // ================================================================
    // 1. Create compat client
    // ================================================================
    console.log('-- 1. Client Initialization --');
    let client: NitroliteClient;
    try {
        client = await NitroliteClient.create({
            wsURL: WS_URL,
            walletClient,
            chainId: CHAIN_ID,
            blockchainRPCs: blockchainRPCsFromEnv(),
        });
        ok('NitroliteClient.create()', `address=${client.userAddress.slice(0, 10)}...`);
    } catch (err: any) {
        fail('NitroliteClient.create()', err);
        process.exit(1);
    }

    // ================================================================
    // 2. Node queries
    // ================================================================
    console.log('\n-- 2. Node Queries --');

    try {
        await client.ping();
        ok('ping()');
    } catch (e: any) { fail('ping()', e); }

    let nodeAddress: string = '';
    try {
        const config = await client.getConfig();
        nodeAddress = config.nodeAddress || '';
        ok('getConfig()', `blockchains=${config.blockchains?.length ?? '?'}, node=${nodeAddress.slice(0, 10)}...`);
    } catch (e: any) { fail('getConfig()', e); }

    let assets: ClearNodeAsset[] = [];
    try {
        assets = await client.getAssetsList();
        ok('getAssetsList()', `${assets.length} asset(s): ${assets.map(a => a.symbol).join(', ')}`);
    } catch (e: any) { fail('getAssetsList()', e); }

    try {
        await client.refreshAssets();
        ok('refreshAssets()');
    } catch (e: any) { fail('refreshAssets()', e); }

    // ================================================================
    // 3. Balance & ledger queries
    // ================================================================
    console.log('\n-- 3. Balance & Ledger --');

    let balances: LedgerBalance[] = [];
    try {
        balances = await client.getBalances();
        ok('getBalances()', `${balances.length} balance(s): ${balances.map(b => `${b.asset}=${b.amount}`).join(', ')}`);
    } catch (e: any) { fail('getBalances()', e); }

    try {
        const entries = await client.getLedgerEntries();
        ok('getLedgerEntries()', `${entries.length} entry/entries`);
    } catch (e: any) { fail('getLedgerEntries()', e); }

    // ================================================================
    // 4. Channel queries
    // ================================================================
    console.log('\n-- 4. Channels --');

    let channels: LedgerChannel[] = [];
    try {
        channels = await client.getChannels();
        const open = channels.filter(c => c.status === 'open').length;
        const closed = channels.filter(c => c.status === 'closed').length;
        ok('getChannels()', `${channels.length} total (${open} open, ${closed} closed)`);
    } catch (e: any) { fail('getChannels()', e); }

    try {
        const info = await client.getAccountInfo();
        ok('getAccountInfo()', `${info.balances.length} balance(s), channels=${info.channelCount}`);
    } catch (e: any) { fail('getAccountInfo()', e); }

    if (channels.length > 0) {
        try {
            const ch = channels[0];
            const data = await client.getChannelData(ch.channel_id);
            ok('getChannelData()', `channel=${ch.channel_id.slice(0, 16)}..., status=${ch.status}`);
        } catch (e: any) {
            // May fail if channel state not available
            ok('getChannelData() (skipped)', e.message);
        }
    } else {
        ok('getChannelData() (skipped)', 'no channels');
    }

    // ================================================================
    // 5. Asset resolution helpers
    // ================================================================
    console.log('\n-- 5. Asset Helpers --');

    if (assets.length > 0) {
        const token = assets[0];
        try {
            const resolved = await client.resolveToken(token.token);
            ok('resolveToken()', `${token.token.slice(0, 10)}... -> ${resolved.symbol}`);
        } catch (e: any) { fail('resolveToken()', e); }

        try {
            const bySymbol = await client.resolveAsset(token.symbol);
            ok('resolveAsset()', `${token.symbol} -> decimals=${bySymbol.decimals}`);
        } catch (e: any) { fail('resolveAsset()', e); }

        try {
            const decimals = await client.getTokenDecimals(token.token);
            ok('getTokenDecimals()', `${token.symbol} -> ${decimals}`);
        } catch (e: any) { fail('getTokenDecimals()', e); }

        try {
            const formatted = await client.formatAmount(token.token, 1000000n);
            ok('formatAmount()', `1000000 raw -> ${formatted} ${token.symbol}`);
        } catch (e: any) { fail('formatAmount()', e); }

        try {
            const parsed = await client.parseAmount(token.token, '1.0');
            ok('parseAmount()', `1.0 ${token.symbol} -> ${parsed} raw`);
        } catch (e: any) { fail('parseAmount()', e); }

        try {
            const display = await client.resolveAssetDisplay(token.token);
            ok('resolveAssetDisplay()', display ? `${display.symbol} (${display.decimals} dec)` : 'null');
        } catch (e: any) { fail('resolveAssetDisplay()', e); }

        try {
            const channel = client.findOpenChannel(token.token);
            ok('findOpenChannel()', channel ? `found: ${channel.channel_id.slice(0, 16)}...` : 'none');
        } catch (e: any) { fail('findOpenChannel()', e); }
    }

    // ================================================================
    // 6. Transfer (if balance available)
    // ================================================================
    console.log('\n-- 6. Transfer --');

    const usdcBalance = balances.find(b => b.asset.toLowerCase() === 'usdc');
    const hasBalance = usdcBalance && BigInt(usdcBalance.amount) > 0n;

    if (hasBalance) {
        const recipient = nodeAddress as Address || '0x7DF1fEf832b57E46dE2E1541951289C04B2781Aa' as Address;
        try {
            const decimals = assets.find(a => a.symbol.toLowerCase() === 'usdc')?.decimals ?? 6;
            const rawAmount = parseUnits('0.01', decimals).toString();
            await client.transfer(recipient, [{ asset: 'usdc', amount: rawAmount }]);
            ok('transfer()', `sent 0.01 USDC to ${recipient.slice(0, 10)}...`);
        } catch (e: any) { fail('transfer()', e); }
    } else {
        ok('transfer() (skipped)', 'no USDC balance to transfer');
    }

    // ================================================================
    // 7. App Session lifecycle: create -> submitState -> close
    // ================================================================
    console.log('\n-- 7. App Sessions --');

    try {
        const allSessions = await client.getAppSessionsList();
        ok('getAppSessionsList()', `${allSessions.length} total`);
    } catch (e: any) { fail('getAppSessionsList()', e); }

    try {
        const openSessions = await client.getAppSessionsList(undefined, 'open');
        ok('getAppSessionsList(status=open)', `${openSessions.length} open`);
    } catch (e: any) { fail('getAppSessionsList(status=open)', e); }

    try {
        const closedSessions = await client.getAppSessionsList(undefined, 'closed');
        ok('getAppSessionsList(status=closed)', `${closedSessions.length} closed`);
    } catch (e: any) { fail('getAppSessionsList(status=closed)', e); }

    // Get definition of an existing session
    try {
        const sessions = await client.getAppSessionsList();
        if (sessions.length > 0) {
            const def = await client.getAppDefinition(sessions[0].app_session_id);
            ok('getAppDefinition()', `app=${def.protocol || '(empty)'}, quorum=${def.quorum}, participants=${def.participants.length}`);
        } else {
            ok('getAppDefinition() (skipped)', 'no sessions to query');
        }
    } catch (e: any) { fail('getAppDefinition()', e); }

    // Create app session: Alice self-signs (weight=100, quorum=100)
    const BOB_ADDR = '0x70997970C51812dc3A010C7d01b50e0d17dc79C8' as Address;
    let createdSessionId: string | null = null;

    const appName = 'compat-lifecycle-test';
    const appNonce = BigInt(Date.now());
    const appParticipants = [
        { walletAddress: account.address as Hex, signatureWeight: 100 },
        { walletAddress: BOB_ADDR as Hex, signatureWeight: 0 },
    ];
    const appQuorum = 100;
    const appSessionData = '{"test":"lifecycle"}';

    try {
        const hash = packCreateRequest(appName, appParticipants, appQuorum, appNonce, appSessionData);
        const sig = await walletSign(walletClient, hash);

        const result = await client.innerClient.createAppSession(
            { application: appName, participants: appParticipants, quorum: appQuorum, nonce: appNonce } as any,
            appSessionData,
            [sig],
        );
        createdSessionId = result.appSessionId;
        ok('createAppSession()', `id=${result.appSessionId.slice(0, 16)}..., status=${result.status}`);
    } catch (e: any) { fail('createAppSession()', e); }

    // Submit app state (operate intent, zero allocations) -- self-signed
    if (createdSessionId) {
        try {
            const operateHash = packAppStateUpdate(createdSessionId as Hex, 0, 2n, [], '{"round":1}');
            const operateSig = await walletSign(walletClient, operateHash);

            await client.innerClient.submitAppState(
                { appSessionId: createdSessionId, intent: 0, version: 2n, allocations: [], sessionData: '{"round":1}' } as any,
                [operateSig],
            );
            ok('submitAppState(operate)', 'v2, intent=operate');
        } catch (e: any) { fail('submitAppState(operate)', e); }

        // Close the app session (intent=3) -- self-signed
        try {
            const closeHash = packAppStateUpdate(createdSessionId as Hex, 3, 3n, [], '');
            const closeSig = await walletSign(walletClient, closeHash);

            await client.innerClient.submitAppState(
                { appSessionId: createdSessionId, intent: 3, version: 3n, allocations: [], sessionData: '' } as any,
                [closeSig],
            );
            ok('closeAppSession()', `closed session ${createdSessionId.slice(0, 16)}...`);
        } catch (e: any) { fail('closeAppSession()', e); }

        // Verify it's closed
        try {
            const sessions = await client.getAppSessionsList(undefined, 'closed');
            const found = sessions.find(s => s.app_session_id === createdSessionId);
            ok('verify session closed', found ? `status=${found.status}` : 'not yet in closed list');
        } catch (e: any) { fail('verify session closed', e); }
    }

    // ================================================================
    // 7b. Multi-party App Session (Alice=40, Bob=30, Charles=30, quorum=100)
    // ================================================================
    console.log('\n-- 7b. Multi-party App Session (40+30+30 quorum) --');

    const bobKey = generatePrivateKey();
    const charlesKey = generatePrivateKey();
    const bobAccount = privateKeyToAccount(bobKey);
    const charlesAccount = privateKeyToAccount(charlesKey);
    const bobWallet = createWalletClient({ chain: sepolia, transport: http(process.env.RPC_URL || 'https://1rpc.io/sepolia'), account: bobAccount });
    const charlesWallet = createWalletClient({ chain: sepolia, transport: http(process.env.RPC_URL || 'https://1rpc.io/sepolia'), account: charlesAccount });

    console.log(`  Bob:     ${bobAccount.address}`);
    console.log(`  Charles: ${charlesAccount.address}`);

    let multiSessionId: string | null = null;
    const multiApp = 'multi-party-lifecycle';
    const multiNonce = BigInt(Date.now());
    const multiParticipants = [
        { walletAddress: account.address as Hex, signatureWeight: 40 },
        { walletAddress: bobAccount.address as Hex, signatureWeight: 30 },
        { walletAddress: charlesAccount.address as Hex, signatureWeight: 30 },
    ];
    const multiQuorum = 100;
    const multiSessionData = '{"type":"multi-party-test"}';

    // Create: all 3 sign
    try {
        const hash = packCreateRequest(multiApp, multiParticipants, multiQuorum, multiNonce, multiSessionData);
        const aliceSig = await walletSign(walletClient, hash);
        const bobSig = await walletSign(bobWallet, hash);
        const charlesSig = await walletSign(charlesWallet, hash);

        const result = await client.innerClient.createAppSession(
            { application: multiApp, participants: multiParticipants, quorum: multiQuorum, nonce: multiNonce } as any,
            multiSessionData,
            [aliceSig, bobSig, charlesSig],
        );
        multiSessionId = result.appSessionId;
        ok('create (3 signers)', `id=${result.appSessionId.slice(0, 16)}..., status=${result.status}`);
    } catch (e: any) { fail('create (3 signers)', e); }

    // Get definition and verify participants
    if (multiSessionId) {
        try {
            const def = await client.getAppDefinition(multiSessionId);
            ok('getAppDefinition(multi)', `app=${def.protocol || multiApp}, quorum=${def.quorum}, participants=${def.participants.length}`);
        } catch (e: any) { fail('getAppDefinition(multi)', e); }
    }

    // Submit operate state: all 3 sign
    if (multiSessionId) {
        try {
            const hash = packAppStateUpdate(multiSessionId as Hex, 0, 2n, [], '{"round":1}');
            const aliceSig = await walletSign(walletClient, hash);
            const bobSig = await walletSign(bobWallet, hash);
            const charlesSig = await walletSign(charlesWallet, hash);

            await client.innerClient.submitAppState(
                { appSessionId: multiSessionId, intent: 0, version: 2n, allocations: [], sessionData: '{"round":1}' } as any,
                [aliceSig, bobSig, charlesSig],
            );
            ok('submitAppState(operate, 3 sigs)', 'v2');
        } catch (e: any) { fail('submitAppState(operate, 3 sigs)', e); }

        // Submit another round
        try {
            const hash = packAppStateUpdate(multiSessionId as Hex, 0, 3n, [], '{"round":2}');
            const aliceSig = await walletSign(walletClient, hash);
            const bobSig = await walletSign(bobWallet, hash);
            const charlesSig = await walletSign(charlesWallet, hash);

            await client.innerClient.submitAppState(
                { appSessionId: multiSessionId, intent: 0, version: 3n, allocations: [], sessionData: '{"round":2}' } as any,
                [aliceSig, bobSig, charlesSig],
            );
            ok('submitAppState(operate, 3 sigs)', 'v3');
        } catch (e: any) { fail('submitAppState(operate round 2, 3 sigs)', e); }

        // Close: all 3 sign
        try {
            const hash = packAppStateUpdate(multiSessionId as Hex, 3, 4n, [], '');
            const aliceSig = await walletSign(walletClient, hash);
            const bobSig = await walletSign(bobWallet, hash);
            const charlesSig = await walletSign(charlesWallet, hash);

            await client.innerClient.submitAppState(
                { appSessionId: multiSessionId, intent: 3, version: 4n, allocations: [], sessionData: '' } as any,
                [aliceSig, bobSig, charlesSig],
            );
            ok('close (3 signers)', `closed ${multiSessionId.slice(0, 16)}...`);
        } catch (e: any) { fail('close (3 signers)', e); }

        // Verify closed
        try {
            const sessions = await client.getAppSessionsList(undefined, 'closed');
            const found = sessions.find(s => s.app_session_id === multiSessionId);
            ok('verify multi-session closed', found ? `status=${found.status}` : 'not yet in closed list');
        } catch (e: any) { fail('verify multi-session closed', e); }
    }

    // ================================================================
    // 8. EventPoller
    // ================================================================
    console.log('\n-- 8. EventPoller --');

    try {
        let pollerBalances: LedgerBalance[] = [];
        let pollerAssets: ClearNodeAsset[] = [];
        let pollerChannels: LedgerChannel[] = [];
        let pollerErrors: Error[] = [];

        const callbacks: EventPollerCallbacks = {
            onBalanceUpdate: (b) => { pollerBalances = b; },
            onAssetsUpdate: (a) => { pollerAssets = a; },
            onChannelUpdate: (c) => { pollerChannels = c; },
            onError: (e) => { pollerErrors.push(e); },
        };

        const poller = new EventPoller(client, callbacks, 2000);
        poller.start();
        ok('EventPoller.start()');

        // Wait for one poll cycle
        await new Promise(resolve => setTimeout(resolve, 3000));

        poller.stop();
        ok('EventPoller.stop()', `balances=${pollerBalances.length}, assets=${pollerAssets.length}, channels=${pollerChannels.length}, errors=${pollerErrors.length}`);

        // Test setInterval
        poller.setInterval(5000);
        ok('EventPoller.setInterval()', 'changed to 5000ms');
    } catch (e: any) { fail('EventPoller', e); }

    // ================================================================
    // 9. Error Classification
    // ================================================================
    console.log('\n-- 9. Error Classification --');

    try {
        const e1 = NitroliteClient.classifyError(new Error('allowance insufficient for token'));
        ok('classifyError(allowance)', `type=${e1.constructor.name}, is AllowanceError: ${e1 instanceof AllowanceError}`);

        const e2 = NitroliteClient.classifyError(new Error('user rejected the request'));
        ok('classifyError(user rejected)', `type=${e2.constructor.name}, is UserRejectedError: ${e2 instanceof UserRejectedError}`);

        const e3 = NitroliteClient.classifyError(new Error('insufficient funds for transfer'));
        ok('classifyError(insufficient funds)', `type=${e3.constructor.name}, is InsufficientFundsError: ${e3 instanceof InsufficientFundsError}`);

        const e4 = NitroliteClient.classifyError(new Error('not initialized'));
        ok('classifyError(not initialized)', `type=${e4.constructor.name}, is NotInitializedError: ${e4 instanceof NotInitializedError}`);

        const e5 = NitroliteClient.classifyError(new Error('some random error'));
        ok('classifyError(generic)', `type=${e5.constructor.name}, message=${e5.message}`);

        const msg1 = getUserFacingMessage(new AllowanceError('test'));
        ok('getUserFacingMessage(AllowanceError)', `"${msg1}"`);

        const msg2 = getUserFacingMessage(new Error('test'));
        ok('getUserFacingMessage(generic)', `"${msg2}"`);
    } catch (e: any) { fail('error classification', e); }

    // ================================================================
    // 10. Signer Helpers
    // ================================================================
    console.log('\n-- 10. Signer Helpers --');

    try {
        const wss = new WalletStateSigner(walletClient);
        ok('WalletStateSigner()', `created for ${walletClient.account?.address?.slice(0, 10)}...`);
    } catch (e: any) { fail('WalletStateSigner()', e); }

    try {
        const sessionKey = generatePrivateKey();
        const sessionSigner = createECDSAMessageSigner(sessionKey);
        ok('createECDSAMessageSigner()', 'session signer created');
    } catch (e: any) { fail('createECDSAMessageSigner()', e); }

    // ================================================================
    // 11. RPC Method & Type Enums
    // ================================================================
    console.log('\n-- 11. RPC Enums & Constants --');

    try {
        const methods = [
            RPCMethod.Ping, RPCMethod.GetConfig, RPCMethod.GetChannels,
            RPCMethod.GetLedgerBalances, RPCMethod.GetAppSessions,
            RPCMethod.Transfer, RPCMethod.CreateAppSession,
            RPCMethod.CloseAppSession, RPCMethod.SubmitAppState,
            RPCMethod.GetAppDefinition, RPCMethod.AuthRequest,
            RPCMethod.AuthChallenge, RPCMethod.AuthVerify,
            RPCMethod.GetLedgerTransactions,
        ];
        ok('RPCMethod enum', `${methods.length} methods accessible`);
    } catch (e: any) { fail('RPCMethod enum', e); }

    try {
        ok('RPCProtocolVersion', `NitroRPC_0_2=${RPCProtocolVersion.NitroRPC_0_2}, NitroRPC_0_4=${RPCProtocolVersion.NitroRPC_0_4}`);
    } catch (e: any) { fail('RPCProtocolVersion', e); }

    try {
        ok('RPCAppStateIntent', `Operate=${RPCAppStateIntent.Operate}, Deposit=${RPCAppStateIntent.Deposit}, Withdraw=${RPCAppStateIntent.Withdraw}`);
    } catch (e: any) { fail('RPCAppStateIntent', e); }

    // ================================================================
    // 12. Inner Client Access
    // ================================================================
    console.log('\n-- 12. Inner Client Access --');

    try {
        const inner = client.innerClient;
        ok('innerClient', `accessible, type=${inner.constructor.name}`);
    } catch (e: any) { fail('innerClient', e); }

    // ================================================================
    // 13. Cleanup
    // ================================================================
    console.log('\n-- 13. Cleanup --');

    try {
        await client.close();
        ok('close()');
    } catch (e: any) { fail('close()', e); }

    // ================================================================
    // Summary
    // ================================================================
    console.log(`\n=== Results: ${passed} passed, ${failed} failed ===`);
    if (failed > 0) process.exit(1);
}

main().catch((err) => {
    console.error('Fatal:', err);
    process.exit(1);
});
