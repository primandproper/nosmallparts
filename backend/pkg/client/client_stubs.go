package client

import (
	"context"

	auditgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/audit"
	authgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/auth"
	dataprivacygrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/dataprivacy"
	identitygrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/identity"
	internalopsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/internalops"
	issuereportsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/issue_reports"
	notificationsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/notifications"
	oauthgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/oauth"
	settingsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/settings"
	uploadedmediagrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/uploaded_media"
	waitlistsgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/waitlists"
	webhooksgrpc "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/grpc/generated/services/webhooks"

	"google.golang.org/grpc"
)

// Stub types so *httpClient can satisfy Client by embedding these and overriding the few implemented methods.

type stubAuthClient struct{}
type stubIdentityClient struct{}
type stubAuditClient struct{}
type stubDataPrivacyClient struct{}
type stubInternalOpsClient struct{}
type stubIssueReportsClient struct{}
type stubNotificationsClient struct{}
type stubOAuthClient struct{}
type stubSettingsClient struct{}
type stubUploadedMediaClient struct{}
type stubWaitlistsClient struct{}
type stubWebhooksClient struct{}

// Auth (stub; httpClient overrides LoginForToken, AdminLoginForToken, VerifyTOTPSecret)
func (stubAuthClient) GetAuthStatus(context.Context, *authgrpc.GetAuthStatusRequest, ...grpc.CallOption) (*authgrpc.GetAuthStatusResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) ExchangeToken(context.Context, *authgrpc.ExchangeTokenRequest, ...grpc.CallOption) (*authgrpc.ExchangeTokenResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) AdminLoginForToken(context.Context, *authgrpc.AdminLoginForTokenRequest, ...grpc.CallOption) (*authgrpc.LoginForTokenResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) CheckPermissions(context.Context, *authgrpc.UserPermissionsRequestInput, ...grpc.CallOption) (*authgrpc.UserPermissionsResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) GetActiveAccount(context.Context, *authgrpc.GetActiveAccountRequest, ...grpc.CallOption) (*authgrpc.GetActiveAccountResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) GetSelf(context.Context, *authgrpc.GetSelfRequest, ...grpc.CallOption) (*authgrpc.GetSelfResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) LoginForToken(context.Context, *authgrpc.LoginForTokenRequest, ...grpc.CallOption) (*authgrpc.LoginForTokenResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) RedeemPasswordResetToken(context.Context, *authgrpc.RedeemPasswordResetTokenRequest, ...grpc.CallOption) (*authgrpc.RedeemPasswordResetTokenResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) RefreshTOTPSecret(context.Context, *authgrpc.RefreshTOTPSecretRequest, ...grpc.CallOption) (*authgrpc.RefreshTOTPSecretResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) RequestEmailVerificationEmail(context.Context, *authgrpc.RequestEmailVerificationEmailRequest, ...grpc.CallOption) (*authgrpc.RequestEmailVerificationEmailResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) RequestPasswordResetToken(context.Context, *authgrpc.RequestPasswordResetTokenRequest, ...grpc.CallOption) (*authgrpc.RequestPasswordResetTokenResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) RequestUsernameReminder(context.Context, *authgrpc.RequestUsernameReminderRequest, ...grpc.CallOption) (*authgrpc.RequestUsernameReminderResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) VerifyEmailAddress(context.Context, *authgrpc.VerifyEmailAddressRequest, ...grpc.CallOption) (*authgrpc.VerifyEmailAddressResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) VerifyTOTPSecret(context.Context, *authgrpc.VerifyTOTPSecretRequest, ...grpc.CallOption) (*authgrpc.VerifyTOTPSecretResponse, error) {
	return nil, errUnimplemented
}
func (stubAuthClient) UpdatePassword(context.Context, *authgrpc.UpdatePasswordRequest, ...grpc.CallOption) (*authgrpc.UpdatePasswordResponse, error) {
	return nil, errUnimplemented
}

