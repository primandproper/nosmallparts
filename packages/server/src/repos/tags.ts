import { sql } from 'drizzle-orm';
import { ulid } from 'ulid';
import { db } from '../db/index.js';
import { scriptTags } from '../db/schema.js';

export async function listTags() {
  return db.select().from(scriptTags).orderBy(scriptTags.name);
}

export async function upsertTagByName(name: string): Promise<string> {
  // Normalize: trim, collapse whitespace, match case-insensitively so
  // "Comedy" and " comedy " merge into one tag (§2.5).
  const trimmed = name.trim().replace(/\s+/g, ' ');
  const [existing] = await db
    .select()
    .from(scriptTags)
    .where(sql`lower(${scriptTags.name}) = ${trimmed.toLowerCase()}`);
  if (existing) return existing.id;
  const id = ulid();
  await db.insert(scriptTags).values({ id, name: trimmed });
  return id;
}
