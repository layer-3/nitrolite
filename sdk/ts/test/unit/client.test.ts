import { Decimal } from 'decimal.js';
import { jest } from '@jest/globals';
import { Client } from '../../src/client.js';

describe('Client.getOnChainBalance', () => {
    it('delegates to the initialized blockchain client for the requested chain', async () => {
        const chainId = 11155111n;
        const wallet = '0x1234567890123456789012345678901234567890' as const;
        const expected = new Decimal('12.345');
        const getTokenBalance = jest.fn().mockResolvedValue(expected);
        const initializeBlockchainClient = jest.fn().mockResolvedValue(undefined);

        const client = Object.create(Client.prototype) as Client & {
            blockchainClients: Map<bigint, { getTokenBalance: typeof getTokenBalance }>;
            initializeBlockchainClient: typeof initializeBlockchainClient;
        };

        client.blockchainClients = new Map([[chainId, { getTokenBalance }]]);
        client.initializeBlockchainClient = initializeBlockchainClient;

        const balance = await client.getOnChainBalance(chainId, 'usdc', wallet);

        expect(initializeBlockchainClient).toHaveBeenCalledWith(chainId);
        expect(getTokenBalance).toHaveBeenCalledWith('usdc', wallet);
        expect(balance).toBe(expected);
    });
});
