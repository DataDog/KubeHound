package vertex

var _ Vertex = (*Node)(nil)

type Node struct {
}

func (v Node) Label() string {
	return ""
}

func (v Node) Traversal() VertexTraversal {
	return nil
}
