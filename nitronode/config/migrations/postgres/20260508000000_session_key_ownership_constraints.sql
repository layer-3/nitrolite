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

ALTER TABLE current_session_key_states_v1
    ADD CONSTRAINT current_session_key_states_v1_key_kind_uniq UNIQUE (session_key, kind);

-- +goose Down
ALTER TABLE current_session_key_states_v1
    DROP CONSTRAINT IF EXISTS current_session_key_states_v1_key_kind_uniq;

ALTER TABLE channel_session_key_states_v1
    DROP COLUMN IF EXISTS session_key_sig;

ALTER TABLE app_session_key_states_v1
    DROP COLUMN IF EXISTS session_key_sig;
