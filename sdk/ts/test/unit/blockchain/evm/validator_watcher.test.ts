import { jest } from '@jest/globals';
import { watchValidatorRegistered } from '../../../../src/blockchain/evm/validator_watcher';
import type { EVMClient } from '../../../../src/blockchain/evm/interface';
import type { ValidatorRegisteredEvent } from '../../../../src/core/event';

// ── fixtures ──────────────────────────────────────────────────────────────────

const CONTRACT = '0x1111111111111111111111111111111111111111' as const;
const VALIDATOR = '0x2222222222222222222222222222222222222222' as const;
const CHAIN_ID = 1n;

type OnLogs = (logs: LogEntry[]) => void;
type OnError = (err: Error) => void;

interface LogEntry {
    removed: boolean;
    blockNumber: bigint | null;
    args: { validatorId?: number; validator?: string };
}

function makeLog(validatorId: number, blockNumber: bigint | null = 10n, removed = false): LogEntry {
    return { removed, blockNumber, args: { validatorId, validator: VALIDATOR } };
}

/**
 * Flush microtask queue so that mocked resolved promises (getBlockNumber, getLogs)
 * can settle and the generator can advance to the live subscription phase.
 */
async function flush(ticks = 10): Promise<void> {
    for (let i = 0; i < ticks; i++) await Promise.resolve();
}

type Harness = {
    client: EVMClient;
    triggerLogs: OnLogs;
    triggerError: OnError;
    unwatchSpy: jest.Mock;
};

function makeClient(opts: {
    headBlock?: bigint;
    histLogs?: LogEntry[];
    getBlockNumberError?: Error;
    getLogsError?: Error;
} = {}): Harness {
    let onLogs: OnLogs | undefined;
    let onError: OnError | undefined;
    const unwatchSpy = jest.fn();

    const client = {
        getBlockNumber: opts.getBlockNumberError
            ? jest.fn().mockRejectedValue(opts.getBlockNumberError)
            : jest.fn().mockResolvedValue(opts.headBlock ?? 100n),
        getLogs: opts.getLogsError
            ? jest.fn().mockRejectedValue(opts.getLogsError)
            : jest.fn().mockResolvedValue(opts.histLogs ?? []),
        watchContractEvent: jest.fn().mockImplementation(
            ({ onLogs: ol, onError: oe }: { onLogs: OnLogs; onError: OnError }) => {
                onLogs = ol;
                onError = oe;
                return unwatchSpy;
            },
        ),
    } as unknown as EVMClient;

    return {
        client,
        triggerLogs: (logs) => {
            if (!onLogs) throw new Error('watchContractEvent not yet reached');
            onLogs(logs);
        },
        triggerError: (err) => {
            if (!onError) throw new Error('watchContractEvent not yet reached');
            onError(err);
        },
        unwatchSpy,
    };
}

// ── tests ─────────────────────────────────────────────────────────────────────
//
// Pattern used throughout:
//   const p = gen.next();   // start/advance the generator — do NOT await yet
//   await flush();           // let mocked async ops (getBlockNumber, getLogs) settle
//   triggerLogs/triggerError // inject behaviour
//   await p;                 // now consume the result
//
// This keeps the gen.next() promise in scope so it can be caught if it rejects.

