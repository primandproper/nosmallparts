package qdrant

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Name   string `env:"NAME"    json:"name"`
	Host   string `env:"HOST"    json:"host"`
	APIKey string `env:"API_KEY" json:"apiKey"`
	Port   uint16 `env:"PORT"    json:"port"`
}

func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		&c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.Host, validation.Required),
		validation.Field(&c.Port, validation.Required),
		validation.Field(&c.APIKey, validation.Required),
	)
}
