/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_NITRONODE_URL?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

interface Window {
  __ENV__?: {
    NITRONODE_URL?: string;
  };
}
