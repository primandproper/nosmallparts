package waitlists

import "github.com/google/wire"

var (
	WaitlistsRepoProviders = wire.NewSet(
		ProvideWaitlistsRepository,
	)
)
