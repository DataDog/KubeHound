package vertex

const (
	podLabel = "Pod"
)

var _ PipelineBuilder = (*Pod)(nil)

type Pod struct {
}

func (v Pod) Label() string {
	return podLabel
}

func (v Pod) BatchSize() int {
	return DefaultBatchSize
}

func (v Pod) Traversal() VertexTraversal {
	return nil
}
