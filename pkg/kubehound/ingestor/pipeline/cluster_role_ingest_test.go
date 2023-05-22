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
	cri := &ClusterRoleIngest{}

	ctx := context.Background()
	fakeRole, err := loadTestObject[types.ClusterRoleType]("testdata/clusterrole.json")
	assert.NoError(t, err)

	client := mockcollect.NewCollectorClient(t)
	client.EXPECT().StreamClusterRoles(ctx, cri).
		RunAndReturn(func(ctx context.Context, i collector.ClusterRoleIngestor) error {
			// Fake the stream of a single cluster role from the collector client
			err := i.IngestClusterRole(ctx, fakeRole)
			if err != nil {
				return err
			}

			return i.Complete(ctx)
		})

	// Cache setup
	c := cache.NewCacheProvider(t)
	cw := cache.NewAsyncWriter(t)
	cwDone := make(chan struct{})
	cw.EXPECT().Queue(ctx, mock.AnythingOfType("*cache.roleCacheKey"), mock.AnythingOfType("string")).Return(nil).Once()
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
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("vertex.Role")).Return(gw, nil)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
		GraphDB:   gdb,
		StoreDB:   sdb,
	}

	// Initialize
	err = cri.Initialize(ctx, deps)
	assert.NoError(t, err)

	go func() {
		// Simulate a delayed flush completion
		time.Sleep(time.Second)
		close(cwDone)
		close(swDone)
		close(gwDone)
	}()

	// Run
	err = cri.Run(ctx)
	assert.NoError(t, err)

	// Close
	err = cri.Close(ctx)
	assert.NoError(t, err)
}
