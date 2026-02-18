package grpc

import "github.com/google/wire"

var (
	IssueReportSvcProviders = wire.NewSet(
		NewService,
		ProvideMethodPermissions,
	)
)
