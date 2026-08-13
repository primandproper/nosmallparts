import { RoomServiceClient, TrackSource, TrackType } from 'livekit-server-sdk';

function getAdminClient(): RoomServiceClient | null {
  const apiKey = process.env.LIVEKIT_API_KEY;
  const apiSecret = process.env.LIVEKIT_API_SECRET;
  const wsUrl = process.env.LIVEKIT_URL;
  if (!apiKey || !apiSecret || !wsUrl) return null;
  const httpUrl = wsUrl.replace(/^wss?:\/\//, (m) => (m === 'wss://' ? 'https://' : 'http://'));
  return new RoomServiceClient(httpUrl, apiKey, apiSecret);
}

export async function livekitRemoveParticipant(roomName: string, identity: string): Promise<void> {
  const client = getAdminClient();
  if (!client) return;
  await client.removeParticipant(roomName, identity).catch((err) => {
    console.warn('[livekit] removeParticipant failed', err);
  });
}

export async function livekitSetMuted(
  roomName: string,
  identity: string,
  muted: boolean
): Promise<void> {
  const client = getAdminClient();
  if (!client) return;
  try {
    const info = await client.getParticipant(roomName, identity);
    for (const track of info.tracks ?? []) {
      if (track.type === TrackType.AUDIO) {
        await client.mutePublishedTrack(roomName, identity, track.sid, muted).catch(() => {});
      }
    }
    // Also restrict publishable sources so a force-muted participant cannot
    // simply re-publish audio (§4.8). Unmuting restores the full set.
    const sources = muted
      ? [TrackSource.CAMERA, TrackSource.SCREEN_SHARE]
      : [
          TrackSource.CAMERA,
          TrackSource.MICROPHONE,
          TrackSource.SCREEN_SHARE,
          TrackSource.SCREEN_SHARE_AUDIO,
        ];
    await client
      .updateParticipant(roomName, identity, undefined, {
        canSubscribe: true,
        canPublish: true,
        canPublishData: true,
        canPublishSources: sources,
      })
      .catch(() => {});
  } catch (err) {
    console.warn('[livekit] setMuted failed', err);
  }
}
