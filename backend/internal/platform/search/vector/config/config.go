package vectorcfg

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/vector"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/vector/pinecone"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/vector/qdrant"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	ProviderQdrant   = "qdrant"
	ProviderPinecone = "pinecone"
)

type Config struct {
	Pinecone *pinecone.Config `env:"PINECONE" envDefault:"qdrant"`
	Qdrant   *qdrant.Config   `env:"QDRANT"   envDefault:"pinecone"`
	Provider string           `env:"PROVIDER" json:"provider"`
}

func (c *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &c,
		validation.Field(&c.Provider, validation.In(ProviderQdrant, ProviderPinecone)),
		validation.Field(&c.Pinecone, validation.When(c.Provider == ProviderPinecone, validation.Required)),
		validation.Field(&c.Qdrant, validation.When(c.Provider == ProviderQdrant, validation.Required)),
	)
}

func (c *Config) ProvideVectorSearcher(logger logging.Logger, tracerProvider tracing.TracerProvider) (vector.Searcher, error) {
	switch c.Provider {
	case ProviderQdrant:
		return qdrant.ProvideQdrantClient(c.Qdrant, logger, tracerProvider)
	case ProviderPinecone:
		return pinecone.ProvidePineconeClient(c.Pinecone, logger, tracerProvider)
	default:
		return nil, fmt.Errorf("unknown provider %q", c.Provider)
	}
}
