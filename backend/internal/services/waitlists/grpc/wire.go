package grpc

import "github.com/google/wire"

var (
	WaitlistsSvcProviders = wire.NewSet(
		NewService,
		ProvideMethodPermissions,
	)
)
