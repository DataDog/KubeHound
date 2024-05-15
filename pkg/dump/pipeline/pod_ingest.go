package pipeline

import (
	"context"
	"path"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"

	corev1 "k8s.io/api/core/v1"
)

func ingestPodPath(pod types.PodType) string {
	return path.Join(pod.Namespace, collector.PodPath)
}

type PodIngestor struct {
	buffer map[string]*corev1.PodList
	writer writer.DumperWriter
}

func NewPodIngestor(ctx context.Context, dumpWriter writer.DumperWriter) *PodIngestor {
	return &PodIngestor{
		buffer: make(map[string]*corev1.PodList),
		writer: dumpWriter,
	}
}

func (d *PodIngestor) IngestPod(ctx context.Context, pod types.PodType) error {
	if ok, err := preflight.CheckPod(pod); !ok {
		return err
	}

	podPath := ingestPodPath(pod)

	return bufferObject[corev1.PodList, types.PodType](ctx, podPath, d.buffer, pod)
}

// Complete() is invoked by the collector when all k8s assets have been streamed.
// The function flushes all writers and waits for completion.
func (d *PodIngestor) Complete(ctx context.Context) error {
	return dumpObj[*corev1.PodList](ctx, d.buffer, d.writer)
}