// Identity (stub; httpClient overrides CreateUser)
func (stubIdentityClient) AdminUpdateUserStatus(context.Context, *identitygrpc.AdminUpdateUserStatusRequest, ...grpc.CallOption) (*identitygrpc.AdminUpdateUserStatusResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) AcceptAccountInvitation(context.Context, *identitygrpc.AcceptAccountInvitationRequest, ...grpc.CallOption) (*identitygrpc.AcceptAccountInvitationResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) ArchiveAccount(context.Context, *identitygrpc.ArchiveAccountRequest, ...grpc.CallOption) (*identitygrpc.ArchiveAccountResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) ArchiveUserMembership(context.Context, *identitygrpc.ArchiveUserMembershipRequest, ...grpc.CallOption) (*identitygrpc.ArchiveUserMembershipResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) ArchiveUser(context.Context, *identitygrpc.ArchiveUserRequest, ...grpc.CallOption) (*identitygrpc.ArchiveUserResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) CancelAccountInvitation(context.Context, *identitygrpc.CancelAccountInvitationRequest, ...grpc.CallOption) (*identitygrpc.CancelAccountInvitationResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) CreateAccount(context.Context, *identitygrpc.CreateAccountRequest, ...grpc.CallOption) (*identitygrpc.CreateAccountResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) CreateAccountInvitation(context.Context, *identitygrpc.CreateAccountInvitationRequest, ...grpc.CallOption) (*identitygrpc.CreateAccountInvitationResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) CreateUser(context.Context, *identitygrpc.CreateUserRequest, ...grpc.CallOption) (*identitygrpc.CreateUserResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) GetAccount(context.Context, *identitygrpc.GetAccountRequest, ...grpc.CallOption) (*identitygrpc.GetAccountResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) GetAccountInvitation(context.Context, *identitygrpc.GetAccountInvitationRequest, ...grpc.CallOption) (*identitygrpc.GetAccountInvitationResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) GetAccounts(context.Context, *identitygrpc.GetAccountsRequest, ...grpc.CallOption) (*identitygrpc.GetAccountsResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) GetAccountsForUser(context.Context, *identitygrpc.GetAccountsForUserRequest, ...grpc.CallOption) (*identitygrpc.GetAccountsForUserResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) GetReceivedAccountInvitations(context.Context, *identitygrpc.GetReceivedAccountInvitationsRequest, ...grpc.CallOption) (*identitygrpc.GetReceivedAccountInvitationsResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) GetSentAccountInvitations(context.Context, *identitygrpc.GetSentAccountInvitationsRequest, ...grpc.CallOption) (*identitygrpc.GetSentAccountInvitationsResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) GetUser(context.Context, *identitygrpc.GetUserRequest, ...grpc.CallOption) (*identitygrpc.GetUserResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) RejectAccountInvitation(context.Context, *identitygrpc.RejectAccountInvitationRequest, ...grpc.CallOption) (*identitygrpc.RejectAccountInvitationResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) GetUsers(context.Context, *identitygrpc.GetUsersRequest, ...grpc.CallOption) (*identitygrpc.GetUsersResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) GetUsersForAccount(context.Context, *identitygrpc.GetUsersForAccountRequest, ...grpc.CallOption) (*identitygrpc.GetUsersForAccountResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) SearchForUsers(context.Context, *identitygrpc.SearchForUsersRequest, ...grpc.CallOption) (*identitygrpc.SearchForUsersResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) SetDefaultAccount(context.Context, *identitygrpc.SetDefaultAccountRequest, ...grpc.CallOption) (*identitygrpc.SetDefaultAccountResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) TransferAccountOwnership(context.Context, *identitygrpc.TransferAccountOwnershipRequest, ...grpc.CallOption) (*identitygrpc.TransferAccountOwnershipResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) UpdateAccount(context.Context, *identitygrpc.UpdateAccountRequest, ...grpc.CallOption) (*identitygrpc.UpdateAccountResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) UpdateAccountMemberPermissions(context.Context, *identitygrpc.UpdateAccountMemberPermissionsRequest, ...grpc.CallOption) (*identitygrpc.UpdateAccountMemberPermissionsResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) UpdateUserDetails(context.Context, *identitygrpc.UpdateUserDetailsRequest, ...grpc.CallOption) (*identitygrpc.UpdateUserDetailsResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) UpdateUserEmailAddress(context.Context, *identitygrpc.UpdateUserEmailAddressRequest, ...grpc.CallOption) (*identitygrpc.UpdateUserEmailAddressResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) UpdateUserUsername(context.Context, *identitygrpc.UpdateUserUsernameRequest, ...grpc.CallOption) (*identitygrpc.UpdateUserUsernameResponse, error) {
	return nil, errUnimplemented
}
func (stubIdentityClient) UploadUserAvatar(context.Context, *identitygrpc.UploadUserAvatarRequest, ...grpc.CallOption) (*identitygrpc.UploadUserAvatarResponse, error) {
	return nil, errUnimplemented
}

