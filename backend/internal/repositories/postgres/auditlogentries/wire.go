package auditlogentries

import "github.com/google/wire"

var (
	AuditRepoProviders = wire.NewSet(
		ProvideAuditLogRepository,
	)
)
