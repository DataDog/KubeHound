package vertex

import (
	"fmt"
	"testing"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
	"github.com/stretchr/testify/assert"
)

func TestContainer_Traversal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want VertexTraversal
		data graph.Container
	}{
		{
			name: "Add containers in JanusGraph",
			// We set the values to all field with non default values
			// so we are sure all are correctly propagated.
			data: graph.Container{
				StoreID:      "test id",
				Name:         "test name",
				Image:        "image",
				Command:      []string{"/usr/bin/sleep"},
				Args:         []string{"600", "lol2ndarguments"},
				Capabilities: []string{"NET_CAP_ADMIN", "NET_RAW_ADMIN"},
				Privileged:   true,
				PrivEsc:      true,
				HostPID:      true,
				HostPath:     true,
				HostIPC:      true,
				HostNetwork:  true,
				RunAsUser:    1234,
				Ports:        []int{1337},
				Pod:          "test pod",
				Node:         "test node",
				Compromised:  1,
				Critical:     true,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			v := Container{}

			g := gremlingo.GraphTraversalSource{}

			vertexTraversal := v.Traversal()
			inserts := []TraversalInput{&tt.data}

			traversal := vertexTraversal(&g, inserts)
			// This is ugly but doesn't need to write to the DB
			// This just makes sure the traversal is correctly returned with the correct values
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test id")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test name")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "image")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "/usr/bin/sleep")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "600")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "lol2ndarguments")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "1234")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "1337")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "NET_CAP_ADMIN")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "NET_RAW_ADMIN")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test pod")
			assert.Contains(t, fmt.Sprintf("%s", traversal.Bytecode), "test node")
		})
	}
}
