import {
    Client,
    ChannelDefaultSigner,
    ChannelSessionKeyStateSigner,
    type StateSigner,
    type TransactionSigner,
} from '@yellow-org/sdk';
import type {
    AppDefinitionV1,
    AppParticipantV1,
    AppAllocationV1,
    AppSessionKeyStateV1,
    AppInfoV1,
    ChannelSessionKeyStateV1,
} from '@yellow-org/sdk';
import type * as core from '@yellow-org/sdk';
import { Decimal } from 'decimal.js';
import { Address, Hex, WalletClient, createPublicClient, http, formatUnits, parseUnits } from 'viem';

import type {
    RPCBalance,
    RPCChannelUpdate,
    RPCAsset,
    RPCAppDefinition,
    RPCAppSessionAllocation,
    TransferAllocation,
    ContractAddresses,
    AccountInfo,
    LedgerChannel,
    LedgerBalance,
    LedgerEntry,
    AppSession,
    ClearNodeAsset,
    SubmitAppStateRequestParams,
    SubmitAppStateRequestParamsV04,
    GetAppDefinitionResponseParams,
    CreateAppSessionRequestParams,
    CloseAppSessionRequestParams,
} from './types.js';
import { RPCAppStateIntent } from './types.js';

import { buildClientOptions, type CompatClientConfig } from './config.js';
import {
    AllowanceError,
    InsufficientFundsError,
    NotInitializedError,
    OngoingStateTransitionError,
    UserRejectedError,
} from './errors.js';

// ---------------------------------------------------------------------------
// WalletClient-based signers for browser (MetaMask) environments
// ---------------------------------------------------------------------------

class WalletMsgSigner implements StateSigner {
    constructor(private walletClient: WalletClient) {}

    getAddress(): Address {
        if (!this.walletClient.account?.address) throw new Error('Wallet has no account');
        return this.walletClient.account.address;
    }

    async signMessage(hash: Hex): Promise<Hex> {
        if (!this.walletClient.account) throw new Error('Wallet has no account');
        return this.walletClient.signMessage({
            account: this.walletClient.account,
            message: { raw: hash },
        });
    }
}

class WalletTxSigner implements TransactionSigner {
    constructor(private walletClient: WalletClient) {}

    getAddress(): Address {
        if (!this.walletClient.account?.address) throw new Error('Wallet has no account');
        return this.walletClient.account.address;
    }

    async sendTransaction(_tx: any): Promise<Hex> {
        throw new Error('Use the blockchain client for transactions');
    }

    async signMessage(message: { raw: Hex }): Promise<Hex> {
        if (!this.walletClient.account) throw new Error('Wallet has no account');
        const chain = this.walletClient.chain;
        return this.walletClient.signTypedData({
            account: this.walletClient.account,
            domain: { name: 'Nitrolite', version: '1', chainId: chain?.id ?? 1 },
            types: { Message: [{ name: 'data', type: 'bytes32' }] },
            primaryType: 'Message',
            message: { data: message.raw },
        });
    }
}

// ---------------------------------------------------------------------------
// Asset resolution helper
// ---------------------------------------------------------------------------

interface AssetInfo {
    symbol: string;
    chainId: bigint;
    decimals: number;
    tokenAddress: string;
}

// ---------------------------------------------------------------------------
// NitroliteClient compat facade
// ---------------------------------------------------------------------------

export interface NitroliteClientConfig {
    wsURL: string;
    walletClient: WalletClient;
    chainId: number;
    blockchainRPCs?: Record<number, string>;
    channelSessionKeySigner?: {
        sessionKeyPrivateKey: Hex;
        walletAddress: Address;
        metadataHash: Hex;
        authSig: Hex;
    };
    /** @deprecated v0.5.3 compat -- ignored, addresses come from get_config */
    addresses?: ContractAddresses;
    /** @deprecated v0.5.3 compat -- ignored */
    challengeDuration?: bigint;
}

export class NitroliteClient {
    /** The underlying v1.0.0 SDK Client -- use for any functionality not wrapped by compat. */
    readonly innerClient: Client;
    readonly userAddress: Address;

    private assetsByToken = new Map<string, AssetInfo>(); // lowercase tokenAddr -> info
    private assetsBySymbol = new Map<string, AssetInfo>(); // lowercase symbol -> info
    private _chainId: bigint;
    private _lastChannels: LedgerChannel[] = [];
    private _lastAppSessionsListError: string | null = null;
    private _lastAppSessionsListErrorLogged: string | null = null;
    private _blockchains: core.Blockchain[] | null = null;
    private _lockingTokenDecimals = new Map<number, number>();
    private _blockchainRPCs: Record<number, string>;

    private constructor(client: Client, userAddress: Address, chainId: number, blockchainRPCs?: Record<number, string>) {
        this.innerClient = client;
        this.userAddress = userAddress;
        this._chainId = BigInt(chainId);
        this._blockchainRPCs = blockchainRPCs ?? {};
    }

    // -----------------------------------------------------------------------
    // Factory
    // -----------------------------------------------------------------------

