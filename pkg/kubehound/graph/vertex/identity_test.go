package vertex

import (
	"fmt"
	"testing"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
	"github.com/stretchr/testify/assert"
)

func TestIdentity_Traversal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want VertexTraversal
		data graph.Identity
	}{
		{
			name: "Add Identities in JanusGraph",
			// We set the values to all field with non default values
			// so we are sure all are correctly propagated.
			data: graph.Identity{
				StoreID:      "test id",
				Name:         "test name identity",
				IsNamespaced: true,
				Namespace:    "lol namespace",
				Type:         "some type",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			v := Identity{}

			g := gremlingo.GraphTraversalSource{}

			vertexTraversal := v.Traversal()
			inserts := []TraversalInput{&tt.data}

			traversal := vertexTraversal(&g, inserts)
			// This is ugly but doesn't need to write to the DB
			// This just makes sure the traversal is correctly returned with the correct values
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test id")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test name identity")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "lol namespace")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "some type")
		})
	}
}
