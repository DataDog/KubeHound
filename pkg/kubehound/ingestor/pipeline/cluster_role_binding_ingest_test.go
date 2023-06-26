package pipeline

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockcollect "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	cache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClusterRoleBindingIngest_Pipeline(t *testing.T) {
	crbi := &ClusterRoleBindingIngest{}

	ctx := context.Background()
	fakeCrb, err := loadTestObject[types.ClusterRoleBindingType]("testdata/clusterrolebinding.json")
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
	c := cache.NewCacheProvider(t)
	c.EXPECT().Get(ctx, mock.AnythingOfType("*cachekey.roleCacheKey")).Return(store.ObjectID().Hex(), nil)

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
	storeId := store.ObjectID()
	isw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.Identity")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.Identity).Id = storeId
			return nil
		}).Once()
	isw.EXPECT().Flush(ctx).Return(nil)
	isw.EXPECT().Close(ctx).Return(nil)
	sdb.EXPECT().BulkWriter(ctx, identities, mock.Anything).Return(isw, nil)

	// Graph setup
	vtxInsert := map[string]any{
		"isNamespaced": false,
		"name":         "app-monitors",
		"namespace":    "test-app",
		"storeID":      storeId.Hex(),
		"type":         "ServiceAccount",
	}
	gdb := graphdb.NewProvider(t)
	gw := graphdb.NewAsyncVertexWriter(t)
	gw.EXPECT().Queue(ctx, vtxInsert).Return(nil).Once()
	gw.EXPECT().Flush(ctx).Return(nil)
	gw.EXPECT().Close(ctx).Return(nil)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("vertex.Identity"), mock.AnythingOfType("graphdb.WriterOption")).Return(gw, nil)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
		GraphDB:   gdb,
		StoreDB:   sdb,
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
