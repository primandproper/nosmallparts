package repositories

import (
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	postgresmigrations "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/migrations"
)

// ProvideMigrator returns a Migrator appropriate for the configured database provider.
func ProvideMigrator(
	cfg *databasecfg.Config,
	logger logging.Logger,
) database.Migrator {
	switch cfg.Provider {
	case databasecfg.ProviderPostgres:
		return postgresmigrations.NewMigrator(logger)
	default:
		return nil
	}
}
