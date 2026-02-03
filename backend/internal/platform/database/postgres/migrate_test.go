package postgres

import (
	"strings"
	"testing"
	"time"

	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/fake"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GuiaBolso/darwin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var queryReplacer = strings.NewReplacer(
	`(`, `\(`,
	`)`, `\)`,
	`$`, `\$`,
)

func TestQuerier_Migrate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		exampleCreationTime := fake.BuildFakeTime()

		ctx := t.Context()
		c, db := buildTestClient(t)

		c.timeFunc = func() time.Time {
			return exampleCreationTime
		}
		c.config = &databasecfg.Config{
			MaxPingAttempts: 1,
			PingWaitPeriod:  time.Second,
			Migrations: []*databasecfg.MigrationSpec{
				{
					Description: "migration 1",
					RawQuery:    "CREATE TABLE users(id serial PRIMARY KEY, name TEXT)",
				},
			},
		}

		pgDialect := &darwin.PostgresDialect{}

		// called by c.IsReady()
		db.ExpectPing()

		db.ExpectBegin()
		db.ExpectExec(queryReplacer.Replace(pgDialect.CreateTableSQL())).WillReturnResult(sqlmock.NewResult(1, 1))
		db.ExpectCommit()

		db.ExpectQuery(queryReplacer.Replace(pgDialect.AllSQL())).WillReturnRows(sqlmock.NewRows([]string{
			"version",
			"description",
			"checksum",
			"applied_at",
			"execution_time",
		}))

		db.ExpectQuery(queryReplacer.Replace(pgDialect.AllSQL())).WillReturnRows(sqlmock.NewRows([]string{
			"version",
			"description",
			"checksum",
			"applied_at",
			"execution_time",
		}))

		for _, migration := range c.config.Migrations {
			db.ExpectBegin()
			db.ExpectExec(queryReplacer.Replace(migration.RawQuery)).WillReturnResult(sqlmock.NewResult(1, 1))
			db.ExpectCommit()

			db.ExpectBegin()
			db.ExpectExec(queryReplacer.Replace(pgDialect.InsertSQL())).WithArgs(
				1.0,
				migration.Description,
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).WillReturnResult(sqlmock.NewResult(1, 1))
			db.ExpectCommit()
		}

		err := c.Migrate(ctx)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, db)
	})
}
