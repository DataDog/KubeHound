package vertex

const (
	volumeLabel = "Volume"
)

var _ Builder = (*Volume)(nil)

type Volume struct {
}

func (v Volume) Label() string {
	return volumeLabel
}

func (v Volume) BatchSize() int {
	return DefaultBatchSize
}

func (v Volume) Traversal() VertexTraversal {
	return nil
}
