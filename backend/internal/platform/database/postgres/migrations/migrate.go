package migrations

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/GuiaBolso/darwin"
)

type Migrator struct {
	logger      logging.Logger
	tracer      tracing.Tracer
	config      *databasecfg.Config
	db          *sql.DB
	migrations  []*databasecfg.MigrationSpec
	migrateOnce sync.Once
}

func NewMigrator(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	db *sql.DB,
	config *databasecfg.Config,
	migrations []*databasecfg.MigrationSpec,
) *Migrator {
	return &Migrator{
		config:     config,
		logger:     logging.EnsureLogger(logger),
		tracer:     tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer("postgres_migrator")),
		db:         db,
		migrations: migrations,
	}
}

func migrationConfigToDarwinMigration(migration *databasecfg.MigrationSpec) (*darwin.Migration, error) {
	query := migration.RawQuery
	if query == "" && migration.Filepath != "" {
		queryFile, err := os.ReadFile(migration.Filepath)
		if err != nil {
			return nil, fmt.Errorf("error reading migration file %q: %w", migration.Filepath, err)
		}
		query = string(queryFile)
	}

	return &darwin.Migration{
		Description: migration.Description,
		Script:      query,
	}, nil
}

// BuildMigrationFunc returns a sync.Once compatible function closure that will migrate a postgres database.
func (m *Migrator) migrationFunc() {
	infoChan := make(chan darwin.MigrationInfo)
	go func() {
		for x := range infoChan {
			m.logger.WithValues(map[string]any{
				"version": x.Migration.Version,
				"status":  x.Status.String(),
			}).Info("database migration handled")
		}
	}()

	startTime := time.Now()
	convertedMigrations := []darwin.Migration{}
	for i, migration := range m.migrations {
		converted, err := migrationConfigToDarwinMigration(migration)
		if err != nil {
			panic(fmt.Errorf("error converting migration %d: %w", i, err))
		}
		converted.Version = float64(i + 1)
		convertedMigrations = append(convertedMigrations, *converted)
	}

	if err := darwin.New(darwin.NewGenericDriver(m.db, darwin.PostgresDialect{}), convertedMigrations, infoChan).Migrate(); err != nil {
		panic(fmt.Errorf("running migration: %w", err))
	}
	m.logger.WithValue("elapsed", time.Since(startTime).Milliseconds()).Info("migration completed")
}

// Migrate is a simple wrapper around the core querier Migrate call.
func (m *Migrator) Migrate(ctx context.Context) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if m.config == nil {
		return database.ErrNilInputProvided
	}

	if !m.IsReady(ctx) {
		return database.ErrDatabaseNotReady
	}

	m.migrateOnce.Do(m.migrationFunc)

	return nil
}

// IsReady returns whether the database is ready for the querier.
func (m *Migrator) IsReady(ctx context.Context) bool {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithValue("connection_url", m.config.ConnectionDetails.String())

	attemptCount := 0
	for {
		if err := m.db.PingContext(ctx); err != nil {
			logger.WithValue("attempt_count", attemptCount).Info("ping failed, waiting for db")
			time.Sleep(m.config.PingWaitPeriod)

			attemptCount++
			if attemptCount >= int(m.config.MaxPingAttempts) {
				break
			}
		} else {
			return true
		}
	}

	return false
}
