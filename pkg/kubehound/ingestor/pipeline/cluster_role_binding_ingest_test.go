//nolint:forcetypeassert
package pipeline

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockcollect "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	mockcache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClusterRoleBindingIngest_Pipeline(t *testing.T) {
	t.Parallel()

	crbi := &ClusterRoleBindingIngest{}

	ctx := context.Background()
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

	// Cache setup
	c := mockcache.NewCacheProvider(t)
	c.EXPECT().Get(ctx, cachekey.Identity("app-monitors-cluster", "")).Return(&cache.CacheResult{
		Value: nil,
		Err:   cache.ErrNoEntry,
	}).Once()
	c.EXPECT().Get(ctx, cachekey.Role("test-reader", "")).Return(&cache.CacheResult{
		Value: *oFakeClusterRole,
		Err:   nil,
	}).Twice()
	cw := mockcache.NewAsyncWriter(t)
	cw.EXPECT().Queue(ctx, mock.AnythingOfType("*cachekey.identityCacheKey"), mock.AnythingOfType("string")).Return(nil).Once()
	cw.EXPECT().Flush(ctx).Return(nil)
	cw.EXPECT().Close(ctx).Return(nil)
	c.EXPECT().BulkWriter(ctx, mock.AnythingOfType("cache.WriterOption")).Return(cw, nil)

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

	// Graph setup
	vtxInsert := map[string]any{
		"critical":     false,
		"isNamespaced": false,
		"name":         "app-monitors-cluster",
		"namespace":    "",
		"storeID":      storeID.Hex(),
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
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Identity"), c, mock.AnythingOfType("graphdb.WriterOption")).Return(gw, nil)

	psVtxInsert := map[string]any{
		"isNamespaced": false,
		"critical":     false,
		"name":         "test-reader::app-monitors-read",
		"namespace":    "",
		"role":         "test-reader",
		"roleBinding":  "app-monitors-read",
		"storeID":      psStoreID.Hex(),
		"team":         "test-team",
		"app":          "test-app",
		"service":      "test-service",
		"cluster":      "test-cluster",
		"runID":        testID.String(),
		"rules":        []interface{}{"API()::R(pods)::N()::V(get,list)", "API()::R(configmaps)::N()::V(get)", "API(apps)::R(statefulsets)::N()::V(get,list)"},
	}

	psgw := graphdb.NewAsyncVertexWriter(t)
	psgw.EXPECT().Queue(ctx, psVtxInsert).Return(nil).Once()
	psgw.EXPECT().Flush(ctx).Return(nil)
	psgw.EXPECT().Close(ctx).Return(nil)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.PermissionSet"), c, mock.AnythingOfType("graphdb.WriterOption")).Return(psgw, nil)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
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
