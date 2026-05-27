// Runtime configuration resolved once at module load.
//
// Production builds load /v1/playground/env.js (rendered by the container
// entrypoint via envsubst) which sets `window.__ENV__`. Vite dev can override
// via `VITE_*` in `.env.local`. The hard-coded fallback keeps `npm run dev`
// working out of the box against sandbox.

const NODE_URL_FALLBACK = 'wss://nitronode-sandbox.yellow.org/v1/ws';

export const NODE_URL: string =
  (typeof window !== 'undefined' && window.__ENV__?.NITRONODE_URL) ||
  import.meta.env.VITE_NITRONODE_URL ||
  NODE_URL_FALLBACK;

// Faucet host defaults to the nitronode host on the matching HTTP scheme,
// since the chart deploys both behind the same ingress. An explicit override
// is honored first so a split-host deploy can point the faucet elsewhere.
function deriveFaucetUrl(nodeUrl: string): string {
  try {
    const u = new URL(nodeUrl);
    u.protocol = u.protocol === 'wss:' ? 'https:' : 'http:';
    u.pathname = '/v1/faucet-app/requestTokens';
    u.search = '';
    u.hash = '';
    return u.toString();
  } catch {
    return 'https://nitronode-sandbox.yellow.org/v1/faucet-app/requestTokens';
  }
}

export const FAUCET_URL: string =
  (typeof window !== 'undefined' && window.__ENV__?.FAUCET_URL) ||
  import.meta.env.VITE_FAUCET_URL ||
  deriveFaucetUrl(NODE_URL);

// Whether the faucet UI surface is enabled. The faucet itself ships only to
// non-prod envs (see helmfile.yaml.gotmpl), so this flag hides the tab on
// hosts where the request would 404. Default `true` keeps sandbox/stress
// behavior — prod overrides via env.js.
function parseEnabled(raw: string | undefined): boolean | null {
  if (raw === undefined || raw === '') return null;
  return raw !== 'false' && raw !== '0';
}
export const FAUCET_ENABLED: boolean =
  parseEnabled(typeof window !== 'undefined' ? window.__ENV__?.FAUCET_ENABLED : undefined) ??
  parseEnabled(import.meta.env.VITE_FAUCET_ENABLED) ??
  true;
