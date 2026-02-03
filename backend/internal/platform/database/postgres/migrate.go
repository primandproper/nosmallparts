package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database"
	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"

	"github.com/GuiaBolso/darwin"
)

// BuildMigrationFunc returns a sync.Once compatible function closure that will migrate a postgres database.
func (q *Client) migrationFunc(migrations []*databasecfg.MigrationSpec) {
	darwinMigrations := []darwin.Migration{}
	for i, m := range migrations {
		var queryToUse string
		if m.RawQuery != "" {
			queryToUse = m.RawQuery
		} else {
			contents, err := os.ReadFile(m.Filepath)
			if err != nil {
				panic(fmt.Errorf("reading migration file: %w", err))
			}
			queryToUse = string(contents)
		}

		darwinMigrations = append(darwinMigrations, darwin.Migration{
			Version:     float64(i + 1),
			Description: m.Description,
			Script:      queryToUse,
		})
	}

	startTime := time.Now()
	if err := darwin.New(darwin.NewGenericDriver(q.db, darwin.PostgresDialect{}), darwinMigrations, nil).Migrate(); err != nil {
		panic(fmt.Errorf("running migration: %w", err))
	}
	q.logger.WithValue("elapsed", time.Since(startTime).Milliseconds()).Info("migration completed")
}

// Migrate is a simple wrapper around the core querier Migrate call.
func (q *Client) Migrate(ctx context.Context) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if q.config == nil {
		return database.ErrNilInputProvided
	}

	if !q.IsReady(ctx) {
		return database.ErrDatabaseNotReady
	}

	q.migrateOnce.Do(func() {
		q.migrationFunc(q.config.Migrations)
	})

	return nil
}