    static async create(config: NitroliteClientConfig): Promise<NitroliteClient> {
        const walletAddress = config.walletClient.account?.address;
        if (!walletAddress) throw new Error('WalletClient must have an account');

        let stateSigner: StateSigner;
        if (config.channelSessionKeySigner) {
            const signerWallet = config.channelSessionKeySigner.walletAddress;
            if (signerWallet.toLowerCase() !== walletAddress.toLowerCase()) {
                throw new Error(
                    `channelSessionKeySigner wallet ${signerWallet} does not match walletClient account ${walletAddress}`,
                );
            }

            stateSigner = new ChannelSessionKeyStateSigner(
                config.channelSessionKeySigner.sessionKeyPrivateKey,
                signerWallet,
                config.channelSessionKeySigner.metadataHash,
                config.channelSessionKeySigner.authSig,
            );
        } else {
            stateSigner = new ChannelDefaultSigner(new WalletMsgSigner(config.walletClient));
        }

        const txSigner = new WalletTxSigner(config.walletClient);

        const opts = buildClientOptions({
            wsURL: config.wsURL,
            blockchainRPCs: config.blockchainRPCs,
        });

        const v1Client = await Client.create(config.wsURL, stateSigner, txSigner, ...opts);

        const compat = new NitroliteClient(v1Client, walletAddress, config.chainId, config.blockchainRPCs);

        try {
            await compat.refreshAssets();
        } catch {
            console.warn('[compat] Could not pre-load asset map; will retry on demand');
        }

        return compat;
    }

    // -----------------------------------------------------------------------
    // Asset mapping
    // -----------------------------------------------------------------------

    async refreshAssets(): Promise<void> {
        const assets = await this.innerClient.getAssets();
        this.assetsByToken.clear();
        this.assetsBySymbol.clear();

        for (const asset of assets) {
            for (const token of asset.tokens) {
                const info: AssetInfo = {
                    symbol: asset.symbol,
                    chainId: token.blockchainId,
                    decimals: asset.decimals,
                    tokenAddress: token.address.toLowerCase(),
                };
                this.assetsByToken.set(info.tokenAddress, info);
                if (token.blockchainId === this._chainId) {
                    this.assetsBySymbol.set(asset.symbol.toLowerCase(), info);
                }
            }
        }
    }

    private async ensureAssets(): Promise<void> {
        if (this.assetsByToken.size === 0) await this.refreshAssets();
    }

    private async getDecimalsForAsset(assetSymbol: string): Promise<number> {
        await this.ensureAssets();
        const info = this.assetsBySymbol.get(assetSymbol.toLowerCase());
        if (!info) {
            console.warn(`[compat] Unknown asset symbol ${assetSymbol}, falling back to 6 decimals`);
        }
        return info?.decimals ?? 6;
    }

    async resolveToken(tokenAddress: Address | string): Promise<AssetInfo> {
        await this.ensureAssets();
        const key = tokenAddress.toString().toLowerCase();
        const info = this.assetsByToken.get(key);
        if (!info) throw new Error(`Unknown token address: ${tokenAddress}`);
        return info;
    }

    async resolveAsset(symbol: string): Promise<AssetInfo> {
        await this.ensureAssets();
        const info = this.assetsBySymbol.get(symbol.toLowerCase());
        if (!info) throw new Error(`Unknown asset: ${symbol}`);
        return info;
    }

    // -----------------------------------------------------------------------
    // Convenience helpers (reduce consumer-side boilerplate)
    // -----------------------------------------------------------------------

    async getTokenDecimals(tokenAddress: Address | string): Promise<number> {
        await this.ensureAssets();
        const key = tokenAddress.toString().toLowerCase();
        const info = this.assetsByToken.get(key);
        if (!info) {
            console.warn(`[compat] Unknown token ${tokenAddress}, falling back to 6 decimals`);
        }
        return info?.decimals ?? 6;
    }

    async formatAmount(tokenAddress: Address | string, rawAmount: bigint): Promise<string> {
        const decimals = await this.getTokenDecimals(tokenAddress);
        return formatUnits(rawAmount, decimals);
    }

    async parseAmount(tokenAddress: Address | string, humanAmount: string): Promise<bigint> {
        const decimals = await this.getTokenDecimals(tokenAddress);
        return parseUnits(humanAmount, decimals);
    }

    async resolveAssetDisplay(tokenAddress: Address | string, _chainId?: number): Promise<{ symbol: string; decimals: number } | null> {
        await this.ensureAssets();
        const key = tokenAddress.toString().toLowerCase();
        const info = this.assetsByToken.get(key);
        if (!info) return null;
        return { symbol: info.symbol, decimals: info.decimals };
    }

    findOpenChannel(tokenAddress: Address | string, chainId?: number): LedgerChannel | null {
        const normalizedToken = tokenAddress.toString().toLowerCase();
        return this._lastChannels.find((ch) => {
            const statusMatch = ch.status === 'open' || ch.status === 'resizing';
            const tokenMatch = ch.token.toLowerCase() === normalizedToken;
            const chainMatch = chainId === undefined || ch.chain_id === chainId;
            return statusMatch && tokenMatch && chainMatch;
        }) ?? null;
    }

