package config

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config manages our body validation.
type Config struct {
	OAuth2ClientCreationDisabled bool `env:"CREATION_DISABLED" json:"creationEnabled"`
}

func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, c)
}
