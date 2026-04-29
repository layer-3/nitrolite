import { Address, Hex } from 'viem';

// =============================================================================
// RPCMethod enum -- mirrors v0.5.3 method names used in the appstore
// =============================================================================

export enum RPCMethod {
    Ping = 'ping',
    GetConfig = 'get_config',
    GetChannels = 'get_channels',
    ChannelsUpdate = 'channels_update',
    ChannelUpdate = 'channel_update',
    BalanceUpdate = 'balance_update',
    GetAssets = 'get_assets',
    Assets = 'assets',
    GetLedgerBalances = 'get_ledger_balances',
    GetLedgerEntries = 'get_ledger_entries',
    GetAppSessions = 'get_app_sessions',
    CreateChannel = 'create_channel',
    CloseChannel = 'close_channel',
    ResizeChannel = 'resize_channel',
    Transfer = 'transfer',
    CreateAppSession = 'create_app_session',
    CloseAppSession = 'close_app_session',
    SubmitAppState = 'submit_app_state',
    GetAppDefinition = 'get_app_definition',
    AuthRequest = 'auth_request',
    AuthChallenge = 'auth_challenge',
    AuthVerify = 'auth_verify',
    Error = 'error',
    GetLedgerTransactions = 'get_ledger_transactions',
    TransferNotification = 'tr',
}

export enum RPCChannelStatus {
    Open = 'open',
    Closed = 'closed',
    Resizing = 'resizing',
    Challenged = 'challenged',
}

export enum RPCProtocolVersion {
    NitroRPC_0_2 = 'NitroRPC/0.2',
    NitroRPC_0_4 = 'NitroRPC/0.4',
}

export enum RPCAppStateIntent {
    Operate = 'operate',
    Deposit = 'deposit',
    Withdraw = 'withdraw',
}

export enum RPCTxType {
    Transfer = 'transfer',
    Deposit = 'deposit',
    Withdrawal = 'withdrawal',
    AppDeposit = 'app_deposit',
    AppWithdrawal = 'app_withdrawal',
    EscrowLock = 'escrow_lock',
    EscrowUnlock = 'escrow_unlock',
}

// =============================================================================
// Wire types -- shapes the appstore expects from v0.5.3 SDK
// =============================================================================

export type NitroliteRPCRequest = [number, string, any, number];
export type MessageSignerPayload = Uint8Array | NitroliteRPCRequest;
export type MessageSigner = (payload: MessageSignerPayload) => Promise<string>;
export type RequestID = number;

export interface NitroliteRPCMessage {
    req: NitroliteRPCRequest;
    sig: string;
}

export interface RPCResponse {
    requestId: number;
    method: string;
    params: any;
}

export interface RPCBalance {
    asset: string;
    amount: string;
}

export interface RPCAsset {
    token: Address;
    chainId: number;
    symbol: string;
    decimals: number;
}

export interface RPCChannelUpdate {
    channelId: string;
    participant: string;
    status: string;
    token: string;
    amount: bigint;
    chainId: number;
    adjudicator: string;
    challenge: number;
    nonce: number;
    version: number;
    createdAt: string;
    updatedAt: string;
}

export interface RPCLedgerEntry {
    id: number;
    account_id: string;
    account_type: number;
    asset: string;
    participant: string;
    credit: string;
    debit: string;
    created_at: string;
}

export type AccountID = string;

// =============================================================================
// Channel operation types
// =============================================================================

export interface ContractAddresses {
    custody: Address | string;
    adjudicator: Address | string;
}

export interface Allocation {
    destination: Address | string;
    token: Address | string;
    amount: bigint;
}

export interface FinalState {
    intent: number;
    channelId: string;
    data: Hex;
    version: bigint;
    allocations: [Allocation, Allocation];
    serverSignature: string;
}

export interface ChannelData {
    lastValidState: any;
    stateData: Hex;
}

export interface CreateChannelResponseParams {
    channel: any;
    state: any;
    serverSignature: string;
}

export interface CloseChannelResponseParams {
    channelId: string;
    state: any;
    serverSignature: string;
}

export interface ResizeChannelRequestParams {
    channel_id: string;
    allocate_amount: bigint;
    resize_amount: bigint;
    funds_destination: Address | string;
}

export interface TransferAllocation {
    asset: string;
    amount: string;
}

// =============================================================================
// App Session types
// =============================================================================

export interface RPCAppDefinition {
    application: string;
    protocol: RPCProtocolVersion;
    participants: Hex[];
    weights: number[];
    quorum: number;
    challenge: number;
    nonce?: number;
}

export interface RPCAppSessionAllocation {
    asset: string;
    amount: string;
    participant: Address;
}

