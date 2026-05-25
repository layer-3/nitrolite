import { useCallback, useEffect, useState } from 'react';
import type { Client, Channel } from '@yellow-org/sdk';
import type { Address } from 'viem';

export interface UseChannelsResult {
  channels: Channel[];
  isLoading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
}

export function useChannels(client: Client | null, address: Address | null): UseChannelsResult {
  const [channels, setChannels] = useState<Channel[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    if (!client || !address) {
      setChannels([]);
      return;
    }
    setIsLoading(true);
    setError(null);
    try {
      const res = await client.getChannels(address);
      if (res.metadata && res.metadata.totalCount > res.channels.length) {
        console.warn(
          `[playground] getChannels returned ${res.channels.length} of ${res.metadata.totalCount} total — pagination not implemented`,
        );
      }
      setChannels(res.channels);
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    } finally {
      setIsLoading(false);
    }
  }, [client, address]);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return { channels, isLoading, error, refresh };
}
