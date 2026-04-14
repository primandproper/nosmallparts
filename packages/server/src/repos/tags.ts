import { eq } from 'drizzle-orm';
import { ulid } from 'ulid';
import { db } from '../db/index.js';
import { scriptTags } from '../db/schema.js';

export async function listTags() {
  return db.select().from(scriptTags).orderBy(scriptTags.name);
}

export async function upsertTagByName(name: string): Promise<string> {
  const trimmed = name.trim();
  const [existing] = await db.select().from(scriptTags).where(eq(scriptTags.name, trimmed));
  if (existing) return existing.id;
  const id = ulid();
  await db.insert(scriptTags).values({ id, name: trimmed });
  return id;
}
