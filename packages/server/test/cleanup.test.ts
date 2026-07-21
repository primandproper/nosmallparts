import { describe, expect, it } from 'vitest';
import { eq } from 'drizzle-orm';
import { app } from '../src/app.js';
import { db } from '../src/db/index.js';
import { stageParticipants, stages } from '../src/db/schema.js';
import { getStageById } from '../src/repos/stages.js';
import { runStageReaperOnce } from '../src/ws/stage-handler.js';
import { guestIdentity, joinClient, makeStage, sleep, userIdentity } from './helpers.js';

const CLEANUP_WAIT = 250; // NSP_STAGE_CLEANUP_DELAY_MS=100 in setup

describe('stage cleanup (§10.2)', () => {
  it('deletes the stage shortly after the last participant leaves', async () => {
    const stage = await makeStage();
    const a = await joinClient(await userIdentity(), stage.id, 'Hosty');
    const b = await joinClient(guestIdentity(), stage.id, 'Guesty');
    await a.disconnect();
    await b.disconnect();
    await sleep(CLEANUP_WAIT);

    expect(await getStageById(stage.id)).toBeNull();
    // Dependents are gone too (FK cascade).
    const parts = await db
      .select()
      .from(stageParticipants)
      .where(eq(stageParticipants.stageId, stage.id));
    expect(parts).toHaveLength(0);

    // "As if it never existed."
    const res = await app.request(`/api/stages/${stage.code}/exists`);
    const body = (await res.json()) as { exists: boolean };
    expect(body.exists).toBe(false);
  });

  it('a joiner during the 5s window cancels cleanup', async () => {
    const stage = await makeStage();
    const a = await joinClient(guestIdentity(), stage.id, 'FirstIn');
    await a.disconnect();
    // Join again immediately, within the cleanup delay.
    const b = await joinClient(guestIdentity(), stage.id, 'SecondIn');
    await sleep(CLEANUP_WAIT);
    expect(await getStageById(stage.id)).not.toBeNull();
    expect(b.participantId).toBeTruthy();
  });

  it('reaper deletes stages idle past the threshold, recovered from DB', async () => {
    // Simulate a stage orphaned before a server restart: rows exist in the DB
    // but there are no live connections and all timestamps are old.
    const stage = await makeStage();
    const old = new Date(Date.now() - 10 * 60_000 * 100);
    await db.update(stages).set({ createdAt: old, updatedAt: old }).where(eq(stages.id, stage.id));

    const fresh = await makeStage();

    const deleted = await runStageReaperOnce();
    expect(deleted).toContain(stage.id);
    expect(await getStageById(stage.id)).toBeNull();
    expect(await getStageById(fresh.id)).not.toBeNull();
  });

  it('reaper spares stages with live connections', async () => {
    const stage = await makeStage();
    await joinClient(guestIdentity(), stage.id, 'StillHere');
    const old = new Date(Date.now() - 10 * 60_000 * 100);
    await db.update(stages).set({ createdAt: old, updatedAt: old }).where(eq(stages.id, stage.id));
    await db
      .update(stageParticipants)
      .set({ createdAt: old })
      .where(eq(stageParticipants.stageId, stage.id));

    const deleted = await runStageReaperOnce();
    expect(deleted).not.toContain(stage.id);
    expect(await getStageById(stage.id)).not.toBeNull();
  });

  it('leftAt is set on disconnect and filtered from stage state', async () => {
    const stage = await makeStage();
    const a = await joinClient(guestIdentity(), stage.id, 'Leaver');
    const b = await joinClient(guestIdentity(), stage.id, 'Stayer');
    await a.disconnect();
    const [row] = await db
      .select()
      .from(stageParticipants)
      .where(eq(stageParticipants.id, a.participantId!));
    expect(row.leftAt).not.toBeNull();
    const state = b.lastState();
    const names = ((state?.participants ?? []) as { name: string }[]).map((p) => p.name);
    expect(names).not.toContain('Leaver');
    expect(names).toContain('Stayer');
  });
});
