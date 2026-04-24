import type { AppSessionKeyStateV1 } from './app/types.js';
import type { ChannelSessionKeyStateV1 } from './rpc/types.js';

function asRecord(raw: unknown, context: string): Record<string, unknown> {
  if (!raw || typeof raw !== 'object') {
    throw new Error(`Invalid ${context}: expected object`);
  }
  return raw as Record<string, unknown>;
}

function requireStringField(raw: unknown, context: string, field: string): string {
  const record = asRecord(raw, context);
  const value = record[field];
  if (typeof value !== 'string') {
    throw new Error(`Invalid ${context}: missing required string field ${field}`);
  }
  return value;
}

function requireStringArrayField(raw: unknown, context: string, field: string): string[] {
  const record = asRecord(raw, context);
  const value = record[field];
  if (!Array.isArray(value) || value.some((item) => typeof item !== 'string')) {
    throw new Error(`Invalid ${context}: expected ${field} to be string[]`);
  }
  return value;
}

export function transformChannelSessionKeyState(
  raw: unknown,
  context = 'channel session key state'
): ChannelSessionKeyStateV1 {
  return {
    user_address: requireStringField(raw, context, 'user_address'),
    session_key: requireStringField(raw, context, 'session_key'),
    version: requireStringField(raw, context, 'version'),
    assets: requireStringArrayField(raw, context, 'assets'),
    expires_at: requireStringField(raw, context, 'expires_at'),
    user_sig: requireStringField(raw, context, 'user_sig'),
  };
}

export function transformAppSessionKeyState(
  raw: unknown,
  context = 'app session key state'
): AppSessionKeyStateV1 {
  return {
    user_address: requireStringField(raw, context, 'user_address'),
    session_key: requireStringField(raw, context, 'session_key'),
    version: requireStringField(raw, context, 'version'),
    application_ids: requireStringArrayField(raw, context, 'application_ids'),
    app_session_ids: requireStringArrayField(raw, context, 'app_session_ids'),
    expires_at: requireStringField(raw, context, 'expires_at'),
    user_sig: requireStringField(raw, context, 'user_sig'),
  };
}
