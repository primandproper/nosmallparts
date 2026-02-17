package grpc

import (
	"github.com/google/wire"
)

var StagesSvcProviders = wire.NewSet(
	NewService,
	ProvideStagesMethodPermissions,
)
