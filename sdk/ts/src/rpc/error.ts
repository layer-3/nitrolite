/**
 * RPC error types and error handling utilities
 */

import { Payload } from './message.js';

/**
 * Standard key for storing error messages in payloads
 */
export const ERROR_PARAM_KEY = 'error';

/**
 * RPCError represents a custom RPC error type for protocol errors
 */
export class RPCError extends Error {
  constructor(message: string) {
    super(message);
    this.name = 'RPCError';
  }
}

/**
 * Standard RPC error types
 */
export const ErrAlreadyConnected = new RPCError('already connected');
export const ErrNotConnected = new RPCError('not connected');
export const ErrConnectionTimeout = new RPCError('connection timeout');
export const ErrReadingMessage = new RPCError('error reading message');
export const ErrNilRequest = new RPCError('nil request');
export const ErrInvalidRequestMethod = new RPCError('invalid request method');
export const ErrMarshalingRequest = new RPCError('error marshaling request');
export const ErrSendingRequest = new RPCError('error sending request');
export const ErrNoResponse = new RPCError('no response received');
export const ErrSendingPing = new RPCError('error sending ping');
export const ErrDialingWebsocket = new RPCError('error dialing websocket');

/**
 * NewErrorPayload creates a standardized error payload
 */
export function newErrorPayload(errMsg: string): Payload {
  return {
    [ERROR_PARAM_KEY]: errMsg,
  };
}

/**
 * Errorf creates a formatted RPC error
 */
export function errorf(format: string, ...args: unknown[]): RPCError {
  const message = format.replace(/%s/g, () => String(args.shift() || ''));
  return new RPCError(message);
}