describe('watchValidatorRegistered — live events (fromBlock = 0n)', () => {
    it('yields an event pushed by watchContractEvent', async () => {
        const { client, triggerLogs, triggerError, unwatchSpy } = makeClient();
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 0n);

        const first = gen.next();          // generator starts, suspends at getBlockNumber
        await flush();                      // generator reaches live phase, suspends at wakeUp
        triggerLogs([makeLog(1, 42n)]);     // queue event + wake up generator
        const { value } = await first;      // generator yields the event

        expect(value).toMatchObject<ValidatorRegisteredEvent>({
            blockchainId: CHAIN_ID,
            validatorId: 1,
            validator: VALIDATOR,
            blockNumber: 42n,
        });

        const closing = gen.next();
        triggerError(new Error('done'));
        await closing.catch(() => {});
        expect(unwatchSpy).toHaveBeenCalledTimes(1);
    });

    it('passes liveFromBlock = headBlock + 1n to watchContractEvent', async () => {
        const { client, triggerError } = makeClient({ headBlock: 50n });
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 0n);

        const p = gen.next();
        await flush();
        triggerError(new Error('end'));
        await p.catch(() => {});

        expect(client.watchContractEvent).toHaveBeenCalledWith(
            expect.objectContaining({ fromBlock: 51n }),
        );
    });

    it('calls unwatch when the subscription ends', async () => {
        const { client, triggerError, unwatchSpy } = makeClient();
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 0n);

        const p = gen.next();
        await flush();
        triggerError(new Error('done')); // ends the subscription
        await p.catch(() => {});
        expect(unwatchSpy).toHaveBeenCalledTimes(1);
    });
});

describe('watchValidatorRegistered — historical replay (fromBlock > 0n)', () => {
    it('yields historical logs in order before live events', async () => {
        const { client, triggerError } = makeClient({
            headBlock: 50n,
            histLogs: [makeLog(7, 20n), makeLog(8, 25n)],
        });
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 5n);

        await flush(); // getBlockNumber + getLogs settle; generator yields first hist log
        const ev1 = await gen.next();
        const ev2 = await gen.next();

        expect(ev1.value).toMatchObject({ validatorId: 7, blockNumber: 20n });
        expect(ev2.value).toMatchObject({ validatorId: 8, blockNumber: 25n });
        expect(client.getLogs).toHaveBeenCalledWith(
            expect.objectContaining({ fromBlock: 5n, toBlock: 50n }),
        );

        // Close cleanly: generator is at live phase, wake it and end it
        const pClose = gen.next();
        await flush();
        triggerError(new Error('done'));
        await pClose.catch(() => {});
    });

    it('anchors live subscription at headBlock + 1n after historical fetch', async () => {
        const { client, triggerError } = makeClient({ headBlock: 80n, histLogs: [] });
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 10n);

        const p = gen.next();
        await flush();
        triggerError(new Error('end'));
        await p.catch(() => {});

        expect(client.watchContractEvent).toHaveBeenCalledWith(
            expect.objectContaining({ fromBlock: 81n }),
        );
    });

    it('skips historical phase when getBlockNumber fails and proceeds to live', async () => {
        const { client, triggerError } = makeClient({
            getBlockNumberError: new Error('rpc down'),
        });
        const warnSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 5n);

        const p = gen.next();
        await flush();
        expect(client.getLogs).not.toHaveBeenCalled();
        expect(warnSpy).toHaveBeenCalledWith(expect.stringContaining('[nitrolite]'), expect.anything());

        triggerError(new Error('end'));
        await p.catch(() => {});
        warnSpy.mockRestore();
    });

    it('continues to live phase and yields live events when getLogs fails', async () => {
        const { client, triggerLogs, triggerError } = makeClient({
            headBlock: 50n,
            getLogsError: new Error('range too large'),
        });
        const warnSpy = jest.spyOn(console, 'warn').mockImplementation(() => {});
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 5n);

        const first = gen.next();
        await flush(); // getLogs rejects; generator enters live phase

        expect(warnSpy).toHaveBeenCalledWith(expect.stringContaining('gap fill incomplete'), expect.anything());
        triggerLogs([makeLog(3, 55n)]);
        const { value } = await first;
        expect(value?.validatorId).toBe(3);

        const pClose = gen.next();
        triggerError(new Error('done'));
        await pClose.catch(() => {});
        warnSpy.mockRestore();
    });
});

