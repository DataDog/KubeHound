package vertex

import "github.com/DataDog/KubeHound/pkg/kubehound/models/graph"

const (
	identityLabel = "Identity"
)

var _ Builder = (*Identity)(nil)

type Identity struct {
	graph.Identity
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
