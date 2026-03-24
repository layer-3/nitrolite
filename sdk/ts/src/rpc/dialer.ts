import { Message, marshalMessage, unmarshalMessage } from './message';
import { ErrAlreadyConnected, ErrNotConnected } from './error';

/**
 * Dialer is the interface for RPC client connections.
 * It provides methods to establish connections, send requests, and receive responses.
 */
export interface Dialer {
  /**
   * Dial establishes a connection to the specified URL.
   * This method is designed to be called asynchronously as it manages the connection lifecycle.
   * The handleClosure callback is invoked when the connection is closed, with an error if any.
   */
  dial(url: string, handleClosure: (err?: Error) => void): Promise<void>;

  /**
   * IsConnected returns true if the dialer has an active connection.
   */
  isConnected(): boolean;

  /**
   * Call sends an RPC request and waits for a response.
   * It returns an error if the request cannot be sent or no response is received.
   * The request can be cancelled using an AbortSignal.
   */
  call(req: Message, signal?: AbortSignal): Promise<Message>;

  /**
   * EventChannel returns an async iterable for receiving unsolicited events from the server.
   * Events are responses that don't match any pending request ID.
   */
  eventChannel(): AsyncIterable<Message>;

  /**
   * Close closes the connection gracefully.
   */
  close(): Promise<void>;
}

/**
 * WebsocketDialerConfig contains configuration options for the WebSocket dialer
 */
export interface WebsocketDialerConfig {
  /**
   * HandshakeTimeout is the duration to wait for the WebSocket handshake to complete (in milliseconds)
   */
  handshakeTimeout: number;

  /**
   * EventChanSize is the buffer size for the event channel
   * A larger buffer prevents blocking when processing many unsolicited events
   */
  eventChanSize: number;
}

/**
 * DefaultWebsocketDialerConfig provides sensible defaults for WebSocket connections
 */
export const DefaultWebsocketDialerConfig: WebsocketDialerConfig = {
  handshakeTimeout: 5000, // 5 seconds
  eventChanSize: 100,
};

/**
 * WebsocketDialer implements the Dialer interface using WebSocket connections.
 * It provides thread-safe RPC communication with automatic ping handling.
 */
export class WebsocketDialer implements Dialer {
  private config: WebsocketDialerConfig;
  private ws: WebSocket | null = null;
  private responseSinks: Map<number, (msg: Message) => void> = new Map();
  private eventQueue: Message[] = [];
  private eventResolvers: Array<(value: Message) => void> = [];
  private closeHandler?: (err?: Error) => void;

  constructor(config: WebsocketDialerConfig = DefaultWebsocketDialerConfig) {
    this.config = config;
  }

  /**
   * Dial establishes a WebSocket connection to the specified URL.
   */
  async dial(url: string, handleClosure: (err?: Error) => void): Promise<void> {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      throw ErrAlreadyConnected;
    }

    this.closeHandler = handleClosure;

    return new Promise((resolve, reject) => {
      const ws = new WebSocket(url);
      const timeout = setTimeout(() => {
        ws.close();
        reject(new Error('WebSocket handshake timeout'));
      }, this.config.handshakeTimeout);

      ws.onopen = () => {
        clearTimeout(timeout);
        this.ws = ws;
        resolve();
      };

      ws.onerror = () => {
        clearTimeout(timeout);
        const error = new Error('WebSocket error');
        if (this.ws === ws) {
          this.handleClose(error);
        }
        reject(error);
      };

      ws.onclose = () => {
        clearTimeout(timeout);
        if (this.ws === ws) {
          this.handleClose();
        }
      };

      ws.onmessage = (event) => {
        this.handleMessage(event.data);
      };
    });
  }

  /**
   * IsConnected returns true if the dialer has an active connection.
   */
  isConnected(): boolean {
    return this.ws !== null && this.ws.readyState === WebSocket.OPEN;
  }

  /**
   * Call sends an RPC request and waits for a response.
   */
  async call(req: Message, signal?: AbortSignal): Promise<Message> {
    if (!this.isConnected()) {
      throw ErrNotConnected;
    }

    return new Promise((resolve, reject) => {
      if (signal?.aborted) {
        reject(new Error('Request aborted'));
        return;
      }

      const abortHandler = () => {
        this.responseSinks.delete(req.requestId);
        reject(new Error('Request aborted'));
      };

      signal?.addEventListener('abort', abortHandler, { once: true });

      this.responseSinks.set(req.requestId, (msg: Message) => {
        signal?.removeEventListener('abort', abortHandler);
        resolve(msg);
      });

      try {
        const data = marshalMessage(req);
        this.ws!.send(data);
      } catch (error) {
        this.responseSinks.delete(req.requestId);
        signal?.removeEventListener('abort', abortHandler);
        reject(error);
      }
    });
  }

  /**
   * EventChannel returns an async iterable for receiving unsolicited events from the server.
   */
  async *eventChannel(): AsyncIterable<Message> {
    while (this.isConnected()) {
      if (this.eventQueue.length > 0) {
        const event = this.eventQueue.shift()!;
        yield event;
      } else {
        // Wait for next event
        const event = await new Promise<Message>((resolve) => {
          this.eventResolvers.push(resolve);
        });
        yield event;
      }
    }
  }

  /**
   * Close closes the connection gracefully.
   */
  async close(): Promise<void> {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
  }

  private handleMessage(data: string): void {
    try {
      const msg: Message = unmarshalMessage(data);

      // Check if this is a response to a pending request
      const sink = this.responseSinks.get(msg.requestId);
      if (sink) {
        this.responseSinks.delete(msg.requestId);
        sink(msg);
      } else {
        // This is an unsolicited event
        if (this.eventResolvers.length > 0) {
          const resolver = this.eventResolvers.shift()!;
          resolver(msg);
        } else {
          this.eventQueue.push(msg);
          // Limit queue size
          if (this.eventQueue.length > this.config.eventChanSize) {
            this.eventQueue.shift();
          }
        }
      }
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error);
    }
  }

  private handleClose(error?: Error): void {
    this.ws = null;
    // Reject all pending requests
    for (const [id, sink] of this.responseSinks.entries()) {
      sink({
        type: 4, // MsgType.RespErr
        requestId: id,
        method: '',
        payload: { error: 'Connection closed' },
        timestamp: Date.now(),
      });
    }
    this.responseSinks.clear();

    if (this.closeHandler) {
      this.closeHandler(error);
    }
  }
}

/**
 * NewWebsocketDialer creates a new WebSocket dialer with the given configuration
 */
export function newWebsocketDialer(config: WebsocketDialerConfig = DefaultWebsocketDialerConfig): WebsocketDialer {
  return new WebsocketDialer(config);
}
