package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockcollect "github.com/DataDog/KubeHound/pkg/collector/mocks"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	cache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestClusterRoleIngest_Pipeline(t *testing.T) {
	ri := &ClusterRoleIngest{}

	ctx := context.Background()
	fakeRole, err := loadTestObject[types.ClusterRoleType]("testdata/clusterrole.json")
	assert.NoError(t, err)

	client := mockcollect.NewCollectorClient(t)
	client.EXPECT().StreamClusterRoles(ctx, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, process collector.ClusterRoleProcessor, complete collector.Complete) error {
			// Fake the stream of a single cluster role from the collector client
			err := process(ctx, &fakeRole)
			if err != nil {
				return err
			}

			return complete(ctx)
		})

	// Cache setup
	c := cache.NewCacheProvider(t)
	cw := cache.NewAsyncWriter(t)
	cwDone := make(chan struct{})
	cw.EXPECT().Queue(ctx, mock.AnythingOfType("*cache.roleCacheKey"), mock.AnythingOfType("primitive.ObjectID")).Return(nil).Once()
	cw.EXPECT().Flush(ctx).Return(cwDone, nil)
	cw.EXPECT().Close(ctx).Return(nil)
	c.EXPECT().BulkWriter(ctx).Return(cw, nil)

	// Store setup
	sdb := storedb.NewProvider(t)
	sw := storedb.NewAsyncWriter(t)
	roles := collections.Role{}
	swDone := make(chan struct{})
	sw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.Role")).Return(nil).Once()
	sw.EXPECT().Flush(ctx).Return(swDone, nil)
	sw.EXPECT().Close(ctx).Return(nil)
	sdb.EXPECT().BulkWriter(ctx, roles).Return(sw, nil)

	// Graph setup
	gdb := graphdb.NewProvider(t)
	gw := graphdb.NewAsyncVertexWriter(t)
	gwDone := make(chan struct{})
	gw.EXPECT().Queue(ctx, mock.AnythingOfType("*graph.Role")).Return(nil).Once()
	gw.EXPECT().Flush(ctx).Return(gwDone, nil)
	gw.EXPECT().Close(ctx).Return(nil)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("vertex.VertexTraversal")).Return(gw, nil)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
		GraphDB:   gdb,
		StoreDB:   sdb,
	}

	// Initialize
	err = ri.Initialize(ctx, deps)
	assert.NoError(t, err)

	go func() {
		// Simulate a delayed flush completion
		time.Sleep(time.Second)
		close(cwDone)
		close(swDone)
		close(gwDone)
	}()

	// Run
	err = ri.Run(ctx)
	assert.NoError(t, err)

	// Close
	err = ri.Close(ctx)
	assert.NoError(t, err)
}
