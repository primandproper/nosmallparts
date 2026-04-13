package maintenance

import (
	"testing"

	pgtesting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuerier_Integration_DeleteExpiredOAuth2ClientTokens(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, container := buildDatabaseClientForTest(t)

	databaseURI, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, databaseURI)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	// fetch as list
	count, err := dbc.DeleteExpiredOAuth2ClientTokens(ctx)
	assert.Zero(t, count)
	assert.NoError(t, err)
}
