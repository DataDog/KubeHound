package vertex

const (
	containerLabel = "Container"
)

var _ Vertex = (*Container)(nil)

type Container struct {
}

func (v Container) Label() string {
	return containerLabel
}

func (v Container) Traversal() VertexTraversal {
	return nil
}
