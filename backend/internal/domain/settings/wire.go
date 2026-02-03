package settings

import (
	"github.com/google/wire"
)

var (
	Providers = wire.NewSet(
		ProvideServiceSettingDataManagerFromRepository,
		ProvideServiceSettingConfigurationDataManagerFromRepository,
	)
)

func ProvideServiceSettingDataManagerFromRepository(r Repository) ServiceSettingDataManager {
	return r
}

func ProvideServiceSettingConfigurationDataManagerFromRepository(r Repository) ServiceSettingConfigurationDataManager {
	return r
}
