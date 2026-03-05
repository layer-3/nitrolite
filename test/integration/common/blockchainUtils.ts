import {
    WaitForTransactionReceiptParameters,
    Hash,
    createPublicClient,
    http,
    Address,
    erc20Abi,
    formatUnits,
    createTestClient,
    TestClient,
    Hex,
    createWalletClient,
    TransactionSerializable,
    serializeTransaction,
    Signature,
    GetTxpoolContentReturnType,
} from 'viem';
import { chain } from './setup';
import { privateKeyToAccount } from 'viem/accounts';
import { custodyAbi } from '@layer-3/nitrolite/dist/abis/generated';

export class BlockchainUtils {
    private client = null;
    private testClient: TestClient = null;
    private lastSnapshotId: Hex | null = null;

    constructor() {
        this.client = createPublicClient({
            chain,
            transport: http(),
        });

        this.testClient = createTestClient({
            chain,
            transport: http(),
            mode: 'anvil',
        });
    }

    async waitForTransaction(
        txHash: Hash,
        timeoutMs: number = 5000,
        confirmations: number = 0
    ): Promise<WaitForTransactionReceiptParameters> {
        try {
            const timeoutPromise = new Promise((_, reject) => {
                setTimeout(() => {
                    reject(new Error(`Transaction wait timeout after ${timeoutMs}ms`));
                }, timeoutMs);
            });

            const receiptPromise = this.client.waitForTransactionReceipt({
                hash: txHash,
                confirmations,
            });

            const receipt = await Promise.race([receiptPromise, timeoutPromise]);
            return receipt as WaitForTransactionReceiptParameters;
        } catch (error) {
            throw new Error(`Error waiting for transaction: ${error.message}`);
        }
    }

    async getBalance(address: `0x${string}`): Promise<bigint> {
        try {
            const balance = await this.client.getBalance({ address });
            return balance;
        } catch (error) {
            throw new Error(`Error getting balance: ${error.message}`);
        }
    }

    async getErc20Balance(
        tokenAddress: Address,
        userAddress: Address,
        decimals?: number
    ): Promise<{ rawBalance: bigint; formattedBalance: string }> {
        try {
            // Get balance
            const balance = await this.client.readContract({
                address: tokenAddress,
                abi: erc20Abi,
                functionName: 'balanceOf',
                args: [userAddress],
            });

            const tokenDecimals =
                decimals ??
                (await this.client.readContract({
                    address: tokenAddress,
                    abi: erc20Abi,
                    functionName: 'decimals',
                }));

            return {
                rawBalance: balance,
                formattedBalance: formatUnits(balance, tokenDecimals),
            };
        } catch (error) {
            throw new Error(`Error getting ERC20 balance: ${error.message}`);
        }
    }

    async getChannelBalance(
        custodyAddress: Address,
        channelId: Hash,
        tokenAddress: Address,
        decimals?: number
    ): Promise<{ rawBalance: bigint; formattedBalance: string }> {
        try {
            // Get channel balance
            const balances = await this.client.readContract({
                address: custodyAddress,
                abi: custodyAbi,
                functionName: 'getChannelBalances',
                args: [channelId, [tokenAddress]],
            });

            const channelBalance = balances[0];

            const tokenDecimals =
                decimals ??
                (await this.client.readContract({
                    address: tokenAddress,
                    abi: erc20Abi,
                    functionName: 'decimals',
                }));

            return {
                rawBalance: channelBalance,
                formattedBalance: formatUnits(channelBalance, tokenDecimals),
            };
        } catch (error) {
            throw new Error(`Error getting channel balance: ${error.message}`);
        }
    }

    async makeSnapshot(): Promise<Hex> {
        try {
            const snapshotId = await this.testClient.snapshot();
            this.lastSnapshotId = snapshotId;

            return snapshotId;
        } catch (error) {
            throw new Error(`Error making snapshot: ${error.message}`);
        }
    }

    async resetSnapshot(snapshotId?: Hex): Promise<void> {
        try {
            if (!snapshotId && !this.lastSnapshotId) {
                throw new Error('No snapshot ID provided and no last snapshot available');
            }

            snapshotId = snapshotId || this.lastSnapshotId!;
            await this.testClient.revert({ id: snapshotId });
        } catch (error) {
            throw new Error(`Error resetting snapshot: ${error.message}`);
        }
    }

    async pauseMining(): Promise<void> {
        try {
            await this.testClient.setAutomine(false);
        } catch (error) {
            throw new Error(`Error pausing mining: ${error.message}`);
        }
    }

    async resumeMining(): Promise<void> {
        try {
            await this.testClient.setAutomine(true);
        } catch (error) {
            throw new Error(`Error resuming mining: ${error.message}`);
        }
    }

    async mineBlock(): Promise<void> {
        try {
            await this.testClient.mine({ blocks: 1 });
        } catch (error) {
            throw new Error(`Error mining block: ${error.message}`);
        }
    }

    async readTxPool(): Promise<GetTxpoolContentReturnType> {
        try {
            const content = await this.testClient.getTxpoolContent();
            return content;
        } catch (error) {
            throw new Error(`Error reading transaction pool: ${error.message}`);
        }
    }

    async sendRawTransactionAs(pk: Hex, tx: TransactionSerializable, sig: Signature): Promise<Hash> {
        try {
            const account = privateKeyToAccount(pk);
            const walletClient = createWalletClient({
                account,
                chain,
                transport: http(),
            });

            const serializedTx = serializeTransaction(tx, sig);

            const txHash = await walletClient.sendRawTransaction({
                serializedTransaction: serializedTx,
            });
            return txHash;
        } catch (error) {
            throw new Error(`Error sending transaction: ${error.message}`);
        }
    }

    async dropTxFromPool(txHash: Hash): Promise<void> {
        try {
            await this.testClient.dropTransaction({ hash: txHash });
        } catch (error) {
            throw new Error(`Error dropping transactions: ${error.message}`);
        }
    }
}
