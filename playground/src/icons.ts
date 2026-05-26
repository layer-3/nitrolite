// Icon paths for tokens and chains.
// Files are stored in public/icons/tokens/ and public/icons/chains/.
// Components must handle img onError with a letter-avatar fallback.

const TOKEN_ICONS: Record<string, string> = {
  eth:    '/icons/tokens/eth.png',
  matic:  '/icons/tokens/matic.png',
  pol:    '/icons/tokens/pol.png',
  bnb:    '/icons/tokens/bnb.png',
  usdt:   '/icons/tokens/usdt.png',
  xrp:    '/icons/tokens/xrp.png',
  yellow: '/icons/tokens/yellow.png',
};

// Testnets reuse their parent mainnet icon.
const CHAIN_ICONS: Record<string, string> = {
  '1':       '/icons/chains/ethereum.webp',
  '11155111':'/icons/chains/ethereum.webp', // Sepolia
  '137':     '/icons/chains/polygon.webp',
  '80002':   '/icons/chains/polygon.webp',  // Amoy
  '8453':    '/icons/chains/base.png',
  '84532':   '/icons/chains/base.png',      // Base Sepolia
  '59144':   '/icons/chains/linea.webp',
  '59141':   '/icons/chains/linea.webp',    // Linea Sepolia
  '56':      '/icons/chains/binance.webp',
  '1449000': '/icons/chains/xrplevm.png',   // XRPL EVM Testnet
};

export function tokenIconUrl(symbol: string): string | null {
  return TOKEN_ICONS[symbol.toLowerCase()] ?? null;
}

export function chainIconUrl(chainId: bigint): string | null {
  return CHAIN_ICONS[chainId.toString()] ?? null;
}
