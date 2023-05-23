package vertex

const (
	podLabel = "Pod"
)

var _ Vertex = (*Pod)(nil)

type Pod struct {
}

func (v Pod) Label() string {
	return podLabel
}

func (v Pod) Traversal() VertexTraversal {
	return nil
}
