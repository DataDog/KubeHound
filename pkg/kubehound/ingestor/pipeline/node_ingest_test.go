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

func TestNodeIngest_Pipeline(t *testing.T) {
	ni := &NodeIngest{}

	ctx := context.Background()
	fakeNode, err := loadTestObject[types.NodeType]("testdata/node.json")
	assert.NoError(t, err)

	client := mockcollect.NewCollectorClient(t)
	client.EXPECT().StreamNodes(mock.Anything, mock.Anything, mock.Anything).
		RunAndReturn(func(ctx context.Context, np collector.NodeProcessor, c collector.Complete) error {
			// Fake the stream of a single node from the collector client
			err := np(ctx, &fakeNode)
			if err != nil {
				return err
			}

			return c(ctx)
		})

	// Cache setup
	c := cache.NewCacheProvider(t)
	cw := cache.NewAsyncWriter(t)
	cwDone := make(chan struct{})
	cw.EXPECT().Queue(mock.Anything, mock.AnythingOfType("*cache.nodeCacheKey"), mock.AnythingOfType("primitive.ObjectID")).Return(nil).Once()
	cw.EXPECT().Flush(mock.Anything).Return(cwDone, nil)
	cw.EXPECT().Close(mock.Anything).Return(nil)
	c.EXPECT().BulkWriter(mock.Anything).Return(cw, nil)

	// Store setup
	sdb := storedb.NewProvider(t)
	sw := storedb.NewAsyncWriter(t)
	nodes := collections.Node{}
	swDone := make(chan struct{})
	sw.EXPECT().Queue(mock.Anything, mock.AnythingOfType("*store.Node")).Return(nil).Once()
	sw.EXPECT().Flush(mock.Anything).Return(swDone, nil)
	sw.EXPECT().Close(mock.Anything).Return(nil)
	sdb.EXPECT().BulkWriter(mock.Anything, nodes).Return(sw, nil)

	// Graph setup
	gdb := graphdb.NewProvider(t)
	gw := graphdb.NewAsyncVertexWriter(t)
	gwDone := make(chan struct{})
	gw.EXPECT().Queue(mock.Anything, mock.AnythingOfType("*graph.Node")).Return(nil).Once()
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
	err = ni.Initialize(ctx, deps)
	assert.NoError(t, err)

	go func() {
		// Simulate a delayed flush completion
		time.Sleep(time.Second)
		close(cwDone)
		close(swDone)
		close(gwDone)
	}()

	// Run
	err = ni.Run(ctx)
	assert.NoError(t, err)

	// Close
	err = ni.Close(ctx)
	assert.NoError(t, err)
}
