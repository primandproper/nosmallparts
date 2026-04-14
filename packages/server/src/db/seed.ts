/**
 * Seed built-in sample plays as public scripts owned by a synthetic `system` user.
 * Idempotent: if the system user and plays already exist, it no-ops.
 */
import { eq } from 'drizzle-orm';
import { ulid } from 'ulid';
import { samplePlays } from '@nsp/shared';
import { db } from './index.js';
import {
  playScriptCastMembers,
  playScriptLines,
  playScripts,
  user,
} from './schema.js';

const SYSTEM_USER_ID = 'system-nosmallparts';

async function ensureSystemUser() {
  const [existing] = await db.select().from(user).where(eq(user.id, SYSTEM_USER_ID));
  if (existing) return;
  const now = new Date();
  await db.insert(user).values({
    id: SYSTEM_USER_ID,
    name: 'No Small Parts',
    email: 'system@nosmallparts.local',
    emailVerified: true,
    createdAt: now,
    updatedAt: now,
  });
}

async function seedPlay(p: (typeof samplePlays)[number]) {
  const [existing] = await db.select().from(playScripts).where(eq(playScripts.id, p.id));
  if (existing) return;

  const now = new Date();
  await db.insert(playScripts).values({
    id: p.id,
    title: p.title,
    description: p.description,
    author: p.author ?? null,
    isPublic: true,
    createdByUser: SYSTEM_USER_ID,
    createdAt: now,
    updatedAt: now,
  });

  for (const [i, c] of p.cast.entries()) {
    await db.insert(playScriptCastMembers).values({
      id: `${p.id}:${c.id}`,
      playScriptId: p.id,
      name: c.name,
      description: c.description ?? null,
      sortOrder: i,
    });
  }
  for (const l of p.lines) {
    await db.insert(playScriptLines).values({
      id: ulid(),
      playScriptId: p.id,
      castMemberId: `${p.id}:${l.castMemberId}`,
      lineText: l.text,
      lineIndex: l.index,
      direction: l.direction ?? null,
      scrollMode: l.scrollMode ?? 'auto',
      autoScrollWpm: l.autoScrollWpm ?? null,
    });
  }
  console.log(`seeded: ${p.title}`);
}

async function run() {
  await ensureSystemUser();
  for (const p of samplePlays) {
    await seedPlay(p);
  }
  console.log('Seed complete.');
}

run().catch((err) => {
  console.error(err);
  process.exit(1);
});
