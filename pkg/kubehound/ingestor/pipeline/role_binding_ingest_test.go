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
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func TestRoleBindingIngest_Pipeline(t *testing.T) {
	t.Parallel()

	ri := &RoleBindingIngest{}
	ctx := t.Context()
	fakeRb, err := loadTestObject[types.RoleBindingType]("testdata/rolebinding.json")
	assert.NoError(t, err)

	fakeRole, err := loadTestObject[types.RoleType]("testdata/role.json")
	assert.NoError(t, err)
	oFakeRole, err := converter.NewStore(testConfig).Role(ctx, fakeRole)
	assert.NoError(t, err)

	client := mockcollect.NewCollectorClient(t)
	client.EXPECT().StreamRoleBindings(ctx, ri).
		RunAndReturn(func(ctx context.Context, i collector.RoleBindingIngestor) error {
			// Fake the stream of a single role binding from the collector client
			err := i.IngestRoleBinding(ctx, fakeRb)
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

	// Insert the role that the converter will look up
	rulesJSON, _ := json.Marshal(oFakeRole.Rules)
	_, err = db.ExecContext(ctx, "INSERT INTO roles (id, name, is_namespaced, namespace, rules, run_id, cluster_name) VALUES (?, ?, ?, ?, ?, ?, ?)",
		oFakeRole.Id, oFakeRole.Name, 1, oFakeRole.Namespace, string(rulesJSON), testID.String(), "test-cluster")
	require.NoError(t, err)

	// Store setup -  rolebindings
	sdb := storedb.NewProvider(t)
	rsw := storedb.NewAsyncWriter(t)
	crbs := collections.RoleBinding{}
	rsw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.RoleBinding")).Return(nil).Once()
	rsw.EXPECT().Flush(ctx).Return(nil)
	rsw.EXPECT().Close(ctx).Return(nil)
	sdb.EXPECT().BulkWriter(ctx, crbs, mock.Anything).Return(rsw, nil)

	// Store setup -  identities
	isw := storedb.NewAsyncWriter(t)

	// Store setup -  permissionsets
	pssw := storedb.NewAsyncWriter(t)
	psbs := collections.PermissionSet{}
	psStoreID := store.ObjectID()
	pssw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.PermissionSet")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.PermissionSet).Id = psStoreID

			return nil
		}).Once()
	pssw.EXPECT().Flush(ctx).Return(nil)
	pssw.EXPECT().Close(ctx).Return(nil)
	sdb.EXPECT().BulkWriter(ctx, psbs, mock.Anything).Return(pssw, nil)

	identities := collections.Identity{}
	storeID := store.ObjectID()
	isw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.Identity")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.Identity).Id = storeID

			return nil
		}).Once()
	isw.EXPECT().Flush(ctx).Return(nil)
	isw.EXPECT().Close(ctx).Return(nil)
	sdb.EXPECT().BulkWriter(ctx, identities, mock.Anything).Return(isw, nil)

	// Graph setup
	vtxInsert := map[string]any{
		"isNamespaced": true,
		"critical":     false,
		"name":         "app-monitors",
		"namespace":    "test-app",
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
		"isNamespaced": true,
		"critical":     false,
		"name":         "test-reader::app-monitors-read",
		"role":         "test-reader",
		"roleBinding":  "app-monitors-read",
		"namespace":    "test-app",
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
	err = ri.Initialize(ctx, deps)
	assert.NoError(t, err)

	// Run
	err = ri.Run(ctx)
	assert.NoError(t, err)

	// Close
	err = ri.Close(ctx)
	assert.NoError(t, err)
}
