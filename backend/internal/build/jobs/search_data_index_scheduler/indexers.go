package searchdataindexscheduler

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/text/indexing"
	coreindexing "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/indexing"
)

func ProvideIndexFunctions(identityRepo identity.Repository) map[string]indexing.Function {
	outputMap := map[string]indexing.Function{}
	coreMap := coreindexing.BuildCoreDataIndexingFunctions(identityRepo)

	for k, v := range coreMap {
		outputMap[k] = v
	}

	return outputMap
}
