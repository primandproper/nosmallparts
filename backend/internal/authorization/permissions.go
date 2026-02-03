package authorization

import (
	"github.com/mikespook/gorbac/v2"
)

var _ gorbac.Permission = (*Permission)(nil)

type (
	// Permission is a simple string alias.
	Permission string
)

// ID implements the gorbac Permission interface.
func (p Permission) ID() string {
	return string(p)
}

// Match implements the gorbac Permission interface.
func (p Permission) Match(perm gorbac.Permission) bool {
	return p.ID() == perm.ID()
}

var (
	// ServiceAdminPermissions is every service admin permission.
	ServiceAdminPermissions = []gorbac.Permission{
		ReadUserDataPermission,
		UpdateUserStatusPermission,
		ReadUserPermission,
		SearchUserPermission,
		ArchiveUserPermission,
		CreateOAuth2ClientsPermission,
		ArchiveOAuth2ClientsPermission,
		ArchiveServiceSettingsPermission,
		CreateUserNotificationsPermission,
		ImpersonateUserPermission,
		PublishArbitraryQueueMessagePermission,
		// only admins can arbitrarily create these via the API, this is exclusively for integration test purposes.
		CreateServiceSettingsPermission,
		CreateWaitlistsPermission,
		UpdateWaitlistsPermission,
		ArchiveWaitlistsPermission,
	}

	// ServiceDataAdminPermissions is every service admin permission.
	ServiceDataAdminPermissions = []gorbac.Permission{}

	// AccountAdminPermissions is every account admin permission.
	AccountAdminPermissions = []gorbac.Permission{
		UpdateAccountPermission,
		ArchiveAccountPermission,
		TransferAccountPermission,
		InviteUserToAccountPermission,
		ModifyMemberPermissionsForAccountPermission,
		RemoveMemberAccountPermission,
		CreateWebhooksPermission,
		UpdateWebhooksPermission,
		ArchiveWebhooksPermission,
		CreateIssueReportsPermission,
		UpdateIssueReportsPermission,
		ArchiveIssueReportsPermission,
		CreateWebhookTriggerEventsPermission,
		ArchiveWebhookTriggerEventsPermission,
	}

	// AccountMemberPermissions is every account member permission.
	AccountMemberPermissions = []gorbac.Permission{
		ReadWebhooksPermission,
		ReadIssueReportsPermission,
		ReadAuditLogEntriesPermission,
		ReadOAuth2ClientsPermission,
		ReadServiceSettingsPermission,
		SearchServiceSettingsPermission,
		CreateUploadedMediaPermission,
		ReadUploadedMediaPermission,
		UpdateUploadedMediaPermission,
		ArchiveUploadedMediaPermission,
		CreateServiceSettingConfigurationsPermission,
		ReadServiceSettingConfigurationsPermission,
		UpdateServiceSettingConfigurationsPermission,
		ArchiveServiceSettingConfigurationsPermission,
		ReadUserNotificationsPermission,
		UpdateUserNotificationsPermission,
		CreateWaitlistSignupsPermission,
		UpdateWaitlistSignupsPermission,
		ArchiveWaitlistSignupsPermission,
		ReadWaitlistsPermission,
		ReadWaitlistSignupsPermission,
		CreateWaitlistSignupsPermission,
		UpdateWaitlistSignupsPermission,
		ArchiveWaitlistSignupsPermission,
	}
)

func init() {
	// assign service admin permissions.
	for _, perm := range ServiceAdminPermissions {
		must(serviceAdmin.Assign(perm))
	}

	// assign service data admin permissions.
	for _, perm := range ServiceDataAdminPermissions {
		must(serviceDataAdmin.Assign(perm))
		must(serviceAdmin.Assign(perm)) // these aren't separate things yet
	}

	// assign account admin permissions.
	for _, perm := range AccountAdminPermissions {
		must(accountAdmin.Assign(perm))
	}

	// assign account member permissions.
	for _, perm := range AccountMemberPermissions {
		must(accountMember.Assign(perm))
	}
}
