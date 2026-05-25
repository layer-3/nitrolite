// Fallback public RPC URLs per chain ID. The Nitronode tells us which chains it supports
// via getConfig(); for each one we need an RPC for the SDK's on-chain reads / writes.
// Until the SDK gains a withEIP1193Provider option, we use these public endpoints.

export const PUBLIC_RPC_URLS: Record<string, string> = {
  '1': 'https://ethereum-rpc.publicnode.com',
  '11155111': 'https://ethereum-sepolia-rpc.publicnode.com',
  '137': 'https://polygon-rpc.com',
  '80002': 'https://rpc-amoy.polygon.technology',
  '8453': 'https://mainnet.base.org',
  '84532': 'https://sepolia.base.org',
  '42161': 'https://arb1.arbitrum.io/rpc',
  '421614': 'https://sepolia-rollup.arbitrum.io/rpc',
  '59144': 'https://rpc.linea.build',
  '59141': 'https://rpc.sepolia.linea.build',
};

export function rpcUrlFor(chainId: bigint): string | undefined {
  return PUBLIC_RPC_URLS[chainId.toString()];
}
