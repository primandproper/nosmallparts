package dbcleaner

import (
	"github.com/google/wire"
)

var (
	// ProvidersDBCleaner are what we provide to dependency injection.
	ProvidersDBCleaner = wire.NewSet(
		NewDBCleaner,
	)
)