    async getAccountInfo(): Promise<AccountInfo> {
        const balances = await this.getBalances();
        return {
            balances,
            channelCount: BigInt(this._lastChannels.length),
        };
    }

    // -----------------------------------------------------------------------
    // On-chain operations (v0.5.3 compat surface)
    // -----------------------------------------------------------------------

    private static readonly MAX_UINT256 = 2n ** 256n - 1n;
    private static readonly DEFAULT_APPROVE_AMOUNT = new Decimal(100000);

    /** Classify raw SDK/wallet errors into typed compat errors. */
    static classifyError(error: unknown): Error {
        const msg = error instanceof Error ? error.message : String(error);
        const lower = msg.toLowerCase();
        if (lower.includes('allowance') && lower.includes('insufficient')) return new AllowanceError(msg);
        if (lower.includes('user rejected') || lower.includes('user denied')) return new UserRejectedError(msg);
        if (lower.includes('insufficient funds') || lower.includes('exceeds balance')) return new InsufficientFundsError(msg);
        if (lower.includes('not initialized') || lower.includes('not connected')) return new NotInitializedError(msg);
        if (lower.includes('ongoing') || lower.includes('state transition')) return new OngoingStateTransitionError(msg);
        return error instanceof Error ? error : new Error(msg);
    }

    private async checkpointWithApproval(symbol: string, chainId: bigint, tokenAddress: string): Promise<any> {
        try {
            return await this.innerClient.checkpoint(symbol);
        } catch (err) {
            const classified = NitroliteClient.classifyError(err);
            if (!(classified instanceof AllowanceError)) throw classified;
            console.log('[compat] Allowance insufficient, requesting token approval…');
            await this.innerClient.approveToken(chainId, tokenAddress, NitroliteClient.DEFAULT_APPROVE_AMOUNT);
            return await this.innerClient.checkpoint(symbol);
        }
    }

    private toHumanAmount(rawAmount: bigint, decimals: number): Decimal {
        return new Decimal(rawAmount.toString()).div(new Decimal(10).pow(decimals));
    }

    async deposit(tokenAddress: Address, amount: bigint): Promise<any> {
        const { symbol, chainId, decimals, tokenAddress: resolvedAddr } = await this.resolveToken(tokenAddress);
        await this.innerClient.setHomeBlockchain(symbol, chainId).catch(() => {});
        const humanAmount = this.toHumanAmount(amount, decimals);
        await this.innerClient.deposit(chainId, symbol, humanAmount);
        return await this.checkpointWithApproval(symbol, chainId, resolvedAddr);
    }

    async depositAndCreateChannel(tokenAddress: Address, amount: bigint, _respParams?: any): Promise<any> {
        return this.deposit(tokenAddress, amount);
    }

    async createChannel(_respParams?: any): Promise<any> {
        console.warn('[compat] createChannel is implicit in v1.0.0 -- use deposit() instead');
    }

    async closeChannel(params?: { tokenAddress?: Address | string } | any): Promise<any> {
        const tokenAddr = params?.tokenAddress?.toString().toLowerCase();

        if (tokenAddr) {
            await this.ensureAssets();
            const info = this.assetsByToken.get(tokenAddr);
            if (!info) throw new Error(`Unknown token address for close: ${params.tokenAddress}`);

            await this.innerClient.closeHomeChannel(info.symbol);
            await this.checkpointWithApproval(info.symbol, info.chainId, info.tokenAddress);
            return;
        }

        const channels = await this.getChannels();
        const openChannels = channels.filter((ch) => ch.status === 'open' || ch.status === 'resizing');

        for (const ch of openChannels) {
            try {
                await this.ensureAssets();
                const info = this.assetsByToken.get(ch.token.toLowerCase());
                const symbol = info?.symbol;
                if (!symbol) continue;

                await this.innerClient.closeHomeChannel(symbol);
                await this.checkpointWithApproval(symbol, info.chainId, info.tokenAddress);
            } catch {
                // channel may already be closing
            }
        }
    }

    async resizeChannel(params: { allocate_amount: bigint; token: Address }): Promise<any> {
        return this.deposit(params.token, params.allocate_amount);
    }

    async challengeChannel(params: { state: any }): Promise<any> {
        return this.innerClient.challenge(params.state);
    }

    async withdrawal(tokenAddress: Address, amount: bigint): Promise<any> {
        const { symbol, chainId, decimals, tokenAddress: resolvedAddr } = await this.resolveToken(tokenAddress);
        await this.innerClient.setHomeBlockchain(symbol, chainId).catch(() => {});
        const humanAmount = this.toHumanAmount(amount, decimals);
        await this.innerClient.withdraw(chainId, symbol, humanAmount);
        return await this.checkpointWithApproval(symbol, chainId, resolvedAddr);
    }