// Audit
func (stubAuditClient) GetAuditLogEntriesForAccount(context.Context, *auditgrpc.GetAuditLogEntriesForAccountRequest, ...grpc.CallOption) (*auditgrpc.GetAuditLogEntriesForAccountResponse, error) {
	return nil, errUnimplemented
}
func (stubAuditClient) GetAuditLogEntriesForUser(context.Context, *auditgrpc.GetAuditLogEntriesForUserRequest, ...grpc.CallOption) (*auditgrpc.GetAuditLogEntriesForUserResponse, error) {
	return nil, errUnimplemented
}
func (stubAuditClient) GetAuditLogEntryByID(context.Context, *auditgrpc.GetAuditLogEntryByIDRequest, ...grpc.CallOption) (*auditgrpc.GetAuditLogEntryByIDResponse, error) {
	return nil, errUnimplemented
}

// DataPrivacy
func (stubDataPrivacyClient) AggregateUserDataReport(context.Context, *dataprivacygrpc.AggregateUserDataReportRequest, ...grpc.CallOption) (*dataprivacygrpc.AggregateUserDataReportResponse, error) {
	return nil, errUnimplemented
}
func (stubDataPrivacyClient) DestroyAllUserData(context.Context, *dataprivacygrpc.DestroyAllUserDataRequest, ...grpc.CallOption) (*dataprivacygrpc.DestroyAllUserDataResponse, error) {
	return nil, errUnimplemented
}
func (stubDataPrivacyClient) FetchUserDataReport(context.Context, *dataprivacygrpc.FetchUserDataReportRequest, ...grpc.CallOption) (*dataprivacygrpc.FetchUserDataReportResponse, error) {
	return nil, errUnimplemented
}

// InternalOps
func (stubInternalOpsClient) PublishArbitraryQueueMessage(context.Context, *internalopsgrpc.PublishArbitraryQueueMessageRequest, ...grpc.CallOption) (*internalopsgrpc.PublishArbitraryQueueMessageResponse, error) {
	return nil, errUnimplemented
}

// IssueReports
func (stubIssueReportsClient) CreateIssueReport(context.Context, *issuereportsgrpc.CreateIssueReportRequest, ...grpc.CallOption) (*issuereportsgrpc.CreateIssueReportResponse, error) {
	return nil, errUnimplemented
}
func (stubIssueReportsClient) GetIssueReport(context.Context, *issuereportsgrpc.GetIssueReportRequest, ...grpc.CallOption) (*issuereportsgrpc.GetIssueReportResponse, error) {
	return nil, errUnimplemented
}
func (stubIssueReportsClient) GetIssueReports(context.Context, *issuereportsgrpc.GetIssueReportsRequest, ...grpc.CallOption) (*issuereportsgrpc.GetIssueReportsResponse, error) {
	return nil, errUnimplemented
}
func (stubIssueReportsClient) GetIssueReportsForAccount(context.Context, *issuereportsgrpc.GetIssueReportsForAccountRequest, ...grpc.CallOption) (*issuereportsgrpc.GetIssueReportsForAccountResponse, error) {
	return nil, errUnimplemented
}
func (stubIssueReportsClient) GetIssueReportsForTable(context.Context, *issuereportsgrpc.GetIssueReportsForTableRequest, ...grpc.CallOption) (*issuereportsgrpc.GetIssueReportsForTableResponse, error) {
	return nil, errUnimplemented
}
func (stubIssueReportsClient) GetIssueReportsForRecord(context.Context, *issuereportsgrpc.GetIssueReportsForRecordRequest, ...grpc.CallOption) (*issuereportsgrpc.GetIssueReportsForRecordResponse, error) {
	return nil, errUnimplemented
}
func (stubIssueReportsClient) UpdateIssueReport(context.Context, *issuereportsgrpc.UpdateIssueReportRequest, ...grpc.CallOption) (*issuereportsgrpc.UpdateIssueReportResponse, error) {
	return nil, errUnimplemented
}
func (stubIssueReportsClient) ArchiveIssueReport(context.Context, *issuereportsgrpc.ArchiveIssueReportRequest, ...grpc.CallOption) (*issuereportsgrpc.ArchiveIssueReportResponse, error) {
	return nil, errUnimplemented
}

