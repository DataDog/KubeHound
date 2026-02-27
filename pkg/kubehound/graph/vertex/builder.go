package vertex

import (
	"database/sql"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	gremlin "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

// Optional syntactic sugar.
var __ = gremlin.T__
var Column = gremlin.Column
var P = gremlin.P

// Builder interface defines objects used to construct vertices within our graph database.
type Builder interface {
	// Initialize intializes a vertex builder from the application config
	Initialize(cfg *config.KubehoundConfig) error

	// Label returns the label for the vertex (convention is all camelcase i.e VertexName)
	Label() string

	// BatchSize returns the batch size of bulk inserts (and threshold for triggering a flush).
	BatchSize() int

	// Query returns the SQL SELECT statement for this vertex type.
	// The query MUST use positional parameters ($1, $2) for run_id and cluster_name.
	Query(runID, clusterName string) string

	// Scanner scans a single row into a map suitable for the gremlin traversal.
	Scanner(rows *sql.Rows) (map[string]any, error)

	// Traversal returns a graph traversal function that enables creating vertices from an input array of TraversalInput objects.
	Traversal() types.VertexTraversal
}
