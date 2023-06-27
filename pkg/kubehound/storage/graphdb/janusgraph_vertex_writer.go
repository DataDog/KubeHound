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
func NewJanusGraphAsyncVertexWriter(ctx context.Context, drc *gremlingo.DriverRemoteConnection,
	v vertex.Builder, opts ...WriterOption) (*JanusGraphAsyncWriter[vertex.Traversal], error) {

	options := &writerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	source := gremlingo.Traversal_().WithRemote(drc)
	tx := source.Tx()
	gtx, err := tx.Begin()
	if err != nil {
		return nil, err
	}

	jw := JanusGraphAsyncWriter[vertex.Traversal]{
		label:           v.Label(),
		gremlin:         v.Traversal(),
		drc:             drc,
		inserts:         make([]types.TraversalInput, 0, v.BatchSize()),
		traversalSource: gtx,
		transaction:     tx,
		batchSize:       v.BatchSize(),
		writingInFlight: &sync.WaitGroup{},
		consumerChan:    make(chan []types.TraversalInput, v.BatchSize()*channelSizeBatchFactor),
		tags:            options.Tags,
		cache:           options.Cache,
	}

	jw.startBackgroundWriter(ctx)

	return &jw, nil
}
