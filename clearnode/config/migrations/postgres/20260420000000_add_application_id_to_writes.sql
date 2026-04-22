-- +goose Up
ALTER TABLE channel_states ADD COLUMN application_id VARCHAR(66);
ALTER TABLE transactions ADD COLUMN application_id VARCHAR(66);

CREATE INDEX idx_channel_states_app_id ON channel_states(application_id);
CREATE INDEX idx_transactions_app_id ON transactions(application_id);

-- +goose Down
DROP INDEX IF EXISTS idx_transactions_app_id;
DROP INDEX IF EXISTS idx_channel_states_app_id;
ALTER TABLE transactions DROP COLUMN application_id;
ALTER TABLE channel_states DROP COLUMN application_id;
