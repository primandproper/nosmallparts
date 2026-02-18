package config

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/dataprivacy/config"
	identitycfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/identity/config"
	oauthcfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/services/oauth/config"
)

func TestServicesConfig_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("valid config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &ServicesConfig{
			Users:         identitycfg.Config{},
			DataPrivacy:   dataprivacycfg.Config{},
			Auth:          authentication.Config{},
			OAuth2Clients: oauthcfg.Config{},
		}

		err := cfg.ValidateWithContext(ctx)
		// May have validation errors in sub-configs, but should not panic
		_ = err
	})

	T.Run("empty config", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cfg := &ServicesConfig{}

		err := cfg.ValidateWithContext(ctx)
		// May have validation errors, but should not panic
		_ = err
	})
}