    async getChannelData(_channelId: string): Promise<any> {
        await this.ensureAssets();
        for (const [, info] of this.assetsBySymbol) {
            try {
                const ch = await this.innerClient.getHomeChannel(this.userAddress, info.symbol);
                if (ch.channelId === _channelId) {
                    return {
                        channel: ch,
                        state: await this.innerClient.getLatestState(this.userAddress, info.symbol, false),
                    };
                }
            } catch {
                // no channel for this asset
            }
        }
        throw new Error(`Channel ${_channelId} not found`);
    }

    // -----------------------------------------------------------------------
    // Off-chain queries (for hooks to call directly)
    // -----------------------------------------------------------------------

    private static readonly STATUS_MAP: Record<number, string> = {
        0: 'void',
        1: 'open',
        2: 'challenged',
        3: 'closed',
    };

    async getChannels(): Promise<LedgerChannel[]> {
        try {
            this._lastChannels = await this.getChannelsViaRPC();
        } catch {
            this._lastChannels = await this.getChannelsViaAssetScan();
        }
        return this._lastChannels;
    }

    private async getChannelsViaRPC(): Promise<LedgerChannel[]> {
        const { channels: sdkChannels } = await this.innerClient.getChannels(this.userAddress);
        const result: LedgerChannel[] = [];

        for (const ch of sdkChannels) {
            let userBalance = 0n;

            if (ch.status === 1) {
                try {
                    const state = await this.innerClient.getLatestState(this.userAddress, ch.asset, false);
                    const raw = state.homeLedger?.userBalance;

                    if (raw) {
                        const info = this.assetsBySymbol.get(ch.asset.toLowerCase());
                        const dec = info?.decimals ?? 6;
                        userBalance = BigInt(raw.mul(new Decimal(10).pow(dec)).toFixed(0));
                    }
                } catch {
                    // state not available yet
                }
            }

            result.push({
                channel_id: ch.channelId,
                participant: ch.userWallet ?? this.userAddress,
                status: NitroliteClient.STATUS_MAP[ch.status] ?? String(ch.status),
                token: (ch.tokenAddress ?? '') as string,
                amount: userBalance,
                chain_id: Number(ch.blockchainId ?? 0),
                adjudicator: '',
                challenge: ch.challengeDuration ?? 0,
                nonce: Number(ch.nonce ?? 0),
                version: Number(ch.stateVersion ?? 0),
                created_at: '',
                updated_at: '',
            });
        }

        return result;
    }

    private async getChannelsViaAssetScan(): Promise<LedgerChannel[]> {
        const assets = await this.innerClient.getAssets();
        const channels: LedgerChannel[] = [];

        for (const asset of assets) {
            try {
                const ch = await this.innerClient.getHomeChannel(this.userAddress, asset.symbol);

                if (ch.channelId) {
                    let userBalance = 0n;

                    try {
                        const state = await this.innerClient.getLatestState(this.userAddress, asset.symbol, false);
                        const raw = state.homeLedger?.userBalance;

                        if (raw) {
                            const info = this.assetsBySymbol.get(asset.symbol.toLowerCase());
                            const dec = info?.decimals ?? asset.decimals;
                            userBalance = BigInt(raw.mul(new Decimal(10).pow(dec)).toFixed(0));
                        }
                    } catch {
                        // state not available yet
                    }

                    const chainId = Number(ch.blockchainId ?? asset.tokens?.[0]?.blockchainId ?? 0);
                    channels.push({
                        channel_id: ch.channelId,
                        participant: this.userAddress,
                        status: NitroliteClient.STATUS_MAP[ch.status] ?? String(ch.status),
                        token: (ch.tokenAddress || asset.tokens?.[0]?.address || '') as string,
                        amount: userBalance,
                        chain_id: chainId,
                        adjudicator: '',
                        challenge: ch.challengeDuration ?? 0,
                        nonce: Number(ch.nonce ?? 0),
                        version: Number(ch.stateVersion ?? 0),
                        created_at: '',
                        updated_at: '',
                    });
                }
            } catch {
                // no channel for this asset
            }
        }

        return channels;
    }

    async getBalances(wallet?: Address): Promise<LedgerBalance[]> {
        const balances = await this.innerClient.getBalances(wallet ?? this.userAddress);
        return balances.map((b) => {
            const info = this.assetsBySymbol.get(b.asset.toLowerCase());
            const dec = info?.decimals ?? 6;
            const rawAmount = b.balance.mul(new Decimal(10).pow(dec)).toFixed(0);
            return {
                asset: b.asset,
                amount: rawAmount,
            };
        });
    }

    async getLedgerEntries(wallet?: Address): Promise<LedgerEntry[]> {
        const { transactions } = await this.innerClient.getTransactions(wallet ?? this.userAddress);
        return transactions.map((tx, i) => ({
            id: i,
            account_id: (wallet ?? this.userAddress) as string,
            account_type: 0,
            asset: tx.asset,
            participant: (wallet ?? this.userAddress) as string,
            credit: tx.amount.greaterThanOrEqualTo(0) ? tx.amount.toString() : '0',
            debit: tx.amount.lessThan(0) ? tx.amount.abs().toString() : '0',
            created_at: tx.createdAt?.toISOString?.() ?? '',
        }));
    }

