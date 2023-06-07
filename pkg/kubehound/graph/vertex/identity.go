package vertex

const (
	identityLabel = "Identity"
)

var _ PipelineBuilder = (*Identity)(nil)

type Identity struct {
}

func (v Identity) Label() string {
	return identityLabel
}

func (v Identity) BatchSize() int {
	return DefaultBatchSize
}

func (v Identity) Traversal() VertexTraversal {
	return nil
}
