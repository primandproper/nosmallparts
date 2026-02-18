package grpc

import "github.com/google/wire"

var (
	DataPrivSvcProviders = wire.NewSet(
		NewDataPrivacyService,
	)
)