    async getAppSessionsList(wallet?: Address, status?: string): Promise<AppSession[]> {
        const mapSessions = (sessions: any[]) => sessions.map((s) => ({
            app_session_id: s.appSessionId,
            nonce: Number(s.nonce ?? 0),
            participants: s.participants.map((p: any) => p.walletAddress),
            protocol: '',
            quorum: s.quorum,
            status: s.isClosed ? 'closed' : 'open',
            version: Number(s.version ?? 0),
            weights: s.participants.map((p: any) => p.signatureWeight),
            allocations: s.allocations?.map((a: any) => {
                const info = this.assetsBySymbol.get(a.asset?.toLowerCase?.() ?? '');
                const dec = info?.decimals ?? 6;
                const rawAmount = a.amount
                    ? a.amount.mul(new Decimal(10).pow(dec)).toFixed(0)
                    : '0';
                return {
                    participant: a.participant as Address,
                    asset: a.asset,
                    amount: rawAmount,
                };
            }) ?? [],
            sessionData: s.sessionData ?? '',
        }));

        const participant = (wallet ?? this.userAddress).toLowerCase() as Address;
        const normalizedStatus = status?.toLowerCase();
        const effectiveStatus = normalizedStatus && normalizedStatus !== 'any'
            ? normalizedStatus
            : undefined;
        const request = effectiveStatus
            ? { wallet: participant, status: effectiveStatus }
            : { wallet: participant };

        try {
            console.info('[compat] getAppSessionsList request', {
                participant,
                status: effectiveStatus ?? 'any',
                rawStatus: status ?? null,
            });
            const { sessions } = await this.innerClient.getAppSessions(request);
            console.info(`[compat] getAppSessionsList success count=${sessions.length}`);
            this._lastAppSessionsListError = null;
            return mapSessions(sessions);
        } catch (err) {
            if (effectiveStatus) {
                try {
                    console.warn(
                        `[compat] getAppSessionsList retrying without status filter participant=${participant} status=${effectiveStatus}`,
                    );
                    const { sessions } = await this.innerClient.getAppSessions({ wallet: participant });
                    const mapped = mapSessions(sessions);
                    const filtered = mapped.filter((session) => session.status === effectiveStatus);
                    console.info(
                        `[compat] getAppSessionsList success count=${filtered.length} (fallback without status)`,
                    );
                    this._lastAppSessionsListError = null;
                    return filtered;
                } catch {
                    // fall through to the original failure handling
                }
            }

            const message = err instanceof Error ? err.message : String(err);
            this._lastAppSessionsListError = message;
            if (this._lastAppSessionsListErrorLogged !== message) {
                console.warn(
                    `[compat] getAppSessionsList failed participant=${participant} status=${effectiveStatus ?? 'any'} error=${message}`,
                );
                this._lastAppSessionsListErrorLogged = message;
            }
            return [];
        }
    }

    getLastAppSessionsListError(): string | null {
        return this._lastAppSessionsListError;
    }

    async getAssetsList(): Promise<ClearNodeAsset[]> {
        const assets = await this.innerClient.getAssets();
        const result: ClearNodeAsset[] = [];
        for (const asset of assets) {
            for (const token of asset.tokens) {
                result.push({
                    token: token.address as Address,
                    chainId: Number(token.blockchainId),
                    symbol: asset.symbol,
                    decimals: asset.decimals,
                });
            }
        }
        return result;
    }

    async getConfig(): Promise<any> {
        return this.innerClient.getConfig();
    }

    // -----------------------------------------------------------------------
    // Session key operations
    // -----------------------------------------------------------------------

    async signChannelSessionKeyState(state: ChannelSessionKeyStateV1): Promise<Hex> {
        return this.innerClient.signChannelSessionKeyState(state);
    }

    async submitChannelSessionKeyState(state: ChannelSessionKeyStateV1): Promise<void> {
        await this.innerClient.submitChannelSessionKeyState(state);
    }

    async getLastChannelKeyStates(
        userAddress: string,
        sessionKey?: string,
    ): Promise<ChannelSessionKeyStateV1[]> {
        return this.innerClient.getLastChannelKeyStates(userAddress, sessionKey);
    }

    async signSessionKeyState(state: AppSessionKeyStateV1): Promise<Hex> {
        return this.innerClient.signSessionKeyState(state);
    }

    async submitSessionKeyState(state: AppSessionKeyStateV1): Promise<void> {
        await this.innerClient.submitSessionKeyState(state);
    }

    async getLastKeyStates(userAddress: string, sessionKey?: string): Promise<AppSessionKeyStateV1[]> {
        return this.innerClient.getLastKeyStates(userAddress, sessionKey);
    }

    // -----------------------------------------------------------------------
    // App session operations
    // -----------------------------------------------------------------------

