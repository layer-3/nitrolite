import {
    Allocation,
    convertRPCToClientChannel,
    convertRPCToClientState,
    createCloseChannelMessage,
    createCreateChannelMessage,
    createResizeChannelMessage,
    NitroliteClient,
    parseChannelUpdateResponse,
    parseCloseChannelResponse,
    parseCreateChannelResponse,
    parseResizeChannelResponse,
    RPCChannelStatus,
    State,
    StateSigner,
} from '@layer-3/nitrolite';
import { Identity } from './identity';
import { Address, createPublicClient, Hex, http } from 'viem';
import { chain, CONFIG } from './setup';
import {
    getChannelUpdatePredicateWithStatus,
    getCloseChannelPredicate,
    getCreateChannelPredicate,
    TestWebSocket,
} from './ws';
import { composeResizeChannelParams } from './testHelpers';

export class TestNitroliteClient extends NitroliteClient {
    constructor(private identity: Identity, stateSigner?: StateSigner) {
        const publicClient = createPublicClient({
            chain,
            transport: http(),
        });

        if (!stateSigner) {
            stateSigner = identity.stateWalletSigner;
        }

        super({
            // @ts-ignore
            publicClient,
            walletClient: identity.walletClient,
            stateSigner: stateSigner,
            account: identity.walletClient.account,
            chainId: chain.id,
            challengeDuration: BigInt(CONFIG.DEFAULT_CHALLENGE_TIMEOUT), // min
            addresses: {
                custody: CONFIG.ADDRESSES.CUSTODY_ADDRESS,
                adjudicator: CONFIG.ADDRESSES.DUMMY_ADJUDICATOR_ADDRESS,
            },
        });
    }

    createAndWaitForChannel = async (
        ws: TestWebSocket,
        { tokenAddress, amount, depositAmount }: { tokenAddress: Address; amount: bigint; depositAmount?: bigint }
    ) => {
        const msg = await createCreateChannelMessage(this.identity.messageWalletSigner, {
            chain_id: chain.id,
            token: tokenAddress,
        });

        const createResponse = await ws.sendAndWaitForResponse(msg, getCreateChannelPredicate(), 5000);
        expect(createResponse).toBeDefined();

        const { params: createParsedResponseParams } = parseCreateChannelResponse(createResponse);


        const openChannelPromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Open),
            undefined,
            5000
        );

        depositAmount = depositAmount ?? amount;
        const { channelId, initialState } = await this.depositAndCreateChannel(tokenAddress, depositAmount, {
            unsignedInitialState: convertRPCToClientState(
                createParsedResponseParams.state,
                createParsedResponseParams.serverSignature
            ),
            channel: convertRPCToClientChannel(createParsedResponseParams.channel),
            serverSignature: createParsedResponseParams.serverSignature,
        });

        // wait for Clearnode to process channel opening
        await openChannelPromise;

        const resizeChannelPromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Open),
            undefined,
            5000
        );

        const {state} = await this.resizeChannelAndWait(ws, channelId, initialState, this.identity.walletAddress, amount);

        const resizeResponse = await resizeChannelPromise;
        const resizeParsedResponse = parseChannelUpdateResponse(resizeResponse);
        const responseChannel = resizeParsedResponse.params;

        return { params: responseChannel, state };
    };

    resizeChannelAndWait = async (
        ws: TestWebSocket,
        channelId: Hex,
        previousState: State,
        fundsDestination: Address,
        resizeAmount: bigint,
        allocateAmount: bigint = -resizeAmount,
    ) => {
        const msg = await createResizeChannelMessage(this.identity.messageWalletSigner, {
            channel_id: channelId,
            resize_amount: resizeAmount,
            allocate_amount: allocateAmount,
            funds_destination: fundsDestination,
        });

        const resizeResponse = await ws.sendAndWaitForResponse(msg, (msg: any) => {
            try {
                const parsed = parseResizeChannelResponse(msg);
                return parsed.method === 'resize_channel';
            } catch {
                return false;
            }
        }, 5000);

        const resizeParsedResponse = parseResizeChannelResponse(resizeResponse);

        const resizeChannelUpdatePromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Open),
            undefined,
            5000
        );

        const resizeStateParams = composeResizeChannelParams(channelId, resizeParsedResponse.params, previousState);
        const { resizeState } = await this.resizeChannel({
            ...resizeStateParams
        });

        const resizeChannelUpdateResponse = await resizeChannelUpdatePromise;
        const resizeChannelUpdateParsedResponse = parseChannelUpdateResponse(resizeChannelUpdateResponse);
        const responseChannel = resizeChannelUpdateParsedResponse.params;

        return { params: responseChannel, state: resizeState };
    };

    closeAndWithdrawChannel = async (ws: TestWebSocket, channelId: Hex) => {
        const msg = await createCloseChannelMessage(
            this.identity.messageWalletSigner,
            channelId,
            this.identity.walletAddress
        );

        const closeResponse = await ws.sendAndWaitForResponse(msg, getCloseChannelPredicate(), 1000);
        const closeParsedResponse = parseCloseChannelResponse(closeResponse);

        const closeChannelUpdateChannelPromise = ws.waitForMessage(
            getChannelUpdatePredicateWithStatus(RPCChannelStatus.Closed),
            undefined,
            5000
        );

        await this.closeChannel({
            finalState: {
                intent: closeParsedResponse.params.state.intent,
                channelId: closeParsedResponse.params.channelId,
                data: closeParsedResponse.params.state.stateData,
                allocations: [
                    {
                        destination: closeParsedResponse.params.state.allocations[0].destination as Address,
                        token: closeParsedResponse.params.state.allocations[0].token as Address,
                        amount: closeParsedResponse.params.state.allocations[0].amount,
                    },
                    {
                        destination: closeParsedResponse.params.state.allocations[1].destination as Address,
                        token: closeParsedResponse.params.state.allocations[1].token as Address,
                        amount: closeParsedResponse.params.state.allocations[1].amount,
                    },
                ] as [Allocation, Allocation],
                version: BigInt(closeParsedResponse.params.state.version),
                serverSignature: closeParsedResponse.params.serverSignature,
            },
            stateData: closeParsedResponse.params.state.stateData,
        });

        const closeChannelUpdateResponse = await closeChannelUpdateChannelPromise;
        const closeChannelUpdateParsedResponse = parseChannelUpdateResponse(closeChannelUpdateResponse);
        const responseChannel = closeChannelUpdateParsedResponse.params;

        return { params: responseChannel };
    };
}
