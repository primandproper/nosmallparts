package authorization

import (
	"github.com/mikespook/gorbac/v2"
)

const (
	// ReadAuditLogEntriesPermission is a service permission.
	ReadAuditLogEntriesPermission Permission = "read.audit_log_entries"
)

var (
	// AuditPermissions contains all audit-related permissions.
	AuditPermissions = []gorbac.Permission{
		ReadAuditLogEntriesPermission,
	}
)
