package grpc

import (
	"github.com/google/wire"
)

var PlayScriptsSvcProviders = wire.NewSet(
	NewService,
	ProvidePlayScriptsMethodPermissions,
)
