package migrations

import (
	"testing"
	"time"

	databasecfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/config"
	pgtesting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres/testing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuerier_Migrate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		c, db, _ := pgtesting.BuildDatabaseContainerForTest(t)

		config := &databasecfg.Config{MaxPingAttempts: 1, PingWaitPeriod: time.Second}

		migrator := NewMigrator(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), db, config, nil)
		migrator.migrateOnce.Do(func() {})

		ctx := t.Context()
		assert.NoError(t, migrator.Migrate(ctx))

		require.NoError(t, c.Terminate(ctx))
	})
}
