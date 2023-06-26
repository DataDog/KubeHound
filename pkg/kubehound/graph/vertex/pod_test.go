package vertex

import (
	"fmt"
	"testing"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/stretchr/testify/assert"
)

func TestPod_Traversal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want Traversal
		data graph.Pod
	}{
		{
			name: "Add Identities in JanusGraph",
			// We set the values to all field with non default values
			// so we are sure all are correctly propagated.
			data: graph.Pod{
				StoreID:                "test id",
				Name:                   "test name pod",
				IsNamespaced:           true,
				Namespace:              "lol namespace",
				Compromised:            1,
				Critical:               true,
				SharedProcessNamespace: true,
				ServiceAccount:         "some service account",
				Node:                   "lol node",
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			v := Pod{}

			g := gremlingo.GraphTraversalSource{}

			vertexTraversal := v.Traversal()
			inserts := []types.TraversalInput{&tt.data}

			traversal := vertexTraversal(&g, inserts)
			// This is ugly but doesn't need to write to the DB
			// This just makes sure the traversal is correctly returned with the correct values
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test id")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test name pod")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "lol namespace")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "some service account")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "lol node")
		})
	}
}
