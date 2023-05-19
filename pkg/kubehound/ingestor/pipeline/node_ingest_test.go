package pipeline

import (
	"context"
	"testing"

	collector "github.com/DataDog/KubeHound/pkg/collector/mocks"
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
	client := collector.NewCollectorClient(t)

	// Cache setup
	c := cache.NewCacheProvider(t)
	cw := cache.NewAsyncWriter(t)
	cwDone := make(chan struct{})
	cw.EXPECT().Flush(mock.Anything).Return(cwDone, nil)
	cw.EXPECT().Close(mock.Anything).Return(nil)
	c.EXPECT().BulkWriter(mock.Anything).Return(cw, nil)

	// Store setup
	sdb := storedb.NewProvider(t)
	sw := storedb.NewAsyncWriter(t)
	nodes := collections.Node{}
	swDone := make(chan struct{})
	sw.EXPECT().Flush(mock.Anything).Return(swDone, nil)
	sw.EXPECT().Close(mock.Anything).Return(nil)
	sdb.EXPECT().BulkWriter(mock.Anything, nodes).Return(sw, nil)

	// Graph setup
	gdb := graphdb.NewProvider(t)
	gw := graphdb.NewAsyncVertexWriter(t)
	gwDone := make(chan struct{})
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
	err := ni.Initialize(ctx, deps)
	assert.NoError(t, err)

	// Run

	// Close
	err = ni.Close(ctx)
	assert.NoError(t, err)
}
