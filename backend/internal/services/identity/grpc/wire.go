package grpc

import "github.com/google/wire"

var (
	IDSvcProviders = wire.NewSet(
		NewService,
		ProvideMethodPermissions,
	)
)
