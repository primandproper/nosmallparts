import { createChannel, createClient, ChannelCredentials } from 'nice-grpc';
import { AuthServiceDefinition } from '@/generated/auth/auth_service';
import { IdentityServiceDefinition } from '@/generated/identity/identity_service';

// Get the gRPC backend URL from environment variable
function getGrpcBackendUrl(): string {
  const url = process.env.GRPC_BACKEND_URL;
  if (!url) {
    return 'localhost:8001';
  }
  return url;
}

// Create singleton channels and clients
let _channel: ReturnType<typeof createChannel> | null = null;
let _authClient: ReturnType<typeof createClient<typeof AuthServiceDefinition>> | null = null;
let _identityClient: ReturnType<typeof createClient<typeof IdentityServiceDefinition>> | null =
  null;

function getChannel() {
  if (!_channel) {
    const target = getGrpcBackendUrl();
    _channel = createChannel(target, ChannelCredentials.createInsecure());
  }
  return _channel;
}

/**
 * Get the AuthService gRPC client.
 * This should only be called from server-side code (API routes).
 */
export function getAuthClient() {
  if (!_authClient) {
    _authClient = createClient(AuthServiceDefinition, getChannel());
  }
  return _authClient;
}

/**
 * Get the IdentityService gRPC client.
 * This should only be called from server-side code (API routes).
 */
export function getIdentityClient() {
  if (!_identityClient) {
    _identityClient = createClient(IdentityServiceDefinition, getChannel());
  }
  return _identityClient;
}
