package keys

const (
	idSuffix = ".id"

	// AccountIDKey is the standard key for referring to an account ID.
	AccountIDKey = "account" + idSuffix
	// AccountInvitationKey is the standard key for referring to an account invitation.
	AccountInvitationKey = "account_invitation"
	// AccountInvitationIDKey is the standard key for referring to an account invitation ID.
	AccountInvitationIDKey = AccountInvitationKey + idSuffix
	// AccountInvitationTokenKey is the standard key for referring to an account invitation token.
	AccountInvitationTokenKey = "account_invitation.token"
	// AuditLogEntryIDKey is the standard key for referring to an audit log entry's ID.
	AuditLogEntryIDKey = "audit_log_entry" + idSuffix
	// AuditLogEntryResourceTypesKey is the standard key for referring to an audit log entry's resource type.
	AuditLogEntryResourceTypesKey = "audit_log_entry.resource_types"
	// IssueReportIDKey is the standard key for referring to an issue report ID.
	IssueReportIDKey = "issue_report" + idSuffix
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
	// PasswordResetTokenKey is the standard key for referring to a password reset token.
	PasswordResetTokenKey = "password_reset_token"
	// PasswordResetTokenIDKey is the standard key for referring to a password reset token's ID.
	PasswordResetTokenIDKey = PasswordResetTokenKey + idSuffix
	// PlayScriptIDKey is the standard key for referring to a play script ID.
	PlayScriptIDKey = "play_script" + idSuffix
	// ServiceSettingIDKey is the standard key for referring to a service setting's ID.
	ServiceSettingIDKey = "service_setting" + idSuffix
	// ServiceSettingNameKey is the standard key for referring to a service setting's name.
	ServiceSettingNameKey = "service_setting.name"
	// ServiceSettingConfigurationIDKey is the standard key for referring to a service setting configuration's ID.
	ServiceSettingConfigurationIDKey = "service_setting_configuration" + idSuffix
	// ServiceSettingConfigurationNameKey is the standard key for referring to a service setting configuration's Name.
	ServiceSettingConfigurationNameKey = "service_setting_configuration.name"
	// StageIDKey is the standard key for referring to a stage ID.
	StageIDKey = "stage" + idSuffix
	// UploadedMediaIDKey is the standard key for referring to an uploaded media ID.
	UploadedMediaIDKey = "uploaded_media" + idSuffix
	// UserEmailAddressKey is the standard key for referring to a user's email address.
	UserEmailAddressKey = "user.email_address"
	// UserEmailVerificationTokenKey is the standard key for referring to a user email verification token.
	// #nosec G101
	UserEmailVerificationTokenKey = "user.email_verification_token"
	// UserDataAggregationReportIDKey is the standard key for referring to a user data aggregation report.
	UserDataAggregationReportIDKey = "user_data_aggregation_report" + idSuffix
	// UserNotificationIDKey is the standard key for referring to a user notification ID.
	UserNotificationIDKey = "user_notification" + idSuffix
	// WaitlistIDKey is the standard key for referring to a waitlist ID.
	WaitlistIDKey = "waitlist" + idSuffix
	// WaitlistSignupIDKey is the standard key for referring to a waitlist signup ID.
	WaitlistSignupIDKey = "waitlist_signup" + idSuffix
	// WebhookIDKey is the standard key for referring to a webhook's ID.
	WebhookIDKey = "webhook" + idSuffix
	// WebhookTriggerEventIDKey is the standard key for referring to a webhook trigger event's ID.
	WebhookTriggerEventIDKey = "webhook_trigger_event" + idSuffix
)
