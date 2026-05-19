-- +goose Up
-- Bind a session_key to a single owner (per kind) and require co-signature at submit time.

-- Co-signature: the session-key holder proves possession at registration and on every update.
-- Nullable to accommodate rows written before this column existed; new submits enforce non-null
-- in application code. Columns add first so a constraint failure below does not leave the
-- session_key_sig schema partially applied.
ALTER TABLE app_session_key_states_v1
    ADD COLUMN session_key_sig TEXT;

ALTER TABLE channel_session_key_states_v1
    ADD COLUMN session_key_sig TEXT;

-- Pre-flight: refuse the migration if duplicate (session_key, kind) rows are present in
-- current_session_key_states_v1. Such rows are evidence of cross-wallet collisions that
-- the old code path allowed; manual remediation is required before the constraint adds.
-- +goose StatementBegin
DO $$
DECLARE
    dup_count bigint;
BEGIN
    SELECT COUNT(*) INTO dup_count
    FROM (
        SELECT session_key, kind
        FROM current_session_key_states_v1
        GROUP BY session_key, kind
        HAVING COUNT(*) > 1
    ) AS dups;

    IF dup_count > 0 THEN
        RAISE EXCEPTION 'duplicate (session_key, kind) rows detected (%); manual remediation required before applying constraint', dup_count;
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE current_session_key_states_v1
    ADD CONSTRAINT current_session_key_states_v1_key_kind_uniq UNIQUE (session_key, kind);

-- Fail closed at the DB layer for new history rows during rolling deploys: a pre-MF-H02 binary
-- (already running the MF-H01 schema where current_session_key_states_v1 exists) would happily
-- insert history rows without session_key_sig, and the new GetAppSessionKeyOwner/GetChannelSessionKeyOwner
-- lookups would then trust those unproven rows as legitimate owners. NOT VALID skips the legacy
-- backfill scan so pre-existing rows are not blocked; only future inserts are checked.
ALTER TABLE app_session_key_states_v1
    ADD CONSTRAINT app_session_key_states_v1_session_key_sig_present_chk
    CHECK (session_key_sig IS NOT NULL AND session_key_sig <> '') NOT VALID;

ALTER TABLE channel_session_key_states_v1
    ADD CONSTRAINT channel_session_key_states_v1_session_key_sig_present_chk
    CHECK (session_key_sig IS NOT NULL AND session_key_sig <> '') NOT VALID;

-- +goose Down
ALTER TABLE channel_session_key_states_v1
    DROP CONSTRAINT IF EXISTS channel_session_key_states_v1_session_key_sig_present_chk;

ALTER TABLE app_session_key_states_v1
    DROP CONSTRAINT IF EXISTS app_session_key_states_v1_session_key_sig_present_chk;

ALTER TABLE current_session_key_states_v1
    DROP CONSTRAINT IF EXISTS current_session_key_states_v1_key_kind_uniq;

ALTER TABLE channel_session_key_states_v1
    DROP COLUMN IF EXISTS session_key_sig;

ALTER TABLE app_session_key_states_v1
    DROP COLUMN IF EXISTS session_key_sig;