// Notifications
func (stubNotificationsClient) GetUserNotification(context.Context, *notificationsgrpc.GetUserNotificationRequest, ...grpc.CallOption) (*notificationsgrpc.GetUserNotificationResponse, error) {
	return nil, errUnimplemented
}
func (stubNotificationsClient) GetUserNotifications(context.Context, *notificationsgrpc.GetUserNotificationsRequest, ...grpc.CallOption) (*notificationsgrpc.GetUserNotificationsResponse, error) {
	return nil, errUnimplemented
}
func (stubNotificationsClient) UpdateUserNotification(context.Context, *notificationsgrpc.UpdateUserNotificationRequest, ...grpc.CallOption) (*notificationsgrpc.UpdateUserNotificationResponse, error) {
	return nil, errUnimplemented
}

// OAuth
func (stubOAuthClient) ArchiveOAuth2Client(context.Context, *oauthgrpc.ArchiveOAuth2ClientRequest, ...grpc.CallOption) (*oauthgrpc.ArchiveOAuth2ClientResponse, error) {
	return nil, errUnimplemented
}
func (stubOAuthClient) CreateOAuth2Client(context.Context, *oauthgrpc.CreateOAuth2ClientRequest, ...grpc.CallOption) (*oauthgrpc.CreateOAuth2ClientResponse, error) {
	return nil, errUnimplemented
}
func (stubOAuthClient) GetOAuth2Client(context.Context, *oauthgrpc.GetOAuth2ClientRequest, ...grpc.CallOption) (*oauthgrpc.GetOAuth2ClientResponse, error) {
	return nil, errUnimplemented
}
func (stubOAuthClient) GetOAuth2Clients(context.Context, *oauthgrpc.GetOAuth2ClientsRequest, ...grpc.CallOption) (*oauthgrpc.GetOAuth2ClientsResponse, error) {
	return nil, errUnimplemented
}

// Settings
func (stubSettingsClient) GetServiceSettingConfigurationsForAccount(context.Context, *settingsgrpc.GetServiceSettingConfigurationsForAccountRequest, ...grpc.CallOption) (*settingsgrpc.GetServiceSettingConfigurationsForAccountResponse, error) {
	return nil, errUnimplemented
}
func (stubSettingsClient) GetServiceSettingConfigurationsForUser(context.Context, *settingsgrpc.GetServiceSettingConfigurationsForUserRequest, ...grpc.CallOption) (*settingsgrpc.GetServiceSettingConfigurationsForUserResponse, error) {
	return nil, errUnimplemented
}
func (stubSettingsClient) SearchForServiceSettings(context.Context, *settingsgrpc.SearchForServiceSettingsRequest, ...grpc.CallOption) (*settingsgrpc.SearchForServiceSettingsResponse, error) {
	return nil, errUnimplemented
}
func (stubSettingsClient) ArchiveServiceSetting(context.Context, *settingsgrpc.ArchiveServiceSettingRequest, ...grpc.CallOption) (*settingsgrpc.ArchiveServiceSettingResponse, error) {
	return nil, errUnimplemented
}
func (stubSettingsClient) ArchiveServiceSettingConfiguration(context.Context, *settingsgrpc.ArchiveServiceSettingConfigurationRequest, ...grpc.CallOption) (*settingsgrpc.ArchiveServiceSettingConfigurationResponse, error) {
	return nil, errUnimplemented
}
func (stubSettingsClient) CreateServiceSetting(context.Context, *settingsgrpc.CreateServiceSettingRequest, ...grpc.CallOption) (*settingsgrpc.CreateServiceSettingResponse, error) {
	return nil, errUnimplemented
}
func (stubSettingsClient) CreateServiceSettingConfiguration(context.Context, *settingsgrpc.CreateServiceSettingConfigurationRequest, ...grpc.CallOption) (*settingsgrpc.CreateServiceSettingConfigurationResponse, error) {
	return nil, errUnimplemented
}
func (stubSettingsClient) GetServiceSetting(context.Context, *settingsgrpc.GetServiceSettingRequest, ...grpc.CallOption) (*settingsgrpc.GetServiceSettingResponse, error) {
	return nil, errUnimplemented
}
func (stubSettingsClient) GetServiceSettingConfigurationByName(context.Context, *settingsgrpc.GetServiceSettingConfigurationByNameRequest, ...grpc.CallOption) (*settingsgrpc.GetServiceSettingConfigurationByNameResponse, error) {
	return nil, errUnimplemented
}
func (stubSettingsClient) GetServiceSettings(context.Context, *settingsgrpc.GetServiceSettingsRequest, ...grpc.CallOption) (*settingsgrpc.GetServiceSettingsResponse, error) {
	return nil, errUnimplemented
}
func (stubSettingsClient) UpdateServiceSettingConfiguration(context.Context, *settingsgrpc.UpdateServiceSettingConfigurationRequest, ...grpc.CallOption) (*settingsgrpc.UpdateServiceSettingConfigurationResponse, error) {
	return nil, errUnimplemented
}

