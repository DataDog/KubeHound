package vertex

import (
	"testing"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
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

			traversal := gremlingo.Traversal_().WithRemote(driver)
			tx := traversal.Tx()
			g, err := tx.Begin()
			assert.NoError(t, err)

			v := Container{}
			// We set the values to all field with non default values so we are sure all are correctly propagated.
			insert := graph.Container{
				StoreId:      "test id",
				Name:         "test name",
				Image:        "image",
				Command:      []string{"/usr/bin/sleep"},
				Args:         []string{"600"},
				Capabilities: []string{"CAPABILITY"},
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
			}
			vertexTraversal := v.Traversal()
			_ = vertexTraversal(g, []TraversalInput{insert})

			err = tx.Commit()
			assert.NoError(t, err)
			// err = tx.Close()
			// assert.NoError(t, err)

			driver, err = gremlingo.NewDriverRemoteConnection(dbHost)
			assert.NoError(t, err)

			traversal = gremlingo.Traversal_().WithRemote(driver)
			tx = traversal.Tx()
			g, err = tx.Begin()

			assert.NoError(t, err)
			test := g.V().HasLabel(v.Label()).Properties()
			res, err := test.Traversal.GetResultSet()
			assert.NoError(t, err)
			data, err := res.All()

			for _, d := range data {
				vertex, err := d.GetVertex()
				assert.NoError(t, err)
				t.Errorf("vertex:  %+v", vertex)
				t.Errorf("vertex elem: %+v", vertex.Element)
				t.Errorf("vertex id: %+v", vertex.Id)
				t.Errorf("vertex label: %+v", vertex.Label)
			}

			err = tx.Commit()
			assert.NoError(t, err)
			// err = tx.Close()
			// assert.NoError(t, err)
			t.Errorf("%+v", res)
			t.Errorf("%+v", data)
		})
	}
}
