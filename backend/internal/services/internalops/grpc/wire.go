package grpc

import "github.com/google/wire"

var (
	InternalOpsSvcProviders = wire.NewSet(
		NewService,
	)
)
