# PostgreSQL Integration Tests

This package contains integration tests that verify all database operations work correctly with PostgreSQL.

## Prerequisites

1. PostgreSQL server running (version 12+)
2. A test database created
3. Migrations applied to the test database

## Setup

### 1. Create Test Database

```bash
# Connect to PostgreSQL
psql -U postgres

# Create test database
CREATE DATABASE nitrolite_test;

# Exit psql
\q
```

### 2. Apply Migrations

Apply the migration schema to your test database:

```bash
# Navigate to migrations directory
cd clearnode/config/migrations/postgres

# Apply the migration manually
psql -U postgres -d nitrolite_test -f 20251222000000_initial_schema.sql
```

Or use a migration tool like goose:

```bash
goose -dir clearnode/config/migrations/postgres postgres "user=postgres dbname=nitrolite_test sslmode=disable" up
```

## Running the Tests

Set the `POSTGRES_DSN` environment variable and run the tests:

```bash
# Set the connection string
export POSTGRES_DSN="host=localhost user=postgres password=postgres dbname=nitrolite_test port=5432 sslmode=disable"

# Run all integration tests
go test ./clearnode/store/database/test -v

# Run specific test
go test ./clearnode/store/database/test -v -run TestPostgres_ChannelOperations

# Run tests with cleanup instructions
go test ./clearnode/store/database/test -v -run TestPostgres_PrintCleanupSQL
```

## Tests Included

The test suite covers:

1. **Channel Operations** (`TestPostgres_ChannelOperations`)
   - Creating channels
   - Retrieving channels by ID
   - Updating channel status and version

2. **State Operations** (`TestPostgres_StateOperations`)
   - Storing user states
   - Retrieving latest states
   - Handling multiple epochs and versions

3. **Transaction Operations** (`TestPostgres_TransactionOperations`)
   - Recording transactions
   - Retrieving user transaction history
   - Filtering by asset and type

4. **App Session Operations** (`TestPostgres_AppSessionOperations`)
   - Creating app sessions with participants
   - Retrieving sessions by participant
   - Updating session status

5. **App Ledger Operations** (`TestPostgres_AppLedgerOperations`)
   - Recording ledger entries
   - Calculating session balances
   - Getting participant allocations

6. **User Balances** (`TestPostgres_UserBalances`)
   - Retrieving user balances across multiple assets

7. **Decimal Precision** (`TestPostgres_DecimalPrecision`)
   - Verifying large decimal values (78 digits, 18 decimal places)
   - Testing negative net flows
   - Ensuring no precision loss

## Cleanup

After running the tests, you can clean up the test data using the SQL commands printed by `TestPostgres_PrintCleanupSQL`:

```bash
# Run the cleanup test to see the SQL commands
go test ./clearnode/store/database/test -v -run TestPostgres_PrintCleanupSQL
```

Or manually clean up:

```sql
-- Delete all test data (preserves schema)
DELETE FROM blockchain_actions;
DELETE FROM app_ledger_v1;
DELETE FROM app_session_participants_v1;
DELETE FROM app_sessions_v1;
DELETE FROM transactions;
DELETE FROM channel_states;
DELETE FROM channels;
DELETE FROM contract_events;
DELETE FROM session_keys;
DELETE FROM rpc_store;
```

Or drop and recreate:

```sql
-- Drop all tables
DROP TABLE IF EXISTS blockchain_actions CASCADE;
DROP TABLE IF EXISTS app_ledger_v1 CASCADE;
DROP TABLE IF EXISTS app_session_participants_v1 CASCADE;
DROP TABLE IF EXISTS app_sessions_v1 CASCADE;
DROP TABLE IF EXISTS transactions CASCADE;
DROP TABLE IF EXISTS channel_states CASCADE;
DROP TABLE IF EXISTS channels CASCADE;
DROP TABLE IF EXISTS contract_events CASCADE;
DROP TABLE IF EXISTS session_keys CASCADE;
DROP TABLE IF EXISTS rpc_store CASCADE;

-- Then re-apply migrations
```

## Troubleshooting

### Tests are skipped

If tests are skipped with message "POSTGRES_DSN environment variable not set", make sure you've exported the environment variable:

```bash
export POSTGRES_DSN="host=localhost user=postgres password=yourpassword dbname=nitrolite_test port=5432 sslmode=disable"
```

### Connection refused

If you get "connection refused" errors:
- Verify PostgreSQL is running: `pg_isready`
- Check PostgreSQL is listening on the correct port: `netstat -an | grep 5432`
- Verify your connection parameters (host, port, user, password)

### Foreign key constraint errors

If you see foreign key constraint violations:
- Make sure all migrations have been applied
- Check that the schema matches the models in the code

### Decimal precision issues

The tests verify that NUMERIC(78, 18) fields work correctly with:
- Large numbers (up to 60 digits before decimal point)
- 18 decimal places of precision
- Negative values for net flows

If decimal tests fail, verify your migration uses `NUMERIC(78, 18)` for all balance and amount fields.
