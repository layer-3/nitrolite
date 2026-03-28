// =============================================================================
// @yellow-org/sdk-compat barrel export
//
// Re-exports everything apps previously imported from '@layer-3/nitrolite'
// (v0.5.3) but backed by the v1.0.0+ SDK (@yellow-org/sdk).
// =============================================================================

// --- Client facade ---
export { NitroliteClient, type NitroliteClientConfig } from './client';

// --- Signers ---
export { WalletStateSigner, createECDSAMessageSigner } from './signers';

// --- Auth helpers ---
export {
    createAuthRequestMessage,
    createAuthVerifyMessage,
    createAuthVerifyMessageWithJWT,
    createEIP712AuthMessageSigner,
    generateRequestId,
    getCurrentTimestamp,
    parseAuthChallengeResponse,
    parseAuthVerifyResponse,
    type AuthRequestParams,
} from './auth';

// --- App session signing helpers ---
export {
    packCreateAppSessionHash,
    packSubmitAppStateHash,
    toWalletQuorumSignature,
    toSessionKeyQuorumSignature,
    type CreateAppSessionHashParticipant,
    type CreateAppSessionHashParams,
    type SubmitAppStateHashAllocation,
    type SubmitAppStateHashParams,
} from './app-signing';

// --- RPC helpers ---
export {
    parseAnyRPCResponse,
    NitroliteRPC,
    createGetChannelsMessage,
    parseGetChannelsResponse,
    createGetLedgerBalancesMessage,
    parseGetLedgerBalancesResponse,
    parseGetLedgerEntriesResponse,
    parseGetAppSessionsResponse,
    createGetAppSessionsMessage,
    createTransferMessage,
    createAppSessionMessage,
    parseCreateAppSessionResponse,
    createCloseAppSessionMessage,
    parseCloseAppSessionResponse,
    createSubmitAppStateMessage,
    parseSubmitAppStateResponse,
    createGetAppDefinitionMessage,
    parseGetAppDefinitionResponse,
    createCreateChannelMessage,
    parseCreateChannelResponse,
    createCloseChannelMessage,
    parseCloseChannelResponse,
    createResizeChannelMessage,
    parseResizeChannelResponse,
    createPingMessage,
    convertRPCToClientChannel,
    convertRPCToClientState,
    parseChannelUpdateResponse,
    parseTransferResponse,
    parseGetLedgerTransactionsResponse,
    createGetLedgerTransactionsMessageV2,
    createGetAppSessionsMessageV2,
    createCleanupSessionKeyCacheMessage,
    createRevokeSessionKeyMessage,
} from './rpc';

// --- Types ---
export {
    RPCMethod,
    RPCChannelStatus,
    RPCProtocolVersion,
    RPCAppStateIntent,
    EIP712AuthTypes,
    type MessageSigner,
    type NitroliteRPCMessage,
    type RPCResponse,
    type RPCBalance,
    type RPCAsset,
    type RPCChannelUpdate,
    type RPCLedgerEntry,
    type AccountID,
    type RPCAppDefinition,
    type RPCAppSessionAllocation,
    type RPCAppSession,
    type CloseAppSessionRequestParams,
    type CreateAppSessionRequestParams,
    type SubmitAppStateRequestParams,
    type SubmitAppStateRequestParamsV02,
    type SubmitAppStateRequestParamsV04,
    type GetAppDefinitionResponseParams,
    type ContractAddresses,
    type Allocation,
    type FinalState,
    type ChannelData,
    type CreateChannelResponseParams,
    type CloseChannelResponseParams,
    type ResizeChannelRequestParams,
    type TransferAllocation,
    type Channel,
    type State,
    type AppLogic,
    RPCTxType,
    type RequestID,
    type RPCTransaction,
    type AuthChallengeResponse,
    type TransferNotificationResponseParams,
    type LedgerAccountType,
    type RPCData,
    type TransferRequestParams,
    type ResizeChannelParams,
    type RPCChannelOperation,
    type GetLedgerTransactionsFilters,
    type MessageSignerPayload,
    type NitroliteRPCRequest,
} from './types';

// --- Clearnode response types (used by consuming apps' stores) ---
export type {
    AccountInfo,
    LedgerChannel,
    LedgerBalance,
    LedgerEntry,
    AppSession,
    ClearNodeAsset,
} from './types';

// --- Errors ---
export {
    CompatError,
    AllowanceError,
    UserRejectedError,
    InsufficientFundsError,
    NotInitializedError,
    OngoingStateTransitionError,
    getUserFacingMessage,
} from './errors';

// --- Events ---
export { EventPoller, type EventPollerCallbacks } from './events';

// --- Config ---
export { buildClientOptions, blockchainRPCsFromEnv, type CompatClientConfig } from './config';

// --- Re-exported from @yellow-org/sdk for integration-test compatibility ---
// Type-only re-export: safe for SSR (erased at compile time, no eager module evaluation).
export type { StateSigner } from '@yellow-org/sdk';

// NOTE: SDK classes (Client, ChannelDefaultSigner, etc.) are intentionally NOT
// re-exported here. Barrel re-exports from '@yellow-org/sdk' trigger eager
// module evaluation of the full SDK, which has side effects that throw during
// SSR / module-load time. Apps needing those classes should import directly
// from '@yellow-org/sdk'.
