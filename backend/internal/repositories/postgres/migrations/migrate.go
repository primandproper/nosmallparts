package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"

	"github.com/GuiaBolso/darwin"
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

// Migrator implements database.Migrator for postgres databases.
type Migrator struct {
	logger logging.Logger
}

var _ database.Migrator = (*Migrator)(nil)

// NewMigrator creates a new postgres Migrator.
func NewMigrator(logger logging.Logger) *Migrator {
	return &Migrator{
		logger: logging.EnsureLogger(logger),
	}
}

// Migrate runs all postgres migrations on the given database connection.
func (m *Migrator) Migrate(ctx context.Context, db *sql.DB) error {
	migrations := []darwin.Migration{
		{Version: 1, Description: "identity tables", Script: fetchMigration("00001_identity")},
		{Version: 2, Description: "audit log entries tables", Script: fetchMigration("00002_auditlogentries")},
		{Version: 3, Description: "auth tables", Script: fetchMigration("00003_auth")},
		{Version: 4, Description: "oauth tables", Script: fetchMigration("00004_oauth")},
		{Version: 5, Description: "settings tables", Script: fetchMigration("00005_settings")},
		{Version: 6, Description: "user notifications table", Script: fetchMigration("00006_notifications")},
		{Version: 7, Description: "webhooks tables", Script: fetchMigration("00007_webhooks")},
		{Version: 8, Description: "waitlist tables", Script: fetchMigration("00008_waitlists")},
		{Version: 9, Description: "issue reports table", Script: fetchMigration("00009_issue_reports")},
		{Version: 10, Description: "uploaded media table", Script: fetchMigration("00010_uploaded_media")},
		{Version: 11, Description: "data privacy tables", Script: fetchMigration("00011_dataprivacy")},
		{Version: 12, Description: "play scripts tables", Script: fetchMigration("00012_play_scripts")},
		{Version: 13, Description: "stages tables", Script: fetchMigration("00013_stages")},
	}

	if err := darwin.New(darwin.NewGenericDriver(db, darwin.PostgresDialect{}), migrations, nil).Migrate(); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	m.logger.Info("migrations completed successfully")
	return nil
}
