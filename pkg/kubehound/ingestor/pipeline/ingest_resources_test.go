package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	collector "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	cache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Shared function to load test objects across all ingests
func loadTestObject[T types.InputType](filename string) (T, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	var output T
	err = decoder.Decode(&output)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func TestIngestResources_Initializer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	client := collector.NewCollectorClient(t)
	c := cache.NewCacheProvider(t)
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
		GraphDB:   gdb,
		StoreDB:   sdb,
		Config: &config.KubehoundConfig{
			Builder: config.BuilderConfig{
				Edge: config.EdgeBuilderConfig{},
			},
		},
	}

	// Test default initialization
	oi, err := CreateResources(ctx, deps)
	assert.NoError(t, err)
	assert.IsType(t, &collector.CollectorClient{}, oi.collect)
	assert.IsType(t, &converter.StoreConverter{}, oi.storeConvert)
	assert.IsType(t, &converter.GraphConverter{}, oi.graphConvert)
	assert.Equal(t, 0, len(oi.cleanup))
	assert.Equal(t, 0, len(oi.flush))

	// Test cache writer mechanics
	cw := cache.NewAsyncWriter(t)
	cw.EXPECT().Flush(ctx).Return(nil)
	cw.EXPECT().Close(ctx).Return(nil)

	c.EXPECT().BulkWriter(ctx).Return(cw, nil)

	oi, err = CreateResources(ctx, deps, WithCacheWriter())
	assert.NoError(t, err)

	assert.NoError(t, oi.flushWriters(ctx))
	assert.NoError(t, oi.cleanupAll(ctx))

	// Test store writer mechanics
	sw := storedb.NewAsyncWriter(t)
	sw.EXPECT().Flush(ctx).Return(nil)
	sw.EXPECT().Close(ctx).Return(nil)

	collection := collections.Node{}
	sdb.EXPECT().BulkWriter(ctx, collection, mock.Anything).Return(sw, nil)

	oi, err = CreateResources(ctx, deps, WithStoreWriter(collection))
	assert.NoError(t, err)

	assert.NoError(t, oi.flushWriters(ctx))
	assert.NoError(t, oi.cleanupAll(ctx))

	// Test graph writer mechanics
	gw := graphdb.NewAsyncVertexWriter(t)
	gw.EXPECT().Flush(ctx).Return(nil)
	gw.EXPECT().Close(ctx).Return(nil)

	vtx := &vertex.Node{}
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Node"), c, mock.AnythingOfType("graphdb.WriterOption")).Return(gw, nil)

	oi, err = CreateResources(ctx, deps, WithGraphWriter(vtx))
	assert.NoError(t, err)

	assert.NoError(t, oi.flushWriters(ctx))
	assert.NoError(t, oi.cleanupAll(ctx))
}

func TestIngestResources_FlushErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := collector.NewCollectorClient(t)
	c := cache.NewCacheProvider(t)
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
		GraphDB:   gdb,
		StoreDB:   sdb,
		Config: &config.KubehoundConfig{
			Dynamic: config.DynamicConfig{
				Cluster: "test-clister",
				RunID:   config.NewRunID(),
			},
		},
	}

	// Set cache to succeed
	cw := cache.NewAsyncWriter(t)
	cw.EXPECT().Flush(ctx).Return(nil)
	c.EXPECT().BulkWriter(ctx).Return(cw, nil)

	// Set store to fail
	sw := storedb.NewAsyncWriter(t)
	sw.EXPECT().Flush(ctx).Return(errors.New("test error"))
	sdb.EXPECT().BulkWriter(ctx, mock.Anything, mock.Anything).Return(sw, nil)

	oi, err := CreateResources(ctx, deps, WithCacheWriter(), WithStoreWriter(collections.Node{}))
	assert.NoError(t, err)

	assert.ErrorContains(t, oi.flushWriters(ctx), "test error")
}

func TestIngestResources_CloseErrors(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := collector.NewCollectorClient(t)
	c := cache.NewCacheProvider(t)
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
		GraphDB:   gdb,
		StoreDB:   sdb,
		Config: &config.KubehoundConfig{
			Builder: config.BuilderConfig{
				Edge: config.EdgeBuilderConfig{},
			},
		},
	}

	// Set cache to succeed
	cw := cache.NewAsyncWriter(t)
	cw.EXPECT().Close(ctx).Return(nil)
	c.EXPECT().BulkWriter(ctx).Return(cw, nil)

	// Set store to fail
	sw := storedb.NewAsyncWriter(t)
	sw.EXPECT().Close(ctx).Return(errors.New("test error"))
	sdb.EXPECT().BulkWriter(ctx, mock.Anything, mock.Anything).Return(sw, nil)

	oi, err := CreateResources(ctx, deps, WithCacheWriter(), WithStoreWriter(collections.Node{}))
	assert.NoError(t, err)

	assert.ErrorContains(t, oi.cleanupAll(ctx), "test error")
}

func TestIngestResources_CloseIdempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := collector.NewCollectorClient(t)
	c := cache.NewCacheProvider(t)
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
		GraphDB:   gdb,
		StoreDB:   sdb,
		Config: &config.KubehoundConfig{
			Builder: config.BuilderConfig{
				Edge: config.EdgeBuilderConfig{},
			},
		},
	}

	cw := cache.NewAsyncWriter(t)
	cw.EXPECT().Close(ctx).Return(nil).Once()
	c.EXPECT().BulkWriter(ctx).Return(cw, nil).Once()

	sw := storedb.NewAsyncWriter(t)
	sw.EXPECT().Close(ctx).Return(nil).Once()
	sdb.EXPECT().BulkWriter(ctx, mock.Anything, mock.Anything).Return(sw, nil).Once()

	oi, err := CreateResources(ctx, deps, WithCacheWriter(), WithStoreWriter(collections.Node{}))
	assert.NoError(t, err)

	assert.NoError(t, oi.cleanupAll(ctx))
	assert.NoError(t, oi.cleanupAll(ctx))
}
