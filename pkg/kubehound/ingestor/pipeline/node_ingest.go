package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
)

const (
	NodeIngestName = "k8s-node-ingest"
)

type NodeIngest struct {
	vertex     vertex.Node
	collection collections.Node
	BaseObjectIngest
}

var _ ObjectIngest = (*NodeIngest)(nil)

func (i NodeIngest) streamCallback(ctx context.Context, node *types.NodeType) error {
	// Normalize node to store object format
	o, err := i.opts.storeConvert.Node(ctx, *node)
	if err != nil {
		return err
	}

	// Async write to store
	if err := i.storeWriter(i.collection).Queue(ctx, o); err != nil {
		return err
	}

	// Async write to cache
	if err := i.opts.cacheWriter.Queue(ctx, cache.NodeKey(o.K8.Name), o.Id); err != nil {
		return err
	}

	// Transform store model to vertex input
	v, err := i.opts.graphConvert.Node(o)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.graphWriter(i.vertex).Queue(ctx, v); err != nil {
		return err
	}

	return nil
}

func (i NodeIngest) completeCallback(ctx context.Context) error {
	return i.flushWriters(ctx)
}

func (i NodeIngest) Name() string {
	return NodeIngestName
}

func (i NodeIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	var err error
	defer func() {
		if err != nil {
			i.cleanup(ctx)
		}
	}()

	i.vertex = vertex.Node{}
	i.collection = collections.Node{}
	err = i.baseInitialize(ctx, deps,
		WithCacheWriter(),
		WithStoreWriter(i.collection),
		WithGraphWriter(i.vertex))

	return err
}

func (i NodeIngest) Run(ctx context.Context) error {
	return i.opts.collect.StreamNodes(ctx, i.streamCallback, i.completeCallback)
}

func (i NodeIngest) Close(ctx context.Context) error {
	return i.cleanup(ctx)
}
