package msgconfig

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
)

func Test_cleanString(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotEmpty(t, cleanString(t.Name()))
	})
}

func TestProvideConsumerProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := &Config{
			Consumer: MessageQueueConfig{
				Provider: ProviderRedis,
			},
		}

		provider, err := ProvideConsumerProvider(ctx, logger, cfg)
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})
}

func TestProvidePublisherProvider(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := logging.NewNoopLogger()
		cfg := &Config{
			Publisher: MessageQueueConfig{
				Provider: ProviderRedis,
			},
		}

		provider, err := ProvidePublisherProvider(ctx, logger, tracing.NewNoopTracerProvider(), cfg)
		assert.NoError(t, err)
		assert.NotNil(t, provider)
	})
}
