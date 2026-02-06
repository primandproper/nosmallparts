package livekit

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config holds configuration for LiveKit integration.
type Config struct {
	_ struct{} `json:"-"`

	APIKey    string `env:"API_KEY"    json:"apiKey,omitempty"`
	APISecret string `env:"API_SECRET" json:"apiSecret,omitempty"`
	URL       string `env:"URL"        json:"url,omitempty"`
}

// Enabled returns true if LiveKit is configured.
func (c *Config) Enabled() bool {
	return c.APIKey != "" && c.APISecret != "" && c.URL != ""
}

// ValidateWithContext validates the config.
func (c *Config) ValidateWithContext(ctx context.Context) error {
	// LiveKit is optional - only validate if any values are provided
	if c.APIKey == "" && c.APISecret == "" && c.URL == "" {
		return nil
	}

	return validation.ValidateStructWithContext(ctx, c,
		validation.Field(&c.APIKey, validation.Required),
		validation.Field(&c.APISecret, validation.Required),
		validation.Field(&c.URL, validation.Required),
	)
}
