package edge

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	BatchSizeDefault       = 250
	BatchSizeMedium        = BatchSizeDefault / 4
	BatchSizeSmall         = BatchSizeDefault / 8
	BatchSizeClusterImpact = 1
)

// Optional syntactic sugar.
var __ = gremlin.T__
var P = gremlin.P

// Edge interface defines objects used to construct edges within our graph database through processing data from the intermediate store.

//go:generate mockery --name Builder --output mocks --case underscore --filename edge.go --with-expecter
type Builder interface {
	// Initialize intializes an edge builder from the application config
	Initialize(cfg *config.EdgeBuilderConfig) error

	// Name returns the unique name for the edge builder. This must be unique.
	Name() string

	// Label returns the label for the edge (convention is all uppercase i.e EDGE_NAME).
	Label() string

	// BatchSize returns the batch size of bulk inserts (and threshold for triggering a flush).
	BatchSize() int

	// Traversal returns a graph traversal function that enables creating edges from an input array of TraversalInput objects.
	Traversal() types.EdgeTraversal

	// Processor transforms an object queued for writing to a format suitable for consumption by the Traversal function.
	Processor(context.Context, *converter.ObjectIdConverter, any) (any, error)

	// Stream will query the store db for the data required to create an edge and stream to graph DB via callbacks.
	// Each query result is encapsulated within an DataContainer and transformed to a TraversalInput via a call to
	// the edge's Processor function. Invoking the complete callback signals the end of the stream.
	Stream(ctx context.Context, store storedb.Provider, cache cache.CacheReader,
		process types.ProcessEntryCallback, complete types.CompleteQueryCallback) error
}
