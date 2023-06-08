package vertex

import (
	"fmt"
	"testing"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/utils"
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
				StoreId:      "test id",
				Name:         "test name",
				Image:        "image",
				Command:      []string{"/usr/bin/sleep"},
				Args:         []string{"600"},
				Capabilities: []string{"NET_CAP_ADMIN", "NET_RAW_ADMINaa"},
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

			dbHost := "ws://localhost:8182/gremlin"
			driver, err := gremlingo.NewDriverRemoteConnection(dbHost)
			assert.NoError(t, err)

			g := gremlingo.Traversal_().WithRemote(driver)
			assert.NoError(t, err)

			insert, err := utils.StructToMap(tt.data)
			assert.NoError(t, err)

			vertexTraversal := v.Traversal()
			inserts := []TraversalInput{insert}

			fmt.Printf("inserts: %v\n", inserts)
			traversal := vertexTraversal(g, inserts)

			// Write to db
			promise := traversal.Iterate()
			err = <-promise
			assert.NoError(t, err)

			// NO IDEA
			// driver, err = gremlingo.NewDriverRemoteConnection(dbHost)
			// assert.NoError(t, err)
			// g = gremlingo.Traversal_().WithRemote(driver)
			// // tx = traversal.Tx()
			// // g, err = tx.Begin()

			// assert.NoError(t, err)
			// test := g.V().HasLabel(v.Label()).ValueMap()
			// res, err := test.Traversal.GetResultSet()
			// assert.NoError(t, err)
			// data, err := res.All()
			// for _, d := range data {
			// 	gotInterface := d.GetInterface().(map[any]any)
			// 	t.Errorf("gotInterface:  %+v", gotInterface)
			// 	container := graph.Container{}
			// 	mapstructure.Decode(gotInterface, &container)
			// 	t.Errorf("container:  %+v", container)
			// }
		})
	}
}
