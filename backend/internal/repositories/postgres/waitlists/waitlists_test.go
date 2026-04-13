package waitlists

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/waitlists/converters"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/waitlists/fakes"
	platformerrors "github.com/verygoodsoftwarenotvirus/platform/v4/errors"
	"github.com/verygoodsoftwarenotvirus/platform/v4/database/filtering"
	pgtesting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/repositories/postgres/testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleQuantity = 3
)

func createWaitlistForTest(t *testing.T, ctx context.Context, exampleWaitlist *types.Waitlist, dbc *repository) *types.Waitlist {
	t.Helper()

	if exampleWaitlist == nil {
		exampleWaitlist = fakes.BuildFakeWaitlist()
	}
	dbInput := converters.ConvertWaitlistToWaitlistDatabaseCreationInput(exampleWaitlist)

	created, err := dbc.CreateWaitlist(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleWaitlist.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleWaitlist.ID, created.ID)
	assert.Equal(t, exampleWaitlist.Name, created.Name)
	assert.Equal(t, exampleWaitlist.Description, created.Description)

	waitlist, err := dbc.GetWaitlist(ctx, created.ID)
	assert.NoError(t, err)
	require.NotNil(t, waitlist)
	exampleWaitlist.CreatedAt = waitlist.CreatedAt
	assert.Equal(t, waitlist.ID, created.ID)

	return created
}

func TestQuerier_Integration_Waitlists(t *testing.T) {
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

	exampleWaitlist := fakes.BuildFakeWaitlist()
	createdWaitlists := []*types.Waitlist{}

	createdWaitlists = append(createdWaitlists, createWaitlistForTest(t, ctx, exampleWaitlist, dbc))

	for i := 0; i < exampleQuantity; i++ {
		input := fakes.BuildFakeWaitlist()
		input.Name = fmt.Sprintf("%s %d", exampleWaitlist.Name, i)
		createdWaitlists = append(createdWaitlists, createWaitlistForTest(t, ctx, input, dbc))
	}

	// fetch as list
	waitlists, err := dbc.GetWaitlists(ctx, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, waitlists.Data)
	assert.Equal(t, len(createdWaitlists), len(waitlists.Data))

	// fetch active waitlists
	activeWaitlists, err := dbc.GetActiveWaitlists(ctx, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, activeWaitlists.Data)

	// check not expired
	notExpired, err := dbc.WaitlistIsNotExpired(ctx, createdWaitlists[0].ID)
	assert.NoError(t, err)
	assert.True(t, notExpired)

	// create signup (requires identity tables - skip or use mock; integration test may need identity setup)
	// For a minimal integration test we just archive waitlists
	for _, wl := range createdWaitlists {
		assert.NoError(t, dbc.ArchiveWaitlist(ctx, wl.ID))

		_, err = dbc.GetWaitlist(ctx, wl.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_WaitlistIsNotExpired(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.WaitlistIsNotExpired(ctx, "")
		assert.Error(t, err)
		assert.False(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestQuerier_GetWaitlist(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetWaitlist(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestQuerier_CreateWaitlist(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateWaitlist(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputProvided)
	})
}

func TestQuerier_UpdateWaitlist(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateWaitlist(ctx, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputProvided)
	})
}

func TestQuerier_ArchiveWaitlist(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveWaitlist(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestQuerier_GetWaitlistSignup(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist signup ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		exampleWaitlistID := fakes.BuildFakeID()

		actual, err := c.GetWaitlistSignup(ctx, "", exampleWaitlistID)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})

	T.Run("with invalid waitlist ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		exampleSignupID := fakes.BuildFakeID()

		actual, err := c.GetWaitlistSignup(ctx, exampleSignupID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestQuerier_GetWaitlistSignupsForWaitlist(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)
		filter := filtering.DefaultQueryFilter()

		actual, err := c.GetWaitlistSignupsForWaitlist(ctx, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestQuerier_CreateWaitlistSignup(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateWaitlistSignup(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputProvided)
	})
}

func TestQuerier_UpdateWaitlistSignup(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateWaitlistSignup(ctx, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputProvided)
	})
}

func TestQuerier_ArchiveWaitlistSignup(T *testing.T) {
	T.Parallel()

	T.Run("with invalid waitlist signup ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveWaitlistSignup(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}
