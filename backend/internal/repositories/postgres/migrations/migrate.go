package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres"
	pgmigrations "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres/migrations"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
)

var (
	//go:embed migration_files/*.sql
	rawMigrations embed.FS
)

func fetchMigration(name string) string {
	file, err := rawMigrations.ReadFile(fmt.Sprintf("migration_files/%s.sql", name))
	if err != nil {
		panic(err)
	}

	return string(file)
}

// ProvideMigratorFactory provides a MigratorFactory for use with the database client.
func ProvideMigratorFactory(logger logging.Logger, tracerProvider tracing.TracerProvider, config *databasecfg.Config) postgres.MigratorFactory {
	return func(db *sql.DB) postgres.Migrator {
		return NewMigrator(logger, tracerProvider, db, config)
	}
}

func NewMigrator(logger logging.Logger, tracerProvider tracing.TracerProvider, db *sql.DB, config *databasecfg.Config) *pgmigrations.Migrator {
	migrations := []*databasecfg.MigrationSpec{
		{
			Description: "identity tables",
			RawQuery:    fetchMigration("00001_identity"),
		},
		{
			Description: "audit log entries tables",
			RawQuery:    fetchMigration("00002_auditlogentries"),
		},
		{
			Description: "auth tables",
			RawQuery:    fetchMigration("00003_auth"),
		},
		{
			Description: "oauth tables",
			RawQuery:    fetchMigration("00004_oauth"),
		},
		{
			Description: "settings tables",
			RawQuery:    fetchMigration("00005_settings"),
		},
		{
			Description: "user notifications table",
			RawQuery:    fetchMigration("00006_notifications"),
		},
		{
			Description: "webhooks tables",
			RawQuery:    fetchMigration("00007_webhooks"),
		},
		{
			Description: "waitlist tables",
			RawQuery:    fetchMigration("00008_waitlists"),
		},
		{
			Description: "issue reports table",
			RawQuery:    fetchMigration("00009_issue_reports"),
		},
		{
			Description: "uploaded media table",
			RawQuery:    fetchMigration("00010_uploaded_media"),
		},
		{
			Description: "play scripts tables",
			RawQuery:    fetchMigration("00011_play_scripts"),
		},
		{
			Description: "stages tables",
			RawQuery:    fetchMigration("00012_stages"),
		},
	}

	return pgmigrations.NewMigrator(logger, tracerProvider, db, config, migrations)
}
