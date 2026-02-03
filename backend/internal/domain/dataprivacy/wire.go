package dataprivacy

import (
	"github.com/google/wire"
)

var (
	Providers = wire.NewSet(
		ProvideDataPrivacyDataManagerFromRepository,
	)
)

func ProvideDataPrivacyDataManagerFromRepository(r Repository) DataPrivacyDataManager {
	return r
}
