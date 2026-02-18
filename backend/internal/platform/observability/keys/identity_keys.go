package keys

const (
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
	// UserEmailAddressKey is the standard key for referring to a user's email address.
	UserEmailAddressKey = "user.email_address"
	// UserIsServiceAdminKey is the standard key for referring to a user's admin status.
	UserIsServiceAdminKey = "user.is_admin"
	// UsernameKey is the standard key for referring to a username.
	UsernameKey = "user.username"
	// #nosec G101 UserEmailVerificationTokenKey is the standard key for referring to a username.
	UserEmailVerificationTokenKey = "user.email_verification_token"
	// UserDataAggregationReportIDKey is the standard key for referring to a user data aggregation report.
	UserDataAggregationReportIDKey = "user_data_aggregation_report" + idSuffix
)
