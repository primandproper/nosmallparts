package grpc

import "github.com/google/wire"

var (
	UploadedMediaSvcProviders = wire.NewSet(
		NewService,
		ProvideMethodPermissions,
	)
)
