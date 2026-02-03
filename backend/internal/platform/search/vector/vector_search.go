package vector

import (
	"context"
)

type (
	// UpsertVector is an abstraction meant to be used in place of vendor-specific upsert structs.
	UpsertVector struct {
		Metadata map[string]any
		ID       string
		Vectors  []float32
	}

	// QueryResult is an abstraction meant to be used in place of vendor-specific query structs.
	QueryResult struct {
		Vectors []float32
		Score   float32
	}

	// Searcher provides an interface for implementing vector search.
	Searcher interface {
		CreateIndex(ctx context.Context, indexName string) error
		UpsertVector(ctx context.Context, indexName string, data UpsertVector) error
		QueryVector(ctx context.Context, indexName string, queryVector []float32, queryMetadata map[string]any) ([]*QueryResult, error)
		DeleteVector(ctx context.Context, indexName, vectorID string) error
	}
)
