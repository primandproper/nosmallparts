package pinecone

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/internalerrors"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/vector"

	"github.com/pinecone-io/go-pinecone/v2/pinecone"
	"google.golang.org/protobuf/types/known/structpb"
)

type client struct {
	logger logging.Logger
	tracer tracing.Tracer
	c      *pinecone.Client
}

func ProvidePineconeClient(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider) (vector.Searcher, error) {
	if cfg == nil {
		return nil, internalerrors.NilConfigError("pinecone")
	}

	pc, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: cfg.APIKey,
	})
	if err != nil {
		return nil, err
	}

	return &client{
		c:      pc,
		logger: logging.EnsureLogger(logger),
		tracer: tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(fmt.Sprintf("%s_pinecone", cfg.Name))),
	}, nil
}

func (c *client) createIndexRequest(indexName string) *pinecone.CreateServerlessIndexRequest {
	// NOTE: this can be configured via the client later
	return &pinecone.CreateServerlessIndexRequest{
		Name:      indexName,
		Dimension: 3,
		Metric:    pinecone.Cosine,
		Cloud:     pinecone.Aws,
		Region:    "us-east-1",
		Tags:      &pinecone.IndexTags{"environment": "development"},
	}
}

func (c *client) CreateIndex(ctx context.Context, indexName string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if _, err := c.c.CreateServerlessIndex(ctx, c.createIndexRequest(indexName)); err != nil {
		return observability.PrepareError(err, span, "creating index")
	}

	return nil
}

func (c *client) UpsertVector(ctx context.Context, indexName string, data vector.UpsertVector) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.IndexNameKey, indexName)

	idx, err := c.c.DescribeIndex(ctx, indexName)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "describing index")
	}

	idxConnection, err := c.c.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "creating index")
	}

	metadata, err := structpb.NewStruct(data.Metadata)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "parsing metadata")
	}

	vecs := []*pinecone.Vector{
		{
			Id:       data.ID,
			Values:   data.Vectors,
			Metadata: metadata,
		},
	}

	if _, err = idxConnection.UpsertVectors(ctx, vecs); err != nil {
		return fmt.Errorf("upserting vectors: %w", err)
	}

	return nil
}

func (c *client) QueryVector(ctx context.Context, indexName string, queryVector []float32, queryMetadata map[string]any) ([]*vector.QueryResult, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	idx, err := c.c.DescribeIndex(ctx, indexName)
	if err != nil {
		return nil, observability.PrepareError(err, span, "describing index")
	}

	idxConnection, err := c.c.Index(pinecone.NewIndexConnParams{Host: idx.Host})
	if err != nil {
		return nil, observability.PrepareError(err, span, "creating IndexConnection for Host")
	}

	metadataFilter, err := structpb.NewStruct(queryMetadata)
	if err != nil {
		return nil, observability.PrepareError(err, span, "creating metadataFilter: %v", err)
	}

	searchResult, err := idxConnection.QueryByVectorValues(ctx, &pinecone.QueryByVectorValuesRequest{
		Vector:         queryVector,
		TopK:           3,
		MetadataFilter: metadataFilter,
		IncludeValues:  true,
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "querying by vector: %v", err)
	}

	output := []*vector.QueryResult{}

	for _, result := range searchResult.Matches {
		output = append(output, &vector.QueryResult{
			Vectors: result.Vector.Values,
			Score:   result.Score,
		})
	}

	return output, nil
}

func (c *client) DeleteVector(ctx context.Context, indexName, vectorID string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	idx, err := c.c.DescribeIndex(ctx, indexName)
	if err != nil {
		return observability.PrepareError(err, span, "describing index %q", indexName)
	}

	idxConnection, err := c.c.Index(pinecone.NewIndexConnParams{Host: idx.Host, Namespace: "example-namespace"})
	if err != nil {
		return observability.PrepareError(err, span, "creating index connection for host %q", idx.Host)
	}

	if err = idxConnection.DeleteVectorsById(ctx, []string{vectorID}); err != nil {
		return observability.PrepareError(err, span, "deleting vector %q", vectorID)
	}

	return nil
}
