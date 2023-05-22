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

func TestRoleIngest_Pipeline(t *testing.T) {
	ri := &RoleIngest{}

	ctx := context.Background()
	fakeRole, err := loadTestObject[types.RoleType]("testdata/role.json")
	assert.NoError(t, err)

	client := mockcollect.NewCollectorClient(t)
	client.EXPECT().StreamRoles(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, process collector.RoleProcessor, complete collector.Complete) error {
			// Fake the stream of a single role from the collector client
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
	cw.EXPECT().Queue(mock.Anything, mock.AnythingOfType("*cache.roleCacheKey"), mock.AnythingOfType("primitive.ObjectID")).Return(nil).Once()
	cw.EXPECT().Flush(mock.Anything).Return(cwDone, nil)
	cw.EXPECT().Close(mock.Anything).Return(nil)
	c.EXPECT().BulkWriter(mock.Anything).Return(cw, nil)

	// Store setup
	sdb := storedb.NewProvider(t)
	sw := storedb.NewAsyncWriter(t)
	roles := collections.Role{}
	swDone := make(chan struct{})
	sw.EXPECT().Queue(mock.Anything, mock.AnythingOfType("*store.Role")).Return(nil).Once()
	sw.EXPECT().Flush(mock.Anything).Return(swDone, nil)
	sw.EXPECT().Close(mock.Anything).Return(nil)
	sdb.EXPECT().BulkWriter(mock.Anything, roles).Return(sw, nil)

	// Graph setup
	gdb := graphdb.NewProvider(t)
	gw := graphdb.NewAsyncVertexWriter(t)
	gwDone := make(chan struct{})
	gw.EXPECT().Queue(mock.Anything, mock.AnythingOfType("*graph.Role")).Return(nil).Once()
	gw.EXPECT().Flush(mock.Anything).Return(gwDone, nil)
	gw.EXPECT().Close(mock.Anything).Return(nil)
	gdb.EXPECT().VertexWriter(mock.Anything, mock.AnythingOfType("vertex.VertexTraversal")).Return(gw, nil)

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
