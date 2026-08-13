import { mkdtempSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));

// Environment must be set BEFORE any src module is imported — db path and
// all NSP_* durations are read at module load.
process.env.DATABASE_URL = join(mkdtempSync(join(tmpdir(), 'nsp-test-')), 'test.db');
process.env.BETTER_AUTH_SECRET = 'test-secret-test-secret-test-secret';
process.env.NODE_ENV = 'test';
process.env.ADMIN_EMAILS = 'admin@example.com, Boss@Example.COM';
process.env.TRUSTED_ORIGINS = 'http://localhost:8081';
// Shrink all timers so integration tests run fast.
process.env.NSP_HOST_GRACE_MS = '80';
process.env.NSP_VACANCY_CLAIM_MS = '150';
process.env.NSP_STAGE_CLEANUP_DELAY_MS = '100';
process.env.NSP_STAGE_IDLE_MS = '300000';
process.env.NSP_STAGE_REAPER_INTERVAL_MS = '60000';
process.env.NSP_COUNTDOWN_MS = '40';

const { migrate } = await import('drizzle-orm/better-sqlite3/migrator');
const { db } = await import('../src/db/index.js');
migrate(db, { migrationsFolder: join(here, '../src/db/migrations') });
