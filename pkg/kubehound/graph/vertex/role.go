package vertex

import "github.com/DataDog/KubeHound/pkg/kubehound/models/graph"

const (
	roleLabel = "Role"
)

var _ Vertex = (*Role)(nil)

type Role struct {
	graph.Role
}

func (v Role) Label() string {
	return roleLabel
}

func (v Role) Traversal() VertexTraversal {
	return nil
}
