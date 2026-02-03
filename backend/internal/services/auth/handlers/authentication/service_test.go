package authentication

import (
	"encoding/base64"
	"net/http"
	"testing"

	mockauthn "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/mock"
	tokenscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/tokens/config"
	identitymock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/identity/mock"
	oauthmock "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/domain/oauth/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/analytics"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/encoding"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/featureflags"
	msgconfig "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/config"
	mockpublishers "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/messagequeue/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/reflection"
	mockrouting "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/routing/mock"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildTestService(t *testing.T) *service {
	t.Helper()

	ctx := t.Context()
	logger := logging.NewNoopLogger()
	encoderDecoder := encoding.ProvideServerEncoderDecoder(logger, tracing.NewNoopTracerProvider(), encoding.ContentTypeJSON)

	cfg := &Config{
		Tokens: tokenscfg.Config{
			Provider:                tokenscfg.ProviderJWT,
			Audience:                "",
			Base64EncodedSigningKey: base64.URLEncoding.EncodeToString([]byte(testutils.Example32ByteKey)),
		},
	}
	queueCfg := &msgconfig.QueuesConfig{DataChangesTopicName: "data_changes"}

	pp := &mockpublishers.PublisherProvider{}
	pp.On(reflection.GetMethodName(pp.ProvidePublisher), queueCfg.DataChangesTopicName).Return(&mockpublishers.Publisher{}, nil)

	rpm := mockrouting.NewRouteParamManager()
	rpm.On(
		"BuildRouteParamStringIDFetcher",
		AuthProviderParamKey,
	).Return(func(*http.Request) string { return "" })

	s, err := ProvideService(
		ctx,
		logger,
		cfg,
		&mockauthn.Authenticator{},
		&oauthmock.RepositoryMock{},
		&identitymock.RepositoryMock{},
		encoderDecoder,
		tracing.NewNoopTracerProvider(),
		pp,
		&featureflags.NoopFeatureFlagManager{},
		analytics.NewNoopEventReporter(),
		rpm,
		queueCfg,
	)
	require.NoError(t, err)

	mock.AssertExpectationsForObjects(t, pp, rpm)

	return s.(*service)
}

func TestProvideService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		encoderDecoder := encoding.ProvideServerEncoderDecoder(logger, tracing.NewNoopTracerProvider(), encoding.ContentTypeJSON)

		cfg := &Config{
			Tokens: tokenscfg.Config{
				Provider:                tokenscfg.ProviderJWT,
				Audience:                "",
				Base64EncodedSigningKey: base64.URLEncoding.EncodeToString([]byte(testutils.Example32ByteKey)),
			},
		}
		queueCfg := &msgconfig.QueuesConfig{DataChangesTopicName: "data_changes"}

		pp := &mockpublishers.PublisherProvider{}
		pp.On(reflection.GetMethodName(pp.ProvidePublisher), queueCfg.DataChangesTopicName).Return(&mockpublishers.Publisher{}, nil)

		rpm := mockrouting.NewRouteParamManager()
		rpm.On(
			"BuildRouteParamStringIDFetcher",
			AuthProviderParamKey,
		).Return(func(*http.Request) string { return "" })

		s, err := ProvideService(
			ctx,
			logger,
			cfg,
			&mockauthn.Authenticator{},
			&oauthmock.RepositoryMock{},
			&identitymock.RepositoryMock{},
			encoderDecoder,
			tracing.NewNoopTracerProvider(),
			pp,
			&featureflags.NoopFeatureFlagManager{},
			analytics.NewNoopEventReporter(),
			rpm,
			queueCfg,
		)

		assert.NotNil(t, s)
		assert.NoError(t, err)
	})
}
