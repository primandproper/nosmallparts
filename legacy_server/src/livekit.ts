import { AccessToken } from 'livekit-server-sdk';

const LIVEKIT_API_KEY = process.env.LIVEKIT_API_KEY || 'APIEKDkNgq7FWcq';
const LIVEKIT_API_SECRET =
  process.env.LIVEKIT_API_SECRET || 'EME1b8dRhSuUy7aLe9mi4W3pfmWw2YabxDOfFRgPxovA';
const LIVEKIT_URL = process.env.LIVEKIT_URL || 'https://nosmallparts-sx2i4bl2.livekit.cloud';

export function isLiveKitConfigured(): boolean {
  return !!(LIVEKIT_API_KEY && LIVEKIT_API_SECRET && LIVEKIT_URL);
}

export async function generateToken(
  stageId: string,
  participantId: string,
  participantName: string
): Promise<string> {
  if (!isLiveKitConfigured()) {
    throw new Error('LiveKit is not configured');
  }

  const token = new AccessToken(LIVEKIT_API_KEY, LIVEKIT_API_SECRET, {
    identity: participantId,
    name: participantName,
  });

  token.addGrant({
    roomJoin: true,
    room: stageId,
    canPublish: true,
    canSubscribe: true,
  });

  return await token.toJwt();
}

export function getLiveKitUrl(): string {
  return LIVEKIT_URL;
}
