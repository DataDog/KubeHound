package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockcollect "github.com/DataDog/KubeHound/pkg/collector/mocks"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	cache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRoleBindingIngest_Pipeline(t *testing.T) {
	ri := &RoleBindingIngest{}

	ctx := context.Background()
	fakeRb, err := loadTestObject[types.RoleBindingType]("testdata/rolebinding.json")
	assert.NoError(t, err)

	client := mockcollect.NewCollectorClient(t)
	client.EXPECT().StreamRoleBindings(ctx, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, process collector.RoleBindingProcessor, complete collector.Complete) error {
			// Fake the stream of a single role binding from the collector client
			err := process(ctx, fakeRb)
			if err != nil {
				return err
			}

			return complete(ctx)
		})

	// Cache setup
	c := cache.NewCacheProvider(t)
	c.EXPECT().Get(ctx, mock.AnythingOfType("*cache.roleCacheKey")).Return(store.ObjectID().Hex(), nil)

	// Store setup -  rolebindings
	sdb := storedb.NewProvider(t)
	rsw := storedb.NewAsyncWriter(t)
	crbs := collections.RoleBinding{}
	rswDone := make(chan struct{})
	rsw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.RoleBinding")).Return(nil).Once()
	rsw.EXPECT().Flush(ctx).Return(rswDone, nil)
	rsw.EXPECT().Close(ctx).Return(nil)
	sdb.EXPECT().BulkWriter(ctx, crbs).Return(rsw, nil)

	// Store setup -  identities
	isw := storedb.NewAsyncWriter(t)
	identities := collections.Identity{}
	iswDone := make(chan struct{})
	isw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.Identity")).Return(nil).Once()
	isw.EXPECT().Flush(ctx).Return(iswDone, nil)
	isw.EXPECT().Close(ctx).Return(nil)
	sdb.EXPECT().BulkWriter(ctx, identities).Return(isw, nil)

	// Graph setup
	gdb := graphdb.NewProvider(t)
	gw := graphdb.NewAsyncVertexWriter(t)
	gwDone := make(chan struct{})
	gw.EXPECT().Queue(ctx, mock.AnythingOfType("*graph.Identity")).Return(nil).Once()
	gw.EXPECT().Flush(ctx).Return(gwDone, nil)
	gw.EXPECT().Close(ctx).Return(nil)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("vertex.Identity")).Return(gw, nil)

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
		close(rswDone)
		close(iswDone)
		close(gwDone)
	}()

	// Run
	err = ri.Run(ctx)
	assert.NoError(t, err)

	// Close
	err = ri.Close(ctx)
	assert.NoError(t, err)
}
