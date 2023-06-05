package vertex

import "github.com/DataDog/KubeHound/pkg/kubehound/models/graph"

const (
	volumeLabel = "Volume"
)

var _ Vertex = (*Volume)(nil)

type Volume struct {
	graph.Volume
}

func (v Volume) Label() string {
	return volumeLabel
}

func (v Volume) Traversal() VertexTraversal {
	return nil
}
