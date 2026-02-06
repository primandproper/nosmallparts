package stages

import "github.com/google/wire"

var (
	StagesRepoProviders = wire.NewSet(
		ProvideStagesRepository,
	)
)
