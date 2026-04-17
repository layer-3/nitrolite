-- +goose Up
ALTER TABLE user_balances ADD COLUMN home_blockchain_id NUMERIC(20,0) NOT NULL DEFAULT 0;
ALTER TABLE user_balances ADD COLUMN enforced NUMERIC(78, 18) NOT NULL DEFAULT 0;

UPDATE user_balances ub
SET home_blockchain_id = c.blockchain_id
FROM (
    SELECT DISTINCT ON (user_wallet, asset)
           user_wallet, asset, blockchain_id, channel_id, state_version
    FROM channels
    WHERE type = 1 AND status <= 1
    ORDER BY user_wallet, asset, state_version DESC, channel_id DESC
) c
WHERE c.user_wallet = ub.user_wallet
  AND c.asset = ub.asset;

UPDATE user_balances ub
SET enforced = COALESCE((
    SELECT s.home_user_balance
    FROM (
        SELECT DISTINCT ON (user_wallet, asset)
               user_wallet, asset, channel_id, state_version
        FROM channels
        WHERE type = 1 AND status <= 1 AND state_version > 0
        ORDER BY user_wallet, asset, state_version DESC, channel_id DESC
    ) c
    JOIN channel_states s ON s.home_channel_id = c.channel_id AND s.version = c.state_version
    WHERE c.user_wallet = ub.user_wallet
      AND c.asset = ub.asset
    ORDER BY s.epoch DESC
    LIMIT 1
), 0);

-- +goose Down
ALTER TABLE user_balances DROP COLUMN enforced;
ALTER TABLE user_balances DROP COLUMN home_blockchain_id;
