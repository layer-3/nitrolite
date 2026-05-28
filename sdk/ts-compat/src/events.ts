import type { NitroliteClient } from './client.js';
import type { LedgerChannel, LedgerBalance, ClearNodeAsset } from './types.js';

export interface EventPollerCallbacks {
    onChannelUpdate?: (channels: LedgerChannel[]) => void;
    onBalanceUpdate?: (balances: LedgerBalance[]) => void;
    /** @deprecated Receives the legacy ClearNodeAsset shape. Kept for
     *  backwards compatibility with v0.5.3 callers; new code should
     *  poll the v1 SDK's `core.Asset` surface directly. */
    onAssetsUpdate?: (assets: ClearNodeAsset[]) => void;
    onError?: (error: Error) => void;
}

/**
 * Polls the v1 Client for state changes and dispatches synthetic events
 * that match the v0.5.3 push event shapes.
 */
export class EventPoller {
    private intervalId: ReturnType<typeof setInterval> | null = null;
    private running = false;

    constructor(
        private client: NitroliteClient,
        private callbacks: EventPollerCallbacks,
        private intervalMs = 5000,
    ) {}

    start(): void {
        if (this.running) return;
        this.running = true;
        this.poll();
        this.intervalId = setInterval(() => this.poll(), this.intervalMs);
    }

    stop(): void {
        this.running = false;
        if (this.intervalId) {
            clearInterval(this.intervalId);
            this.intervalId = null;
        }
    }

    setInterval(ms: number): void {
        this.intervalMs = ms;
        if (this.running) {
            this.stop();
            this.start();
        }
    }

    private async poll(): Promise<void> {
        const [channels, balances, assets] = await Promise.allSettled([
            this.client.getChannels(),
            this.client.getBalances(),
            this.client.getAssetsList(),
        ]);

        if (channels.status === 'fulfilled') {
            this.callbacks.onChannelUpdate?.(channels.value);
        } else {
            this.callbacks.onError?.(channels.reason instanceof Error ? channels.reason : new Error(String(channels.reason)));
        }
        if (balances.status === 'fulfilled') {
            this.callbacks.onBalanceUpdate?.(balances.value);
        } else {
            this.callbacks.onError?.(balances.reason instanceof Error ? balances.reason : new Error(String(balances.reason)));
        }
        if (assets.status === 'fulfilled') {
            this.callbacks.onAssetsUpdate?.(assets.value);
        } else {
            this.callbacks.onError?.(assets.reason instanceof Error ? assets.reason : new Error(String(assets.reason)));
        }
    }
}
