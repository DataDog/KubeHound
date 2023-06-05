package vertex

import "github.com/DataDog/KubeHound/pkg/kubehound/models/graph"

const (
	podLabel = "Pod"
)

var _ Vertex = (*Pod)(nil)

type Pod struct {
	graph.Pod
}

func (v Pod) Label() string {
	return podLabel
}

func (v Pod) Traversal() VertexTraversal {
	return nil
}
