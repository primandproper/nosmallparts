import { createChannel, createClient, ChannelCredentials } from 'nice-grpc';
import { PlayScriptsServiceDefinition } from '@/generated/play_scripts/play_scripts_service';

// Get the gRPC backend URL from environment variable
// This should only be used server-side (in API routes)
function getGrpcBackendUrl(): string {
  const url = process.env.GRPC_BACKEND_URL;
  if (!url) {
    // Default to localhost for development
    return 'localhost:50051';
  }
  return url;
}

// Create a singleton channel and client
// Note: This is lazily initialized to avoid issues during build time
let _channel: ReturnType<typeof createChannel> | null = null;
let _playScriptsClient: ReturnType<
  typeof createClient<typeof PlayScriptsServiceDefinition>
> | null = null;

function getChannel() {
  if (!_channel) {
    const target = getGrpcBackendUrl();
    // Use insecure credentials for local development
    _channel = createChannel(target, ChannelCredentials.createInsecure());
  }
  return _channel;
}

/**
 * Get the PlayScriptsService gRPC client.
 * This should only be called from server-side code (API routes).
 */
export function getPlayScriptsClient() {
  if (!_playScriptsClient) {
    _playScriptsClient = createClient(PlayScriptsServiceDefinition, getChannel());
  }
  return _playScriptsClient;
}
