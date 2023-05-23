package vertex

const (
	volumeLabel = "Volume"
)

var _ Vertex = (*Volume)(nil)

type Volume struct {
}

func (v Volume) Label() string {
	return volumeLabel
}

func (v Volume) Traversal() VertexTraversal {
	return nil
}
