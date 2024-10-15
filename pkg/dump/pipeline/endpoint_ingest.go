package pipeline

import (
	"context"
	"path"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	discoveryv1 "k8s.io/api/discovery/v1"
)

type EndpointIngestor struct {
	buffer map[string]*discoveryv1.EndpointSliceList
	writer writer.DumperWriter
}

func ingestEndpointPath(endpoint types.EndpointType) string {
	return path.Join(endpoint.Namespace, collector.EndpointPath)
}

func NewEndpointIngestor(ctx context.Context, dumpWriter writer.DumperWriter) *EndpointIngestor {
	return &EndpointIngestor{
		buffer: make(map[string]*discoveryv1.EndpointSliceList),
		writer: dumpWriter,
	}
}

func (d *EndpointIngestor) IngestEndpoint(ctx context.Context, endpoint types.EndpointType) error {
	if ok, err := preflight.CheckEndpoint(ctx, endpoint); !ok {
		return err
	}

	endpointPath := ingestEndpointPath(endpoint)

	return bufferObject[discoveryv1.EndpointSliceList, types.EndpointType](ctx, endpointPath, d.buffer, endpoint)
}

// Complete() is invoked by the collector when all k8s assets have been streamed.
// The function flushes all writers and waits for completion.
func (d *EndpointIngestor) Complete(ctx context.Context) error {
	return dumpObj[*discoveryv1.EndpointSliceList](ctx, d.buffer, d.writer)
}
