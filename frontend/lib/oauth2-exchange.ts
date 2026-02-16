/**
 * Exchanges a JWT (from LoginForToken) for an OAuth2 access token.
 * The gRPC auth interceptor only accepts OAuth2 tokens stored in oauth2_client_tokens,
 * so we must perform this exchange after login.
 */

function randomString(length: number): string {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789';
  let result = '';
  const randomValues = new Uint8Array(length);
  crypto.getRandomValues(randomValues);
  for (let i = 0; i < length; i++) {
    result += chars[randomValues[i]! % chars.length];
  }
  return result;
}

export interface OAuth2TokenResponse {
  accessToken: string;
  refreshToken?: string;
  expiresAt: string;
}

/**
 * Exchange JWT for OAuth2 token via the backend's OAuth2 authorize + token flow.
 */
export async function exchangeJwtForOAuth2Token(
  authServerBase: string,
  clientId: string,
  clientSecret: string,
  jwt: string
): Promise<OAuth2TokenResponse> {
  const state = randomString(32);
  // For PKCE "plain", code_challenge equals code_verifier
  const codeVerifier = randomString(64);
  const params = new URLSearchParams({
    client_id: clientId,
    response_type: 'code',
    state,
    redirect_uri: authServerBase,
    code_challenge: codeVerifier,
    code_challenge_method: 'plain',
  });
  const authUrl = `${authServerBase}/oauth2/authorize?${params.toString()}`;

  const authRes = await fetch(authUrl, {
    method: 'GET',
    headers: { Authorization: `Bearer ${jwt}` },
    redirect: 'manual',
  });

  const location = authRes.headers.get('Location');
  if (!location) {
    throw new Error('OAuth2 authorize did not return redirect with code');
  }
  const locationUrl = new URL(location, authServerBase);
  const code = locationUrl.searchParams.get('code');
  if (!code) {
    throw new Error('OAuth2 authorize redirect missing code');
  }

  const tokenParams = new URLSearchParams({
    grant_type: 'authorization_code',
    code,
    redirect_uri: authServerBase,
    client_id: clientId,
    client_secret: clientSecret,
    code_verifier: codeVerifier,
  });

  const tokenRes = await fetch(`${authServerBase}/oauth2/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: tokenParams.toString(),
  });

  if (!tokenRes.ok) {
    const text = await tokenRes.text();
    throw new Error(`OAuth2 token exchange failed: ${tokenRes.status} ${text}`);
  }

  const tokenData = (await tokenRes.json()) as {
    access_token: string;
    refresh_token?: string;
    expires_in?: number;
  };

  const expiresIn = tokenData.expires_in ?? 24 * 60 * 60; // default 24h
  const expiresAt = new Date(Date.now() + expiresIn * 1000).toISOString();

  return {
    accessToken: tokenData.access_token,
    refreshToken: tokenData.refresh_token,
    expiresAt,
  };
}
