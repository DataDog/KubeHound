package edge

import (
	"context"
	"database/sql"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

// Optional syntactic sugar.
var __ = gremlin.T__
var P = gremlin.P

// Builder interface defines objects used to construct edges within our graph database through processing data from the intermediate store.

//go:generate mockery --name Builder --output mocks --case underscore --filename edge.go --with-expecter
type Builder interface {
	// Initialize intializes an edge builder from the application config
	Initialize(cfg *config.EdgeBuilderConfig, runtime *config.DynamicConfig) error

	// Name returns the unique name for the edge builder. This must be unique.
	Name() string

	// Label returns the label for the edge (convention is all uppercase i.e EDGE_NAME).
	Label() string

	// AttckTechniqueID returns the ATT&CK technique ID for the edge.
	AttckTechniqueID() AttckTechniqueID

	// AttckTacticID returns the ATT&CK tactic ID for the edge.
	AttckTacticID() AttckTacticID

	// BatchSize returns the batch size of bulk inserts (and threshold for triggering a flush).
	BatchSize() int

	// Traversal returns a graph traversal function that enables creating edges from an input array of TraversalInput objects.
	Traversal() types.EdgeTraversal

	// Stream queries the store db, processes each row, and writes edges to the graph db writer.
	Stream(ctx context.Context, db *sql.DB, w types.EdgeWriter) error
}

// DependentBuilder interface defines objects used to construct edges with dependencies on other edges in the graph.
// Dependent edges are built last and their dependencies cannot be dependent edges themselves.
type DependentBuilder interface {
	Builder

	// Dependencies returns the edge labels of all dependencies.
	Dependencies() []string
}
