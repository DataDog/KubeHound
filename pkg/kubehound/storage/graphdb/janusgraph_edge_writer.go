package graphdb

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/edge"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

var _ AsyncPathWriter = (*JanusGraphAsyncWriter[edge.Traversal])(nil)

func NewJanusGraphAsyncEdgeWriter(ctx context.Context, drc *gremlingo.DriverRemoteConnection,
	e edge.Builder, opts ...WriterOption) (*JanusGraphAsyncWriter[edge.Traversal], error) {

	options := &writerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	source := gremlingo.Traversal_().WithRemote(drc)
	jw := JanusGraphAsyncWriter[edge.Traversal]{
		label:           e.Label(),
		gremlin:         e.Traversal(),
		inserts:         make([]types.TraversalInput, 0, e.BatchSize()),
		traversalSource: source,
		batchSize:       e.BatchSize(),
		writingInFlight: &sync.WaitGroup{},
		consumerChan:    make(chan []types.TraversalInput, e.BatchSize()*channelSizeBatchFactor),
	}

	jw.startBackgroundWriter(ctx)

	return &jw, nil
}
