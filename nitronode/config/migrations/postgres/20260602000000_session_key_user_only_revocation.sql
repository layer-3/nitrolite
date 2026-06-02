-- +goose Up
-- User-only session-key revocation: a submit with expires_at <= now deactivates a delegation
-- and is authorized by the wallet's user_sig alone, so the session-key holder cannot veto
-- revocation of a lost, unavailable, or compromised key. Such revocation rows carry an empty
-- session_key_sig, which the *_session_key_sig_present_chk CHECK constraints added in
-- 20260508000000_session_key_ownership_constraints reject. Drop them.
--
-- The ownership guarantee those constraints protected is preserved in application code: submits
-- that activate, extend, or rotate a key (expires_at > now) still require a valid session_key_sig,
-- so every *active* history row remains co-signed by the session-key holder. Owner and auth
-- lookups filter expires_at > now, so revocation rows with an empty session_key_sig are never
-- trusted as a session key's owner.
ALTER TABLE app_session_key_states_v1
    DROP CONSTRAINT IF EXISTS app_session_key_states_v1_session_key_sig_present_chk;

ALTER TABLE channel_session_key_states_v1
    DROP CONSTRAINT IF EXISTS channel_session_key_states_v1_session_key_sig_present_chk;

-- +goose Down
-- Re-add as NOT VALID: skip the legacy backfill scan (revocation rows with an empty
-- session_key_sig may now exist) so the down migration cannot fail on pre-existing data; only
-- future inserts are checked. Matches how 20260508000000 originally added them.
ALTER TABLE app_session_key_states_v1
    ADD CONSTRAINT app_session_key_states_v1_session_key_sig_present_chk
    CHECK (session_key_sig IS NOT NULL AND session_key_sig <> '') NOT VALID;

ALTER TABLE channel_session_key_states_v1
    ADD CONSTRAINT channel_session_key_states_v1_session_key_sig_present_chk
    CHECK (session_key_sig IS NOT NULL AND session_key_sig <> '') NOT VALID;
