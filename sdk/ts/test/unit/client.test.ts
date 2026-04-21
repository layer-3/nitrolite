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

    it('waits for blockchain client initialization before reading the balance', async () => {
        const chainId = 11155111n;
        const wallet = '0x1234567890123456789012345678901234567890' as const;
        const expected = new Decimal('7.5');
        let resolveInit: (() => void) | undefined;
        const initializeBlockchainClient = jest.fn().mockImplementation(
            () =>
                new Promise<void>((resolve) => {
                    resolveInit = () => {
                        client.blockchainClients.set(chainId, { getTokenBalance });
                        resolve();
                    };
                })
        );
        const getTokenBalance = jest.fn().mockResolvedValue(expected);

        const client = Object.create(Client.prototype) as Client & {
            blockchainClients: Map<bigint, { getTokenBalance: typeof getTokenBalance }>;
            initializeBlockchainClient: typeof initializeBlockchainClient;
        };

        client.blockchainClients = new Map();
        client.initializeBlockchainClient = initializeBlockchainClient;

        const balancePromise = client.getOnChainBalance(chainId, 'usdc', wallet);

        expect(initializeBlockchainClient).toHaveBeenCalledWith(chainId);
        expect(getTokenBalance).not.toHaveBeenCalled();

        resolveInit?.();

        await expect(balancePromise).resolves.toBe(expected);
        expect(getTokenBalance).toHaveBeenCalledWith('usdc', wallet);
    });

    it('propagates balance-read failures from the blockchain client', async () => {
        const chainId = 11155111n;
        const wallet = '0x1234567890123456789012345678901234567890' as const;
        const expectedError = new Error('unknown asset');
        const getTokenBalance = jest.fn().mockRejectedValue(expectedError);
        const initializeBlockchainClient = jest.fn().mockResolvedValue(undefined);

        const client = Object.create(Client.prototype) as Client & {
            blockchainClients: Map<bigint, { getTokenBalance: typeof getTokenBalance }>;
            initializeBlockchainClient: typeof initializeBlockchainClient;
        };

        client.blockchainClients = new Map([[chainId, { getTokenBalance }]]);
        client.initializeBlockchainClient = initializeBlockchainClient;

        await expect(client.getOnChainBalance(chainId, 'usdc', wallet)).rejects.toThrow(
            'unknown asset'
        );
        expect(initializeBlockchainClient).toHaveBeenCalledWith(chainId);
        expect(getTokenBalance).toHaveBeenCalledWith('usdc', wallet);
    });
});
