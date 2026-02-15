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

	// CreateScriptTagsPermission is a service admin permission for creating script tags.
	CreateScriptTagsPermission Permission = "create.script_tags"
	// AddTagToPlayScriptPermission is a service admin permission for tagging play scripts.
	AddTagToPlayScriptPermission Permission = "add_tag_to.play_script"
	// RemoveTagFromPlayScriptPermission is a service admin permission for removing tags from play scripts.
	RemoveTagFromPlayScriptPermission Permission = "remove_tag_from.play_script"
)

var (
	// PlayScriptsPermissions contains all play script-related permissions.
	PlayScriptsPermissions = []gorbac.Permission{
		CreatePlayScriptsPermission,
		ReadPlayScriptsPermission,
		UpdatePlayScriptsPermission,
		ArchivePlayScriptsPermission,
		CreateScriptTagsPermission,
		AddTagToPlayScriptPermission,
		RemoveTagFromPlayScriptPermission,
	}
)
