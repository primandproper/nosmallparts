package authorization

import (
	"github.com/mikespook/gorbac/v2"
)

const (
	// CreateIssueReportsPermission is an account admin permission.
	CreateIssueReportsPermission Permission = "create.issue_reports"
	// ReadIssueReportsPermission is an account admin permission.
	ReadIssueReportsPermission Permission = "read.issue_reports"
	// UpdateIssueReportsPermission is an account admin permission.
	UpdateIssueReportsPermission Permission = "update.issue_reports"
	// ArchiveIssueReportsPermission is an account admin permission.
	ArchiveIssueReportsPermission Permission = "archive.issue_reports"
)

var (
	// IssueReportsPermissions contains all issue report-related permissions.
	IssueReportsPermissions = []gorbac.Permission{
		CreateIssueReportsPermission,
		ReadIssueReportsPermission,
		UpdateIssueReportsPermission,
		ArchiveIssueReportsPermission,
	}
)