export interface CloseAppSessionRequestParams {
    app_session_id: string;
    allocations: RPCAppSessionAllocation[];
    version?: number;
    session_data?: string;
    quorum_sigs: string[];
}

export interface CreateAppSessionRequestParams {
    definition: RPCAppDefinition;
    allocations: RPCAppSessionAllocation[];
    session_data?: string;
    quorum_sigs: string[];
    owner_sig?: string;
}

export interface SubmitAppStateRequestParamsV02 {
    app_session_id: Hex;
    allocations: RPCAppSessionAllocation[];
    session_data?: string;
    quorum_sigs?: string[];
}

export interface SubmitAppStateRequestParamsV04 {
    app_session_id: Hex;
    intent: RPCAppStateIntent;
    version: number;
    allocations: RPCAppSessionAllocation[];
    session_data?: string;
    quorum_sigs?: string[];
}

export type SubmitAppStateRequestParams = SubmitAppStateRequestParamsV02 | SubmitAppStateRequestParamsV04;

export interface GetAppDefinitionResponseParams {
    protocol: string;
    participants: Address[];
    weights: number[];
    quorum: number;
    challenge: number;
    nonce: number;
}

export interface RPCAppSession {
    appSessionId: Hex;
    application: string;
    status: RPCChannelStatus;
    participants: Address[];
    protocol: RPCProtocolVersion;
    challenge: number;
    weights: number[];
    quorum: number;
    version: number;
    nonce: number;
    createdAt: Date;
    updatedAt: Date;
    sessionData?: string;
}

// =============================================================================
// Channel / State types used by on-chain operations
// =============================================================================

export interface Channel {
    channelId: string;
    participants: Address[];
    adjudicator: Address;
    challenge: number;
    nonce: bigint;
    version: bigint;
}

export interface State {
    channelId: string;
    version: bigint;
    data: Hex;
    allocations: Allocation[];
}

export interface AppLogic<T = bigint> {
    encode(data: T): Hex;
    decode(encoded: Hex): T;
    validateTransition(channel: Channel, prevState: T, nextState: T): boolean;
    provideProofs(channel: Channel, state: T, previousStates: State[]): State[];
    isFinal(state: T): boolean;
    getAdjudicatorAddress(): Address;
    getAdjudicatorType(): string;
}

// =============================================================================
// Nitronode response types (previously in appstore/src/store/types.ts)
// =============================================================================

export interface AccountInfo {
    /** Per-asset balances in raw units (avoids incorrect cross-decimal summing). */
    balances: LedgerBalance[];
    channelCount: bigint;
}

export interface LedgerBalance {
    asset: string;
    amount: string;
}

export interface LedgerChannel {
    channel_id: string;
    participant: string;
    status: string;
    token: string;
    amount: bigint;
    chain_id: number;
    adjudicator: string;
    challenge: number;
    nonce: number;
    version: number;
    created_at: string;
    updated_at: string;
}

export interface AppSession {
    app_session_id: string;
    nonce: number;
    participants: string[];
    protocol: string;
    quorum: number;
    status: string;
    version: number;
    weights: number[];
    allocations?: RPCAppSessionAllocation[];
    sessionData?: string;
}

export interface LedgerEntry {
    id: number;
    account_id: string;
    account_type: number;
    asset: string;
    participant: string;
    credit: string;
    debit: string;
    created_at: string;
}

export interface ClearNodeAsset {
    token: Address;
    chainId: number;
    symbol: string;
    decimals: number;
}

// =============================================================================
// v0.5.3 transaction / notification types (used by beatwav)
// =============================================================================

export type LedgerAccountType = Address | `0x${string}`;

export interface RPCTransaction {
    id: number;
    txType: RPCTxType;
    fromAccount: LedgerAccountType;
    fromAccountTag?: string;
    toAccount: LedgerAccountType;
    toAccountTag?: string;
    asset: string;
    amount: string;
    createdAt: Date;
}

export interface AuthChallengeResponse {
    method: RPCMethod.AuthChallenge;
    params: {
        challengeMessage: string;
    };
}

export interface TransferNotificationResponseParams {
    transactions: RPCTransaction[];
}

// =============================================================================
// EIP-712 Auth Types (v0.5.3 compatible)
// =============================================================================

export const EIP712AuthTypes = {
    Policy: [
        { name: 'challenge', type: 'string' },
        { name: 'scope', type: 'string' },
        { name: 'wallet', type: 'address' },
        { name: 'session_key', type: 'address' },
        { name: 'expires_at', type: 'uint64' },
        { name: 'allowances', type: 'Allowance[]' },
    ],
    Allowance: [
        { name: 'asset', type: 'string' },
        { name: 'amount', type: 'string' },
    ],
} as const;
