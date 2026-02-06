package livekit

import (
	"github.com/google/wire"
)

var (
	// Providers represents what this package offers to the dependency graph.
	Providers = wire.NewSet(
		NewHandler,
	)
)
