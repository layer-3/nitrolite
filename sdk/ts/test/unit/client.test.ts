import { Decimal } from 'decimal.js';
import { jest } from '@jest/globals';
import { Client } from '../../src/client.js';
import * as core from '../../src/core/index.js';

const USER_WALLET = '0x1234567890123456789012345678901234567890' as const;
const RECIPIENT_WALLET = '0x3333333333333333333333333333333333333333' as const;
const NODE_ADDRESS = '0x1111111111111111111111111111111111111111' as const;
const TOKEN_ADDRESS = '0x2222222222222222222222222222222222222222' as const;
const HOME_CHANNEL_ID = '0xaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa';
const USER_SIGNATURE = '0x00';
const NODE_SIGNATURE = '0x01';

function createHighLevelClient(latestState?: core.State | null, latestStateError?: Error) {
    const getLatestState = jest.fn();
    if (latestStateError) {
        getLatestState.mockRejectedValue(latestStateError);
    } else {
        // Default to null (absence) when no state provided.
        getLatestState.mockResolvedValue(latestState ?? null);
    }

    const client = Object.create(Client.prototype) as any;
    client.getUserAddress = jest.fn(() => USER_WALLET);
    client.getLatestState = getLatestState;
    client.getSupportedSigValidatorsBitmap = jest.fn().mockResolvedValue('0x00');
    client.homeBlockchains = new Map([['usdc', 11155111n]]);
    client.assetStore = {
        getSuggestedBlockchainId: jest.fn().mockResolvedValue(11155111n),
        getTokenAddress: jest.fn().mockResolvedValue(TOKEN_ADDRESS),
    };
    client.getNodeAddress = jest.fn().mockResolvedValue(NODE_ADDRESS);
    client.signState = jest.fn().mockResolvedValue(USER_SIGNATURE);
    client.requestChannelCreation = jest.fn().mockResolvedValue(NODE_SIGNATURE);
    client.signAndSubmitState = jest.fn().mockImplementation(async (_current: core.State, proposed: core.State) => {
        proposed.userSig = USER_SIGNATURE;
        proposed.nodeSig = NODE_SIGNATURE;
        return NODE_SIGNATURE;
    });

    return client as Client & Record<string, any>;
}

function receivedOffchainState(): core.State {
    const state = core.newVoidState('usdc', USER_WALLET);
    state.version = 3n;
    state.homeLedger.userBalance = new Decimal(5);
    state.homeLedger.userNetFlow = new Decimal(5);
    state.homeLedger.nodeBalance = new Decimal(0);
    state.homeLedger.nodeNetFlow = new Decimal(0);
    return state;
}

function openChannelState(): core.State {
    const state = receivedOffchainState();
    state.homeChannelId = HOME_CHANNEL_ID;
    return state;
}

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

describe('Client.acknowledge', () => {
    it('creates a channel with acknowledgement when no latest state exists', async () => {
        const client = createHighLevelClient(null);

        const state = await client.acknowledge('usdc');

        expect(client.requestChannelCreation).toHaveBeenCalledTimes(1);
        expect(client.signAndSubmitState).not.toHaveBeenCalled();
        expect(state.homeChannelId).toBeDefined();
        expect(state.transition.type).toBe(core.TransitionType.Acknowledgement);
        expect(state.userSig).toBe(USER_SIGNATURE);
        expect(state.nodeSig).toBe(NODE_SIGNATURE);
    });

    it('creates a channel with acknowledgement when received off-chain funds have no home channel', async () => {
        const latestState = receivedOffchainState();
        const client = createHighLevelClient(latestState);

        const state = await client.acknowledge('usdc');

        expect(client.requestChannelCreation).toHaveBeenCalledTimes(1);
        expect(client.signAndSubmitState).not.toHaveBeenCalled();
        expect(state.version).toBe(latestState.version + 1n);
        expect(state.homeChannelId).toBeDefined();
        expect(state.homeLedger.userBalance.equals(latestState.homeLedger.userBalance)).toBe(true);
        expect(state.homeLedger.userNetFlow.equals(latestState.homeLedger.userNetFlow)).toBe(true);
        expect(state.transition.type).toBe(core.TransitionType.Acknowledgement);
        expect(state.userSig).toBe(USER_SIGNATURE);
        expect(state.nodeSig).toBe(NODE_SIGNATURE);
    });

    it('submits an acknowledgement when latest state already has a home channel', async () => {
        const latestState = openChannelState();
        const client = createHighLevelClient(latestState);

        const state = await client.acknowledge('usdc');

        expect(client.requestChannelCreation).not.toHaveBeenCalled();
        expect(client.signAndSubmitState).toHaveBeenCalledTimes(1);
        expect(state.homeChannelId).toBe(HOME_CHANNEL_ID);
        expect(state.transition.type).toBe(core.TransitionType.Acknowledgement);
        expect(state.userSig).toBe(USER_SIGNATURE);
        expect(state.nodeSig).toBe(NODE_SIGNATURE);
    });

    it('rejects an already acknowledged state on an existing home channel', async () => {
        const latestState = openChannelState();
        latestState.userSig = USER_SIGNATURE;
        const client = createHighLevelClient(latestState);

        await expect(client.acknowledge('usdc')).rejects.toThrow('state already acknowledged by user');

        expect(client.requestChannelCreation).not.toHaveBeenCalled();
        expect(client.signAndSubmitState).not.toHaveBeenCalled();
    });

    it('rejects an already acknowledged state before creating a home channel', async () => {
        const latestState = receivedOffchainState();
        latestState.userSig = USER_SIGNATURE;
        const client = createHighLevelClient(latestState);

        await expect(client.acknowledge('usdc')).rejects.toThrow('state already acknowledged by user');

        expect(client.requestChannelCreation).not.toHaveBeenCalled();
        expect(client.signAndSubmitState).not.toHaveBeenCalled();
    });
});

