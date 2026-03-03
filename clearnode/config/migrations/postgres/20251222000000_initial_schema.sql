-- +goose Up

-- Channels table: Represents state channels between user and node
CREATE TABLE channels (
    channel_id CHAR(66) PRIMARY KEY,
    user_wallet CHAR(42) NOT NULL,
    asset VARCHAR(20) NOT NULL,
    type SMALLINT NOT NULL, -- ChannelType enum: 0=void, 1=home, 2=escrow
    blockchain_id NUMERIC(20,0) NOT NULL,
    token CHAR(42) NOT NULL,
    challenge_duration BIGINT NOT NULL DEFAULT 0,
    challenge_expires_at TIMESTAMPTZ,
    nonce NUMERIC(20,0) NOT NULL DEFAULT 0,
    approved_sig_validators VARCHAR(66) NOT NULL DEFAULT 0,
    status SMALLINT NOT NULL, -- ChannelStatus enum: 0=void, 1=open, 2=challenged, 3=closed
    state_version NUMERIC(20,0) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_channels_wallet_asset_type_status ON channels(user_wallet, asset, type, status);
CREATE INDEX idx_channels_asset_status ON channels(asset, status);

-- Channel States table: Immutable state records
CREATE TABLE channel_states (
    id CHAR(66) PRIMARY KEY, -- Deterministic hash: Hash(UserWallet, Asset, Epoch, Version)
    asset VARCHAR(20) NOT NULL,
    user_wallet CHAR(42) NOT NULL,
    epoch NUMERIC(20,0) NOT NULL,
    version NUMERIC(20,0) NOT NULL,

    transition_type SMALLINT NOT NULL, -- TransactionType enum for the transition that led to this state
    transition_tx_id CHAR(66), -- Transaction that caused this state transition
    transition_account_id VARCHAR(66), -- Account (wallet or channel) that initiated the transition
    transition_amount NUMERIC(78, 18) NOT NULL DEFAULT 0, -- Amount involved in the transition (positive for credits to user, negative for debits)

    -- Optional channel references
    home_channel_id CHAR(66),
    escrow_channel_id CHAR(66),

    -- Home Channel balances and flows (balances are positive only, net flows can be negative)
    home_user_balance NUMERIC(78, 18) NOT NULL DEFAULT 0,
    home_user_net_flow NUMERIC(78, 18) NOT NULL DEFAULT 0,
    home_node_balance NUMERIC(78, 18) NOT NULL DEFAULT 0,
    home_node_net_flow NUMERIC(78, 18) NOT NULL DEFAULT 0,

    -- Escrow Channel balances and flows (balances are positive only, net flows can be negative)
    escrow_user_balance NUMERIC(78, 18) NOT NULL DEFAULT 0,
    escrow_user_net_flow NUMERIC(78, 18) NOT NULL DEFAULT 0,
    escrow_node_balance NUMERIC(78, 18) NOT NULL DEFAULT 0,
    escrow_node_net_flow NUMERIC(78, 18) NOT NULL DEFAULT 0,

    user_sig TEXT, -- TODO: consider using fixed char length
    node_sig TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Uniqueness constraint to catch bugs in version computation or concurrency issues
ALTER TABLE channel_states ADD CONSTRAINT uq_channel_states_wallet_asset_epoch_version
    UNIQUE (user_wallet, asset, epoch, version);

-- Covering index for "latest state" queries: WHERE user_wallet=? AND asset=? ORDER BY epoch DESC, version DESC LIMIT 1
CREATE INDEX idx_channel_states_latest ON channel_states(user_wallet, asset, epoch DESC, version DESC);

-- Partial index for "signed latest state" queries
CREATE INDEX idx_channel_states_latest_signed ON channel_states(user_wallet, asset, epoch DESC, version DESC)
    WHERE user_sig IS NOT NULL AND node_sig IS NOT NULL;

-- Improved channel ID indexes with epoch/version DESC for "latest by channel" queries
CREATE INDEX idx_channel_states_home_channel_id ON channel_states(home_channel_id, epoch DESC, version DESC)
    WHERE home_channel_id IS NOT NULL;
CREATE INDEX idx_channel_states_escrow_channel_id ON channel_states(escrow_channel_id, epoch DESC, version DESC)
    WHERE escrow_channel_id IS NOT NULL;

-- Transactions table: Records all transactions with optional state references
CREATE TABLE transactions (
    id CHAR(66) PRIMARY KEY, -- Deterministic hash
    tx_type SMALLINT NOT NULL, -- TransactionType enum
    asset_symbol VARCHAR(20) NOT NULL,
    from_account VARCHAR(66) NOT NULL, -- Can be wallet (42) or channel ID (66)
    to_account VARCHAR(66) NOT NULL, -- Can be wallet (42) or channel ID (66)
    sender_new_state_id CHAR(66),
    receiver_new_state_id CHAR(66),
    amount NUMERIC(78, 18) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_type ON transactions(tx_type);
CREATE INDEX idx_transactions_from_account ON transactions(from_account);
CREATE INDEX idx_transactions_to_account ON transactions(to_account);
CREATE INDEX idx_transactions_from_to_type ON transactions(from_account, to_account, tx_type);
CREATE INDEX idx_transactions_from_comp ON transactions(from_account, asset_symbol, created_at DESC);
CREATE INDEX idx_transactions_to_comp ON transactions(to_account, asset_symbol, created_at DESC);

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

-- App Sessions table: Application sessions
CREATE TABLE app_sessions_v1 (
    id CHAR(66) PRIMARY KEY,
    application_id VARCHAR NOT NULL,
    nonce NUMERIC(20,0) NOT NULL,
    session_data TEXT NOT NULL,
    quorum SMALLINT NOT NULL DEFAULT 100,
    version NUMERIC(20,0) NOT NULL DEFAULT 1,
    status SMALLINT NOT NULL, -- AppSessionStatus enum
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_app_sessions_v1_application ON app_sessions_v1(application);
CREATE INDEX idx_app_sessions_v1_status ON app_sessions_v1(status);

-- App Session Participants table: Participants in application sessions
CREATE TABLE app_session_participants_v1 (
    app_session_id CHAR(66) NOT NULL,
    wallet_address CHAR(42) NOT NULL,
    signature_weight SMALLINT NOT NULL,
    PRIMARY KEY (app_session_id, wallet_address),
    FOREIGN KEY (app_session_id) REFERENCES app_sessions_v1(id) ON DELETE CASCADE
);

CREATE INDEX idx_app_session_participants_v1_wallet ON app_session_participants_v1(wallet_address);

-- App Ledger table: Internal ledger entries for application sessions
CREATE TABLE app_ledger_v1 (
    id CHAR(36) PRIMARY KEY, -- UUID
    account_id CHAR(66) NOT NULL, -- Session ID
    asset_symbol VARCHAR(20) NOT NULL,
    wallet CHAR(42) NOT NULL,
    credit NUMERIC(78, 18) NOT NULL DEFAULT 0,
    debit NUMERIC(78, 18) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_app_ledger_v1_account_asset ON app_ledger_v1(account_id, asset_symbol);
CREATE INDEX idx_app_ledger_v1_wallet ON app_ledger_v1(wallet);

-- Contract events table: Blockchain event logs
CREATE TABLE contract_events (
    id BIGSERIAL PRIMARY KEY,
    contract_address CHAR(42) NOT NULL,
    blockchain_id NUMERIC(20,0) NOT NULL,
    name VARCHAR(255) NOT NULL,
    block_number NUMERIC(20,0) NOT NULL,
    transaction_hash VARCHAR(255) NOT NULL,
    log_index BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX contract_events_tx_log_chain_idx ON contract_events (transaction_hash, log_index, blockchain_id);
CREATE INDEX idx_contract_events_latest ON contract_events(blockchain_id, contract_address, block_number DESC, log_index DESC);

-- Blockchain actions table: Pending blockchain operations
CREATE TABLE blockchain_actions (
    id BIGSERIAL PRIMARY KEY,
    action_type SMALLINT NOT NULL,
    state_id CHAR(66),
    blockchain_id NUMERIC(20,0) NOT NULL,
    action_data JSONB,
    status SMALLINT NOT NULL DEFAULT 0,
    retry_count SMALLINT NOT NULL DEFAULT 0,
    last_error TEXT,
    transaction_hash CHAR(66),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    FOREIGN KEY (state_id) REFERENCES channel_states(id) ON DELETE CASCADE
);

CREATE INDEX idx_blockchain_actions_pending ON blockchain_actions(status, created_at) WHERE status = 0;
CREATE INDEX idx_blockchain_actions_state_id ON blockchain_actions(state_id);

-- Session key states: Stores session key delegation metadata signed by the user
-- ID is Hash(user_address + session_key + version)
CREATE TABLE app_session_key_states_v1 (
    id CHAR(66) PRIMARY KEY,
    user_address CHAR(42) NOT NULL,
    session_key CHAR(42) NOT NULL,
    version NUMERIC(20,0) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    user_sig TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_address, session_key, version)
);

CREATE INDEX idx_app_session_key_states_v1_user ON app_session_key_states_v1(user_address);
CREATE INDEX idx_app_session_key_states_v1_expires ON app_session_key_states_v1(expires_at);

-- Session key application IDs: Links session keys to application IDs
CREATE TABLE app_session_key_applications_v1 (
    session_key_state_id CHAR(66) NOT NULL,
    application_id VARCHAR(66) NOT NULL,
    PRIMARY KEY (session_key_state_id, application_id),
    FOREIGN KEY (session_key_state_id) REFERENCES app_session_key_states_v1(id) ON DELETE CASCADE
);

CREATE INDEX idx_app_session_key_applications_v1_app_id ON app_session_key_applications_v1(application_id);

-- Session key app session IDs: Links session keys to app session IDs
CREATE TABLE app_session_key_app_sessions_v1 (
    session_key_state_id CHAR(66) NOT NULL,
    app_session_id CHAR(66) NOT NULL,
    PRIMARY KEY (session_key_state_id, app_session_id),
    FOREIGN KEY (session_key_state_id) REFERENCES app_session_key_states_v1(id) ON DELETE CASCADE
);

CREATE INDEX idx_app_session_key_app_sessions_v1_session_id ON app_session_key_app_sessions_v1(app_session_id);

-- Channel session key states: Stores channel session key delegation metadata signed by the user
-- ID is Hash(user_address + session_key + version)
CREATE TABLE channel_session_key_states_v1 (
    id CHAR(66) PRIMARY KEY,
    user_address CHAR(42) NOT NULL,
    session_key CHAR(42) NOT NULL,
    version NUMERIC(20,0) NOT NULL,
    metadata_hash CHAR(66) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    user_sig TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_address, session_key, version)
);

CREATE INDEX idx_channel_session_key_states_v1_user ON channel_session_key_states_v1(user_address);
CREATE INDEX idx_channel_session_key_states_v1_expires ON channel_session_key_states_v1(expires_at);

-- Channel session key assets: Links channel session keys to permitted assets
CREATE TABLE channel_session_key_assets_v1 (
    session_key_state_id CHAR(66) NOT NULL,
    asset VARCHAR(20) NOT NULL,
    PRIMARY KEY (session_key_state_id, asset),
    FOREIGN KEY (session_key_state_id) REFERENCES channel_session_key_states_v1(id) ON DELETE CASCADE
);

CREATE INDEX idx_channel_session_key_assets_v1_asset ON channel_session_key_assets_v1(asset);

-- User balances table: Stores aggregated user balances per asset
-- This table is used to quickly retrieve user balances without querying channel_states
-- and provides row-level locking for concurrent balance updates
CREATE TABLE user_balances (
    user_wallet CHAR(42) NOT NULL,
    asset VARCHAR(20) NOT NULL,
    balance NUMERIC(78, 18) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_wallet, asset)
);

CREATE INDEX idx_user_balances_user_wallet ON user_balances(user_wallet);

-- +goose Down
DROP INDEX IF EXISTS idx_user_balances_user_wallet;
DROP TABLE IF EXISTS user_balances;
DROP INDEX IF EXISTS idx_channel_session_key_assets_v1_asset;
DROP TABLE IF EXISTS channel_session_key_assets_v1;
DROP INDEX IF EXISTS idx_channel_session_key_states_v1_expires;
DROP INDEX IF EXISTS idx_channel_session_key_states_v1_user;
DROP TABLE IF EXISTS channel_session_key_states_v1;
DROP INDEX IF EXISTS idx_app_session_key_app_sessions_v1_session_id;
DROP TABLE IF EXISTS app_session_key_app_sessions_v1;
DROP INDEX IF EXISTS idx_app_session_key_applications_v1_app_id;
DROP TABLE IF EXISTS app_session_key_applications_v1;
DROP INDEX IF EXISTS idx_app_session_key_states_v1_expires;
DROP INDEX IF EXISTS idx_app_session_key_states_v1_user;
DROP TABLE IF EXISTS app_session_key_states_v1;
DROP INDEX IF EXISTS idx_blockchain_actions_state_id;
DROP INDEX IF EXISTS idx_blockchain_actions_pending;
DROP TABLE IF EXISTS blockchain_actions;
DROP INDEX IF EXISTS idx_contract_events_latest;
DROP INDEX IF EXISTS contract_events_tx_log_chain_idx;
DROP TABLE IF EXISTS contract_events;
DROP INDEX IF EXISTS idx_app_ledger_v1_wallet;
DROP INDEX IF EXISTS idx_app_ledger_v1_account_asset;
DROP TABLE IF EXISTS app_ledger_v1;
DROP INDEX IF EXISTS idx_app_session_participants_v1_wallet;
DROP TABLE IF EXISTS app_session_participants_v1;
DROP INDEX IF EXISTS idx_app_sessions_v1_status;
DROP INDEX IF EXISTS idx_app_sessions_v1_application;
DROP TABLE IF EXISTS app_sessions_v1;
DROP INDEX IF EXISTS idx_apps_v1_owner_wallet;
DROP TABLE IF EXISTS apps_v1;
DROP INDEX IF EXISTS idx_transactions_to_comp;
DROP INDEX IF EXISTS idx_transactions_from_comp;
DROP INDEX IF EXISTS idx_transactions_from_to_type;
DROP INDEX IF EXISTS idx_transactions_to_account;
DROP INDEX IF EXISTS idx_transactions_from_account;
DROP INDEX IF EXISTS idx_transactions_type;
DROP TABLE IF EXISTS transactions;
DROP INDEX IF EXISTS idx_channel_states_escrow_channel_id;
DROP INDEX IF EXISTS idx_channel_states_home_channel_id;
DROP INDEX IF EXISTS idx_channel_states_latest_signed;
DROP INDEX IF EXISTS idx_channel_states_latest;
ALTER TABLE channel_states DROP CONSTRAINT IF EXISTS uq_channel_states_wallet_asset_epoch_version;
DROP TABLE IF EXISTS channel_states;
DROP INDEX IF EXISTS idx_channels_asset_status;
DROP INDEX IF EXISTS idx_channels_wallet_asset_type_status;
DROP TABLE IF EXISTS channels;
