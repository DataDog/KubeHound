package graphdb

import (
	"context"
	"sync"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/path"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
)

var _ AsyncPathWriter = (*JanusGraphAsyncWriter[path.Traversal])(nil)

// NewJanusGraphAsyncPathWriter creates a new bulk path writer instance.
func NewJanusGraphAsyncPathWriter(ctx context.Context, drc *gremlingo.DriverRemoteConnection,
	p path.Builder, opts ...WriterOption) (*JanusGraphAsyncWriter[path.Traversal], error) {

	options := &writerOptions{}
	for _, opt := range opts {
		opt(options)
	}

	jw := JanusGraphAsyncWriter[path.Traversal]{
		label:           p.Label(),
		gremlin:         p.Traversal(),
		drc:             drc,
		inserts:         make([]types.TraversalInput, 0, p.BatchSize()),
		traversalSource: gremlingo.Traversal_().WithRemote(drc),
		batchSize:       p.BatchSize(),
		writingInFlight: &sync.WaitGroup{},
		consumerChan:    make(chan []types.TraversalInput, p.BatchSize()*channelSizeBatchFactor),
		tags:            options.Tags,
	}

	jw.startBackgroundWriter(ctx)

	return &jw, nil
}
