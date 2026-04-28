import { and, avg, count, desc, eq, inArray, or } from 'drizzle-orm';
import { ulid } from 'ulid';
import type { Play } from '@nsp/shared';
import { db } from '../db/index.js';
import {
  playScriptCastMembers,
  playScriptLineSpeakers,
  playScriptLines,
  playScriptRatings,
  playScriptTags,
  playScripts,
  scriptTags,
} from '../db/schema.js';

/** Input shape for create / full replace. */
export interface PlayScriptInput {
  title: string;
  description?: string;
  author?: string | null;
  isPublic?: boolean;
  cast: Array<{ id?: string; name: string; description?: string | null }>;
  lines: Array<{
    castMemberIds: string[];
    text: string;
    index: number;
    direction?: string | null;
    scrollMode?: 'auto' | 'manual' | null;
    autoScrollWpm?: number | null;
  }>;
  tagIds?: string[];
}

/** Convert DB rows to the frontend `Play` shape. */
async function hydrate(playScriptId: string): Promise<Play | null> {
  const [ps] = await db.select().from(playScripts).where(eq(playScripts.id, playScriptId));
  if (!ps) return null;

  const cm = await db
    .select()
    .from(playScriptCastMembers)
    .where(eq(playScriptCastMembers.playScriptId, playScriptId));
  cm.sort((a, b) => a.sortOrder - b.sortOrder);

  const lns = await db
    .select()
    .from(playScriptLines)
    .where(eq(playScriptLines.playScriptId, playScriptId));

  const lineIds = lns.map((l) => l.id);
  const speakerRows = lineIds.length
    ? await db
        .select()
        .from(playScriptLineSpeakers)
        .where(inArray(playScriptLineSpeakers.lineId, lineIds))
    : [];
  const speakersByLine = new Map<string, string[]>();
  for (const r of speakerRows) {
    const arr = speakersByLine.get(r.lineId) ?? [];
    arr.push(r.castMemberId);
    speakersByLine.set(r.lineId, arr);
  }

  const tagJoin = await db
    .select({ tag: scriptTags })
    .from(playScriptTags)
    .innerJoin(scriptTags, eq(playScriptTags.scriptTagId, scriptTags.id))
    .where(eq(playScriptTags.playScriptId, playScriptId));

  const ratingAgg = await db
    .select({
      avg: avg(playScriptRatings.rating).mapWith(Number),
      cnt: count(playScriptRatings.id),
    })
    .from(playScriptRatings)
    .where(eq(playScriptRatings.playScriptId, playScriptId));

  const stats = ratingAgg[0];

  return {
    id: ps.id,
    title: ps.title,
    description: ps.description ?? '',
    author: ps.author ?? undefined,
    isPublic: ps.isPublic,
    cast: cm.map((c) => ({
      id: c.id,
      name: c.name,
      description: c.description ?? undefined,
    })),
    lines: lns
      .map((l) => ({
        castMemberIds: speakersByLine.get(l.id) ?? [],
        text: l.lineText,
        index: l.lineIndex,
        direction: l.direction ?? undefined,
        scrollMode: (l.scrollMode as 'auto' | 'manual' | null) ?? undefined,
        autoScrollWpm: l.autoScrollWpm ?? undefined,
      }))
      .sort((a, b) => a.index - b.index),
    tags: tagJoin.map(({ tag }) => ({
      id: tag.id,
      name: tag.name,
      description: tag.description ?? undefined,
    })),
    ratingStats:
      stats && stats.cnt > 0
        ? { averageRating: stats.avg ?? 0, ratingCount: stats.cnt }
        : undefined,
  };
}

export async function getPlayScript(id: string): Promise<Play | null> {
  return hydrate(id);
}

export async function listPlayScripts(opts: {
  userId?: string | null;
  search?: string;
}): Promise<Play[]> {
  const { userId, search } = opts;

  // Public scripts OR scripts owned by the current user.
  const where = userId
    ? or(eq(playScripts.isPublic, true), eq(playScripts.createdByUser, userId))
    : eq(playScripts.isPublic, true);

  const rows = await db
    .select({ id: playScripts.id })
    .from(playScripts)
    .where(where)
    .orderBy(desc(playScripts.updatedAt));

  const plays = await Promise.all(rows.map((r) => hydrate(r.id)));
  let result = plays.filter((p): p is Play => p !== null);

  if (search) {
    const q = search.toLowerCase().trim();
    result = result.filter(
      (p) =>
        p.title.toLowerCase().includes(q) ||
        p.description.toLowerCase().includes(q) ||
        (p.author || '').toLowerCase().includes(q) ||
        p.cast.some((c) => c.name.toLowerCase().includes(q)) ||
        (p.tags || []).some((t) => t.name.toLowerCase().includes(q))
    );
  }

  return result;
}

