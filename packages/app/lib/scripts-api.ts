import type { Play, ScriptTag } from '@nsp/shared';
import { api } from './api';

export interface PlayPayload {
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
  tagNames?: string[];
}

export async function listPlayScripts(search?: string): Promise<Play[]> {
  const qs = search ? `?search=${encodeURIComponent(search)}` : '';
  const res = await api<{ plays: Play[] }>(`/api/play-scripts${qs}`);
  return res.plays;
}

export async function getPlayScriptById(id: string): Promise<Play> {
  const res = await api<{ play: Play }>(`/api/play-scripts/${id}`);
  return res.play;
}

export async function createPlayScript(p: PlayPayload): Promise<Play> {
  const res = await api<{ play: Play }>('/api/play-scripts', {
    method: 'POST',
    body: JSON.stringify(p),
  });
  return res.play;
}

export async function updatePlayScript(id: string, p: PlayPayload): Promise<Play> {
  const res = await api<{ play: Play }>(`/api/play-scripts/${id}`, {
    method: 'PUT',
    body: JSON.stringify(p),
  });
  return res.play;
}

export async function deletePlayScript(id: string): Promise<void> {
  await api<{ ok: true }>(`/api/play-scripts/${id}`, { method: 'DELETE' });
}

export async function ratePlayScript(
  id: string,
  rating: number
): Promise<{ averageRating: number; ratingCount: number; userRating: number }> {
  return api(`/api/play-scripts/${id}/rate`, {
    method: 'POST',
    body: JSON.stringify({ rating }),
  });
}

export async function listTags(): Promise<ScriptTag[]> {
  const res = await api<{ tags: ScriptTag[] }>('/api/tags');
  return res.tags;
}
