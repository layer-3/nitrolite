import { createAuthSessionWithClearnode } from '@/auth';
import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { CONFIG } from '@/setup';
import { composeResizeChannelParams } from '@/testHelpers';
import { getCloseChannelPredicate, getResizeChannelPredicate, TestWebSocket } from '@/ws';
import {
    createResizeChannelMessage,
    parseResizeChannelResponse,
    createCloseChannelMessage,
    parseCloseChannelResponse,
} from '@layer-3/nitrolite';
import { Hex, parseUnits } from 'viem';

describe('Close channel', () => {
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

    it('should create nitrolite client to close channels', async () => {
        client = new TestNitroliteClient(identity);

        expect(client).toBeDefined();
        expect(client).toHaveProperty('depositAndCreateChannel');
        expect(client).toHaveProperty('closeChannel');
        expect(client).toHaveProperty('withdrawal');
    });

    it('should resize, close channel and withdraw funds', async () => {
        const preFundBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            identity.walletAddress
        );

        const { params, state } = await client.createAndWaitForChannel(ws, {
            tokenAddress: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: depositAmount,
        });

        const postFundBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            identity.walletAddress
        );

        expect(postFundBalance.rawBalance).toBe(preFundBalance.rawBalance - depositAmount);

        let msg = await createResizeChannelMessage(identity.messageSKSigner, {
            channel_id: params.channelId,
            allocate_amount: depositAmount,
            resize_amount: -depositAmount,
            funds_destination: identity.walletAddress,
        });

        const resizeResponse = await ws.sendAndWaitForResponse(msg, getResizeChannelPredicate(), 1000);
        const { params: resizeResponseParams } = parseResizeChannelResponse(resizeResponse);

        expect(resizeResponseParams.state.allocations).toBeDefined();
        expect(resizeResponseParams.state.allocations).toHaveLength(2);
        expect(String(resizeResponseParams.state.allocations[0].destination)).toBe(identity.walletAddress);
        expect(String(resizeResponseParams.state.allocations[0].amount)).toBe('0');
        expect(String(resizeResponseParams.state.allocations[1].destination)).toBe(CONFIG.ADDRESSES.CLEARNODE_ADDRESS);
        expect(String(resizeResponseParams.state.allocations[1].amount)).toBe('0');

        const {txHash: resizeChannelTxHash} = await client.resizeChannel({
            ...composeResizeChannelParams(
                resizeResponseParams.channelId as Hex,
                resizeResponseParams,
                state,
            ),
        });
        expect(resizeChannelTxHash).toBeDefined();

        const resizeReceipt = await blockUtils.waitForTransaction(resizeChannelTxHash);
        expect(resizeReceipt).toBeDefined();

        msg = await createCloseChannelMessage(identity.messageSKSigner, params.channelId, identity.walletAddress);
        const closeResponse = await ws.sendAndWaitForResponse(msg, getCloseChannelPredicate(), 1000);
        expect(closeResponse).toBeDefined();

        const closeParsedResponse = parseCloseChannelResponse(closeResponse);

        const closeChannelTxHash = await client.closeChannel({
            finalState: {
                intent: closeParsedResponse.params.state.intent,
                channelId: closeParsedResponse.params.channelId,
                data: closeParsedResponse.params.state.stateData as Hex,
                allocations: closeParsedResponse.params.state.allocations,
                version: BigInt(closeParsedResponse.params.state.version),
                serverSignature: closeParsedResponse.params.serverSignature,
            },
            stateData: closeParsedResponse.params.state.stateData as Hex,
        });
        expect(closeChannelTxHash).toBeDefined();

        const closeReceipt = await blockUtils.waitForTransaction(closeChannelTxHash);
        expect(closeReceipt).toBeDefined();

        // Close should not change wallet balance
        expect(postFundBalance.rawBalance).toBe(preFundBalance.rawBalance - depositAmount);

        const withdrawalTxHash = await client.withdrawal(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS, depositAmount);
        expect(withdrawalTxHash).toBeDefined();

        const withdrawalReceipt = await blockUtils.waitForTransaction(withdrawalTxHash);
        expect(withdrawalReceipt).toBeDefined();

        const postWithdrawalBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            identity.walletAddress
        );
        expect(postWithdrawalBalance.rawBalance).toBe(preFundBalance.rawBalance);
    });
});