    async createAppSession(
        definitionOrParams: RPCAppDefinition | CreateAppSessionRequestParams,
        allocations?: RPCAppSessionAllocation[],
        quorumSigs?: string[],
        opts?: { ownerSig?: string },
    ): Promise<{ appSessionId: string; version: string; status: string }> {
        const def = 'definition' in definitionOrParams ? definitionOrParams.definition : definitionOrParams;
        const allocs = 'definition' in definitionOrParams ? definitionOrParams.allocations : (allocations ?? []);
        const quorumSignatures = 'definition' in definitionOrParams
            ? (definitionOrParams.quorum_sigs ?? [])
            : (quorumSigs ?? []);
        const sessionData = 'definition' in definitionOrParams
            ? (definitionOrParams.session_data ?? JSON.stringify({ allocations: allocs }))
            : JSON.stringify({ allocations: allocs });
        const ownerSig = 'definition' in definitionOrParams
            ? (definitionOrParams.owner_sig ?? opts?.ownerSig)
            : opts?.ownerSig;

        const v1Def: AppDefinitionV1 = {
            applicationId: def.application || '',
            participants: def.participants.map((addr, i) => ({
                walletAddress: addr as Address,
                signatureWeight: def.weights[i] ?? 1,
            })) as AppParticipantV1[],
            quorum: def.quorum,
            nonce: BigInt(def.nonce ?? Date.now()),
        };

        const v1Opts = ownerSig ? { ownerSig } : undefined;
        const result = await this.innerClient.createAppSession(v1Def, sessionData, quorumSignatures, v1Opts);
        return { appSessionId: result.appSessionId, version: result.version, status: result.status };
    }

    async closeAppSession(
        appSessionIdOrParams: string | CloseAppSessionRequestParams,
        allocations?: RPCAppSessionAllocation[],
        quorumSigs: string[] = [],
    ): Promise<{ appSessionId: string }> {
        const appSessionId = typeof appSessionIdOrParams === 'string'
            ? appSessionIdOrParams
            : appSessionIdOrParams.app_session_id;
        const closeAllocations = typeof appSessionIdOrParams === 'string'
            ? (allocations ?? [])
            : appSessionIdOrParams.allocations;
        const closeVersion = typeof appSessionIdOrParams === 'string'
            ? undefined
            : appSessionIdOrParams.version;
        const closeSessionData = typeof appSessionIdOrParams === 'string'
            ? undefined
            : appSessionIdOrParams.session_data;
        const closeQuorumSignatures = typeof appSessionIdOrParams === 'string'
            ? quorumSigs
            : (appSessionIdOrParams.quorum_sigs ?? quorumSigs);

        const { sessions } = await this.innerClient.getAppSessions({ appSessionId });
        if (sessions.length === 0) throw new Error(`App session ${appSessionId} not found`);

        const session = sessions[0];
        const v1Allocations: AppAllocationV1[] = [];
        for (const a of closeAllocations) {
            const decimals = await this.getDecimalsForAsset(a.asset);
            const humanAmount = new Decimal(a.amount).div(new Decimal(10).pow(decimals));
            v1Allocations.push({
                participant: a.participant as Address,
                asset: a.asset,
                amount: humanAmount,
            });
        }

        const appUpdate = {
            appSessionId,
            intent: NitroliteClient.INTENT_MAP['close'],
            version: closeVersion !== undefined ? BigInt(closeVersion) : session.version + 1n,
            allocations: v1Allocations,
            sessionData: closeSessionData ?? '',
        };

        await this.innerClient.submitAppState(appUpdate, closeQuorumSignatures);
        return { appSessionId };
    }

    async getAppDefinition(appSessionId: string): Promise<GetAppDefinitionResponseParams> {
        const def = await this.innerClient.getAppDefinition(appSessionId);
        return {
            protocol: def.applicationId,
            participants: def.participants.map((p) => p.walletAddress),
            weights: def.participants.map((p) => p.signatureWeight),
            quorum: def.quorum,
            challenge: 0,
            nonce: Number(def.nonce ?? 0),
        };
    }

    private static readonly INTENT_MAP: Record<string, number> = {
        operate: 0,
        deposit: 1,
        withdraw: 2,
        close: 3,
    };

