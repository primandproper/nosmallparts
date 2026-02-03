package migrations

import (
	"testing"
	"time"

	pgtesting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/database/postgres/testing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/pointer"

	"github.com/stretchr/testify/require"
)

func TestQuerier_Migrate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		container, db, config := pgtesting.BuildDatabaseContainerForTest(t)
		require.NoError(t, NewMigrator(logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), db, config).Migrate(ctx))

		if err := container.Stop(ctx, pointer.To(time.Second*10)); err != nil {
			t.Log(err)
		}
	})
}
