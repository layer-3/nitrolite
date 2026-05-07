-- +goose Up
-- Pointer table holding the latest version per (user_address, session_key, kind).
-- Reads of the get_last_key_states endpoints filter this table by user_address (+ optional
-- session_key) and JOIN the corresponding history table on (user_address, session_key, version).
-- This eliminates the GROUP BY scan over history that grows with version churn and bounds
-- per-request DB work to O(distinct keys for user, kind).
--
-- kind values (SessionKeyKind enum on the Go side):
--   1 = channel
--   2 = app_session
CREATE TABLE current_session_key_states_v1 (
    user_address CHAR(42) NOT NULL,
    session_key CHAR(42) NOT NULL,
    kind SMALLINT NOT NULL,
    version NUMERIC(20,0) NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_address, session_key, kind)
);

CREATE INDEX idx_current_session_key_states_v1_user_kind
    ON current_session_key_states_v1(user_address, kind);

-- Backfill from app session key history: latest version per (user_address, session_key).
INSERT INTO current_session_key_states_v1 (user_address, session_key, kind, version, updated_at)
SELECT user_address, session_key, 2, MAX(version), NOW()
FROM app_session_key_states_v1
GROUP BY user_address, session_key
ON CONFLICT (user_address, session_key, kind) DO NOTHING;

-- Backfill from channel session key history: latest version per (user_address, session_key).
INSERT INTO current_session_key_states_v1 (user_address, session_key, kind, version, updated_at)
SELECT user_address, session_key, 1, MAX(version), NOW()
FROM channel_session_key_states_v1
GROUP BY user_address, session_key
ON CONFLICT (user_address, session_key, kind) DO NOTHING;

-- +goose Down
DROP INDEX IF EXISTS idx_current_session_key_states_v1_user_kind;
DROP TABLE IF EXISTS current_session_key_states_v1;