// UploadedMedia (includes streaming; we only stub the unary and stream type)
func (stubUploadedMediaClient) Upload(context.Context, ...grpc.CallOption) (grpc.ClientStreamingClient[uploadedmediagrpc.UploadRequest, uploadedmediagrpc.UploadResponse], error) {
	return nil, errUnimplemented
}
func (stubUploadedMediaClient) CreateUploadedMedia(context.Context, *uploadedmediagrpc.CreateUploadedMediaRequest, ...grpc.CallOption) (*uploadedmediagrpc.CreateUploadedMediaResponse, error) {
	return nil, errUnimplemented
}
func (stubUploadedMediaClient) GetUploadedMedia(context.Context, *uploadedmediagrpc.GetUploadedMediaRequest, ...grpc.CallOption) (*uploadedmediagrpc.GetUploadedMediaResponse, error) {
	return nil, errUnimplemented
}
func (stubUploadedMediaClient) GetUploadedMediaWithIDs(context.Context, *uploadedmediagrpc.GetUploadedMediaWithIDsRequest, ...grpc.CallOption) (*uploadedmediagrpc.GetUploadedMediaWithIDsResponse, error) {
	return nil, errUnimplemented
}
func (stubUploadedMediaClient) GetUploadedMediaForUser(context.Context, *uploadedmediagrpc.GetUploadedMediaForUserRequest, ...grpc.CallOption) (*uploadedmediagrpc.GetUploadedMediaForUserResponse, error) {
	return nil, errUnimplemented
}
func (stubUploadedMediaClient) UpdateUploadedMedia(context.Context, *uploadedmediagrpc.UpdateUploadedMediaRequest, ...grpc.CallOption) (*uploadedmediagrpc.UpdateUploadedMediaResponse, error) {
	return nil, errUnimplemented
}
func (stubUploadedMediaClient) ArchiveUploadedMedia(context.Context, *uploadedmediagrpc.ArchiveUploadedMediaRequest, ...grpc.CallOption) (*uploadedmediagrpc.ArchiveUploadedMediaResponse, error) {
	return nil, errUnimplemented
}

