import { createAuthSessionWithClearnode } from '@/auth';
import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG } from '@/setup';
import { getChannelUpdatePredicateWithStatus, TestWebSocket } from '@/ws';
import {
    parseChannelUpdateResponse,
    RPCChannelStatus,
} from '@layer-3/nitrolite';
import { parseUnits } from 'viem';

describe('Challenge channel', () => {
    const depositAmount = parseUnits('100', 6); // 100 USDC (decimals = 6)

    let ws: TestWebSocket;
    let identity: Identity;
    let client: TestNitroliteClient;
    let blockUtils: BlockchainUtils;
    let databaseUtils: DatabaseUtils;

    beforeAll(async () => {
        blockUtils = new BlockchainUtils();
        databaseUtils = new DatabaseUtils();
        identity = new Identity(CONFIG.IDENTITIES[0].WALLET_PK, CONFIG.IDENTITIES[0].SESSION_PK);
        ws = new TestWebSocket(CONFIG.CLEARNODE_URL, CONFIG.DEBUG_MODE);
        client = new TestNitroliteClient(identity);
    });

    beforeEach(async () => {
        await ws.connect();
        await createAuthSessionWithClearnode(ws, identity);
        await blockUtils.makeSnapshot();
    });

    afterEach(async () => {
        ws.close();
        await databaseUtils.resetClearnodeState();
        await blockUtils.resetSnapshot();
    });

    afterAll(() => {
        databaseUtils.close();
    });

    it('should mark channel as challenged when valid current state is challenged', async () => {
        const { params: createResponse, state: initialState } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: depositAmount * BigInt(10),
        });

        const channelId = createResponse.channelId;
        expect(createResponse.version).toBe(1); // 1 because channel was resized as well

        const challengeReceipt = await client.challengeChannel({
            channelId: channelId,
            candidateState: {
                intent: initialState.intent,
                version: BigInt(createResponse.version),
                data: initialState.data,
                allocations: initialState.allocations,
                sigs: initialState.sigs,
            },
        });

        expect(challengeReceipt).toBeDefined();

        const challengedChannelUpdatePromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Challenged),
            5000
        );

        const challengeConfirmation = await blockUtils.waitForTransaction(challengeReceipt);
        expect(challengeConfirmation).toBeDefined();

        const challengedChannelUpdateResponse = await challengedChannelUpdatePromise;
        const challengedChannelUpdate = parseChannelUpdateResponse(challengedChannelUpdateResponse);

        // Verify channel is marked as challenged
        expect(challengedChannelUpdate.params.channelId).toBe(channelId);
        expect(challengedChannelUpdate.params.status).toBe(RPCChannelStatus.Challenged);
        expect(challengedChannelUpdate.params.version).toBe(createResponse.version);

        // Verify no blockchain actions were created in the database
        const blockchainActions = await databaseUtils.getBlockchainActions({
            channel_id: channelId,
            action_type: 'checkpoint',
        });

        expect(blockchainActions).toHaveLength(0);
    }, 30000);
});
