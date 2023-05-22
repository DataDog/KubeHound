package vertex

const (
	nodeLabel = "Node"
)

var _ Vertex = (*Node)(nil)

type Node struct {
}

func (v Node) Label() string {
	return nodeLabel
}

func (v Node) Traversal() VertexTraversal {
	return nil
}
