package indexing

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/fakes"
	identitymock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/reflection"
	textsearch "github.com/verygoodsoftwarenotvirus/platform/v4/search/text"
	mocksearch "github.com/verygoodsoftwarenotvirus/platform/v4/search/text/mock"
	"github.com/verygoodsoftwarenotvirus/platform/v4/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleIndexRequest(T *testing.T) {
	T.Parallel()

	T.Run("user index type", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		identityRepo := &identitymock.RepositoryMock{}
		identityRepo.On(reflection.GetMethodName(identityRepo.GetUser), testutils.ContextMatcher, exampleUser.ID).Return(exampleUser, nil)
		identityRepo.On(reflection.GetMethodName(identityRepo.MarkUserAsIndexed), testutils.ContextMatcher, exampleUser.ID).Return(nil)

		uss := ConvertUserToUserSearchSubset(exampleUser)

		mim := &mocksearch.IndexManager[UserSearchSubset]{}
		mim.On(reflection.GetMethodName(mim.Index), testutils.ContextMatcher, exampleUser.ID, uss).Return(nil)

		cdi := NewCoreDataIndexer(
			logger,
			tracingnoop.NewTracerProvider(),
			identityRepo,
			mim,
		)

		indexReq := &textsearch.IndexRequest{
			RowID:     exampleUser.ID,
			IndexType: IndexTypeUsers,
			Delete:    false,
		}

		assert.NoError(t, cdi.HandleIndexRequest(ctx, indexReq))

		mock.AssertExpectationsForObjects(t, identityRepo, mim)
	})

	T.Run("deleting user index type", func(t *testing.T) {
		t.Parallel()

		exampleUser := fakes.BuildFakeUser()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		identityRepo := &identitymock.RepositoryMock{}
		identityRepo.On(reflection.GetMethodName(identityRepo.GetUser), testutils.ContextMatcher, exampleUser.ID).Return(exampleUser, nil)

		mim := &mocksearch.IndexManager[UserSearchSubset]{}
		mim.On(reflection.GetMethodName(mim.Delete), testutils.ContextMatcher, exampleUser.ID).Return(nil)

		cdi := NewCoreDataIndexer(
			logger,
			tracingnoop.NewTracerProvider(),
			identityRepo,
			mim,
		)

		indexReq := &textsearch.IndexRequest{
			RowID:     exampleUser.ID,
			IndexType: IndexTypeUsers,
			Delete:    true,
		}

		assert.NoError(t, cdi.HandleIndexRequest(ctx, indexReq))

		mock.AssertExpectationsForObjects(t, identityRepo, mim)
	})
}