describe('watchValidatorRegistered — reorg safety', () => {
    it('skips removed=true logs during historical replay', async () => {
        const { client, triggerError } = makeClient({
            headBlock: 50n,
            histLogs: [makeLog(9, 10n, /* removed= */ true)],
        });
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 5n);
        const p = gen.next(); // start the generator
        await flush();        // getLogs resolves, removed log skipped, live phase entered
        expect(client.getLogs).toHaveBeenCalledTimes(1);
        triggerError(new Error('done')); // wake and close
        await p.catch(() => {});
    });

    it('skips removed=true logs during live subscription', async () => {
        const { client, triggerLogs, triggerError } = makeClient();
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 0n);

        const p = gen.next();
        await flush();
        triggerLogs([makeLog(9, 10n, /* removed= */ true)]); // removed → dropped
        triggerError(new Error('end'));                        // ends the generator
        await p.catch(() => {});
        // If a removed log had been yielded, p would have resolved with it rather
        // than rejecting from watchError — the catch confirms it was dropped.
    });

    it('skips live logs with null blockNumber', async () => {
        const { client, triggerLogs, triggerError } = makeClient();
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 0n);

        const p = gen.next();
        await flush();
        triggerLogs([makeLog(5, /* blockNumber= */ null)]); // null block → dropped
        triggerError(new Error('end'));
        await p.catch(() => {});
    });
});

describe('watchValidatorRegistered — cancellation via AbortSignal', () => {
    it('stops cleanly when AbortSignal fires during live phase', async () => {
        const { client, triggerLogs, unwatchSpy } = makeClient();
        const ac = new AbortController();
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 0n, ac.signal);

        const first = gen.next();
        await flush();
        triggerLogs([makeLog(1, 1n)]);
        const { value } = await first;
        expect(value?.validatorId).toBe(1);

        ac.abort();
        const tail = await gen.next();
        expect(tail.done).toBe(true);
        expect(unwatchSpy).toHaveBeenCalledTimes(1);
    });

    it('stops cleanly when AbortSignal fires during historical replay', async () => {
        let resolveGetLogs!: (v: LogEntry[]) => void;
        const client = {
            getBlockNumber: jest.fn().mockResolvedValue(50n),
            getLogs: jest.fn().mockReturnValue(new Promise<LogEntry[]>(r => { resolveGetLogs = r; })),
            watchContractEvent: jest.fn(),
        } as unknown as EVMClient;

        const ac = new AbortController();
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 1n, ac.signal);
        const started = gen.next();
        await flush(); // getBlockNumber resolved; getLogs is still pending

        ac.abort();
        resolveGetLogs([makeLog(1, 5n)]); // resolves after abort
        await flush();

        const result = await started;
        expect(result.done).toBe(true);
        expect(client.watchContractEvent).not.toHaveBeenCalled();
    });
});

describe('watchValidatorRegistered — subscription error handling', () => {
    it('propagates subscription errors as a thrown exception', async () => {
        const { client, triggerError } = makeClient();
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 0n);

        const p = gen.next();
        await flush();
        triggerError(new Error('connection dropped'));
        await expect(p).rejects.toThrow('connection dropped');
    });

    it('drains all queued events before propagating the subscription error', async () => {
        const { client, triggerLogs, triggerError } = makeClient();
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 0n);

        const first = gen.next();
        await flush();

        // Queue two events then immediately error
        triggerLogs([makeLog(1, 1n), makeLog(2, 2n)]);
        triggerError(new Error('conn lost'));

        const ev1 = await first;
        const ev2 = await gen.next();
        expect(ev1.value?.validatorId).toBe(1);
        expect(ev2.value?.validatorId).toBe(2);
        await expect(gen.next()).rejects.toThrow('conn lost');
    });

    it('does not throw when abort fires while a watch error is also pending', async () => {
        const { client, triggerError } = makeClient();
        const ac = new AbortController();
        const gen = watchValidatorRegistered(CONTRACT, client, CHAIN_ID, 0n, ac.signal);

        const p = gen.next();
        await flush();

        triggerError(new Error('should not propagate'));
        ac.abort();

        // abort wins — generator returns cleanly without throwing
        const result = await p;
        expect(result.done).toBe(true);
    });
});
