package pipeline

import (
	"context"
)

// ObjectIngest represents an ingestion pipeline that receives an input object from a collector implementation,
// processes and persists all resulting KubeHound objects (store models, cache entries, graph vertices, etc).
//
//go:generate mockery --name ObjectIngest --output mocks --case underscore --filename object_ingest.go --with-expecter
type ObjectIngest interface {
	// Name returns the name of the object ingest pipeline.
	Name() string

	// Initialize intializes an object ingest pipeline with any servic dependencies.
	Initialize(ctx context.Context, deps *Dependencies) error

	// Run executes the ingest pipeline, returning when all are complete.
	Run(ctx context.Context) error

	// Close cleans up any resources held in the ingest pipeline on completion/error.
	Close(ctx context.Context) error
}
