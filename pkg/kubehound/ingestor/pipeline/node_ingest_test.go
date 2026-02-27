//nolint:forcetypeassert
package pipeline

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockcollect "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/libkube"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	khstoredb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestNodeIngest_Pipeline(t *testing.T) {
	t.Parallel()
	ni := &NodeIngest{}

	ctx := t.Context()
	fakeNode, err := loadTestObject[types.NodeType]("testdata/node.json")
	assert.NoError(t, err)

	client := mockcollect.NewCollectorClient(t)
	client.EXPECT().StreamNodes(ctx, ni).
		RunAndReturn(func(ctx context.Context, i collector.NodeIngestor) error {
			// Fake the stream of a single node from the collector client
			err := i.IngestNode(ctx, fakeNode)
			if err != nil {
				return err
			}

			return i.Complete(ctx)
		})

	// SQLite in-memory DB setup for converter lookups
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()
	err = khstoredb.InitSchema(db)
	require.NoError(t, err)

	// Reset the sync.Once in libkube so the DefaultNodeIdentity lookup runs fresh
	libkube.ResetOnce()

	// Insert the system:nodes identity that the node converter will look up
	nodesIdentityID := store.ObjectID()
	_, err = db.ExecContext(ctx, "INSERT INTO identities (id, name, is_namespaced, namespace, type, run_id, cluster_name) VALUES (?, ?, ?, ?, ?, ?, ?)",
		nodesIdentityID, "system:nodes", 0, "", "Group", testID.String(), "test-cluster")
	require.NoError(t, err)

	// Store setup
	sdb := storedb.NewProvider(t)
	storeID := store.ObjectID()
	sdb.EXPECT().Write(ctx, mock.AnythingOfType("*store.Node")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.Node).Id = storeID

			return nil
		}).Once()

	// Graph setup
	vtxInsert := map[string]any{
		"compromised":  float64(0), // weird conversion to float by processor
		"critical":     false,
		"isNamespaced": false,
		"name":         "node-1",
		"namespace":    "",
		"storeID":      store.Hex(storeID),
		"app":          "",
		"service":      "",
		"team":         "test-team",
		"cluster":      "test-cluster",
		"runID":        testID.String(),
	}
	gdb := graphdb.NewProvider(t)
	gw := graphdb.NewAsyncVertexWriter(t)
	gw.EXPECT().Queue(ctx, vtxInsert).Return(nil).Once()
	gw.EXPECT().Flush(ctx).Return(nil)
	gw.EXPECT().Close(ctx).Return(nil)
	sdb.EXPECT().Reader().Return(db)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Node"), mock.AnythingOfType("*sql.DB"), mock.AnythingOfType("graphdb.WriterOption")).Return(gw, nil)

	deps := &Dependencies{
		Collector: client,
		GraphDB:   gdb,
		StoreDB:   sdb,
		Config: &config.KubehoundConfig{
			Builder: config.BuilderConfig{
				Edge: config.EdgeBuilderConfig{},
			},
			Dynamic: config.DynamicConfig{
				RunID: testID,
				Cluster: config.DynamicClusterInfo{
					Name: "test-cluster",
				},
			},
		},
	}

	// Initialize
	err = ni.Initialize(ctx, deps)
	assert.NoError(t, err)

	// Run
	err = ni.Run(ctx)
	assert.NoError(t, err)

	// Close
	err = ni.Close(ctx)
	assert.NoError(t, err)
}
