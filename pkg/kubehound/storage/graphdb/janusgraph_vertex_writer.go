package graphdb

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

var _ AsyncVertexWriter = (*JanusGraphAsyncWriter[vertex.Traversal])(nil)

// NewJanusGraphAsyncVertexWriter creates a new bulk vertex writer instance.
func NewJanusGraphAsyncVertexWriter(ctx context.Context, dcp *DriverConnectionPool,
	v vertex.Builder, opts ...WriterOption) (*JanusGraphAsyncWriter[vertex.Traversal], error) {

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

	jw := JanusGraphAsyncWriter[vertex.Traversal]{
		label:           v.Label(),
		gremlin:         v.Traversal(),
		dcp:             dcp,
		inserts:         make([]types.TraversalInput, 0, v.BatchSize()),
		traversalSource: gtx,
		transaction:     tx,
		batchSize:       v.BatchSize(),
		writingInFlight: &sync.WaitGroup{},
		consumerChan:    make(chan []types.TraversalInput, v.BatchSize()*channelSizeBatchFactor),
		tags:            options.Tags,
	}

	jw.startBackgroundWriter(ctx)

	return &jw, nil
}
