import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';

export default defineConfig({
  resolve: {
    alias: {
      // Run tests against @nsp/shared's TypeScript sources so the suite does
      // not depend on the package having been built first.
      '@nsp/shared': fileURLToPath(new URL('../shared/src/index.ts', import.meta.url)),
    },
  },
  test: {
    environment: 'node',
    setupFiles: ['./test/setup.ts'],
    // Each test file gets its own process (and therefore its own SQLite file
    // + in-memory connection maps) so files can run in parallel safely.
    pool: 'forks',
    testTimeout: 15_000,
    hookTimeout: 15_000,
  },
});
