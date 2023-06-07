package vertex

const (
	roleLabel = "Role"
)

var _ PipelineBuilder = (*Role)(nil)

type Role struct {
}

func (v Role) Label() string {
	return roleLabel
}

func (v Role) BatchSize() int {
	return DefaultBatchSize
}

func (v Role) Traversal() VertexTraversal {
	return nil
}
