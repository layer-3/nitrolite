// Resolves the nitronode WebSocket URL at runtime.
//
// Production builds load /playground/env.js (rendered by the container
// entrypoint via envsubst) which sets `window.__ENV__.NITRONODE_URL`. Vite dev
// can override via `VITE_NITRONODE_URL` in `.env.local`. The hard-coded
// fallback keeps `npm run dev` working out of the box.
const FALLBACK = 'wss://nitronode-sandbox.yellow.org/v1/ws';

export const NODE_URL: string =
  (typeof window !== 'undefined' && window.__ENV__?.NITRONODE_URL) ||
  import.meta.env.VITE_NITRONODE_URL ||
  FALLBACK;
