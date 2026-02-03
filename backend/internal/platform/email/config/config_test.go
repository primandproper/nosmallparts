package emailcfg

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/circuitbreaking"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/email/mailgun"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/email/mailjet"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/email/sendgrid"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Sendgrid: &sendgrid.Config{APIToken: t.Name()},
		}

		require.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with invalid token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &Config{
			Provider: "sendgrid",
		}

		require.Error(t, cfg.ValidateWithContext(ctx))
	})
}

func TestConfig_ProvideEmailer(T *testing.T) {
	T.Parallel()

	providers := []string{
		ProviderSendgrid,
		ProviderMailgun,
		ProviderMailjet,
	}

	for _, provider := range providers {
		T.Run(fmt.Sprintf("with %s", provider), func(t *testing.T) {
			t.Parallel()

			logger := logging.NewNoopLogger()
			cfg := &Config{
				Provider: provider,
				Sendgrid: &sendgrid.Config{APIToken: t.Name()},
				Mailgun:  &mailgun.Config{PrivateAPIKey: t.Name(), Domain: t.Name()},
				Mailjet:  &mailjet.Config{APIKey: t.Name(), SecretKey: t.Name()},
			}

			actual, err := cfg.ProvideEmailer(logger, tracing.NewNoopTracerProvider(), &http.Client{}, circuitbreaking.NewNoopCircuitBreaker())
			assert.NotNil(t, actual)
			assert.NoError(t, err)
		})
	}

	T.Run("with invalid provider", func(t *testing.T) {
		t.Parallel()

		logger := logging.NewNoopLogger()
		cfg := &Config{
			Provider: "",
		}

		actual, err := cfg.ProvideEmailer(logger, tracing.NewNoopTracerProvider(), &http.Client{}, circuitbreaking.NewNoopCircuitBreaker())
		assert.NotNil(t, actual)
		assert.NoError(t, err)
	})
}
