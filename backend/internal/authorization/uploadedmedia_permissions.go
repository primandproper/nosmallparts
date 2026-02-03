package authorization

import (
	"github.com/mikespook/gorbac/v2"
)

const (
	// CreateUploadedMediaPermission is an account member permission.
	CreateUploadedMediaPermission Permission = "create.uploaded_media"
	// ReadUploadedMediaPermission is an account member permission.
	ReadUploadedMediaPermission Permission = "read.uploaded_media"
	// UpdateUploadedMediaPermission is an account member permission.
	UpdateUploadedMediaPermission Permission = "update.uploaded_media"
	// ArchiveUploadedMediaPermission is an account member permission.
	ArchiveUploadedMediaPermission Permission = "archive.uploaded_media"
)

var (
	// UploadedMediaPermissions contains all uploaded media-related permissions.
	UploadedMediaPermissions = []gorbac.Permission{
		CreateUploadedMediaPermission,
		ReadUploadedMediaPermission,
		UpdateUploadedMediaPermission,
		ArchiveUploadedMediaPermission,
	}
)
