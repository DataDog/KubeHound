package vertex

import (
	"fmt"
	"testing"

	"github.com/DataDog/KubeHound/pkg/kubehound/graph/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/stretchr/testify/assert"
)

func TestVolume_Traversal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want types.VertexTraversal
		data graph.Volume
	}{
		{
			name: "Add Identities in JanusGraph",
			// We set the values to all field with non default values
			// so we are sure all are correctly propagated.
			data: graph.Volume{
				StoreID:    "test id",
				Name:       "test name volume",
				Type:       "test type",
				SourcePath: "some path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := Volume{}

			g := gremlingo.GraphTraversalSource{}

			vertexTraversal := v.Traversal()
			inserts := []any{&tt.data}

			traversal := vertexTraversal(&g, inserts)
			// This is ugly but doesn't need to write to the DB
			// This just makes sure the traversal is correctly returned with the correct values
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test id")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test name volume")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test type")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "some path")
		})
	}
}
