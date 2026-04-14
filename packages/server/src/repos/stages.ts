import { and, eq, isNull } from 'drizzle-orm';
import { ulid } from 'ulid';
import type { Participant, StagePhase, StageState } from '@nsp/shared';
import { db } from '../db/index.js';
import {
  stageParticipantRoles,
  stageParticipants,
  stages,
} from '../db/schema.js';
import { getPlayScript } from './play-scripts.js';

function generateStageCode(): string {
  // 6-char uppercase alphanumeric (no ambiguous 0/O/1/I/L).
  const alphabet = 'ABCDEFGHJKMNPQRSTUVWXYZ23456789';
  let code = '';
  for (let i = 0; i < 6; i++) {
    code += alphabet[Math.floor(Math.random() * alphabet.length)];
  }
  return code;
}

export async function createStage(playScriptId?: string | null): Promise<{ id: string; code: string }> {
  const id = ulid();
  let code = generateStageCode();
  // Retry a few times if we collide with an existing code.
  for (let attempt = 0; attempt < 5; attempt++) {
    const [existing] = await db.select().from(stages).where(eq(stages.code, code));
    if (!existing) break;
    code = generateStageCode();
  }
  const now = new Date();
  await db.insert(stages).values({
    id,
    code,
    playScriptId: playScriptId ?? null,
    currentLineIndex: 0,
    phase: 'lobby',
    createdAt: now,
    updatedAt: now,
  });
  return { id, code };
}

export async function getStageByCode(code: string) {
  const [row] = await db.select().from(stages).where(eq(stages.code, code));
  return row ?? null;
}

export async function getStageById(id: string) {
  const [row] = await db.select().from(stages).where(eq(stages.id, id));
  return row ?? null;
}

export async function getStageByIdOrCode(idOrCode: string) {
  return (await getStageById(idOrCode)) ?? (await getStageByCode(idOrCode));
}

export async function joinStage(
  idOrCode: string,
  participantName: string,
  userId: string | null
): Promise<{ stageId: string; participantId: string }> {
  const stage = await getStageByIdOrCode(idOrCode);
  if (!stage) throw new Error('stage not found');
  const participantId = ulid();
  await db.insert(stageParticipants).values({
    id: participantId,
    stageId: stage.id,
    userId,
    displayName: participantName,
    isDirector: false,
    isReady: false,
    isAudience: false,
    createdAt: new Date(),
  });
  return { stageId: stage.id, participantId };
}

export async function leaveStage(stageId: string, participantId: string): Promise<void> {
  await db
    .update(stageParticipants)
    .set({ leftAt: new Date() })
    .where(
      and(
        eq(stageParticipants.id, participantId),
        eq(stageParticipants.stageId, stageId)
      )
    );
}

export async function setSelectedPlay(stageId: string, playScriptId: string | null) {
  await db
    .update(stages)
    .set({ playScriptId, updatedAt: new Date() })
    .where(eq(stages.id, stageId));
  // Clear cast-member role selections because the play (and therefore cast ids) changed.
  const parts = await db
    .select()
    .from(stageParticipants)
    .where(eq(stageParticipants.stageId, stageId));
  for (const p of parts) {
    await db
      .delete(stageParticipantRoles)
      .where(eq(stageParticipantRoles.stageParticipantId, p.id));
  }
}

export async function toggleParticipantRole(
  stageId: string,
  participantId: string,
  castMemberId: string
): Promise<void> {
  // Upsert semantics: if the (participant, castMember) row exists, remove it; otherwise insert.
  const [existing] = await db
    .select()
    .from(stageParticipantRoles)
    .where(
      and(
        eq(stageParticipantRoles.stageParticipantId, participantId),
        eq(stageParticipantRoles.castMemberId, castMemberId)
      )
    );
  if (existing) {
    await db
      .delete(stageParticipantRoles)
      .where(
        and(
          eq(stageParticipantRoles.stageParticipantId, participantId),
          eq(stageParticipantRoles.castMemberId, castMemberId)
        )
      );
  } else {
    // Ensure no other participant on this stage currently holds that role (roles are exclusive).
    const parts = await db
      .select({ id: stageParticipants.id })
      .from(stageParticipants)
      .where(eq(stageParticipants.stageId, stageId));
    for (const p of parts) {
      if (p.id === participantId) continue;
      await db
        .delete(stageParticipantRoles)
        .where(
          and(
            eq(stageParticipantRoles.stageParticipantId, p.id),
            eq(stageParticipantRoles.castMemberId, castMemberId)
          )
        );
    }
    await db.insert(stageParticipantRoles).values({
      stageParticipantId: participantId,
      castMemberId,
    });
  }
}

