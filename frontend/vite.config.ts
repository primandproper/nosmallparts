import tailwindcss from '@tailwindcss/vite';
import react from '@vitejs/plugin-react';
import { defineConfig } from 'vitest/config';

// The Go server is the one origin in production: it serves this app, gRPC-web
// under /grpc, the guest identity route and the stage socket. In development
// Vite forwards those to it, so the app is same-origin there too and the
// refresh-token cookie works on localhost.
const backend = process.env.NSP_BACKEND_URL ?? 'http://localhost:3001';

export default defineConfig({
  plugins: [react(), tailwindcss()],
  // The workspace hoists for Expo, whose older React sits at the root, so a
  // hoisted dependency like react-router gets a copy of its own; two Reacts
  // in one page break every hook.
  resolve: {
    dedupe: ['react', 'react-dom'],
  },
  server: {
    proxy: {
      '/grpc': { target: backend },
      '/api': { target: backend },
      '/ws': { target: backend.replace(/^http/, 'ws'), ws: true },
    },
  },
  build: {
    // React, the platform client and protobuf come to ~140 kB gzipped, which
    // is fine for this app; LiveKit is split out on its own.
    chunkSizeWarningLimit: 700,
  },
  test: {
    environment: 'node',
    // e2e/ is Playwright's, run against a server by `make e2e`.
    include: ['src/**/*.test.ts'],
  },
});
