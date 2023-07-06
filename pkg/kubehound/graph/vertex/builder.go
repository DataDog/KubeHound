package vertex

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

const (
	BatchSizeDefault = 500
)

// Optional syntactic sugar.
var __ = gremlin.T__
var Column = gremlin.Column
var P = gremlin.P

// Builder interface defines objects used to construct vertices within our graph database through processing data from an ingestion pipeline.
type Builder interface {
	// Label returns the label for the vertex (convention is all camelcase i.e VertexName)
	Label() string

	// BatchSize returns the batch size of bulk inserts (and threshold for triggering a flush).
	BatchSize() int

	// Processor transforms an object queued for writing to a format suitable for consumption by the Traversal function.
	Processor(context.Context, any) (any, error)

	// Traversal returns a graph traversal function that enables creating vertices from an input array of TraversalInput objects.
	Traversal() types.VertexTraversal
}
