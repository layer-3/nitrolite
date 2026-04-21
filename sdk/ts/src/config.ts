/**
 * Configuration for the Nitrolite SDK Client
 */

/**
 * Config holds the configuration options for the Nitrolite client.
 */
export interface Config {
  /** WebSocket URL of the Nitrolite server */
  url: string;

  /** Maximum time to wait for initial connection (in milliseconds) */
  handshakeTimeout?: number;

  /** Called when connection errors occur */
  errorHandler?: (error: Error) => void;

  /** Maps blockchain IDs to their RPC endpoints */
  blockchainRPCs?: Map<bigint, string>;

  pingInterval?: number;

  /**
   * Advisory origin tag sent to the clearnode as the "app_id" WebSocket query
   * parameter. The clearnode stamps this value on records produced by requests
   * from this client. Empty / undefined means no tag is sent.
   */
  applicationID?: string;
}

/**
 * Option is a functional option for configuring the Client.
 */
export type Option = (config: Config) => void;

/**
 * Default error handler logs errors to console.
 */
function defaultErrorHandler(err: Error): void {
  if (err) {
    console.error('[nitrolite]', 'connection error:', err);
  }
}

/**
 * DefaultConfig returns the default configuration with sensible defaults.
 */
export const DefaultConfig: Partial<Config> = {
  handshakeTimeout: 5000, // 5 seconds
  errorHandler: defaultErrorHandler,
  blockchainRPCs: new Map(),
};

/**
 * WithHandshakeTimeout sets the maximum time to wait for initial connection.
 */
export function withHandshakeTimeout(timeout: number): Option {
  return (config: Config) => {
    config.handshakeTimeout = timeout;
  };
}

/**
 * WithErrorHandler sets a custom error handler for connection errors.
 * The handler is called when the connection encounters an error or is closed.
 */
export function withErrorHandler(handler: (error: Error) => void): Option {
  return (config: Config) => {
    config.errorHandler = handler;
  };
}

/**
 * The URL query parameter name used to declare the client's application
 * identity during the WebSocket upgrade. Kept in sync with
 * pkg/rpc.ApplicationIDQueryParam on the server.
 */
export const APPLICATION_ID_QUERY_PARAM = 'app_id';

/**
 * withApplicationID sets the application ID sent to the clearnode as the
 * `app_id` WebSocket query parameter. Advisory origin tag only.
 */
export function withApplicationID(appID: string): Option {
  return (config: Config) => {
    config.applicationID = appID;
  };
}

/**
 * appendApplicationIDQueryParam returns `wsURL` with the `app_id` query
 * parameter set to `applicationID`. If `applicationID` is empty the URL is
 * returned unchanged. Any existing `app_id` value is overwritten. Throws a
 * descriptive error if `wsURL` cannot be parsed.
 */
export function appendApplicationIDQueryParam(wsURL: string, applicationID?: string): string {
  if (!applicationID) {
    return wsURL;
  }
  let parsed: URL;
  try {
    parsed = new URL(wsURL);
  } catch (err) {
    const cause = err instanceof Error ? err.message : String(err);
    throw new Error(`cannot append ${APPLICATION_ID_QUERY_PARAM}: invalid url ${JSON.stringify(wsURL)} (${cause})`);
  }
  parsed.searchParams.set(APPLICATION_ID_QUERY_PARAM, applicationID);
  return parsed.toString();
}

/**
 * WithBlockchainRPC configures the RPC endpoint for a specific blockchain.
 * This is required for operations that interact with the blockchain (deposit, withdraw, etc.).
 *
 * @param chainId - The blockchain ID (e.g., 80002n for Polygon Amoy)
 * @param rpcUrl - The RPC endpoint URL
 */
export function withBlockchainRPC(chainId: bigint, rpcUrl: string): Option {
  return (config: Config) => {
    if (!config.blockchainRPCs) {
      config.blockchainRPCs = new Map();
    }
    config.blockchainRPCs.set(chainId, rpcUrl);
  };
}
