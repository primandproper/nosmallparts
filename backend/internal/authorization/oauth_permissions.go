package authorization

import (
	"github.com/mikespook/gorbac/v2"
)

const (
	// CreateOAuth2ClientsPermission is an account admin permission.
	CreateOAuth2ClientsPermission Permission = "create.oauth2_clients"
	// ReadOAuth2ClientsPermission is an account admin permission.
	ReadOAuth2ClientsPermission Permission = "read.oauth2_clients"
	// ArchiveOAuth2ClientsPermission is an account admin permission.
	ArchiveOAuth2ClientsPermission Permission = "archive.oauth2_clients"
)

var (
	// OAuthPermissions contains all OAuth-related permissions.
	OAuthPermissions = []gorbac.Permission{
		CreateOAuth2ClientsPermission,
		ReadOAuth2ClientsPermission,
		ArchiveOAuth2ClientsPermission,
	}
)
