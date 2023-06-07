package vertex

import (
	"testing"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/utils"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
	"github.com/mitchellh/mapstructure"
	"github.com/stretchr/testify/assert"
)

func TestContainer_Traversal(t *testing.T) {
	tests := []struct {
		name string
		v    Container
		want VertexTraversal
	}{
		{
			name: "Add containers in JanusGraph",
			v:    Container{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dbHost := "ws://localhost:8182/gremlin"
			driver, err := gremlingo.NewDriverRemoteConnection(dbHost)
			assert.NoError(t, err)

			g := gremlingo.Traversal_().WithRemote(driver)
			// tx := traversal.Tx()
			// g, err := tx.Begin()
			assert.NoError(t, err)

			v := Container{}
			// We set the values to all field with non default values so we are sure all are correctly propagated.
			insert, err := utils.StructToMap(graph.Container{
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
			})
			assert.NoError(t, err)

			vertexTraversal := v.Traversal()
			inserts := []TraversalInput{insert}
			t.Errorf("inserts: %v", inserts)
			_ = vertexTraversal(g, inserts)

			// err = tx.Commit()
			// assert.NoError(t, err)
			// err = tx.Close()
			// assert.NoError(t, err)

			driver, err = gremlingo.NewDriverRemoteConnection(dbHost)
			assert.NoError(t, err)

			g = gremlingo.Traversal_().WithRemote(driver)
			// tx = traversal.Tx()
			// g, err = tx.Begin()

			assert.NoError(t, err)
			test := g.V().HasLabel(v.Label()).ValueMap()
			res, err := test.Traversal.GetResultSet()
			assert.NoError(t, err)
			data, err := res.All()
			for _, d := range data {
				thefuck := d.GetInterface()
				t.Errorf("thefuck:  %+v", thefuck)
				container := graph.Container{}
				mapstructure.Decode(thefuck, &container)
				t.Errorf("container:  %+v", container)
			}
		})
	}
}
