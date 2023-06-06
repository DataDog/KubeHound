package edge

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlin "github.com/apache/tinkerpop/gremlin-go/driver"
)

const (
	DefaultBatchSize = 200
)

// An object to encapsulate the raw data required to create one or more edges. For example a pod id and a node id.
type DataContainer any

// An object to be consumed by an edge traversal function to insert an edge into the graph database. This should contain
// the requisite information to identify vertices to link together and any attributes to be attached to the edge.
type TraversalInput any

// ProcessEntryCallback is a callback provided by the the edge builder that will convert edge query results into graph database writes.
type ProcessEntryCallback func(ctx context.Context, model DataContainer) error

// CompleteQueryCallback is a callback provided by the the edge builder that will flush any outstanding graph database writes.
type CompleteQueryCallback func(ctx context.Context) error

// EdgeTraversal returns the function to create a graph database edge insert from an array of input objects.
type EdgeTraversal func(source *gremlin.GraphTraversalSource, inserts []TraversalInput) *gremlin.GraphTraversal

// Edge interface defines objects used to construct edges within our graph database through processing data from the intermediate store.

//go:generate mockery --name Builder --output mocks --case underscore --filename edge.go --with-expecter
type Builder interface {
	// Label returns the label for the edge (convention is all uppercase i.e EDGE_NAME)
	Label() string

	// BatchSize returns the batch size of bulk inserts (and threshold for triggering a flush).
	BatchSize() int

	// Traversal returns a graph traversal function that enables creating edges from an input array of TraversalInput objects.
	Traversal() EdgeTraversal

	// Stream will query the store db for the data required to create an edge and stream to graph DB via callbacks.
	// Each query result is encapsulated within an DataContainer and transformed to a TraversalInput via a call to
	// the edge's Processor function. Invoking the complete callback signals the end of the stream.
	Stream(ctx context.Context, store storedb.Provider, process ProcessEntryCallback, complete CompleteQueryCallback) error

	// Processor translates an DataContainer retrieved from the data store into a TraversalInput to pass to the traversal.
	Processor(ctx context.Context, model DataContainer) (TraversalInput, error)
}

// EdgeRegistry holds details of edges (i.e attacks) registered in KubeHound.
type EdgeRegistry map[string]Builder

// EdgeRegistry singleton support
var registryInstance EdgeRegistry
var erOnce sync.Once

// Registry returns the EdgeRegistry singleton.
func Registry() EdgeRegistry {
	erOnce.Do(func() {
		registryInstance = make(EdgeRegistry)
	})

	return registryInstance
}

// Register loads the provided edge into the registry.
func Register(edge Builder) {
	log.I.Infof("Registering edge %s", edge.Label())
	Registry()[edge.Label()] = edge
}