    async submitAppState(
        params: SubmitAppStateRequestParams,
    ): Promise<{ appSessionId: string; version: number; status: string }> {
        const isV04 = 'intent' in params;
        const intentStr = isV04 ? (params as SubmitAppStateRequestParamsV04).intent : RPCAppStateIntent.Operate;
        const intentNum = NitroliteClient.INTENT_MAP[intentStr] ?? 0;
        console.info('[compat] submitAppState request', {
            appSessionId: params.app_session_id,
            intent: intentStr,
            allocationCount: params.allocations.length,
            hasQuorumSigs: (params.quorum_sigs?.length ?? 0) > 0,
            quorumSigCount: params.quorum_sigs?.length ?? 0,
        });

        const { sessions } = await this.innerClient.getAppSessions({ appSessionId: params.app_session_id });
        if (sessions.length === 0) throw new Error(`App session ${params.app_session_id} not found`);
        const session = sessions[0];

        const version = isV04
            ? BigInt((params as SubmitAppStateRequestParamsV04).version)
            : session.version + 1n;

        const v1Allocations: AppAllocationV1[] = [];
        for (const a of params.allocations) {
            const decimals = await this.getDecimalsForAsset(a.asset);
            const humanAmount = new Decimal(a.amount).div(new Decimal(10).pow(decimals));
            v1Allocations.push({
                participant: a.participant as Address,
                asset: a.asset,
                amount: humanAmount,
            });
        }

        const appUpdate = {
            appSessionId: params.app_session_id,
            intent: intentNum,
            version,
            allocations: v1Allocations,
            sessionData: params.session_data ?? '',
        };

        if (intentStr === RPCAppStateIntent.Deposit) {
            const userAddress = this.userAddress.toLowerCase();
            const currentByParticipantAndAsset = new Map<string, Decimal>();
            for (const allocation of session.allocations ?? []) {
                currentByParticipantAndAsset.set(
                    `${allocation.participant.toLowerCase()}::${allocation.asset.toLowerCase()}`,
                    allocation.amount,
                );
            }

            type PositiveDelta = { participant: string; asset: string; amount: Decimal };
            const positiveDeltas: PositiveDelta[] = [];
            const negativeDeltas: PositiveDelta[] = [];

            const nextByParticipantAndAsset = new Map<string, PositiveDelta>();
            for (const allocation of v1Allocations) {
                const key = `${allocation.participant.toLowerCase()}::${allocation.asset.toLowerCase()}`;
                nextByParticipantAndAsset.set(key, {
                    participant: allocation.participant.toLowerCase(),
                    asset: allocation.asset,
                    amount: allocation.amount,
                });
            }

            const allKeys = new Set<string>([
                ...currentByParticipantAndAsset.keys(),
                ...nextByParticipantAndAsset.keys(),
            ]);

            for (const key of allKeys) {
                const currentAmount = currentByParticipantAndAsset.get(key) ?? new Decimal(0);
                const nextAllocation = nextByParticipantAndAsset.get(key);
                const nextAmount = nextAllocation?.amount ?? new Decimal(0);
                const [participant, asset] = key.split('::');
                const delta = nextAmount.minus(currentAmount);
                if (delta.greaterThan(0)) {
                    positiveDeltas.push({
                        participant,
                        asset: nextAllocation?.asset ?? asset,
                        amount: delta,
                    });
                } else if (delta.lessThan(0)) {
                    negativeDeltas.push({
                        participant,
                        asset: nextAllocation?.asset ?? asset,
                        amount: delta,
                    });
                }
            }

            if (positiveDeltas.length === 0) {
                throw new Error('Deposit intent requires at least one positive allocation delta');
            }
            if (positiveDeltas.length > 1) {
                throw new Error('Deposit intent currently supports exactly one deposited asset delta');
            }
            if (negativeDeltas.length > 0) {
                throw new Error('Deposit intent cannot decrease existing app-session allocations');
            }

            const [delta] = positiveDeltas;
            if (delta.participant !== userAddress) {
                throw new Error(
                    `Deposit must be submitted by depositor ${delta.participant}; connected wallet is ${userAddress}`,
                );
            }
            console.info('[compat] submitAppState deposit delta', {
                appSessionId: params.app_session_id,
                depositor: delta.participant,
                asset: delta.asset,
                amount: delta.amount.toString(),
                negativeDeltaCount: negativeDeltas.length,
            });

            await this.innerClient.submitAppSessionDeposit(
                appUpdate,
                params.quorum_sigs ?? [],
                delta.asset,
                delta.amount,
            );
        } else {
            await this.innerClient.submitAppState(appUpdate, params.quorum_sigs ?? []);
        }
        console.info('[compat] submitAppState success', {
            appSessionId: params.app_session_id,
            intent: intentStr,
            version: Number(version),
        });
        return {
            appSessionId: params.app_session_id,
            version: Number(version),
            status: intentNum === 3 ? 'closed' : 'open',
        };
    }

    // -----------------------------------------------------------------------
    // Transfer
    // -----------------------------------------------------------------------

    /** Transfers are executed sequentially per allocation and are not atomic; a mid-loop failure leaves prior transfers committed. */
    async transfer(destination: Address, allocations: TransferAllocation[]): Promise<void> {
        for (const alloc of allocations) {
            const decimals = await this.getDecimalsForAsset(alloc.asset);
            const humanAmount = new Decimal(alloc.amount).div(new Decimal(10).pow(decimals));
            await this.innerClient.transfer(destination, alloc.asset, humanAmount);
        }
    }

    // -----------------------------------------------------------------------
    // Lifecycle
    // -----------------------------------------------------------------------

    async close(): Promise<void> {
        await this.innerClient.close();
    }

    async ping(): Promise<void> {
        await this.innerClient.ping();
    }

    waitForClose(): Promise<void> {
        return this.innerClient.waitForClose();
    }

    // -----------------------------------------------------------------------
    // Acknowledge
    // -----------------------------------------------------------------------

    async acknowledge(tokenAddress: Address): Promise<void> {
        const { symbol } = await this.resolveToken(tokenAddress);
        await this.innerClient.acknowledge(symbol);
    }

    // -----------------------------------------------------------------------
    // Token allowance
    // -----------------------------------------------------------------------