export async function updateParticipantFlags(
  participantId: string,
  patch: {
    isDirector?: boolean;
    isAudience?: boolean;
    isReady?: boolean;
  }
) {
  await db
    .update(stageParticipants)
    .set(patch)
    .where(eq(stageParticipants.id, participantId));

  if (patch.isAudience === true) {
    // Moving to audience clears cast assignments.
    await db
      .delete(stageParticipantRoles)
      .where(eq(stageParticipantRoles.stageParticipantId, participantId));
  }
}

export async function setStagePhase(
  stageId: string,
  phase: StagePhase,
  currentLineIndex?: number
) {
  const patch: { phase: StagePhase; updatedAt: Date; currentLineIndex?: number } = {
    phase,
    updatedAt: new Date(),
  };
  if (currentLineIndex !== undefined) patch.currentLineIndex = currentLineIndex;
  await db.update(stages).set(patch).where(eq(stages.id, stageId));
}

export async function incrementLine(stageId: string): Promise<{
  currentLineIndex: number;
  playEnded: boolean;
}> {
  const stage = await getStageById(stageId);
  if (!stage) throw new Error('stage not found');
  if (!stage.playScriptId) return { currentLineIndex: stage.currentLineIndex, playEnded: false };

  const play = await getPlayScript(stage.playScriptId);
  if (!play) return { currentLineIndex: stage.currentLineIndex, playEnded: false };
  const indices = Array.from(new Set(play.lines.map((l) => l.index))).sort((a, b) => a - b);

  const nextIndex = stage.currentLineIndex + 1;
  if (nextIndex >= indices.length) {
    await db
      .update(stages)
      .set({ phase: 'discussion', updatedAt: new Date() })
      .where(eq(stages.id, stageId));
    return { currentLineIndex: stage.currentLineIndex, playEnded: true };
  }
  await db
    .update(stages)
    .set({ currentLineIndex: nextIndex, updatedAt: new Date() })
    .where(eq(stages.id, stageId));
  return { currentLineIndex: nextIndex, playEnded: false };
}

export async function resetStage(stageId: string) {
  await db
    .update(stages)
    .set({ phase: 'lobby', currentLineIndex: 0, updatedAt: new Date() })
    .where(eq(stages.id, stageId));
  const parts = await db
    .select()
    .from(stageParticipants)
    .where(eq(stageParticipants.stageId, stageId));
  for (const p of parts) {
    await db
      .update(stageParticipants)
      .set({ isReady: false })
      .where(eq(stageParticipants.id, p.id));
  }
}

export async function getStageState(stageId: string): Promise<StageState | null> {
  const stage = await getStageById(stageId);
  if (!stage) return null;

  const parts = await db
    .select()
    .from(stageParticipants)
    .where(
      and(
        eq(stageParticipants.stageId, stageId),
        isNull(stageParticipants.leftAt)
      )
    );

  const participants: Participant[] = [];
  for (const p of parts) {
    const roles = await db
      .select()
      .from(stageParticipantRoles)
      .where(eq(stageParticipantRoles.stageParticipantId, p.id));
    participants.push({
      id: p.id,
      name: p.displayName,
      castMemberIds: roles.map((r) => r.castMemberId),
      isDirector: p.isDirector,
      isReady: p.isReady,
      isAudience: p.isAudience,
    });
  }

  const selectedPlay = stage.playScriptId ? await getPlayScript(stage.playScriptId) : null;

  return {
    id: stage.id,
    code: stage.code,
    participants,
    selectedPlay,
    currentLineIndex: stage.currentLineIndex,
    phase: stage.phase as StagePhase,
  };
}
