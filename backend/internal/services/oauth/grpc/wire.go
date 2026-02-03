package grpc

import "github.com/google/wire"

var (
	OAuthSvcProviders = wire.NewSet(
		NewService,
		ProvideMethodPermissions,
	)
)
