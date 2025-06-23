package ingestion

import (
	"context"
	"errors"
)

var (
	// ErrNoResult is returned when no result is found.
	ErrNoResult = errors.New("no result found")
)

// Reader is the interface for reading container data.
type Reader interface {
	// List returns the list of ingestions.
	List(ctx context.Context, filter ListFilter) ([]Ingestion, error)
	// GetEdgeCountPerClasses returns the count of edges per classes.
	GetEdgeCountPerClasses(ctx context.Context) (map[string]int64, error)
	// GetVertexCountPerClasses returns the count of vertices per classes.
	GetVertexCountPerClasses(ctx context.Context) (map[string]int64, error)
}

// ListFilter represents the filter for ingestions.
type ListFilter struct {
	// Cluster is the cluster to filter by.
	Cluster *string
	// RunID is the run identifier to filter by.
	RunID *string
}
