package tokenscfg

import (
	"encoding/base64"
	"testing"

	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
	"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/platform/v4/random"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider:                ProviderJWT,
			Audience:                t.Name(),
			Base64EncodedSigningKey: base64.URLEncoding.EncodeToString(random.MustGenerateRawBytes(ctx, 32)),
		}

		require.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with missing key", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: ProviderJWT,
		}

		require.Error(t, cfg.ValidateWithContext(ctx))
	})
}

func TestConfig_ProvideTokenIssuer(T *testing.T) {
	T.Parallel()

	T.Run("with SendGrid", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		cfg := &Config{
			Provider:                ProviderJWT,
			Audience:                t.Name(),
			Base64EncodedSigningKey: base64.URLEncoding.EncodeToString(random.MustGenerateRawBytes(ctx, 32)),
		}

		actual, err := cfg.ProvideTokenIssuer(logger, tracingnoop.NewTracerProvider())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		cfg := &Config{
			Provider: "",
		}

		actual, err := cfg.ProvideTokenIssuer(logger, tracingnoop.NewTracerProvider())
		assert.Nil(t, actual)
		assert.Error(t, err)
	})
}
