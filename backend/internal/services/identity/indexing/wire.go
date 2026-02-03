package indexing

import (
	"github.com/google/wire"
)

var (
	Providers = wire.NewSet(
		NewCoreDataIndexer,
	)
)
