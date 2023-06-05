package vertex

import "github.com/DataDog/KubeHound/pkg/kubehound/models/graph"

const (
	containerLabel = "Container"
)

var _ Vertex = (*Container)(nil)

type Container struct {
	graph.Container
}

func (v Container) Label() string {
	return containerLabel
}

func (v Container) Traversal() VertexTraversal {
	return nil
}
