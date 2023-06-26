package vertex

import (
	"fmt"
	"testing"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/stretchr/testify/assert"
)

func TestRole_Traversal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want Traversal
		data graph.Role
	}{
		{
			name: "Add Identities in JanusGraph",
			// We set the values to all field with non default values
			// so we are sure all are correctly propagated.
			data: graph.Role{
				StoreID:      "test id",
				Name:         "test name role",
				IsNamespaced: true,
				Namespace:    "lol namespace",
				Rules:        []string{"rule1", "rule2"},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			v := Role{}

			g := gremlingo.GraphTraversalSource{}

			vertexTraversal := v.Traversal()
			inserts := []types.TraversalInput{&tt.data}

			traversal := vertexTraversal(&g, inserts)
			// This is ugly but doesn't need to write to the DB
			// This just makes sure the traversal is correctly returned with the correct values
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test id")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test name role")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "lol namespace")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "rule1")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "rule2")
		})
	}
}
