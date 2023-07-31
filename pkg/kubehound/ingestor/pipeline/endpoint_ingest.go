package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
)

const (
	EndpointIngestName = "k8s-endpoint-ingest"
)

type EndpointIngest struct {
	vertex     *vertex.Endpoint
	collection collections.Endpoint
	r          *IngestResources
}

var _ ObjectIngest = (*EndpointIngest)(nil)

func (i *EndpointIngest) Name() string {
	return EndpointIngestName
}

func (i *EndpointIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	var err error

	i.vertex = &vertex.Endpoint{}
	i.collection = collections.Endpoint{}

	i.r, err = CreateResources(ctx, deps,
		WithCacheWriter(),
		WithStoreWriter(i.collection),
		WithGraphWriter(i.vertex))
	if err != nil {
		return err
	}

	return nil
}

// IngestEndpoint is invoked by the collector for each endpoint slice collected.
// The function ingests an input endpoint slice into the cache/store/graph databases asynchronously.
func (i *EndpointIngest) IngestEndpoint(ctx context.Context, eps types.EndpointType) error {
	if ok, err := preflight.CheckEndpoint(eps); !ok {
		return err
	}

	// Ports specifies the list of network ports exposed by each endpoint in this slice
	// We want to create one store entry per ports#
	for _, ep := range eps.Ports {
		// Normalize endpoint to store object format
		o, err := i.r.storeConvert.Endpoint(ctx, ep, eps)
		if err != nil {
			return err
		}

		// Async write to store
		if err := i.r.writeStore(ctx, i.collection, o); err != nil {
			return err
		}

		// Async write to cache
		ck := cachekey.Endpoint("TODO", int(*ep.Port))
		if err := i.r.writeCache(ctx, ck, o.Id.Hex()); err != nil {
			return err
		}

		// Transform store model to vertex input
		insert, err := i.r.graphConvert.Endpoint(o)
		if err != nil {
			return err
		}

		// Aysnc write to graph
		if err := i.r.writeVertex(ctx, i.vertex, insert); err != nil {
			return err
		}
	}

	return nil
}

// Complete is invoked by the collector when all nodes have been streamed.
// The function flushes all writers and waits for completion.
func (i *EndpointIngest) Complete(ctx context.Context) error {
	return i.r.flushWriters(ctx)
}

func (i *EndpointIngest) Run(ctx context.Context) error {
	return i.r.collect.StreamEndpoints(ctx, i)
}

func (i *EndpointIngest) Close(ctx context.Context) error {
	return i.r.cleanupAll(ctx)
}
