package keys

const (
	// OAuth2ClientIDKey is the standard key for referring to an OAuth2 client's database ID.
	OAuth2ClientIDKey = "oauth2_clients" + idSuffix
	// OAuth2ClientClientIDKey is the standard key for referring to an OAuth2 client's client ID.
	OAuth2ClientClientIDKey = "oauth2_clients.client_id"
	// OAuth2ClientTokenIDKey is the standard key for referring to an OAuth2 client token's ID.
	/* #nosec G101 */
	OAuth2ClientTokenIDKey = "oauth2_client_tokens" + idSuffix
	// OAuth2ClientTokenCodeKey is the standard key for referring to an OAuth2 client token's code.
	/* #nosec G101 */
	OAuth2ClientTokenCodeKey = "oauth2_client_tokens.code"
	// OAuth2ClientTokenAccessKey is the standard key for referring to an OAuth2 client token's access.
	/* #nosec G101 */
	OAuth2ClientTokenAccessKey = "oauth2_client_tokens.access"
	// OAuth2ClientTokenRefreshKey is the standard key for referring to an OAuth2 client token's refresh.
	/* #nosec G101 */
	OAuth2ClientTokenRefreshKey = "oauth2_client_tokens.refresh"
)
