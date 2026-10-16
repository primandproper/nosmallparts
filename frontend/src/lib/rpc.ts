import { methodsOf } from '@primandproper/platform-client';

import { LiveKitServiceDefinition } from '../generated/nosmallparts/livekit/v1/livekit';
import {
  type Play,
  type PlayInput,
  type PlayRatingStats,
  PlayScriptsServiceDefinition,
  type ScriptTag,
} from '../generated/nosmallparts/scripts/v1/scripts';
import { StagesServiceDefinition } from '../generated/nosmallparts/stage/v1/stage';
import { session } from './session';

const PlayScripts = methodsOf(PlayScriptsServiceDefinition);
const Stages = methodsOf(StagesServiceDefinition);
const LiveKit = methodsOf(LiveKitServiceDefinition);

// The reads answer anybody, and differently for somebody signed in, so they go
// with the credential when there is one. The writes need a user.

export async function listPlayScripts(search?: string): Promise<Play[]> {
  const res = await session.callOptionallyAuthenticated(PlayScripts.listPlayScripts, {
    search: search?.trim() || undefined,
  });
  return res.plays;
}

export async function getPlayScript(id: string): Promise<Play> {
  const res = await session.callOptionallyAuthenticated(PlayScripts.getPlayScript, { id });
  if (!res.play) throw new Error('Not found');
  return res.play;
}

export async function createPlayScript(play: PlayInput): Promise<Play> {
  const res = await session.call(PlayScripts.createPlayScript, { play });
  if (!res.play) throw new Error('The server answered with no script.');
  return res.play;
}

export async function updatePlayScript(id: string, play: PlayInput): Promise<Play> {
  const res = await session.call(PlayScripts.updatePlayScript, { id, play });
  if (!res.play) throw new Error('The server answered with no script.');
  return res.play;
}

export async function ratePlayScript(id: string, rating: number): Promise<PlayRatingStats | undefined> {
  const res = await session.call(PlayScripts.ratePlayScript, { id, rating });
  return res.ratingStats;
}

export async function listTags(): Promise<ScriptTag[]> {
  const res = await session.callOptionallyAuthenticated(PlayScripts.listTags, {});
  return res.tags;
}

/** createStage opens a stage, with a play preselected when given one. */
export async function createStage(playScriptId?: string): Promise<{ id: string; code: string }> {
  return session.call(Stages.createStage, { playScriptId });
}

/** stageExists resolves a join code, matched without regard to case. */
export async function stageExists(code: string): Promise<{ exists: boolean; id: string; code: string }> {
  return session.callAnonymous(Stages.stageExists, { code });
}

export interface LiveKitCredentials {
  token: string;
  url: string;
  room: string;
}

/** mintVideoToken admits a participant to their stage's video room. */
export async function mintVideoToken(stageId: string, participantId: string): Promise<LiveKitCredentials> {
  return session.callOptionallyAuthenticated(LiveKit.mintToken, { stageId, participantId });
}
