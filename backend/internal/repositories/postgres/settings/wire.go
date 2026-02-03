package settings

import "github.com/google/wire"

var (
	SettingsRepoProviders = wire.NewSet(
		ProvideSettingsRepository,
	)
)
