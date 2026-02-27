package pipeline

import (
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"testing"

	collector "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
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

	ctx := t.Context()

	client := collector.NewCollectorClient(t)
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	deps := &Dependencies{
		Collector: client,
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

	// Test graph writer mechanics
	gw := graphdb.NewAsyncVertexWriter(t)
	gw.EXPECT().Flush(ctx).Return(nil)
	gw.EXPECT().Close(ctx).Return(nil)

	vtx := &vertex.Node{}
	sdb.EXPECT().Reader().Return(&sql.DB{})
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Node"), mock.AnythingOfType("*sql.DB"), mock.AnythingOfType("graphdb.WriterOption")).Return(gw, nil)

	oi, err = CreateResources(ctx, deps, WithGraphWriter(vtx))
	assert.NoError(t, err)

	assert.NoError(t, oi.flushWriters(ctx))
	assert.NoError(t, oi.cleanupAll(ctx))
}

func TestIngestResources_FlushErrors(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := collector.NewCollectorClient(t)
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	deps := &Dependencies{
		Collector: client,
		GraphDB:   gdb,
		StoreDB:   sdb,
		Config: &config.KubehoundConfig{
			Dynamic: config.DynamicConfig{
				Cluster: config.DynamicClusterInfo{
					Name: "test-cluster",
				},
				RunID: config.NewRunID(),
			},
		},
	}

	// Set graph writer to fail on flush
	gw := graphdb.NewAsyncVertexWriter(t)
	gw.EXPECT().Flush(ctx).Return(errors.New("test error"))
	vtx := &vertex.Node{}
	sdb.EXPECT().Reader().Return(&sql.DB{})
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Node"), mock.AnythingOfType("*sql.DB"), mock.AnythingOfType("graphdb.WriterOption")).Return(gw, nil)

	oi, err := CreateResources(ctx, deps, WithGraphWriter(vtx))
	assert.NoError(t, err)

	assert.ErrorContains(t, oi.flushWriters(ctx), "test error")
}

func TestIngestResources_CloseErrors(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := collector.NewCollectorClient(t)
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	deps := &Dependencies{
		Collector: client,
		GraphDB:   gdb,
		StoreDB:   sdb,
		Config: &config.KubehoundConfig{
			Builder: config.BuilderConfig{
				Edge: config.EdgeBuilderConfig{},
			},
		},
	}

	// Set graph writer to fail on close
	gw := graphdb.NewAsyncVertexWriter(t)
	gw.EXPECT().Close(ctx).Return(errors.New("test error"))
	vtx := &vertex.Node{}
	sdb.EXPECT().Reader().Return(&sql.DB{})
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Node"), mock.AnythingOfType("*sql.DB"), mock.AnythingOfType("graphdb.WriterOption")).Return(gw, nil)

	oi, err := CreateResources(ctx, deps, WithGraphWriter(vtx))
	assert.NoError(t, err)

	assert.ErrorContains(t, oi.cleanupAll(ctx), "test error")
}

func TestIngestResources_CloseIdempotent(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	client := collector.NewCollectorClient(t)
	gdb := graphdb.NewProvider(t)
	sdb := storedb.NewProvider(t)

	deps := &Dependencies{
		Collector: client,
		GraphDB:   gdb,
		StoreDB:   sdb,
		Config: &config.KubehoundConfig{
			Builder: config.BuilderConfig{
				Edge: config.EdgeBuilderConfig{},
			},
		},
	}

	gw := graphdb.NewAsyncVertexWriter(t)
	gw.EXPECT().Close(ctx).Return(nil).Once()
	vtx := &vertex.Node{}
	sdb.EXPECT().Reader().Return(&sql.DB{})
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Node"), mock.AnythingOfType("*sql.DB"), mock.AnythingOfType("graphdb.WriterOption")).Return(gw, nil).Once()

	oi, err := CreateResources(ctx, deps, WithGraphWriter(vtx))
	assert.NoError(t, err)

	assert.NoError(t, oi.cleanupAll(ctx))
	assert.NoError(t, oi.cleanupAll(ctx))
}
