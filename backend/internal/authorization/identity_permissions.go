package authorization

import (
	"github.com/mikespook/gorbac/v2"
)

const (
	// UpdateAccountPermission is an account admin permission.
	UpdateAccountPermission Permission = "update.account"
	// ArchiveAccountPermission is an account admin permission.
	ArchiveAccountPermission Permission = "archive.account"
	// InviteUserToAccountPermission is an account admin permission.
	InviteUserToAccountPermission Permission = "account.add.member"
	// ModifyMemberPermissionsForAccountPermission is an account admin permission.
	ModifyMemberPermissionsForAccountPermission Permission = "account.membership.modify"
	// RemoveMemberAccountPermission is an account admin permission.
	RemoveMemberAccountPermission Permission = "remove_member.account"
	// TransferAccountPermission is an account admin permission.
	TransferAccountPermission Permission = "transfer.account"
)

var (
	// IdentityPermissions contains all identity-related permissions (accounts, users, memberships).
	IdentityPermissions = []gorbac.Permission{
		UpdateAccountPermission,
		ArchiveAccountPermission,
		InviteUserToAccountPermission,
		ModifyMemberPermissionsForAccountPermission,
		RemoveMemberAccountPermission,
		TransferAccountPermission,
	}
)
