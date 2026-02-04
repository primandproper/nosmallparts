package authorization

import (
	"github.com/mikespook/gorbac/v2"
)

const (
	// CreatePlayScriptsPermission is an account admin permission.
	CreatePlayScriptsPermission Permission = "create.play_scripts"
	// ReadPlayScriptsPermission is an account admin permission.
	ReadPlayScriptsPermission Permission = "read.play_scripts"
	// UpdatePlayScriptsPermission is an account admin permission.
	UpdatePlayScriptsPermission Permission = "update.play_scripts"
	// ArchivePlayScriptsPermission is an account admin permission.
	ArchivePlayScriptsPermission Permission = "archive.play_scripts"
)

var (
	// PlayScriptsPermissions contains all play script-related permissions.
	PlayScriptsPermissions = []gorbac.Permission{
		CreatePlayScriptsPermission,
		ReadPlayScriptsPermission,
		UpdatePlayScriptsPermission,
		ArchivePlayScriptsPermission,
	}
)
