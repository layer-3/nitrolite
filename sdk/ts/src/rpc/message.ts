/**
 * Core message structure for RPC communication
 */

import { newErrorPayload, ERROR_PARAM_KEY } from './error';

export enum MsgType {
  Req = 1,
  Resp = 2,
  Event = 3,
  RespErr = 4,
}

/**
 * Payload is a flexible map for method-specific parameters
 */
export type Payload = Record<string, unknown>;

/**
 * Message represents the core data structure for RPC communication.
 * It contains all the information needed to process an RPC call, response, or event.
 *
 * Messages are encoded as JSON arrays for compact transmission:
 * [Type, RequestID, Method, Params, Timestamp]
 */
export interface Message {
  /** Type of the message (request, response, event, or error) */
  type: MsgType;

  /**
   * RequestID is a unique identifier for tracking requests and matching responses.
   * Clients should generate unique IDs to prevent collisions and enable proper
   * request-response correlation.
   */
  requestId: number; // uint64

  /**
   * Method specifies the RPC method to be invoked (e.g., "channels.v1.get_home_channel").
   * Method names should follow a consistent naming convention.
   */
  method: string;

  /**
   * Payload contains the method-specific parameters as a flexible map.
   * This allows different methods to have different parameter structures.
   */
  payload: Payload;

  /**
   * Timestamp is the Unix timestamp in milliseconds when the payload was created.
   * This is used for replay protection and request expiration checks.
   */
  timestamp: number; // uint64 (milliseconds)
}

/**
 * Helper to convert BigInt to string during serialization
 * Also ensures blockchain_id, epoch, version, and nonce fields are strings
 */
function bigIntReplacer(key: string, value: any): any {
  // Convert bigint to string
  if (typeof value === 'bigint') {
    return value.toString();
  }

  // Ensure uint64 fields are strings (in case they're numbers)
  if (key === 'blockchain_id' || key === 'epoch' || key === 'version' || key === 'nonce') {
    if (typeof value === 'number' || typeof value === 'bigint') {
      return value.toString();
    }
  }

  return value;
}

/**
 * NewPayload creates a Payload from any JSON-serializable value
 */
export function newPayload(v: unknown): Payload {
  if (v === null || v === undefined) {
    return {};
  }
  // Always apply bigIntReplacer to ensure bigint values are converted to strings
  return JSON.parse(JSON.stringify(v, bigIntReplacer)) as Payload;
}

/**
 * Translate extracts payload data into a typed object
 */
export function translatePayload<T>(payload: Payload): T {
  return payload as T;
}

/**
 * Extract error from payload if present
 */
export function payloadError(payload: Payload): Error | null {
  const errorMsg = payload[ERROR_PARAM_KEY];
  if (typeof errorMsg === 'string') {
    return new Error(errorMsg);
  }
  return null;
}

/**
 * Extract error from Message if it's an error response
 */
export function messageError(message: Message): Error | null {
  if (message.type !== MsgType.RespErr) {
    return null;
  }
  return payloadError(message.payload);
}

/**
 * NewMessage creates a new Message with the given request ID, type, method, and parameters.
 * The timestamp is automatically set to the current time in Unix milliseconds.
 */
export function newMessage(type: MsgType, requestId: number, method: string, payload: Payload = {}): Message {
  return {
    type,
    requestId,
    method,
    payload,
    timestamp: Date.now(),
  };
}

/**
 * NewRequest creates a new Request message with the given request ID, method, and parameters.
 * The message type is automatically set to MsgType.Req and the timestamp is set to the current time.
 */
export function newRequest(requestId: number, method: string, payload: Payload = {}): Message {
  return newMessage(MsgType.Req, requestId, method, payload);
}

/**
 * NewResponse creates a new Response message with the given request ID, method, and parameters.
 * The message type is automatically set to MsgType.Resp and the timestamp is set to the current time.
 */
export function newResponse(requestId: number, method: string, payload: Payload = {}): Message {
  return newMessage(MsgType.Resp, requestId, method, payload);
}

/**
 * NewEvent creates a new Event message with the given request ID, method, and parameters.
 * The message type is automatically set to MsgType.Event and the timestamp is set to the current time.
 */
export function newEvent(requestId: number, method: string, payload: Payload = {}): Message {
  return newMessage(MsgType.Event, requestId, method, payload);
}

/**
 * NewErrorResponse creates an error Response message containing an error message.
 * The message type is set to MsgTypeRespErr and the error is stored in the payload.
 */
export function newErrorResponse(requestId: number, method: string, errMsg: string): Message {
  const errPayload = newErrorPayload(errMsg);
  return newMessage(MsgType.RespErr, requestId, method, errPayload);
}

/**
 * MarshalJSON serializes Message to JSON array format
 */
export function marshalMessage(message: Message): string {
  const arr = [message.type, message.requestId, message.method, message.payload, message.timestamp];
  return JSON.stringify(arr, bigIntReplacer);
}

/**
 * UnmarshalJSON deserializes Message from JSON array format
 */
export function unmarshalMessage(data: string): Message {
  const arr = JSON.parse(data);
  if (!Array.isArray(arr) || arr.length !== 5) {
    throw new Error('invalid RPCData: expected 5 elements in array');
  }

  const [type, requestId, method, payload, timestamp] = arr;

  if (typeof type !== 'number') {
    throw new Error('invalid type: expected number');
  }
  if (typeof requestId !== 'number') {
    throw new Error('invalid requestId: expected number');
  }
  if (typeof method !== 'string') {
    throw new Error('invalid method: expected string');
  }
  if (typeof payload !== 'object' || payload === null) {
    throw new Error('invalid payload: expected object');
  }
  if (typeof timestamp !== 'number') {
    throw new Error('invalid timestamp: expected number');
  }

  return {
    type: type as MsgType,
    requestId,
    method,
    payload: payload as Payload,
    timestamp,
  };
}
