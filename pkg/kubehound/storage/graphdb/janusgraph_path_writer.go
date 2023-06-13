package graphdb

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/path"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
)

var _ AsyncPathWriter = (*JanusGraphAsyncWriter[path.Traversal])(nil)

func NewJanusGraphAsyncPathWriter(ctx context.Context, drc *gremlingo.DriverRemoteConnection,
	p path.Builder, opts ...WriterOption) (*JanusGraphAsyncWriter[path.Traversal], error) {

	options := &writerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	source := gremlingo.Traversal_().WithRemote(drc)
	jw := JanusGraphAsyncWriter[path.Traversal]{
		label:           p.Label(),
		gremlin:         p.Traversal(),
		inserts:         make([]types.TraversalInput, 0, p.BatchSize()),
		traversalSource: source,
		batchSize:       p.BatchSize(),
		writingInFlight: &sync.WaitGroup{},
		consumerChan:    make(chan []types.TraversalInput, p.BatchSize()*channelSizeBatchFactor),
	}

	jw.startBackgroundWriter(ctx)

	return &jw, nil
}
