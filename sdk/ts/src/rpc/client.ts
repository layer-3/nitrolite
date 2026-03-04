/**
 * High-level client for interacting with the Nitrolite Node RPC server
 * This file implements the V1 API client with versioned request/response types
 */

import { Dialer } from './dialer';
import { Message, newRequest, newPayload, messageError, translatePayload } from './message';
import * as Methods from './methods';
import * as API from './api';

/**
 * RPCClient provides a high-level interface for interacting with the Nitrolite Node V1 RPC API.
 * It wraps a Dialer to provide convenient methods for all V1 RPC operations.
 */
export class RPCClient {
  private dialer: Dialer;

  constructor(dialer: Dialer) {
    this.dialer = dialer;
  }

  /**
   * Start establishes a connection to the RPC server.
   * This is a convenience method that wraps the dialer's dial method.
   */
  async start(url: string, handleClosure: (err?: Error) => void): Promise<void> {
    return this.dialer.dial(url, handleClosure);
  }

  /**
   * Call sends an RPC request with the specified method and parameters.
   * Returns the response payload or throws an error if the RPC call fails.
   */
  private async call<TReq, TResp>(
    method: string,
    req: TReq,
    signal?: AbortSignal
  ): Promise<TResp> {
    // Generate unique request ID
    const requestId = Math.floor(Math.random() * Number.MAX_SAFE_INTEGER);

    // Create payload from request
    const payload = newPayload(req);
    const message = newRequest(requestId, method, payload);

    // Send request and await response
    const response = await this.dialer.call(message, signal);

    // Check if response contains an error
    const err = messageError(response);
    if (err) {
      throw new Error(`rpc returned error: ${err.message}`);
    }

    // Translate response payload to typed object
    return translatePayload<TResp>(response.payload);
  }

  // ============================================================================
  // Channels Group - V1 API Methods
  // ============================================================================

  async channelsV1GetHomeChannel(
    req: API.ChannelsV1GetHomeChannelRequest,
    signal?: AbortSignal
  ): Promise<API.ChannelsV1GetHomeChannelResponse> {
    return this.call(Methods.ChannelsV1GetHomeChannelMethod, req, signal);
  }

  async channelsV1GetEscrowChannel(
    req: API.ChannelsV1GetEscrowChannelRequest,
    signal?: AbortSignal
  ): Promise<API.ChannelsV1GetEscrowChannelResponse> {
    return this.call(Methods.ChannelsV1GetEscrowChannelMethod, req, signal);
  }

  async channelsV1GetChannels(
    req: API.ChannelsV1GetChannelsRequest,
    signal?: AbortSignal
  ): Promise<API.ChannelsV1GetChannelsResponse> {
    return this.call(Methods.ChannelsV1GetChannelsMethod, req, signal);
  }

  async channelsV1GetLatestState(
    req: API.ChannelsV1GetLatestStateRequest,
    signal?: AbortSignal
  ): Promise<API.ChannelsV1GetLatestStateResponse> {
    return this.call(Methods.ChannelsV1GetLatestStateMethod, req, signal);
  }

  async channelsV1GetStates(
    req: API.ChannelsV1GetStatesRequest,
    signal?: AbortSignal
  ): Promise<API.ChannelsV1GetStatesResponse> {
    return this.call(Methods.ChannelsV1GetStatesMethod, req, signal);
  }

  async channelsV1RequestCreation(
    req: API.ChannelsV1RequestCreationRequest,
    signal?: AbortSignal
  ): Promise<API.ChannelsV1RequestCreationResponse> {
    return this.call(Methods.ChannelsV1RequestCreationMethod, req, signal);
  }

  async channelsV1SubmitState(
    req: API.ChannelsV1SubmitStateRequest,
    signal?: AbortSignal
  ): Promise<API.ChannelsV1SubmitStateResponse> {
    return this.call(Methods.ChannelsV1SubmitStateMethod, req, signal);
  }

  // ============================================================================
  // Channel Session Key State - V1 API Methods
  // ============================================================================

  async channelsV1SubmitSessionKeyState(
    req: API.ChannelsV1SubmitSessionKeyStateRequest,
    signal?: AbortSignal
  ): Promise<API.ChannelsV1SubmitSessionKeyStateResponse> {
    return this.call(Methods.ChannelsV1SubmitSessionKeyStateMethod, req, signal);
  }

  async channelsV1GetLastKeyStates(
    req: API.ChannelsV1GetLastKeyStatesRequest,
    signal?: AbortSignal
  ): Promise<API.ChannelsV1GetLastKeyStatesResponse> {
    return this.call(Methods.ChannelsV1GetLastKeyStatesMethod, req, signal);
  }

  // ============================================================================
  // App Sessions Group - V1 API Methods
  // ============================================================================

  async appSessionsV1SubmitDepositState(
    req: API.AppSessionsV1SubmitDepositStateRequest,
    signal?: AbortSignal
  ): Promise<API.AppSessionsV1SubmitDepositStateResponse> {
    return this.call(Methods.AppSessionsV1SubmitDepositStateMethod, req, signal);
  }

  async appSessionsV1SubmitAppState(
    req: API.AppSessionsV1SubmitAppStateRequest,
    signal?: AbortSignal
  ): Promise<API.AppSessionsV1SubmitAppStateResponse> {
    return this.call(Methods.AppSessionsV1SubmitAppStateMethod, req, signal);
  }

