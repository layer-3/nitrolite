import { Pool } from 'pg';
import { CONFIG } from './setup';
import { createCleanupSessionKeyCacheMessage, createECDSAMessageSigner } from '@layer-3/nitrolite';
import { getCleanupSessionKeyCachePredicate, TestWebSocket } from './ws';
import { generatePrivateKey } from 'viem/accounts';

export class DatabaseUtils {
    private pool: Pool;

    constructor() {
        this.pool = new Pool({
            database: CONFIG.DATABASE_NAME,
            user: CONFIG.DATABASE_USER,
            password: CONFIG.DATABASE_PASSWORD,
            host: CONFIG.DATABASE_HOST,
            port: CONFIG.DATABASE_PORT,
        });
    }

    async cleanupDatabaseData(): Promise<void> {
        try {
            const tables = ['app_sessions', 'channels', 'contract_events', 'ledger', 'rpc_store', 'session_keys', 'ledger_transactions', 'blockchain_actions'];

            const client = await this.pool.connect();
            try {
                await client.query('BEGIN');

                await client.query('SET session_replication_role = replica');

                for (const tableName of tables) {

                    await client.query(`TRUNCATE TABLE "${tableName}" CASCADE`);
                }

                await client.query('SET session_replication_role = DEFAULT');

                await client.query('COMMIT');
            } catch (error) {
                await client.query('ROLLBACK');
                throw error;
            } finally {
                client.release();
            }
        } catch (error) {
            console.error('Error during database data cleanup:', error);
            throw error;
        }
    }

    async cleanupSessionKeyCache(): Promise<void> {
        const ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        await ws.connect();

        const randomSigner = createECDSAMessageSigner(generatePrivateKey());
        const msg = await createCleanupSessionKeyCacheMessage(randomSigner);

        try {
            await ws.sendAndWaitForResponse(msg, getCleanupSessionKeyCachePredicate(), 5000);
        } catch (error) {
            console.error('Error during cleanup session key cache:', error);
            throw error;
        } finally {
            ws.close();
        }
    }

    async resetClearnodeState(): Promise<void> {
        await this.cleanupDatabaseData();
        await this.cleanupSessionKeyCache();
    }

    async seedAsset(token: string, chainId: number, symbol: string, decimals: number): Promise<void> {
        const client = await this.pool.connect();
        try {
            await client.query(
                `INSERT INTO assets (token, chain_id, symbol, decimals)
                 VALUES ($1, $2, $3, $4)
                 ON CONFLICT (token, chain_id) DO NOTHING`,
                [token, chainId, symbol, decimals]
            );
        } finally {
            client.release();
        }
    }

    async seedLedger(walletAddress: string, accountId: string, accountType: number, assetSymbol: string, amount: number): Promise<void> {
        let credit = '0';
        let debit = '0';

        if (amount >= 0) {
            credit = amount.toString();
        } else {
            debit = Math.abs(amount).toString();
        }

        const client = await this.pool.connect();
        try {
            await client.query(
                `INSERT INTO ledger (wallet, account_id, account_type, asset_symbol, credit, debit)
                 VALUES ($1, $2, $3, $4, $5, $6)`,
                [walletAddress, accountId, accountType, assetSymbol, credit, debit]
            );
        } finally {
            client.release();
        }
    }

    async getBlockchainActions(filters: { channel_id?: string; action_type?: string }): Promise<any[]> {
        const client = await this.pool.connect();
        try {
            let query = 'SELECT * FROM blockchain_actions WHERE 1=1';
            const values: any[] = [];
            let paramIndex = 1;

            if (filters.channel_id) {
                query += ` AND channel_id = $${paramIndex++}`;
                values.push(filters.channel_id);
            }

            if (filters.action_type) {
                query += ` AND action_type = $${paramIndex++}`;
                values.push(filters.action_type);
            }

            const result = await client.query(query, values);
            return result.rows;
        } finally {
            client.release();
        }
    }

    async close(): Promise<void> {
        await this.pool.end();
    }
}