describe('Client absence semantics for GET methods', () => {
    it('getHomeChannel returns null when the RPC response has no channel', async () => {
        const client = Object.create(Client.prototype) as any;
        client.rpcClient = {
            channelsV1GetHomeChannel: jest.fn().mockResolvedValue({}),
        };

        const result = await client.getHomeChannel(USER_WALLET, 'usdc');

        expect(result).toBeNull();
        expect(client.rpcClient.channelsV1GetHomeChannel).toHaveBeenCalledWith({
            wallet: USER_WALLET,
            asset: 'usdc',
        });
    });

    it('getEscrowChannel returns null when the RPC response has no channel', async () => {
        const client = Object.create(Client.prototype) as any;
        client.rpcClient = {
            channelsV1GetEscrowChannel: jest.fn().mockResolvedValue({}),
        };

        const result = await client.getEscrowChannel('0xEscrow');

        expect(result).toBeNull();
        expect(client.rpcClient.channelsV1GetEscrowChannel).toHaveBeenCalledWith({
            escrow_channel_id: '0xEscrow',
        });
    });

    it('getAppDefinition returns null when the RPC response has no definition', async () => {
        const client = Object.create(Client.prototype) as any;
        client.rpcClient = {
            appSessionsV1GetAppDefinition: jest.fn().mockResolvedValue({}),
        };

        const result = await client.getAppDefinition('0xSession');

        expect(result).toBeNull();
        expect(client.rpcClient.appSessionsV1GetAppDefinition).toHaveBeenCalledWith({
            app_session_id: '0xSession',
        });
    });
});

describe('Client nil-state guards', () => {
    it('closeHomeChannel throws when no latest state exists', async () => {
        const client = Object.create(Client.prototype) as any;
        client.getUserAddress = jest.fn(() => USER_WALLET);
        client.getLatestState = jest.fn().mockResolvedValue(null);

        await expect(client.closeHomeChannel('usdc')).rejects.toThrow(
            'no channel exists for asset usdc'
        );
    });

    it('checkpoint throws when no signed state exists', async () => {
        const client = Object.create(Client.prototype) as any;
        client.getUserAddress = jest.fn(() => USER_WALLET);
        client.getLatestState = jest.fn().mockResolvedValue(null);

        await expect(client.checkpoint('usdc')).rejects.toThrow(
            'no signed state exists for asset usdc'
        );
    });

    it('submitAppSessionDeposit throws when no current state exists', async () => {
        const client = Object.create(Client.prototype) as any;
        client.getUserAddress = jest.fn(() => USER_WALLET);
        client.getLatestState = jest.fn().mockResolvedValue(null);

        const appStateUpdate = {
            appSessionId: '0xSession',
            intent: 'deposit',
            version: 2n,
            allocations: [],
        } as any;

        await expect(
            client.submitAppSessionDeposit(appStateUpdate, ['sig1'], 'usdc', new Decimal(10))
        ).rejects.toThrow('no channel state to advance for AppSession');
    });
});

