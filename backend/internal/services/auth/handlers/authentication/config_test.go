package authentication

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := &Config{
			Debug:                 false,
			EnableUserSignup:      false,
			MinimumUsernameLength: 123,
			MinimumPasswordLength: 123,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})
}

func TestOAuth2Config_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := OAuth2Config{
			Domain:               "example.com",
			AccessTokenLifespan:  time.Hour,
			RefreshTokenLifespan: 24 * time.Hour,
			Debug:                false,
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with missing domain", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := OAuth2Config{
			Domain:               "",
			AccessTokenLifespan:  time.Hour,
			RefreshTokenLifespan: 24 * time.Hour,
			Debug:                false,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with missing access token lifespan", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := OAuth2Config{
			Domain:               "example.com",
			AccessTokenLifespan:  0,
			RefreshTokenLifespan: 24 * time.Hour,
			Debug:                false,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with missing refresh token lifespan", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := OAuth2Config{
			Domain:               "example.com",
			AccessTokenLifespan:  time.Hour,
			RefreshTokenLifespan: 0,
			Debug:                false,
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}