    async checkTokenAllowance(chainId: number, tokenAddress: Address): Promise<bigint> {
        return this.innerClient.checkTokenAllowance(
            BigInt(chainId),
            tokenAddress,
            this.userAddress,
        );
    }

    // -----------------------------------------------------------------------
    // Additional queries
    // -----------------------------------------------------------------------

    async getBlockchains(): Promise<core.Blockchain[]> {
        return this.ensureBlockchains();
    }

    private async ensureBlockchains(): Promise<core.Blockchain[]> {
        if (!this._blockchains) {
            this._blockchains = await this.innerClient.getBlockchains();
        }
        return this._blockchains;
    }

    async getActionAllowances(wallet?: Address): Promise<core.ActionAllowance[]> {
        return this.innerClient.getActionAllowances(wallet ?? this.userAddress);
    }

    async getEscrowChannel(escrowChannelId: string): Promise<core.Channel> {
        return this.innerClient.getEscrowChannel(escrowChannelId);
    }

    // -----------------------------------------------------------------------
    // App registry
    // -----------------------------------------------------------------------

    async getApps(options?: {
        appId?: string;
        ownerWallet?: string;
        page?: number;
        pageSize?: number;
    }): Promise<{ apps: AppInfoV1[]; metadata: core.PaginationMetadata }> {
        return this.innerClient.getApps(options);
    }

    async registerApp(
        appID: string,
        metadata: string,
        creationApprovalNotRequired: boolean,
    ): Promise<void> {
        await this.innerClient.registerApp(appID, metadata, creationApprovalNotRequired);
    }

    // -----------------------------------------------------------------------
    // Security token locking
    // -----------------------------------------------------------------------

    async lockSecurityTokens(
        targetWallet: Address,
        chainId: number,
        amount: bigint,
    ): Promise<string> {
        if (amount <= 0n) throw new Error('amount must be positive');
        const decimals = await this.getLockingTokenDecimals(chainId);
        const humanAmount = this.toHumanAmount(amount, decimals);
        return this.innerClient.escrowSecurityTokens(
            targetWallet,
            BigInt(chainId),
            humanAmount,
        );
    }

    async initiateSecurityTokensWithdrawal(chainId: number): Promise<string> {
        return this.innerClient.initiateSecurityTokensWithdrawal(BigInt(chainId));
    }

    async cancelSecurityTokensWithdrawal(chainId: number): Promise<string> {
        return this.innerClient.cancelSecurityTokensWithdrawal(BigInt(chainId));
    }

    async withdrawSecurityTokens(
        chainId: number,
        destination: Address,
    ): Promise<string> {
        return this.innerClient.withdrawSecurityTokens(BigInt(chainId), destination);
    }

    async approveSecurityToken(chainId: number, amount: bigint): Promise<string> {
        if (amount <= 0n) throw new Error('amount must be positive');
        const decimals = await this.getLockingTokenDecimals(chainId);
        const humanAmount = this.toHumanAmount(amount, decimals);
        return this.innerClient.approveSecurityToken(BigInt(chainId), humanAmount);
    }

    async getLockedBalance(chainId: number, wallet?: Address): Promise<bigint> {
        const balance = await this.innerClient.getLockedBalance(
            BigInt(chainId),
            wallet ?? this.userAddress,
        );
        const decimals = await this.getLockingTokenDecimals(chainId);
        return BigInt(balance.mul(new Decimal(10).pow(decimals)).toFixed(0));
    }

    private static readonly LOCKING_ASSET_ABI = [
        { type: 'function', name: 'asset', inputs: [], outputs: [{ type: 'address' }], stateMutability: 'view' },
    ] as const;

    private static readonly ERC20_DECIMALS_ABI = [
        { type: 'function', name: 'decimals', inputs: [], outputs: [{ type: 'uint8' }], stateMutability: 'view' },
    ] as const;

    private async getLockingTokenDecimals(chainId: number): Promise<number> {
        const cached = this._lockingTokenDecimals.get(chainId);
        if (cached !== undefined) return cached;

        const blockchains = await this.ensureBlockchains();
        const chain = blockchains.find((b) => b.id === BigInt(chainId));
        if (!chain?.lockingContractAddress) {
            throw new Error(`No locking contract configured for chain ${chainId}`);
        }

        const rpcUrl = this._blockchainRPCs[chainId];
        if (!rpcUrl) {
            throw new Error(
                `No RPC URL configured for chain ${chainId}. ` +
                'Pass blockchainRPCs in NitroliteClientConfig to use locking methods.',
            );
        }

        const publicClient = createPublicClient({ transport: http(rpcUrl) });

        const tokenAddress = await publicClient.readContract({
            address: chain.lockingContractAddress,
            abi: NitroliteClient.LOCKING_ASSET_ABI,
            functionName: 'asset',
        }) as Address;

        const decimals = await publicClient.readContract({
            address: tokenAddress,
            abi: NitroliteClient.ERC20_DECIMALS_ABI,
            functionName: 'decimals',
        }) as number;

        this._lockingTokenDecimals.set(chainId, decimals);
        return decimals;
    }
}
