package authorization

import (
	"github.com/mikespook/gorbac/v2"
)

const (
	// CreateWaitlistsPermission is an account admin permission.
	CreateWaitlistsPermission Permission = "create.waitlists"
	// ReadWaitlistsPermission is an account admin permission.
	ReadWaitlistsPermission Permission = "read.waitlists"
	// UpdateWaitlistsPermission is an account admin permission.
	UpdateWaitlistsPermission Permission = "update.waitlists"
	// ArchiveWaitlistsPermission is an account admin permission.
	ArchiveWaitlistsPermission Permission = "archive.waitlists"
	// CreateWaitlistSignupsPermission is an account admin permission.
	CreateWaitlistSignupsPermission Permission = "create.waitlist_signups"
	// ReadWaitlistSignupsPermission is an account admin permission.
	ReadWaitlistSignupsPermission Permission = "read.waitlist_signups"
	// UpdateWaitlistSignupsPermission is an account admin permission.
	UpdateWaitlistSignupsPermission Permission = "update.waitlist_signups"
	// ArchiveWaitlistSignupsPermission is an account admin permission.
	ArchiveWaitlistSignupsPermission Permission = "archive.waitlist_signups"
)

var (
	// WaitlistsPermissions contains all waitlist-related permissions.
	WaitlistsPermissions = []gorbac.Permission{
		CreateWaitlistsPermission,
		ReadWaitlistsPermission,
		UpdateWaitlistsPermission,
		ArchiveWaitlistsPermission,
		CreateWaitlistSignupsPermission,
		ReadWaitlistSignupsPermission,
		UpdateWaitlistSignupsPermission,
		ArchiveWaitlistSignupsPermission,
	}
)
