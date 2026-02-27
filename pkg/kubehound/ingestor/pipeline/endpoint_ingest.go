package pipeline

import (
	"context"
	"errors"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
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
		WithStoreWriter(i.collection),
		WithGraphWriter(i.vertex))
	if err != nil {
		return err
	}

	return nil
}

// IngestEndpoint is invoked by the collector for each endpoint slice collected.
// The function ingests an input endpoint slice into the store/graph databases.
func (i *EndpointIngest) IngestEndpoint(ctx context.Context, eps types.EndpointType) error {
	if ok, err := preflight.CheckEndpoint(ctx, eps); !ok {
		return err
	}

	// We want to create one store entry per port and per address.
	for _, port := range eps.Ports {
		for _, addr := range eps.Endpoints {
			// Normalize endpoint to store object format
			o, err := i.r.storeConvert.Endpoint(ctx, addr, port, eps)
			if err != nil {
				if errors.Is(err, converter.ErrEndpointTarget) {
					log.Trace(ctx).Debugf("Endpoint dropped: %s: %s", err.Error(), addr.TargetRef)

					return nil
				}

				return err
			}

			// Write to store
			if err := i.r.writeStore(ctx, i.collection, o); err != nil {
				return err
			}

			// Transform store model to vertex input
			insert, err := i.r.graphConvert.Endpoint(o)
			if err != nil {
				return err
			}

			// Write to graph
			if err := i.r.writeVertex(ctx, i.vertex, insert); err != nil {
				return err
			}
		}
	}

	return nil
}

// Complete is invoked by the collector when all endpoints have been streamed.
func (i *EndpointIngest) Complete(ctx context.Context) error {
	return i.r.flushWriters(ctx)
}

func (i *EndpointIngest) Run(ctx context.Context) error {
	return i.r.collect.StreamEndpoints(ctx, i)
}

func (i *EndpointIngest) Close(ctx context.Context) error {
	return i.r.cleanupAll(ctx)
}
