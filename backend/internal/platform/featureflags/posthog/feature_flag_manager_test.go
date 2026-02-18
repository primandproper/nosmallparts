package posthog

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	mockauthn "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/mocks"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/pointer"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/reflection"

	"github.com/posthog/posthog-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewFeatureFlagManager(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		cfg := &Config{
			ProjectAPIKey:  t.Name(),
			PersonalAPIKey: t.Name(),
		}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker())
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})

	T.Run("with nil config", func(t *testing.T) {
		actual, err := NewFeatureFlagManager(nil, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with missing API key", func(t *testing.T) {
		cfg := &Config{}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with missing API key", func(t *testing.T) {
		cfg := &Config{
			ProjectAPIKey: t.Name(),
		}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker())
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid config", func(t *testing.T) {
		cfg := &Config{
			ProjectAPIKey:  t.Name(),
			PersonalAPIKey: t.Name(),
		}

		actual, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker(), func(config *posthog.Config) {
			config.Interval = -1
		})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestFeatureFlagManager_CanUseFeature(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		t.SkipNow()

		ctx := t.Context()
		exampleUsername := "username"

		flagName := t.Name()
		cfg := &Config{ProjectAPIKey: t.Name(), PersonalAPIKey: t.Name()}

		ts := httptest.NewTLSServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			response, err := json.Marshal(&posthog.FeatureFlagsResponse{
				Flags: []posthog.FeatureFlag{
					{
						Key:          flagName,
						IsSimpleFlag: true,
						Active:       true,
						Filters: posthog.Filter{
							Groups: []posthog.FeatureFlagCondition{
								{
									Properties:        nil,
									RolloutPercentage: nil,
									Variant:           nil,
								},
							},
						},
					},
				},
				GroupTypeMapping: pointer.To(map[string]string{}),
			})
			require.NoError(t, err)

			_, err = res.Write(response)
			require.NoError(t, err)
		}))

		ffm, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker(), func(config *posthog.Config) {
			config.Transport = ts.Client().Transport
			config.Endpoint = ts.URL
		})
		require.NoError(t, err)

		actual, err := ffm.CanUseFeature(ctx, exampleUsername, flagName)
		assert.NoError(t, err)
		assert.True(t, actual)
	})

	T.Run("with error executing request", func(t *testing.T) {
		ctx := t.Context()
		exampleUsername := "username"

		cfg := &Config{ProjectAPIKey: t.Name(), PersonalAPIKey: t.Name()}

		ts := httptest.NewTLSServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusForbidden)
		}))

		ffm, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker(), func(config *posthog.Config) {
			config.Transport = ts.Client().Transport
			config.Endpoint = ts.URL
		})
		require.NoError(t, err)

		actual, err := ffm.CanUseFeature(ctx, exampleUsername, t.Name())
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestFeatureFlagManager_Identify(T *testing.T) {
	T.Run("standard", func(t *testing.T) {
		ctx := t.Context()

		user := mockauthn.NewMockUser()
		user.On(reflection.GetMethodName(user.GetID)).Return("ID").Twice()
		user.On(reflection.GetMethodName(user.GetUsername)).Return("Username")
		user.On(reflection.GetMethodName(user.GetFirstName)).Return("FirstName")
		user.On(reflection.GetMethodName(user.GetLastName)).Return("LastName")

		cfg := &Config{ProjectAPIKey: t.Name(), PersonalAPIKey: t.Name()}

		ts := httptest.NewTLSServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusOK)
		}))

		ffm, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker(), func(config *posthog.Config) {
			config.Transport = ts.Client().Transport
			config.Endpoint = ts.URL
		})
		require.NoError(t, err)

		assert.NoError(t, ffm.Identify(ctx, user))

		mock.AssertExpectationsForObjects(t, user)
	})

	T.Run("with nil user", func(t *testing.T) {
		ctx := t.Context()

		cfg := &Config{ProjectAPIKey: t.Name(), PersonalAPIKey: t.Name()}

		ffm, err := NewFeatureFlagManager(cfg, logging.NewNoopLogger(), tracing.NewNoopTracerProvider(), circuitbreaking.NewNoopCircuitBreaker())
		require.NoError(t, err)

		assert.Error(t, ffm.Identify(ctx, nil))
	})
}
