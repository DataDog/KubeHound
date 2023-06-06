package vertex

import "github.com/DataDog/KubeHound/pkg/kubehound/models/graph"

const (
	containerLabel = "Container"
)

var _ Builder = (*Container)(nil)

type Container struct {
	graph.Container
}

func (v Container) Label() string {
	return containerLabel
}

func (v Container) BatchSize() int {
	return DefaultBatchSize
}

func (v Container) Traversal() VertexTraversal {
	return nil
}
