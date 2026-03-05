import { createAuthSessionWithClearnode } from '@/auth';
import { BlockchainUtils } from '@/blockchainUtils';
import { DatabaseUtils } from '@/databaseUtils';
import { Identity } from '@/identity';
import { TestNitroliteClient } from '@/nitroliteClient';
import { chain, CONFIG } from '@/setup';
import { composeResizeChannelParams } from '@/testHelpers';
import { getChannelUpdatePredicateWithStatus, getCreateChannelPredicate, getResizeChannelPredicate, TestWebSocket } from '@/ws';
import {
    convertRPCToClientChannel,
    convertRPCToClientState,
    createCreateChannelMessage,
    createResizeChannelMessage,
    generateRequestId,
    getCurrentTimestamp,
    NitroliteRPC,
    NitroliteRPCMessage,
    parseChannelUpdateResponse,
    parseCreateChannelResponse,
    parseResizeChannelResponse,
    RPCChannelStatus,
    RPCData,
    RPCMethod,
} from '@layer-3/nitrolite';
import { Hex, parseUnits } from 'viem';

describe('Create channel', () => {
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

    it('should create nitrolite client to open channels', async () => {
        client = new TestNitroliteClient(identity);

        expect(client).toBeDefined();
        expect(client).toHaveProperty('approveTokens');
        expect(client).toHaveProperty('deposit');
        expect(client).toHaveProperty('createChannel');
        expect(client).toHaveProperty('depositAndCreateChannel');
    });

    it('returned state must have zero first allocation even if amount was specified in the request', async () => {
        // manually create ws request
        const requestId = generateRequestId();
        const timestamp = getCurrentTimestamp();
        const params = {
            chain_id: chain.id,
            token: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            amount: 42,  // invalid, allocation must be zero
        };
        const signatures: Hex[] = [];
        const requestData: RPCData = [requestId, RPCMethod.CreateChannel, params, timestamp];
        const message: NitroliteRPCMessage = { req: requestData, sig: signatures };

        const signedRequest = await NitroliteRPC.signRequestMessage(message, identity.messageWalletSigner);
        const msg = JSON.stringify(signedRequest, (_, value) => (typeof value === 'bigint' ? value.toString() : value));

        const msgResponse = await ws.sendAndWaitForResponse(msg, getCreateChannelPredicate(), 5000);
        const response = parseCreateChannelResponse(msgResponse); // to ensure response is valid RPC

        expect(response).toBeDefined();
        expect(response.params).toBeDefined();
        expect(response.params.state).toBeDefined();
        expect(response.params.state.allocations).toBeDefined();
        expect(response.params.state.allocations[0]).toBeDefined();
        expect(response.params.state.allocations[0].amount).toBe(BigInt(0)); // must be zero
    });

    it('should approve, deposit, and create channel (with zero allocation) in one operation', async () => {
        const prevBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            identity.walletAddress
        );

        const msg = await createCreateChannelMessage(identity.messageWalletSigner, {
            chain_id: chain.id,
            token: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
        });
        const createResponse = await ws.sendAndWaitForResponse(msg, getCreateChannelPredicate(), 5000);
        expect(createResponse).toBeDefined();

        const { params: createParsedResponseParams } = parseCreateChannelResponse(createResponse);

        const openChannelPromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Open),
            undefined,
            5000
        );

        const { channelId, initialState, txHash } = await client.depositAndCreateChannel(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            depositAmount,
            {
                unsignedInitialState: convertRPCToClientState(
                    createParsedResponseParams.state,
                    createParsedResponseParams.serverSignature
                ),
                channel: convertRPCToClientChannel(createParsedResponseParams.channel),
                serverSignature: createParsedResponseParams.serverSignature,
            }
        );

        expect(channelId).toBeDefined();
        expect(initialState).toBeDefined();
        expect(txHash).toBeDefined();

        const receipt = await blockUtils.waitForTransaction(txHash);
        expect(receipt).toBeDefined();

        const openResponse = await openChannelPromise;
        expect(openResponse).toBeDefined();

        const openParsedResponse = parseChannelUpdateResponse(openResponse);
        const responseChannel = openParsedResponse.params;

        expect(responseChannel.adjudicator).toBe(CONFIG.ADDRESSES.DUMMY_ADJUDICATOR_ADDRESS);
        expect(responseChannel.amount).toBe(BigInt(0)); // initial channel amount is zero
        expect(responseChannel.chainId).toBe(CONFIG.CHAIN_ID);
        expect(responseChannel.challenge).toBe(CONFIG.DEFAULT_CHALLENGE_TIMEOUT);
        expect(responseChannel.channelId).toBe(channelId);
        expect(responseChannel.participant).toBe(identity.walletAddress);
        expect(responseChannel.status).toBe(RPCChannelStatus.Open);
        expect(responseChannel.token).toBe(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);

        const postBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            identity.walletAddress
        );

        expect(postBalance.rawBalance).toBe(prevBalance.rawBalance - depositAmount);

        const channelBalance = await blockUtils.getChannelBalance(
            CONFIG.ADDRESSES.CUSTODY_ADDRESS,
            channelId,
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS
        );

        expect(channelBalance.rawBalance).toBe(BigInt(0)); // channel balance must be zero
    });

    it('should approve, deposit, create channel and resize (with NON zero allocation) in several operation', async () => {
        const prevBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            identity.walletAddress
        );

        const approveTxHash = await client.approveTokens(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS, depositAmount);

        const approveReceipt = await blockUtils.waitForTransaction(approveTxHash);
        expect(approveReceipt).toBeDefined();

        const depositTxHash = await client.deposit(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS, depositAmount);
        expect(depositTxHash).toBeDefined();

        const depositReceipt = await blockUtils.waitForTransaction(depositTxHash);
        expect(depositReceipt).toBeDefined();

        const postBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            identity.walletAddress
        );

        expect(postBalance.rawBalance).toBe(prevBalance.rawBalance - depositAmount);

        const msg = await createCreateChannelMessage(identity.messageWalletSigner, {
            chain_id: chain.id,
            token: CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
        });
        const createResponse = await ws.sendAndWaitForResponse(msg, getCreateChannelPredicate(), 5000);
        expect(createResponse).toBeDefined();

        const { params: createParsedResponseParams } = parseCreateChannelResponse(createResponse);

        const openChannelPromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Open),
            undefined,
            5000
        );

        const {
            txHash: createChannelTxHash,
            channelId,
            initialState,
        } = await client.createChannel({
            unsignedInitialState: convertRPCToClientState(
                createParsedResponseParams.state,
                createParsedResponseParams.serverSignature
            ),
            channel: convertRPCToClientChannel(createParsedResponseParams.channel),
            serverSignature: createParsedResponseParams.serverSignature,
        });

        expect(channelId).toBeDefined();
        expect(initialState).toBeDefined();
        expect(createChannelTxHash).toBeDefined();

        const createChannelReceipt = await blockUtils.waitForTransaction(createChannelTxHash);
        expect(createChannelReceipt).toBeDefined();

        const openResponse = await openChannelPromise;
        expect(openResponse).toBeDefined();

        const openParsedResponse = parseChannelUpdateResponse(openResponse);
        const responseChannel = openParsedResponse.params;

        expect(responseChannel.adjudicator).toBe(CONFIG.ADDRESSES.DUMMY_ADJUDICATOR_ADDRESS);
        expect(responseChannel.amount).toBe(BigInt(0)); // initial channel amount is zero
        expect(responseChannel.chainId).toBe(CONFIG.CHAIN_ID);
        expect(responseChannel.challenge).toBe(CONFIG.DEFAULT_CHALLENGE_TIMEOUT);
        expect(responseChannel.channelId).toBe(channelId);
        expect(responseChannel.participant).toBe(identity.walletAddress);
        expect(responseChannel.status).toBe(RPCChannelStatus.Open);
        expect(responseChannel.token).toBe(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS);

        const resizeMsgParams = {
            channel_id: channelId,
            resize_amount: depositAmount,
            funds_destination: identity.walletAddress,
        }
        const resizeMsg = await createResizeChannelMessage(identity.messageWalletSigner, {
            ...resizeMsgParams
        });
        const resizeResponse = await ws.sendAndWaitForResponse(resizeMsg, getResizeChannelPredicate(), 5000);
        expect(resizeResponse).toBeDefined();

        const { params: resizeParsedResponseParams } = parseResizeChannelResponse(resizeResponse);
        const resizeParams = composeResizeChannelParams(
            channelId,
            resizeParsedResponseParams,
            initialState,
        );

        const {txHash} = await client.resizeChannel(resizeParams);

        expect(txHash).toBeDefined();

        const resizeReceipt = await blockUtils.waitForTransaction(txHash);
        expect(resizeReceipt).toBeDefined();

        const afterResizeBalance = await blockUtils.getErc20Balance(
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
            identity.walletAddress
        );

        expect(afterResizeBalance.rawBalance).toBe(prevBalance.rawBalance - depositAmount);

        const channelBalance = await blockUtils.getChannelBalance(
            CONFIG.ADDRESSES.CUSTODY_ADDRESS,
            channelId,
            CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS
        );

        expect(channelBalance.rawBalance).toBe(depositAmount);
    });

    // TODO: find a way to know that broker decided not to join channel
    // it('should restrict opening several channels', async () => {
    //     const depositTxHash = await client.deposit(CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS, depositAmount * BigInt(3));
    //     expect(depositTxHash).toBeDefined();

    //     const depositReceipt = await blockUtils.waitForTransaction(depositTxHash);
    //     expect(depositReceipt).toBeDefined();

    //     // previously we deposited 3 * depositAmount, so we can create two channels
    //     for (const amount of [depositAmount, depositAmount * BigInt(2)]) {
    //         const openChannelPromise = ws.waitForMessage(getChannelUpdatePredicateWithStatus(RPCChannelStatus.Open), 5000);

    //         const { txHash: createChannelTxHash, channelId } = await client.createChannel(
    //             CONFIG.ADDRESSES.USDC_TOKEN_ADDRESS,
    //             {
    //                 initialAllocationAmounts: [amount, BigInt(0)],
    //                 stateData: '0x',
    //             }
    //         );

    //         expect(channelId).toBeDefined();
    //         expect(createChannelTxHash).toBeDefined();

    //         const createChannelReceipt = await blockUtils.waitForTransaction(createChannelTxHash);
    //         expect(createChannelReceipt).toBeDefined();

    //         const openResponse = await openChannelPromise;

    //         if (amount === depositAmount) {
    //             expect(openResponse).toBeDefined();
    //         } else {
    //             expect(openResponse).toBeUndefined();
    //             continue; // Skip further checks for the second channel
    //         }

    //         expect(openResponse).toBeDefined();

    //         const openParsedResponse = parseRPCResponse(openResponse) as ChannelsUpdateRPCResponse;
    //         const responseChannel = openParsedResponse.params[0];

    //         expect(responseChannel.amount).toBe(Number(amount));
    //         expect(responseChannel.channel_id).toBe(channelId);
    //     }
    // });
});
