import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';
import path from 'path';

export default defineConfig({
  // App is deployed behind the same hostname as nitronode, mounted at
  // `/v1/playground`. Asset URLs and the SPA entry are rewritten by Vite to
  // include the prefix; nginx serves files from the same prefix in-cluster.
  base: '/v1/playground/',
  plugins: [react()],
  server: {
    port: 3001,
  },
  define: {
    global: 'globalThis',
  },
  resolve: {
    alias: {
      '@': path.resolve(__dirname, './src'),
    },
  },
});