export async function createPlayScript(userId: string, input: PlayScriptInput): Promise<Play> {
  const id = ulid();
  const now = new Date();

  await db.insert(playScripts).values({
    id,
    title: input.title,
    description: input.description ?? '',
    author: input.author ?? null,
    isPublic: input.isPublic ?? false,
    createdByUser: userId,
    createdAt: now,
    updatedAt: now,
  });

  // Remap cast ids to new ULIDs so the returned Play has stable ids.
  const castIdMap = new Map<string, string>();
  for (const [i, c] of input.cast.entries()) {
    const newId = ulid();
    castIdMap.set(c.id ?? String(i), newId);
    await db.insert(playScriptCastMembers).values({
      id: newId,
      playScriptId: id,
      name: c.name,
      description: c.description ?? null,
      sortOrder: i,
    });
  }

  for (const l of input.lines) {
    const lineId = ulid();
    await db.insert(playScriptLines).values({
      id: lineId,
      playScriptId: id,
      lineText: l.text,
      lineIndex: l.index,
      direction: l.direction ?? null,
      scrollMode: l.scrollMode ?? 'auto',
      autoScrollWpm: l.autoScrollWpm ?? null,
    });
    const speakerIds = Array.from(
      new Set((l.castMemberIds || []).map((sid) => castIdMap.get(sid) ?? sid).filter(Boolean))
    );
    for (const sid of speakerIds) {
      await db.insert(playScriptLineSpeakers).values({ lineId, castMemberId: sid });
    }
  }

  if (input.tagIds?.length) {
    for (const tagId of input.tagIds) {
      await db.insert(playScriptTags).values({ playScriptId: id, scriptTagId: tagId });
    }
  }

  const p = await hydrate(id);
  if (!p) throw new Error('insert succeeded but hydrate returned null');
  return p;
}

export async function updatePlayScript(
  userId: string,
  id: string,
  input: PlayScriptInput
): Promise<Play | null> {
  const [existing] = await db
    .select()
    .from(playScripts)
    .where(and(eq(playScripts.id, id), eq(playScripts.createdByUser, userId)));
  if (!existing) return null;

  await db
    .update(playScripts)
    .set({
      title: input.title,
      description: input.description ?? '',
      author: input.author ?? null,
      isPublic: input.isPublic ?? false,
      updatedAt: new Date(),
    })
    .where(eq(playScripts.id, id));

  // Replace cast + lines. Delete lines first (they FK to cast members).
  await db.delete(playScriptLines).where(eq(playScriptLines.playScriptId, id));
  await db.delete(playScriptCastMembers).where(eq(playScriptCastMembers.playScriptId, id));

  const castIdMap = new Map<string, string>();
  for (const [i, c] of input.cast.entries()) {
    const newId = ulid();
    castIdMap.set(c.id ?? String(i), newId);
    await db.insert(playScriptCastMembers).values({
      id: newId,
      playScriptId: id,
      name: c.name,
      description: c.description ?? null,
      sortOrder: i,
    });
  }
  for (const l of input.lines) {
    const lineId = ulid();
    await db.insert(playScriptLines).values({
      id: lineId,
      playScriptId: id,
      lineText: l.text,
      lineIndex: l.index,
      direction: l.direction ?? null,
      scrollMode: l.scrollMode ?? 'auto',
      autoScrollWpm: l.autoScrollWpm ?? null,
    });
    const speakerIds = Array.from(
      new Set((l.castMemberIds || []).map((sid) => castIdMap.get(sid) ?? sid).filter(Boolean))
    );
    for (const sid of speakerIds) {
      await db.insert(playScriptLineSpeakers).values({ lineId, castMemberId: sid });
    }
  }

  if (input.tagIds) {
    await db.delete(playScriptTags).where(eq(playScriptTags.playScriptId, id));
    for (const tagId of input.tagIds) {
      await db.insert(playScriptTags).values({ playScriptId: id, scriptTagId: tagId });
    }
  }

  return hydrate(id);
}

export async function deletePlayScript(userId: string, id: string): Promise<boolean> {
  const result = await db
    .delete(playScripts)
    .where(and(eq(playScripts.id, id), eq(playScripts.createdByUser, userId)))
    .returning({ id: playScripts.id });
  return result.length > 0;
}

export async function ownedBy(userId: string, id: string): Promise<boolean> {
  const [row] = await db
    .select({ id: playScripts.id })
    .from(playScripts)
    .where(and(eq(playScripts.id, id), eq(playScripts.createdByUser, userId)));
  return !!row;
}

export async function ratePlayScript(
  userId: string,
  playScriptId: string,
  rating: number
): Promise<{ averageRating: number; ratingCount: number; userRating: number }> {
  const now = new Date();
  // Upsert via delete-then-insert for simplicity (small scale).
  await db
    .delete(playScriptRatings)
    .where(
      and(
        eq(playScriptRatings.playScriptId, playScriptId),
        eq(playScriptRatings.userId, userId)
      )
    );
  await db.insert(playScriptRatings).values({
    id: ulid(),
    playScriptId,
    userId,
    rating,
    createdAt: now,
    updatedAt: now,
  });

  const [stats] = await db
    .select({
      avg: avg(playScriptRatings.rating).mapWith(Number),
      cnt: count(playScriptRatings.id),
    })
    .from(playScriptRatings)
    .where(eq(playScriptRatings.playScriptId, playScriptId));

  return {
    averageRating: stats?.avg ?? 0,
    ratingCount: stats?.cnt ?? 0,
    userRating: rating,
  };
}

export async function getUserRatings(
  userId: string,
  playScriptIds: string[]
): Promise<Map<string, number>> {
  if (playScriptIds.length === 0) return new Map();
  const rows = await db
    .select()
    .from(playScriptRatings)
    .where(
      and(
        eq(playScriptRatings.userId, userId),
        inArray(playScriptRatings.playScriptId, playScriptIds)
      )
    );
  return new Map(rows.map((r) => [r.playScriptId, r.rating]));
}