describe('Client.transfer', () => {
    it('creates a channel with transfer when no latest state exists', async () => {
        const client = createHighLevelClient(null);
        const amount = new Decimal(1);

        const state = await client.transfer(RECIPIENT_WALLET, 'usdc', amount);

        expect(client.requestChannelCreation).toHaveBeenCalledTimes(1);
        expect(client.signAndSubmitState).not.toHaveBeenCalled();
        expect(state.homeChannelId).toBeDefined();
        expect(state.transition.type).toBe(core.TransitionType.TransferSend);
        expect(state.transition.accountId).toBe(RECIPIENT_WALLET);
        expect(state.userSig).toBe(USER_SIGNATURE);
        expect(state.nodeSig).toBe(NODE_SIGNATURE);
    });

    it('creates a channel with transfer when received off-chain funds have no home channel', async () => {
        const latestState = receivedOffchainState();
        const client = createHighLevelClient(latestState);
        const amount = new Decimal(2);

        const state = await client.transfer(RECIPIENT_WALLET, 'usdc', amount);

        expect(client.requestChannelCreation).toHaveBeenCalledTimes(1);
        expect(client.signAndSubmitState).not.toHaveBeenCalled();
        expect(state.version).toBe(latestState.version + 1n);
        expect(state.homeChannelId).toBeDefined();
        expect(state.homeLedger.userBalance.equals(latestState.homeLedger.userBalance.sub(amount))).toBe(true);
        expect(state.homeLedger.userNetFlow.equals(latestState.homeLedger.userNetFlow)).toBe(true);
        expect(state.transition.type).toBe(core.TransitionType.TransferSend);
        expect(state.transition.accountId).toBe(RECIPIENT_WALLET);
        expect(state.userSig).toBe(USER_SIGNATURE);
        expect(state.nodeSig).toBe(NODE_SIGNATURE);
    });

    it('submits a transfer when latest state already has a home channel', async () => {
        const latestState = openChannelState();
        const client = createHighLevelClient(latestState);
        const amount = new Decimal(2);

        const state = await client.transfer(RECIPIENT_WALLET, 'usdc', amount);

        expect(client.requestChannelCreation).not.toHaveBeenCalled();
        expect(client.signAndSubmitState).toHaveBeenCalledTimes(1);
        expect(state.homeChannelId).toBe(HOME_CHANNEL_ID);
        expect(state.homeLedger.userBalance.equals(latestState.homeLedger.userBalance.sub(amount))).toBe(true);
        expect(state.transition.type).toBe(core.TransitionType.TransferSend);
        expect(state.transition.accountId).toBe(RECIPIENT_WALLET);
        expect(state.userSig).toBe(USER_SIGNATURE);
        expect(state.nodeSig).toBe(NODE_SIGNATURE);
    });
});

describe('Client.deposit cross-chain guard', () => {
    it('rejects deposit when active home channel is on a different chain', async () => {
        const latestState = openChannelState();
        latestState.homeLedger.blockchainId = 137n;
        const client = createHighLevelClient(latestState);

        await expect(client.deposit(8453n, 'usdc', new Decimal(1))).rejects.toThrow(
            'active home channel for asset "usdc" is on chain 137, cannot deposit on chain 8453'
        );
        expect(client.signAndSubmitState).not.toHaveBeenCalled();
        expect(client.requestChannelCreation).not.toHaveBeenCalled();
    });

    it('allows deposit when active home channel matches the supplied chain', async () => {
        const latestState = openChannelState();
        latestState.homeLedger.blockchainId = 8453n;
        latestState.homeLedger.tokenAddress = TOKEN_ADDRESS;
        const client = createHighLevelClient(latestState);

        await client.deposit(8453n, 'usdc', new Decimal(1));
        expect(client.signAndSubmitState).toHaveBeenCalledTimes(1);
        expect(client.requestChannelCreation).not.toHaveBeenCalled();
    });
});

describe('Client.withdraw cross-chain guard', () => {
    it('rejects withdraw when active home channel is on a different chain', async () => {
        const latestState = openChannelState();
        latestState.homeLedger.blockchainId = 137n;
        const client = createHighLevelClient(latestState);

        await expect(client.withdraw(8453n, 'usdc', new Decimal(1))).rejects.toThrow(
            'active home channel for asset "usdc" is on chain 137, cannot withdraw on chain 8453'
        );
        expect(client.signAndSubmitState).not.toHaveBeenCalled();
        expect(client.requestChannelCreation).not.toHaveBeenCalled();
    });

    it('allows withdraw when active home channel matches the supplied chain', async () => {
        const latestState = openChannelState();
        latestState.homeLedger.blockchainId = 8453n;
        latestState.homeLedger.tokenAddress = TOKEN_ADDRESS;
        const client = createHighLevelClient(latestState);

        await client.withdraw(8453n, 'usdc', new Decimal(1));
        expect(client.signAndSubmitState).toHaveBeenCalledTimes(1);
        expect(client.requestChannelCreation).not.toHaveBeenCalled();
    });
});
