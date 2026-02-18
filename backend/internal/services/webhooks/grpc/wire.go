package grpc

import "github.com/google/wire"

var (
	WebhookSvcProviders = wire.NewSet(
		NewService,
		ProvideMethodPermissions,
	)
)
