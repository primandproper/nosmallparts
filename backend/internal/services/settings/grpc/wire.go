package grpc

import "github.com/google/wire"

var (
	SettingSvcProviders = wire.NewSet(
		NewService,
		ProvideMethodPermissions,
	)
)
