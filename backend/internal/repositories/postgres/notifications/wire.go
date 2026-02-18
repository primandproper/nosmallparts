package notifications

import "github.com/google/wire"

var (
	NotifRepoProviders = wire.NewSet(
		ProvideNotificationsRepository,
	)
)
