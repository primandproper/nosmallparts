package grpc

import "github.com/google/wire"

var (
	NotifsSvcProviders = wire.NewSet(
		NewService,
		ProvideMethodPermissions,
	)
)
