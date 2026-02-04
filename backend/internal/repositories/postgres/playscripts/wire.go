package playscripts

import "github.com/google/wire"

var (
	PlayScriptsRepoProviders = wire.NewSet(
		ProvidePlayScriptsRepository,
	)
)
