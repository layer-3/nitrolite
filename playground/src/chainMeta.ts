// Canonical display names and sort order for known chains.
// Used to override the node-provided names in the UI and to order chains consistently.

export const CHAIN_DISPLAY_NAMES: Record<string, string> = {
  '1':       'Ethereum',
  '11155111':'Ethereum Sepolia',
  '137':     'Polygon',
  '80002':   'Polygon Amoy',
  '8453':    'Base',
  '84532':   'Base Sepolia',
  '42161':   'Arbitrum One',
  '421614':  'Arbitrum Sepolia',
  '59144':   'Linea',
  '59141':   'Linea Sepolia',
  '10':      'Optimism',
  '56':      'BNB Chain',
  '43114':   'Avalanche',
  '11235':   'HAQQ',
  '1449000': 'XRPL EVM Testnet',
};

/** Returns our canonical display name, falling back to the node-provided name. */
export function chainDisplayName(chainId: bigint, nodeName?: string): string {
  return CHAIN_DISPLAY_NAMES[chainId.toString()] ?? nodeName ?? `Chain ${chainId}`;
}
