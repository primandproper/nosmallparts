package pinecone

import (
	"context"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Config struct {
	Name   string `env:"NAME"    json:"name"`
	APIKey string `env:"API_KEY" json:"apiKey"`
}

func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		&c,
		validation.Field(&c.Name, validation.Required),
		validation.Field(&c.APIKey, validation.Required),
	)
}
