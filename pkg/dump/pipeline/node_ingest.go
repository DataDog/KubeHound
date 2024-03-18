package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"

	corev1 "k8s.io/api/core/v1"
)

type NodeIngestor struct {
	buffer map[string]*corev1.NodeList
	writer writer.DumperWriter
}

func NewNodeIngestor(ctx context.Context, dumpWriter writer.DumperWriter) *NodeIngestor {
	return &NodeIngestor{
		buffer: make(map[string]*corev1.NodeList),
		writer: dumpWriter,
	}
}

func (d *NodeIngestor) IngestNode(ctx context.Context, node types.NodeType) error {
	if ok, err := preflight.CheckNode(node); !ok {
		return err
	}

	return bufferObject[corev1.NodeList, types.NodeType](ctx, collector.NodePath, d.buffer, node)
}

// Complete() is invoked by the collector when all k8s assets have been streamed.
// The function flushes all writers and waits for completion.
func (d *NodeIngestor) Complete(ctx context.Context) error {
	return dumpObj[*corev1.NodeList](ctx, d.buffer, d.writer)
}
