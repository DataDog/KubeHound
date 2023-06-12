package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

const (
	NodeIngestName = "k8s-node-ingest"
)

type NodeIngest struct {
	vertex     vertex.Node
	collection collections.Node
	r          *IngestResources
}

var _ ObjectIngest = (*NodeIngest)(nil)

func (i *NodeIngest) Name() string {
	return NodeIngestName
}

func (i *NodeIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	var err error

	i.vertex = vertex.Node{}
	i.collection = collections.Node{}

	i.r, err = CreateResources(ctx, deps,
		WithCacheWriter(),
		WithStoreWriter(i.collection),
		WithGraphWriter(i.vertex))
	if err != nil {
		return err
	}

	return nil
}

// streamCallback is invoked by the collector for each node collected.
// The function ingests an input node into the cache/store/graph databases asynchronously.
func (i *NodeIngest) IngestNode(ctx context.Context, node types.NodeType) error {
	log.I.Infof("+++++++ INGEST NODE: (%s) ", node.Name)
	// Normalize node to store object format
	o, err := i.r.storeConvert.Node(ctx, node)
	if err != nil {
		return err
	}

	// Async write to store
	if err := i.r.storeWriter(i.collection).Queue(ctx, o); err != nil {
		return err
	}

	// Async write to cache
	if err := i.r.cacheWriter.Queue(ctx, cache.NodeKey(o.K8.Name), o.Id.Hex()); err != nil {
		return err
	}

	// Transform store model to vertex input
	v, err := i.r.graphConvert.Node(o)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.r.graphWriter(i.vertex).Queue(ctx, v); err != nil {
		return err
	}

	return nil
}

// completeCallback is invoked by the collector when all nodes have been streamed.
// The function flushes all writers and waits for completion.
func (i *NodeIngest) Complete(ctx context.Context) error {
	return i.r.flushWriters(ctx)
}

func (i *NodeIngest) Run(ctx context.Context) error {
	log.I.Infof("!!!!!! run NodeIngest !!!!!!!")
	return i.r.collect.StreamNodes(ctx, i)
}

func (i *NodeIngest) Close(ctx context.Context) error {
	return i.r.cleanupAll(ctx)
}
