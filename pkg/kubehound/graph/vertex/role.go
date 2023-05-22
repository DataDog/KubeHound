package vertex

const (
	roleLabel = "Role"
)

var _ Vertex = (*Role)(nil)

type Role struct {
}

func (v Role) Label() string {
	return roleLabel
}

func (v Role) Traversal() VertexTraversal {
	return nil
}