// Waitlists
func (stubWaitlistsClient) CreateWaitlist(context.Context, *waitlistsgrpc.CreateWaitlistRequest, ...grpc.CallOption) (*waitlistsgrpc.CreateWaitlistResponse, error) {
	return nil, errUnimplemented
}
func (stubWaitlistsClient) GetWaitlist(context.Context, *waitlistsgrpc.GetWaitlistRequest, ...grpc.CallOption) (*waitlistsgrpc.GetWaitlistResponse, error) {
	return nil, errUnimplemented
}
func (stubWaitlistsClient) GetWaitlists(context.Context, *waitlistsgrpc.GetWaitlistsRequest, ...grpc.CallOption) (*waitlistsgrpc.GetWaitlistsResponse, error) {
	return nil, errUnimplemented
}
func (stubWaitlistsClient) GetActiveWaitlists(context.Context, *waitlistsgrpc.GetActiveWaitlistsRequest, ...grpc.CallOption) (*waitlistsgrpc.GetActiveWaitlistsResponse, error) {
	return nil, errUnimplemented
}
func (stubWaitlistsClient) UpdateWaitlist(context.Context, *waitlistsgrpc.UpdateWaitlistRequest, ...grpc.CallOption) (*waitlistsgrpc.UpdateWaitlistResponse, error) {
	return nil, errUnimplemented
}
func (stubWaitlistsClient) ArchiveWaitlist(context.Context, *waitlistsgrpc.ArchiveWaitlistRequest, ...grpc.CallOption) (*waitlistsgrpc.ArchiveWaitlistResponse, error) {
	return nil, errUnimplemented
}
func (stubWaitlistsClient) WaitlistIsNotExpired(context.Context, *waitlistsgrpc.WaitlistIsNotExpiredRequest, ...grpc.CallOption) (*waitlistsgrpc.WaitlistIsNotExpiredResponse, error) {
	return nil, errUnimplemented
}
func (stubWaitlistsClient) CreateWaitlistSignup(context.Context, *waitlistsgrpc.CreateWaitlistSignupRequest, ...grpc.CallOption) (*waitlistsgrpc.CreateWaitlistSignupResponse, error) {
	return nil, errUnimplemented
}
func (stubWaitlistsClient) GetWaitlistSignup(context.Context, *waitlistsgrpc.GetWaitlistSignupRequest, ...grpc.CallOption) (*waitlistsgrpc.GetWaitlistSignupResponse, error) {
	return nil, errUnimplemented
}
func (stubWaitlistsClient) GetWaitlistSignupsForWaitlist(context.Context, *waitlistsgrpc.GetWaitlistSignupsForWaitlistRequest, ...grpc.CallOption) (*waitlistsgrpc.GetWaitlistSignupsForWaitlistResponse, error) {
	return nil, errUnimplemented
}
func (stubWaitlistsClient) UpdateWaitlistSignup(context.Context, *waitlistsgrpc.UpdateWaitlistSignupRequest, ...grpc.CallOption) (*waitlistsgrpc.UpdateWaitlistSignupResponse, error) {
	return nil, errUnimplemented
}
func (stubWaitlistsClient) ArchiveWaitlistSignup(context.Context, *waitlistsgrpc.ArchiveWaitlistSignupRequest, ...grpc.CallOption) (*waitlistsgrpc.ArchiveWaitlistSignupResponse, error) {
	return nil, errUnimplemented
}

// Webhooks
func (stubWebhooksClient) ArchiveWebhook(context.Context, *webhooksgrpc.ArchiveWebhookRequest, ...grpc.CallOption) (*webhooksgrpc.ArchiveWebhookResponse, error) {
	return nil, errUnimplemented
}
func (stubWebhooksClient) ArchiveWebhookTriggerEvent(context.Context, *webhooksgrpc.ArchiveWebhookTriggerEventRequest, ...grpc.CallOption) (*webhooksgrpc.ArchiveWebhookTriggerEventResponse, error) {
	return nil, errUnimplemented
}
func (stubWebhooksClient) CreateWebhook(context.Context, *webhooksgrpc.CreateWebhookRequest, ...grpc.CallOption) (*webhooksgrpc.CreateWebhookResponse, error) {
	return nil, errUnimplemented
}
func (stubWebhooksClient) AddWebhookTriggerEvent(context.Context, *webhooksgrpc.AddWebhookTriggerEventRequest, ...grpc.CallOption) (*webhooksgrpc.AddWebhookTriggerEventResponse, error) {
	return nil, errUnimplemented
}
func (stubWebhooksClient) GetWebhook(context.Context, *webhooksgrpc.GetWebhookRequest, ...grpc.CallOption) (*webhooksgrpc.GetWebhookResponse, error) {
	return nil, errUnimplemented
}
func (stubWebhooksClient) GetWebhooks(context.Context, *webhooksgrpc.GetWebhooksRequest, ...grpc.CallOption) (*webhooksgrpc.GetWebhooksResponse, error) {
	return nil, errUnimplemented
}
