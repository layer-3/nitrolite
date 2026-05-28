// =============================================================================
// @yellow-org/sdk-compat barrel export
//
// Curated migration-layer exports for apps moving off '@layer-3/nitrolite'
// (v0.5.3) onto the v1 runtime (@yellow-org/sdk).
//
// This barrel preserves selected app-facing surfaces. It is not full package
// parity with the published v0.5.3 package, and it is not a promise that every
// legacy helper here is a runtime-faithful one-to-one v1 protocol mapping.
// =============================================================================

// --- Client facade ---
export { NitroliteClient, type NitroliteClientConfig } from './client.js';

// --- Signers ---
export { WalletStateSigner, createECDSAMessageSigner } from './signers.js';

// --- Auth helpers ---
export {
    createAuthRequestMessage,
    createAuthVerifyMessage,
    createAuthVerifyMessageWithJWT,
    createEIP712AuthMessageSigner,
    type AuthRequestParams,
} from './auth.js';

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
} from './app-signing.js';

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
} from './rpc.js';

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
} from './types.js';

// --- Nitronode response types (used by consuming apps' stores) ---
export type {
    AccountInfo,
    LedgerChannel,
    LedgerBalance,
    LedgerEntry,
    AppSession,
    ClearNodeAsset,
} from './types.js';

// --- Errors ---
export {
    CompatError,
    AllowanceError,
    UserRejectedError,
    InsufficientFundsError,
    NotInitializedError,
    OngoingStateTransitionError,
    getUserFacingMessage,
} from './errors.js';

// --- Events ---
export { EventPoller, type EventPollerCallbacks } from './events.js';

// --- Config ---
export { buildClientOptions, blockchainRPCsFromEnv, type CompatClientConfig } from './config.js';

// NOTE: SDK classes (Client, ChannelDefaultSigner, etc.) are intentionally NOT
// re-exported here. Barrel re-exports from '@yellow-org/sdk' trigger eager
// module evaluation of the full SDK, which has side effects that throw during
// SSR / module-load time. Apps needing those classes should import directly
// from '@yellow-org/sdk'.
