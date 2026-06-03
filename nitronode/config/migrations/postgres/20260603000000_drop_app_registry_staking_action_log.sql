-- +goose Up

-- Drop the off-chain app registry, user staking, and action-log (rate limiting)
-- subsystems. Their tables and indexes are no longer referenced by the node.

DROP TABLE IF EXISTS apps_v1;
DROP TABLE IF EXISTS action_log_v1;
DROP TABLE IF EXISTS user_staked_v1;

-- +goose Down

-- Application registry
CREATE TABLE apps_v1 (
    id VARCHAR(66) PRIMARY KEY,
    owner_wallet CHAR(42) NOT NULL,
    metadata TEXT NOT NULL,
    version NUMERIC(20,0) NOT NULL DEFAULT 1,
    creation_approval_not_required BOOLEAN NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_apps_v1_owner_wallet ON apps_v1(owner_wallet);

-- User staked table: Stores staked amounts per user per blockchain
CREATE TABLE user_staked_v1 (
    user_wallet CHAR(42) NOT NULL,
    blockchain_id NUMERIC(20,0) NOT NULL,
    amount NUMERIC(78, 18) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_wallet, blockchain_id)
);

CREATE INDEX idx_user_staked_v1_user_wallet ON user_staked_v1(user_wallet);

-- Action log table: Records user actions for rate limiting and auditing
CREATE TABLE action_log_v1 (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_wallet CHAR(42) NOT NULL,
    gated_action SMALLINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_action_log_v1_wallet_gated_action_created ON action_log_v1(user_wallet, gated_action, created_at DESC);
