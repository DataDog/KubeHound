//nolint:forcetypeassert
package pipeline

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockcollect "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	khstoredb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestClusterRoleBindingIngest_Pipeline(t *testing.T) {
	t.Parallel()

	crbi := &ClusterRoleBindingIngest{}

	ctx := t.Context()
	fakeCrb, err := loadTestObject[types.ClusterRoleBindingType]("testdata/clusterrolebinding.json")
	assert.NoError(t, err)

	fakeClusterRole, err := loadTestObject[types.ClusterRoleType]("testdata/clusterrole.json")
	assert.NoError(t, err)
	oFakeClusterRole, err := converter.NewStore(testConfig).ClusterRole(ctx, fakeClusterRole)
	assert.NoError(t, err)

	client := mockcollect.NewCollectorClient(t)
	client.EXPECT().StreamClusterRoleBindings(ctx, crbi).
		RunAndReturn(func(ctx context.Context, i collector.ClusterRoleBindingIngestor) error {
			// Fake the stream of a single cluster role binding from the collector client
			err := i.IngestClusterRoleBinding(ctx, fakeCrb)
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

	// Insert the cluster role that the converter will look up
	rulesJSON, _ := json.Marshal(oFakeClusterRole.Rules)
	_, err = db.ExecContext(ctx, "INSERT INTO roles (id, name, is_namespaced, namespace, rules, run_id, cluster_name) VALUES (?, ?, ?, ?, ?, ?, ?)",
		oFakeClusterRole.Id, oFakeClusterRole.Name, 0, oFakeClusterRole.Namespace, string(rulesJSON), testID.String(), "test-cluster")
	require.NoError(t, err)

	// Store setup
	sdb := storedb.NewProvider(t)
	storeID := store.ObjectID()
	psStoreID := store.ObjectID()

	sdb.EXPECT().Write(ctx, mock.AnythingOfType("*store.RoleBinding")).Return(nil).Once()

	sdb.EXPECT().Write(ctx, mock.AnythingOfType("*store.Identity")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.Identity).Id = storeID

			return nil
		}).Once()

	sdb.EXPECT().Write(ctx, mock.AnythingOfType("*store.PermissionSet")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.PermissionSet).Id = psStoreID

			return nil
		}).Once()

	// Graph setup
	vtxInsert := map[string]any{
		"critical":     false,
		"isNamespaced": false,
		"name":         "app-monitors-cluster",
		"namespace":    "",
		"storeID":      store.Hex(storeID),
		"type":         "ServiceAccount",
		"team":         "test-team",
		"app":          "test-app",
		"service":      "test-service",
		"cluster":      "test-cluster",
		"runID":        testID.String(),
	}
	gdb := graphdb.NewProvider(t)
	gw := graphdb.NewAsyncVertexWriter(t)
	gw.EXPECT().Queue(ctx, vtxInsert).Return(nil).Once()
	gw.EXPECT().Flush(ctx).Return(nil)
	gw.EXPECT().Close(ctx).Return(nil)
	sdb.EXPECT().Reader().Return(db)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Identity"), mock.AnythingOfType("*sql.DB"), mock.AnythingOfType("graphdb.WriterOption")).Return(gw, nil)

	psVtxInsert := map[string]any{
		"isNamespaced": false,
		"critical":     false,
		"name":         "test-reader::app-monitors-read",
		"namespace":    "",
		"role":         "test-reader",
		"roleBinding":  "app-monitors-read",
		"storeID":      store.Hex(psStoreID),
		"team":         "",
		"app":          "",
		"service":      "",
		"cluster":      "test-cluster",
		"runID":        testID.String(),
		"rules":        []interface{}{"API()::R(pods)::N()::V(get,list)", "API()::R(configmaps)::N()::V(get)", "API(apps)::R(statefulsets)::N()::V(get,list)"},
	}

	psgw := graphdb.NewAsyncVertexWriter(t)
	psgw.EXPECT().Queue(ctx, psVtxInsert).Return(nil).Once()
	psgw.EXPECT().Flush(ctx).Return(nil)
	psgw.EXPECT().Close(ctx).Return(nil)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.PermissionSet"), mock.AnythingOfType("*sql.DB"), mock.AnythingOfType("graphdb.WriterOption")).Return(psgw, nil)

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
	err = crbi.Initialize(ctx, deps)
	assert.NoError(t, err)

	// Run
	err = crbi.Run(ctx)
	assert.NoError(t, err)

	// Close
	err = crbi.Close(ctx)
	assert.NoError(t, err)
}
