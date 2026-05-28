// Icon paths for tokens and chains.
// Files are stored in public/icons/tokens/ and public/icons/chains/.
// Paths are relative to Vite's `base` (resolved via `import.meta.env.BASE_URL`)
// so the same bundle works both at the dev server root and behind the
// production `/v1/playground/` prefix. Components must handle `img onError`
// with a letter-avatar fallback.

const BASE = import.meta.env.BASE_URL;

const TOKEN_ICONS: Record<string, string> = {
  eth:    'icons/tokens/eth.png',
  matic:  'icons/tokens/matic.png',
  pol:    'icons/tokens/pol.png',
  bnb:    'icons/tokens/bnb.png',
  usdt:   'icons/tokens/usdt.png',
  xrp:    'icons/tokens/xrp.png',
  yellow: 'icons/tokens/yellow.png',
};

// Testnets reuse their parent mainnet icon.
const CHAIN_ICONS: Record<string, string> = {
  '1':       'icons/chains/ethereum.webp',
  '11155111':'icons/chains/ethereum.webp', // Sepolia
  '137':     'icons/chains/polygon.webp',
  '80002':   'icons/chains/polygon.webp',  // Amoy
  '8453':    'icons/chains/base.png',
  '84532':   'icons/chains/base.png',      // Base Sepolia
  '59144':   'icons/chains/linea.webp',
  '59141':   'icons/chains/linea.webp',    // Linea Sepolia
  '56':      'icons/chains/binance.webp',
  '1449000': 'icons/chains/xrplevm.png',   // XRPL EVM Testnet
};

export function tokenIconUrl(symbol: string): string | null {
  const p = TOKEN_ICONS[symbol.toLowerCase()];
  return p ? BASE + p : null;
}

export function chainIconUrl(chainId: bigint): string | null {
  const p = CHAIN_ICONS[chainId.toString()];
  return p ? BASE + p : null;
}
