// =============================================================================
// @layer-3/nitrolite-compat barrel export
//
// Re-exports everything apps previously imported from '@layer-3/nitrolite'
// (v0.5.3) but backed by the v1.0.0 SDK.
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
    getUserFacingMessage,
} from './errors';

// --- Events ---
export { EventPoller, type EventPollerCallbacks } from './events';

// --- Config ---
export { buildClientOptions, blockchainRPCsFromEnv, type CompatClientConfig } from './config';

// NOTE: SDK classes (Client, ChannelDefaultSigner, etc.) are intentionally NOT
// re-exported here. Barrel re-exports from '@layer-3/nitrolite' trigger eager
// module evaluation of the full SDK, which has side effects that throw during
// SSR / module-load time. Apps needing those classes should import directly
// from '@layer-3/nitrolite'.
