package issue_reports

import "github.com/google/wire"

var (
	IssueReportsRepoProviders = wire.NewSet(
		ProvideIssueReportsRepository,
	)
)
