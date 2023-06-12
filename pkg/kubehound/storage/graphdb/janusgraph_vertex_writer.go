package graphdb

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

var _ AsyncVertexWriter = (*JanusGraphAsyncWriter[vertex.Traversal])(nil)

func NewJanusGraphAsyncVertexWriter(ctx context.Context, drc *gremlingo.DriverRemoteConnection,
	v vertex.Builder, opts ...WriterOption) (*JanusGraphAsyncWriter[vertex.Traversal], error) {

	options := &writerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	source := gremlingo.Traversal_().WithRemote(drc)
	jw := JanusGraphAsyncWriter[vertex.Traversal]{
		label:           v.Label(),
		gremlin:         v.Traversal(),
		inserts:         make([]types.TraversalInput, 0, v.BatchSize()),
		traversalSource: source,
		batchSize:       v.BatchSize(),
		writingInFlight: &sync.WaitGroup{},
		consumerChan:    make(chan []types.TraversalInput, v.BatchSize()*channelSizeBatchFactor),
	}

	jw.startBackgroundWriter(ctx)

	return &jw, nil
}
