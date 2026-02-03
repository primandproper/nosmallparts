package dataprivacy

import "github.com/google/wire"

var (
	DataPrivProviders = wire.NewSet(
		ProvideDataPrivacyRepository,
	)
)
