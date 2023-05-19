package pipeline

import (
	"context"
	"errors"
	"testing"

	collector "github.com/DataDog/KubeHound/pkg/collector/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	cache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestObjectIngest_Initializer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	oi := &BaseObjectIngest{}

	client := collector.NewCollectorClient(t)
	c := cache.NewCacheProvider(t)
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
		GraphDB:   gdb,
		StoreDB:   sdb,
	}

	// Test default initialization
	err := oi.baseInitialize(ctx, deps)
	assert.NoError(t, err)
	assert.IsType(t, &converter.StoreConverter{}, oi.opts.storeConvert)
	assert.IsType(t, &converter.GraphConverter{}, oi.opts.graphConvert)
	assert.Equal(t, 0, len(oi.opts.cleanup))
	assert.Equal(t, 0, len(oi.opts.flush))

	// Test cache writer mechanics
	oi = &BaseObjectIngest{}
	cw := cache.NewAsyncWriter(t)
	cwDone := make(chan struct{})
	cw.EXPECT().Flush(mock.Anything).Return(cwDone, nil)
	cw.EXPECT().Close(mock.Anything).Return(nil)

	c.EXPECT().BulkWriter(mock.Anything).Return(cw, nil)

	err = oi.baseInitialize(ctx, deps, WithCacheWriter())
	assert.NoError(t, err)

	close(cwDone)
	assert.NoError(t, oi.flushWriters(ctx))
	assert.NoError(t, oi.cleanup(ctx))

	// Test store writer mechanics
	oi = &BaseObjectIngest{}
	sw := storedb.NewAsyncWriter(t)
	swDone := make(chan struct{})
	sw.EXPECT().Flush(mock.Anything).Return(swDone, nil)
	sw.EXPECT().Close(mock.Anything).Return(nil)

	collection := collections.Node{}
	sdb.EXPECT().BulkWriter(mock.Anything, collection).Return(sw, nil)

	err = oi.baseInitialize(ctx, deps, WithStoreWriter(collection))
	assert.NoError(t, err)

	close(swDone)
	assert.NoError(t, oi.flushWriters(ctx))
	assert.NoError(t, oi.cleanup(ctx))

	// Test graph writer mechanics
	oi = &BaseObjectIngest{}
	gw := graphdb.NewAsyncVertexWriter(t)
	gwDone := make(chan struct{})
	gw.EXPECT().Flush(mock.Anything).Return(gwDone, nil)
	gw.EXPECT().Close(mock.Anything).Return(nil)

	vtx := vertex.Node{}
	gdb.EXPECT().VertexWriter(mock.Anything, mock.AnythingOfType("vertex.VertexTraversal")).Return(gw, nil)

	err = oi.baseInitialize(ctx, deps, WithGraphWriter(vtx))
	assert.NoError(t, err)

	close(gwDone)
	assert.NoError(t, oi.flushWriters(ctx))
	assert.NoError(t, oi.cleanup(ctx))
}

func TestObjectIngest_FlushErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	oi := &BaseObjectIngest{}

	client := collector.NewCollectorClient(t)
	c := cache.NewCacheProvider(t)
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
		GraphDB:   gdb,
		StoreDB:   sdb,
	}

	// Set cache to succeed
	cw := cache.NewAsyncWriter(t)
	cwDone := make(chan struct{})
	cw.EXPECT().Flush(mock.Anything).Return(cwDone, nil)
	c.EXPECT().BulkWriter(mock.Anything).Return(cw, nil)

	// Set store to fail
	sw := storedb.NewAsyncWriter(t)
	swDone := make(chan struct{})
	sw.EXPECT().Flush(mock.Anything).Return(swDone, errors.New("test error"))
	sdb.EXPECT().BulkWriter(mock.Anything, mock.Anything).Return(sw, nil)

	err := oi.baseInitialize(ctx, deps, WithCacheWriter(), WithStoreWriter(collections.Node{}))
	assert.NoError(t, err)

	close(cwDone)
	close(swDone)
	assert.ErrorContains(t, oi.flushWriters(ctx), "test error")
}

func TestObjectIngest_CloseErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	oi := &BaseObjectIngest{}

	client := collector.NewCollectorClient(t)
	c := cache.NewCacheProvider(t)
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
		GraphDB:   gdb,
		StoreDB:   sdb,
	}

	// Set cache to succeed
	cw := cache.NewAsyncWriter(t)
	cw.EXPECT().Close(mock.Anything).Return(nil)
	c.EXPECT().BulkWriter(mock.Anything).Return(cw, nil)

	// Set store to fail
	sw := storedb.NewAsyncWriter(t)
	sw.EXPECT().Close(mock.Anything).Return(errors.New("test error"))
	sdb.EXPECT().BulkWriter(mock.Anything, mock.Anything).Return(sw, nil)

	err := oi.baseInitialize(ctx, deps, WithCacheWriter(), WithStoreWriter(collections.Node{}))
	assert.NoError(t, err)

	assert.ErrorContains(t, oi.cleanup(ctx), "test error")
}
