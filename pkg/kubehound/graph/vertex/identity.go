package vertex

const (
	identityLabel = "Identity"
)

var _ Vertex = (*Identity)(nil)

type Identity struct {
}

func (v Identity) Label() string {
	return identityLabel
}

func (v Identity) Traversal() VertexTraversal {
	return nil
}
