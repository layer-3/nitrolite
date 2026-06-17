-- +goose Up

ALTER TABLE contract_events ADD COLUMN block_hash CHAR(66) NOT NULL DEFAULT '';

-- +goose Down

ALTER TABLE contract_events DROP COLUMN block_hash;
