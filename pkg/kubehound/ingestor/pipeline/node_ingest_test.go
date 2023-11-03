//nolint:forcetypeassert
package pipeline

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockcollect "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	mockcache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNodeIngest_Pipeline(t *testing.T) {
	t.Parallel()
	ni := &NodeIngest{}

	ctx := context.Background()
	fakeNode, err := loadTestObject[types.NodeType]("testdata/node.json")
	assert.NoError(t, err)

	client := mockcollect.NewCollectorClient(t)
	client.EXPECT().StreamNodes(ctx, ni).
		RunAndReturn(func(ctx context.Context, i collector.NodeIngestor) error {
			// Fake the stream of a single node from the collector client
			err := i.IngestNode(ctx, fakeNode)
			if err != nil {
				return err
			}

			return i.Complete(ctx)
		})

	// Cache setup
	c := mockcache.NewCacheProvider(t)
	cw := mockcache.NewAsyncWriter(t)
	cw.EXPECT().Queue(ctx, mock.AnythingOfType("*cachekey.nodeCacheKey"), mock.AnythingOfType("string")).Return(nil).Once()
	cw.EXPECT().Flush(ctx).Return(nil)
	cw.EXPECT().Close(ctx).Return(nil)
	c.EXPECT().BulkWriter(ctx).Return(cw, nil)
	c.EXPECT().Get(ctx, cachekey.Identity("system:node:node-1", "")).Return(&cache.CacheResult{
		Value: nil,
		Err:   cache.ErrNoEntry,
	}).Once()
	c.EXPECT().Get(ctx, cachekey.Identity("system:nodes", "")).Return(&cache.CacheResult{
		Value: store.ObjectID().Hex(),
		Err:   nil,
	}).Once()

	// Store setup
	sdb := storedb.NewProvider(t)
	sw := storedb.NewAsyncWriter(t)
	nodes := collections.Node{}
	storeID := store.ObjectID()
	sw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.Node")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.Node).Id = storeID

			return nil
		}).Once()
	sw.EXPECT().Flush(ctx).Return(nil)
	sw.EXPECT().Close(ctx).Return(nil)
	sdb.EXPECT().BulkWriter(ctx, nodes, mock.Anything).Return(sw, nil)

	// Graph setup
	vtxInsert := map[string]any{
		"compromised":  float64(0), // weird conversion to float by processor
		"critical":     false,
		"isNamespaced": false,
		"name":         "node-1",
		"namespace":    "",
		"storeID":      storeID.Hex(),
		"app":          "",
		"service":      "",
		"team":         "test-team",
		"cluster":      "test-cluster",
		"runID":        testID.String(),
	}
	gdb := graphdb.NewProvider(t)
	gw := graphdb.NewAsyncVertexWriter(t)
	gw.EXPECT().Queue(ctx, vtxInsert).Return(nil).Once()
	gw.EXPECT().Flush(ctx).Return(nil)
	gw.EXPECT().Close(ctx).Return(nil)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Node"), c, mock.AnythingOfType("graphdb.WriterOption")).Return(gw, nil)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
		GraphDB:   gdb,
		StoreDB:   sdb,
		Config: &config.KubehoundConfig{
			Builder: config.BuilderConfig{
				Edge: config.EdgeBuilderConfig{},
			},
			Dynamic: config.DynamicConfig{
				RunID:   testID,
				Cluster: "test-cluster",
			},
		},
	}

	// Initialize
	err = ni.Initialize(ctx, deps)
	assert.NoError(t, err)

	// Run
	err = ni.Run(ctx)
	assert.NoError(t, err)

	// Close
	err = ni.Close(ctx)
	assert.NoError(t, err)
}
