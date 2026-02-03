package authcfg

import (
	"context"
	"time"

	tokenscfg "github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/authentication/tokens/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	GoogleSSOConfig struct {
		_ struct{} `json:"-"`

		ClientID     string `env:"CLIENT_ID"     json:"clientID,omitempty"`
		ClientSecret string `env:"CLIENT_SECRET" json:"clientSecret,omitempty"`
		CallbackURL  string `env:"CALLBACK_URL"  json:"callbackURL,omitempty"`
	}

	SSOConfigs struct {
		Google GoogleSSOConfig `envPrefix:"GOOGLE_SSO_" json:"google,omitempty"`
	}

	TokenRefreshConfig struct {
		MaxAccessTokenLifetime  time.Duration `env:"MAX_ACCESS_TOKEN_LIFETIME"  json:"maxAccessTokenLifetime"`
		MaxRefreshTokenLifetime time.Duration `env:"MAX_REFRESH_TOKEN_LIFETIME" json:"maxRefreshTokenLifetime"`
	}

	// Config is our configuration.
	Config struct {
		_ struct{} `json:"-"`

		SSO                   SSOConfigs       `envPrefix:"SSO_CONFIG_"       json:"sso,omitempty"`
		Tokens                tokenscfg.Config `envPrefix:"TOKENS_"           json:"tokens"`
		Debug                 bool             `env:"DEBUG"                   json:"debug,omitempty"`
		EnableUserSignup      bool             `env:"ENABLE_USER_SIGNUP"      json:"enableUserSignup,omitempty"`
		MinimumUsernameLength uint8            `env:"MINIMUM_USERNAME_LENGTH" json:"minimumUsernameLength,omitempty"`
		MinimumPasswordLength uint8            `env:"MINIMUM_PASSWORD_LENGTH" json:"minimumPasswordLength,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.MinimumUsernameLength, validation.Required),
		validation.Field(&cfg.MinimumPasswordLength, validation.Required),
		validation.Field(&cfg.Tokens, validation.Required),
	)
}
