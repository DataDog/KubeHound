package edge

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

type ProcessEntryCallback func(ctx context.Context, model any) error
type CompleteQueryCallback func(ctx context.Context) error
type EdgeTraversal func(g *gremlingo.GraphTraversalSource, inserts []any) *gremlingo.GraphTraversal

type Edge interface {
	// Label returns the label for the edge (convention is all uppercase i.e EDGE_NAME)
	Label() string

	// Traversal returns a graph traversal function that enables creating edges from an input array of data types.
	Traversal() EdgeTraversal

	//
	Stream(ctx context.Context, store storedb.Provider, process ProcessEntryCallback, complete CompleteQueryCallback) error

	// Processors translates a object retrieved from the data store into and input data type to pass to the traversal.
	Processor(ctx context.Context, model any) (any, error)
}

// EdgeRegistry holds details of edges (i.e attacks) registered in KubeHound.
type EdgeRegistry map[string]Edge

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
func Register(edge Edge) {
	log.I.Infof("Registering edge %s", edge.Label())
	Registry()[edge.Label()] = edge
}
