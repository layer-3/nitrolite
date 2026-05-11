import { Address, getAddress } from 'viem';
import { EVMClient } from './interface.js';
import { ChannelHubAbi } from './channel_hub_abi.js';
import { ValidatorRegisteredEvent } from '../../core/event.js';

// Typed single-event ABI slice used for getLogs (historical replay).
// watchContractEvent infers types from the full ChannelHubAbi + eventName.
const VALIDATOR_REGISTERED_ABI = [
    {
        type: 'event',
        name: 'ValidatorRegistered',
        inputs: [
            { name: 'validatorId', type: 'uint8', indexed: true, internalType: 'uint8' as const },
            { name: 'validator', type: 'address', indexed: true, internalType: 'address' as const },
        ],
        anonymous: false,
    },
] as const;

/**
 * Subscribes to ValidatorRegistered events emitted by the ChannelHub contract
 * and yields them as an async stream.
 *
 * Historical replay: when fromBlock > 0n the generator fetches all matching
 * logs from fromBlock to the current chain head before switching to live events,
 * filling any gap caused by a prior outage. Pass fromBlock = 0n on the first
 * call and lastEvent.blockNumber + 1n on each reconnect.
 *
 * Reorg safety: logs with removed = true are skipped.
 *
 * Cancellation: pass an AbortSignal to stop the generator cleanly. On abort the
 * generator returns without throwing, so no error is logged for normal shutdown.
 *
 * With an HTTP transport viem polls getLogs at the configured interval (default
 * 4 s). With a WebSocket transport (wss:// URL) viem uses push subscriptions.
 */
export async function* watchValidatorRegistered(
    contractAddress: Address,
    client: EVMClient,
    blockchainId: bigint,
    fromBlock: bigint,
    signal?: AbortSignal,
): AsyncGenerator<ValidatorRegisteredEvent> {
    // Historical phase: replay events emitted while the subscription was down.
    if (fromBlock > 0n) {
        let currentBlock: bigint;
        try {
            currentBlock = await client.getBlockNumber();
        } catch (err) {
            if (!signal?.aborted) {
                console.warn('[nitrolite] watchValidatorRegistered: failed to fetch block number for historical replay, skipping gap fill', err);
            }
            currentBlock = 0n;
        }

        if (currentBlock >= fromBlock) {
            try {
                const logs = await client.getLogs({
                    address: contractAddress,
                    event: VALIDATOR_REGISTERED_ABI[0],
                    fromBlock,
                    toBlock: currentBlock,
                    strict: true,
                });
                for (const log of logs) {
                    if (log.removed) continue;
                    if (signal?.aborted) return;
                    yield {
                        blockchainId,
                        validatorId: log.args.validatorId,
                        validator: getAddress(log.args.validator),
                        blockNumber: log.blockNumber ?? currentBlock,
                    };
                }
            } catch (err) {
                if (!signal?.aborted) {
                    console.warn('[nitrolite] watchValidatorRegistered: failed to fetch historical logs, gap fill incomplete', err);
                }
            }
        }
    }

    if (signal?.aborted) return;

    // Live phase: bridge watchContractEvent callbacks into the async generator
    // using a promise queue so callers can use standard for-await-of syntax.
    const queue: ValidatorRegisteredEvent[] = [];
    let wakeUp: (() => void) | null = null;
    let watchError: Error | null = null;
    let done = false;

    const notify = (): void => {
        const resolve = wakeUp;
        wakeUp = null;
        resolve?.();
    };

    const unwatch = client.watchContractEvent({
        address: contractAddress,
        abi: ChannelHubAbi,
        eventName: 'ValidatorRegistered',
        onLogs(logs) {
            for (const log of logs) {
                if (log.removed) continue;
                const { validatorId, validator } = log.args;
                if (validatorId === undefined || !validator) continue;
                queue.push({
                    blockchainId,
                    validatorId,
                    validator: getAddress(validator),
                    blockNumber: log.blockNumber ?? 0n,
                });
            }
            notify();
        },
        onError(err) {
            watchError = err;
            done = true;
            notify();
        },
    });

    const onAbort = (): void => {
        done = true;
        notify();
    };
    signal?.addEventListener('abort', onAbort);

    try {
        while (!done || queue.length > 0) {
            while (queue.length > 0) {
                if (signal?.aborted) return;
                yield queue.shift()!;
            }
            if (!done) {
                await new Promise<void>((resolve) => {
                    wakeUp = resolve;
                });
            }
        }
        // Propagate subscription errors but not clean aborts.
        if (watchError && !signal?.aborted) {
            throw watchError;
        }
    } finally {
        signal?.removeEventListener('abort', onAbort);
        unwatch();
    }
}
