-- +goose Up

-- block_hash was added as CHAR(66) NOT NULL DEFAULT '', so rows predating the
-- column (and any default insert) hold a space-padded empty string rather than a
-- real hash. Drop the NOT NULL constraint and the empty default, then normalize
-- those padded empties to NULL so the reconciler's empty-hash guard recognizes
-- "no recorded hash" instead of treating it as a reorged-out block.
ALTER TABLE contract_events ALTER COLUMN block_hash DROP DEFAULT;
ALTER TABLE contract_events ALTER COLUMN block_hash DROP NOT NULL;
UPDATE contract_events SET block_hash = NULL WHERE TRIM(block_hash) = '';

-- +goose Down

UPDATE contract_events SET block_hash = '' WHERE block_hash IS NULL;
ALTER TABLE contract_events ALTER COLUMN block_hash SET DEFAULT '';
ALTER TABLE contract_events ALTER COLUMN block_hash SET NOT NULL;
