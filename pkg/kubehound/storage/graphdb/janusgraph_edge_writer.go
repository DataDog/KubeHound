package graphdb

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

var _ AsyncPathWriter = (*JanusGraphAsyncWriter[edge.Traversal])(nil)

// NewJanusGraphAsyncEdgeWriter creates a new bulk edge writer instance.
func NewJanusGraphAsyncEdgeWriter(ctx context.Context, dcp *DriverConnectionPool,
	e edge.Builder, opts ...WriterOption) (*JanusGraphAsyncWriter[edge.Traversal], error) {

	options := &writerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	// Creating a transaction modifies the connection pool, acquire the lock
	dcp.Lock.Lock()
	defer dcp.Lock.Unlock()

	source := gremlingo.Traversal_().WithRemote(dcp.Driver)
	tx := source.Tx()
	gtx, err := tx.Begin()
	if err != nil {
		return nil, err
	}

	jw := JanusGraphAsyncWriter[edge.Traversal]{
		label:           e.Label(),
		gremlin:         e.Traversal(),
		dcp:             dcp,
		inserts:         make([]types.TraversalInput, 0, e.BatchSize()),
		traversalSource: gtx,
		transaction:     tx,
		batchSize:       e.BatchSize(),
		writingInFlight: &sync.WaitGroup{},
		consumerChan:    make(chan []types.TraversalInput, e.BatchSize()*channelSizeBatchFactor),
		tags:            options.Tags,
	}

	jw.startBackgroundWriter(ctx)

	return &jw, nil
}
