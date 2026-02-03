package grpc

import "github.com/google/wire"

var (
	AuditSvcProviders = wire.NewSet(
		NewService,
	)
)