  async appSessionsV1RebalanceAppSessions(
    req: API.AppSessionsV1RebalanceAppSessionsRequest,
    signal?: AbortSignal
  ): Promise<API.AppSessionsV1RebalanceAppSessionsResponse> {
    return this.call(Methods.AppSessionsV1RebalanceAppSessionsMethod, req, signal);
  }

  async appSessionsV1GetAppDefinition(
    req: API.AppSessionsV1GetAppDefinitionRequest,
    signal?: AbortSignal
  ): Promise<API.AppSessionsV1GetAppDefinitionResponse> {
    return this.call(Methods.AppSessionsV1GetAppDefinitionMethod, req, signal);
  }

  async appSessionsV1GetAppSessions(
    req: API.AppSessionsV1GetAppSessionsRequest,
    signal?: AbortSignal
  ): Promise<API.AppSessionsV1GetAppSessionsResponse> {
    return this.call(Methods.AppSessionsV1GetAppSessionsMethod, req, signal);
  }

  async appSessionsV1CreateAppSession(
    req: API.AppSessionsV1CreateAppSessionRequest,
    signal?: AbortSignal
  ): Promise<API.AppSessionsV1CreateAppSessionResponse> {
    return this.call(Methods.AppSessionsV1CreateAppSessionMethod, req, signal);
  }

  async appSessionsV1CloseAppSession(
    req: API.AppSessionsV1CloseAppSessionRequest,
    signal?: AbortSignal
  ): Promise<API.AppSessionsV1CloseAppSessionResponse> {
    return this.call(Methods.AppSessionsV1CloseAppSessionMethod, req, signal);
  }

  // ============================================================================
  // App Session Key State - V1 API Methods
  // ============================================================================

  async appSessionsV1SubmitSessionKeyState(
    req: API.AppSessionsV1SubmitSessionKeyStateRequest,
    signal?: AbortSignal
  ): Promise<API.AppSessionsV1SubmitSessionKeyStateResponse> {
    return this.call(Methods.AppSessionsV1SubmitSessionKeyStateMethod, req, signal);
  }

  async appSessionsV1GetLastKeyStates(
    req: API.AppSessionsV1GetLastKeyStatesRequest,
    signal?: AbortSignal
  ): Promise<API.AppSessionsV1GetLastKeyStatesResponse> {
    return this.call(Methods.AppSessionsV1GetLastKeyStatesMethod, req, signal);
  }

  // ============================================================================
  // Apps Group - V1 API Methods
  // ============================================================================

  async appsV1GetApps(
    req: API.AppsV1GetAppsRequest,
    signal?: AbortSignal
  ): Promise<API.AppsV1GetAppsResponse> {
    return this.call(Methods.AppsV1GetAppsMethod, req, signal);
  }

  async appsV1SubmitAppVersion(
    req: API.AppsV1SubmitAppVersionRequest,
    signal?: AbortSignal
  ): Promise<API.AppsV1SubmitAppVersionResponse> {
    return this.call(Methods.AppsV1SubmitAppVersionMethod, req, signal);
  }

  // ============================================================================
  // User Group - V1 API Methods
  // ============================================================================

  async userV1GetBalances(
    req: API.UserV1GetBalancesRequest,
    signal?: AbortSignal
  ): Promise<API.UserV1GetBalancesResponse> {
    return this.call(Methods.UserV1GetBalancesMethod, req, signal);
  }

  async userV1GetTransactions(
    req: API.UserV1GetTransactionsRequest,
    signal?: AbortSignal
  ): Promise<API.UserV1GetTransactionsResponse> {
    return this.call(Methods.UserV1GetTransactionsMethod, req, signal);
  }

  async userV1GetActionAllowances(
    req: API.UserV1GetActionAllowancesRequest,
    signal?: AbortSignal
  ): Promise<API.UserV1GetActionAllowancesResponse> {
    return this.call(Methods.UserV1GetActionAllowancesMethod, req, signal);
  }

  // ============================================================================
  // Node Group - V1 API Methods
  // ============================================================================

  async nodeV1Ping(signal?: AbortSignal): Promise<void> {
    await this.call(Methods.NodeV1PingMethod, {}, signal);
  }

  async nodeV1GetConfig(signal?: AbortSignal): Promise<API.NodeV1GetConfigResponse> {
    return this.call(Methods.NodeV1GetConfigMethod, {}, signal);
  }

  async nodeV1GetAssets(
    req: API.NodeV1GetAssetsRequest,
    signal?: AbortSignal
  ): Promise<API.NodeV1GetAssetsResponse> {
    return this.call(Methods.NodeV1GetAssetsMethod, req, signal);
  }

  // ============================================================================
  // Utility Methods
  // ============================================================================

  /**
   * Close closes the connection gracefully.
   */
  async close(): Promise<void> {
    return this.dialer.close();
  }

  /**
   * IsConnected returns true if the client has an active connection.
   */
  isConnected(): boolean {
    return this.dialer.isConnected();
  }

  /**
   * EventChannel returns an async iterable for receiving unsolicited events.
   */
  eventChannel(): AsyncIterable<Message> {
    return this.dialer.eventChannel();
  }
}

/**
 * NewRPCClient creates a new V1 RPC client using the provided dialer.
 */
export function newRPCClient(dialer: Dialer): RPCClient {
  return new RPCClient(dialer);
}
