package maintenance

import "github.com/google/wire"

var (
	Providers = wire.NewSet(
		ProvideMaintenanceRepository,
	)
)
