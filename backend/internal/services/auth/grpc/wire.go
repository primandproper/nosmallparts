package grpc

import "github.com/google/wire"

var (
	AuthSvcProviders = wire.NewSet(
		NewAuthService,
		ProvideMethodPermissions,
	)
)
