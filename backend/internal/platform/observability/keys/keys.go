package keys

const (
	idSuffix = ".id"

	// RequesterIDKey is the standard key for referring to a requesting user's ID.
	RequesterIDKey = "request.made_by"
	// AccountIDKey is the standard key for referring to an account ID.
	AccountIDKey = "account" + idSuffix
	// AccountInvitationKey is the standard key for referring to an account ID.
	AccountInvitationKey = "account_invitation"
	// AccountInvitationIDKey is the standard key for referring to an account ID.
	AccountInvitationIDKey = AccountInvitationKey + idSuffix
	// AccountInvitationTokenKey is the standard key for referring to an account invitation token.
	AccountInvitationTokenKey = "account_invitation.token"
	// ActiveAccountIDKey is the standard key for referring to an active account ID.
	ActiveAccountIDKey = "active_account" + idSuffix
	// UserIDKey is the standard key for referring to a user ID.
	UserIDKey = "user" + idSuffix
	// UserNotificationIDKey is the standard key for referring to a user notification ID.
	UserNotificationIDKey = "user_notification" + idSuffix
	// UserEmailAddressKey is the standard key for referring to a user's email address.
	UserEmailAddressKey = "user.email_address"
	// UserIsServiceAdminKey is the standard key for referring to a user's admin status.
	UserIsServiceAdminKey = "user.is_admin"
	// UsernameKey is the standard key for referring to a username.
	UsernameKey = "user.username"
	// #nosec G101 UserEmailVerificationTokenKey is the standard key for referring to a username.
	UserEmailVerificationTokenKey = "user.email_verification_token"
	// NameKey is the standard key for referring to a name.
	NameKey = "name"
	// SpanIDKey is the standard key for referring to a span ID.
	SpanIDKey = "span" + idSuffix
	// TraceIDKey is the standard key for referring to a trace ID.
	TraceIDKey = "trace" + idSuffix
	// FilterCreatedAfterKey is the standard key for referring to a types.QueryFilter's CreatedAfter field.
	FilterCreatedAfterKey = "query_filter.created_after"
	// FilterCreatedBeforeKey is the standard key for referring to a types.QueryFilter's CreatedBefore field.
	FilterCreatedBeforeKey = "query_filter.created_before"
	// FilterUpdatedAfterKey is the standard key for referring to a types.QueryFilter's UpdatedAfter field.
	FilterUpdatedAfterKey = "query_filter.updated_after"
	// FilterUpdatedBeforeKey is the standard key for referring to a types.QueryFilter's UpdatedAfter field.
	FilterUpdatedBeforeKey = "query_filter.updated_before"
	// FilterSortByKey is the standard key for referring to a types.QueryFilter's SortBy field.
	FilterSortByKey = "query_filter.sort_by"
	// FilterCursorKey is the standard key for referring to a types.QueryFilter's next cursor.
	FilterCursorKey = "query_filter.cursor"
	// FilterLimitKey is the standard key for referring to a types.QueryFilter's limit.
	FilterLimitKey = "query_filter.limit"
	// FilterIsNilKey is the standard key for referring to a types.QueryFilter's null status.
	FilterIsNilKey = "query_filter.is_nil"
	// WebhookIDKey is the standard key for referring to a webhook's ID.
	WebhookIDKey = "webhook" + idSuffix
	// WebhookTriggerEventIDKey is the standard key for referring to a webhook trigger event's ID.
	WebhookTriggerEventIDKey = "webhook_trigger_event" + idSuffix
	// WaitlistIDKey is the standard key for referring to a waitlist ID.
	WaitlistIDKey = "waitlist" + idSuffix
	// WaitlistSignupIDKey is the standard key for referring to a waitlist signup ID.
	WaitlistSignupIDKey = "waitlist_signup" + idSuffix
	// IssueReportIDKey is the standard key for referring to an issue report ID.
	IssueReportIDKey = "issue_report" + idSuffix
	// UploadedMediaIDKey is the standard key for referring to an uploaded media ID.
	UploadedMediaIDKey = "uploaded_media" + idSuffix
	// PlayScriptIDKey is the standard key for referring to a play script ID.
	PlayScriptIDKey = "play_script" + idSuffix
	// AuditLogEntryIDKey is the standard key for referring to an audit log entry's ID.
	AuditLogEntryIDKey = "audit_log_entry" + idSuffix
	// AuditLogEntryResourceTypesKey is the standard key for referring to an audit log entry's resource type.
	AuditLogEntryResourceTypesKey = "audit_log_entry.resource_types"
	// URLKey is the standard key for referring to a URL.
	URLKey = "url"
	// PasswordResetTokenKey is the standard key for referring to a password reset token's ID.
	PasswordResetTokenKey = "password_reset_token"
	// PasswordResetTokenIDKey is the standard key for referring to a password reset token's ID.
	PasswordResetTokenIDKey = PasswordResetTokenKey + idSuffix
	// RequestHeadersKey is the standard key for referring to a http.Request's Headers.
	RequestHeadersKey = "request.headers"
	// RequestIDKey is the standard key for referring to a http.Request's ID.
	RequestIDKey = "request" + idSuffix
	// RequestMethodKey is the standard key for referring to a http.Request's Method.
	RequestMethodKey = "request.method"
	// RequestURIKey is the standard key for referring to a http.Request's URI.
	RequestURIKey = "request.uri"
	// ResponseStatusKey is the standard key for referring to a http.Request's status.
	ResponseStatusKey = "response.status"
	// ResponseHeadersKey is the standard key for referring to a http.Response's Headers.
	ResponseHeadersKey = "response.headers"
	// ReasonKey is the standard key for referring to a reason for a change.
	ReasonKey = "reason"
	// URLQueryKey is the standard key for referring to a URL query.
	URLQueryKey = "url.query"
	// SearchQueryKey is the standard key for referring to a search query parameter value.
	SearchQueryKey = "search_query"
	// UserAgentOSKey is the standard key for referring to a user agent's OS.
	UserAgentOSKey = "os"
	// UserAgentBotKey is the standard key for referring to a user agent's bot status.
	UserAgentBotKey = "is_bot"
	// UserAgentMobileKey is the standard key for referring to user agent's mobile status.
	UserAgentMobileKey = "is_mobile"
	// ValidationErrorKey is the standard key for referring to a struct validation error.
	ValidationErrorKey = "validation_error"
	// UserDataAggregationReportIDKey is the standard key for referring to a user data aggregation report.
	UserDataAggregationReportIDKey = "user_data_aggregation_report" + idSuffix
	// IndexNameKey is the standard key for referring to a given search index.
	IndexNameKey = "index.name"
	// UseDatabaseKey is the standard key for referring to whether or not the database was used in search.
	UseDatabaseKey = "use_database"

	// ServiceSettingIDKey is the standard key for referring to a service setting's ID.
	ServiceSettingIDKey = "service_setting" + idSuffix

	// ServiceSettingNameKey is the standard key for referring to a service setting's ID.
	ServiceSettingNameKey = "service_setting.name"

	// ServiceSettingConfigurationIDKey is the standard key for referring to a service setting configuration's ID.
	ServiceSettingConfigurationIDKey = "service_setting_configuration" + idSuffix

	// ServiceSettingConfigurationNameKey is the standard key for referring to a service setting configuration's Name.
	ServiceSettingConfigurationNameKey = "service_setting_configuration.name"

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
