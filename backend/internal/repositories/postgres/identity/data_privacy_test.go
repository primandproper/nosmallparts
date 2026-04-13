package identity

import (
	"fmt"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQuerier_Integration_DataPrivacy(t *testing.T) {
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

	exampleUser := fakes.BuildFakeUser()
	exampleUser.Username = fmt.Sprintf("%d", pgtesting.HashStringToNumberForTest(t, exampleUser.Username))
	exampleUser.TwoFactorSecretVerifiedAt = nil

	// create
	createdUser := createUserForTest(t, ctx, exampleUser, dbc)

	assert.NoError(t, dbc.DeleteUser(ctx, createdUser.ID))
}

func TestQuerier_DeleteUser(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.DeleteUser(ctx, ""))
	})
}
