package authentication

import (
	"errors"
	"testing"

	types "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth/fakes"
	oauthmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewOAuth2ClientStore(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		domain := "example.com"
		logger := logging.NewNoopLogger()
		tracer := tracing.NewTracer(tracing.NewNoopTracerProvider().Tracer("test"))
		dataManager := &oauthmock.RepositoryMock{}

		store := newOAuth2ClientStore(domain, logger, tracer, dataManager)

		assert.NotNil(t, store)
		impl, ok := store.(*oauth2ClientStoreImpl)
		assert.True(t, ok)
		assert.Equal(t, domain, impl.domain)
		assert.NotNil(t, impl.tracer)
		assert.NotNil(t, impl.logger)
		assert.Equal(t, dataManager, impl.dataManager)
	})
}

func TestOAuth2ClientStoreImpl_GetByID(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		domain := "example.com"
		logger := logging.NewNoopLogger()
		tracer := tracing.NewTracer(tracing.NewNoopTracerProvider().Tracer("test"))

		client := fakes.BuildFakeOAuth2Client()
		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"GetOAuth2ClientByClientID",
			testutils.ContextMatcher,
			client.ID,
		).Return(client, nil)

		store := &oauth2ClientStoreImpl{
			domain:      domain,
			logger:      logger,
			tracer:      tracer,
			dataManager: dataManager,
		}

		result, err := store.GetByID(ctx, client.ID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, client.ID, result.GetID())
		assert.Equal(t, client.ClientSecret, result.GetSecret())
		assert.Equal(t, domain, result.GetDomain())
		assert.False(t, result.IsPublic())

		mock.AssertExpectationsForObjects(t, dataManager)
	})

	T.Run("with error getting client", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		domain := "example.com"
		logger := logging.NewNoopLogger()
		tracer := tracing.NewTracer(tracing.NewNoopTracerProvider().Tracer("test"))

		clientID := "test-client-id"
		dataManager := &oauthmock.RepositoryMock{}
		dataManager.On(
			"GetOAuth2ClientByClientID",
			testutils.ContextMatcher,
			clientID,
		).Return((*types.OAuth2Client)(nil), errors.New("database error"))

		store := &oauth2ClientStoreImpl{
			domain:      domain,
			logger:      logger,
			tracer:      tracer,
			dataManager: dataManager,
		}

		result, err := store.GetByID(ctx, clientID)

		assert.Error(t, err)
		assert.Nil(t, result)

		mock.AssertExpectationsForObjects(t, dataManager)
	})
}
