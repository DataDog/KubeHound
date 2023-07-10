package vertex

import (
	"fmt"
	"testing"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/stretchr/testify/assert"
)

func TestNode_Traversal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want types.VertexTraversal
		data graph.Node
	}{
		{
			name: "Add Identities in JanusGraph",
			// We set the values to all field with non default values
			// so we are sure all are correctly propagated.
			data: graph.Node{
				StoreID:      "test id",
				Name:         "test name node",
				IsNamespaced: true,
				Namespace:    "lol namespace",
				Compromised:  1,
				Critical:     true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			v := Node{}

			g := gremlingo.GraphTraversalSource{}

			vertexTraversal := v.Traversal()
			inserts := []types.TraversalInput{&tt.data}

			traversal := vertexTraversal(&g, inserts)
			// This is ugly but doesn't need to write to the DB
			// This just makes sure the traversal is correctly returned with the correct values
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test id")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test name node")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "lol namespace")
		})
	}
}
