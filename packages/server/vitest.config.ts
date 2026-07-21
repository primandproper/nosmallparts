import { defineConfig } from 'vitest/config';

export default defineConfig({
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
