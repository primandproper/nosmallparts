package qdrant

import (
	"context"
	"fmt"
	"strconv"

	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/cryptography/hashing/fnv"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/internalerrors"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/keys"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/logging"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/observability/tracing"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/pointer"
	"github.com/verygoodsoftwarenotvirus/nosmallparts/backend/internal/platform/search/vector"

	"github.com/qdrant/go-client/qdrant"
)

type client struct {
	logger logging.Logger
	tracer tracing.Tracer
	c      *qdrant.Client
}

func ProvideQdrantClient(cfg *Config, logger logging.Logger, tracerProvider tracing.TracerProvider) (vector.Searcher, error) {
	if cfg == nil {
		return nil, internalerrors.NilConfigError("qdrant")
	}

	pc, err := qdrant.NewClient(&qdrant.Config{
		Host:   cfg.Host,
		Port:   int(cfg.Port),
		APIKey: cfg.APIKey,
		UseTLS: true,
	})
	if err != nil {
		return nil, err
	}

	return &client{
		c:      pc,
		logger: logging.EnsureLogger(logger),
		tracer: tracing.NewTracer(tracing.EnsureTracerProvider(tracerProvider).Tracer(fmt.Sprintf("%s_qdrant", cfg.Name))),
	}, nil
}

func (c *client) createIndexRequest(indexName string) *qdrant.CreateCollection {
	// NOTE: this can be configured via the client later
	return &qdrant.CreateCollection{
		CollectionName: indexName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     4,
			Distance: qdrant.Distance_Cosine,
		}),
	}
}

func (c *client) CreateIndex(ctx context.Context, indexName string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if err := c.c.CreateCollection(ctx, c.createIndexRequest(indexName)); err != nil {
		return observability.PrepareError(err, span, "creating index")
	}

	return nil
}

func idToNumber(id string) (uint64, error) {
	fnvHash, err := fnv.NewFNVHasher().Hash(id)
	if err != nil {
		return 0, err
	}

	parsed, err := strconv.ParseUint(fnvHash, 10, 64)
	if err != nil {
		return 0, err
	}

	return parsed, nil
}

func convertUpsertVectorToUpsertPoints(indexName string, data vector.UpsertVector) (*qdrant.UpsertPoints, error) {
	convertedID, err := idToNumber(data.ID)
	if err != nil {
		return nil, err
	}

	points := []*qdrant.PointStruct{
		{
			Id:      qdrant.NewIDNum(convertedID),
			Vectors: qdrant.NewVectors(data.Vectors...),
			Payload: qdrant.NewValueMap(data.Metadata),
		},
	}

	return &qdrant.UpsertPoints{
		CollectionName: indexName,
		Points:         points,
		Wait:           pointer.To(true),
	}, nil
}

func (c *client) UpsertVector(ctx context.Context, indexName string, data vector.UpsertVector) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	convertedData, err := convertUpsertVectorToUpsertPoints(indexName, data)
	if err != nil {
		return observability.PrepareError(err, span, "converting upsert vector")
	}

	if _, err = c.c.Upsert(ctx, convertedData); err != nil {
		return observability.PrepareError(err, span, "upserting vector")
	}

	return nil
}

func (c *client) QueryVector(ctx context.Context, indexName string, queryVector []float32, _ map[string]any) ([]*vector.QueryResult, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue(keys.IndexNameKey, indexName)

	searchResult, err := c.c.Query(ctx, &qdrant.QueryPoints{
		CollectionName: indexName,
		Query:          qdrant.NewQuery(queryVector...),
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "querying vectors")
	}

	logger = logger.WithValue("count", len(searchResult))
	logger.Info("vector query results received")

	output := []*vector.QueryResult{}

	for _, result := range searchResult {
		output = append(output, &vector.QueryResult{
			Vectors: result.Vectors.GetVector().Data,
			Score:   result.Score,
		})
	}

	return output, nil
}

func (c *client) DeleteVector(ctx context.Context, indexName, vectorID string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	if _, err := c.c.DeleteVectors(ctx, &qdrant.DeletePointVectors{
		CollectionName: indexName,
		Vectors: &qdrant.VectorsSelector{
			Names: []string{vectorID},
		},
		PointsSelector: qdrant.NewPointsSelector(qdrant.NewID(vectorID)),
		Wait:           pointer.To(true),
	}); err != nil {
		return observability.PrepareError(err, span, "deleting vector")
	}

	return nil
}
