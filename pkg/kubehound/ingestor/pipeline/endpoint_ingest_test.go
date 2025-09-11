//nolint:forcetypeassert
package pipeline

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockcollect "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	mockcache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestEndpointSlice_Pipeline(t *testing.T) {
	t.Parallel()

	ei := &EndpointIngest{}

	ctx := context.Background()
	fakeEp, err := loadTestObject[types.EndpointType]("testdata/endpointslice.json")
	assert.NoError(t, err)

	client := mockcollect.NewCollectorClient(t)
	client.EXPECT().StreamEndpoints(ctx, ei).
		RunAndReturn(func(ctx context.Context, i collector.EndpointIngestor) error {
			// Fake the stream of a single role binding from the collector client
			err := i.IngestEndpoint(ctx, fakeEp)
			if err != nil {
				return err
			}

			return i.Complete(ctx)
		})

	// Cache setup
	c := mockcache.NewCacheProvider(t)
	cw := mockcache.NewAsyncWriter(t)
	cw.EXPECT().Queue(ctx, mock.AnythingOfType("*cachekey.endpointCacheKey"), true).Return(nil).Times(2)
	cw.EXPECT().Flush(ctx).Return(nil)
	cw.EXPECT().Close(ctx).Return(nil)
	c.EXPECT().BulkWriter(ctx, mock.AnythingOfType("cache.WriterOption")).Return(cw, nil)

	// Store setup
	sdb := storedb.NewProvider(t)
	sw := storedb.NewAsyncWriter(t)
	endpoints := collections.Endpoint{}
	storeID := store.ObjectID()
	sw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.Endpoint")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.Endpoint).Id = storeID

			return nil
		}).Times(2)

	sw.EXPECT().Flush(ctx).Return(nil)
	sw.EXPECT().Close(ctx).Return(nil)
	sdb.EXPECT().BulkWriter(ctx, endpoints, mock.Anything).Return(sw, nil)

	// Graph setup
	vtx1 := map[string]interface{}{
		"addressType":     "IPv4",
		"addresses":       []interface{}{"10.1.1.1"},
		"app":             "cassandra",
		"compromised":     float64(0),
		"exposure":        float64(shared.EndpointExposureExternal),
		"isNamespaced":    true,
		"name":            "cassandra-temporal-dev-kmwfp::TCP::cql",
		"namespace":       "cassandra-temporal-dev",
		"port":            float64(9042),
		"portName":        "cql",
		"protocol":        "TCP",
		"service":         "cassandra",
		"serviceDns":      "cassandra-temporal-dev.cassandra-temporal-dev",
		"serviceEndpoint": "cassandra-temporal-dev",
		"storeID":         storeID.Hex(),
		"team":            "workflow-engine",
		"cluster":         "test-cluster",
		"runID":           testID.String(),
	}
	vtx2 := map[string]interface{}{
		"addressType":     "IPv4",
		"addresses":       []interface{}{"10.1.1.1"},
		"app":             "cassandra",
		"compromised":     float64(0),
		"exposure":        float64(shared.EndpointExposureExternal),
		"isNamespaced":    true,
		"name":            "cassandra-temporal-dev-kmwfp::TCP::jmx",
		"namespace":       "cassandra-temporal-dev",
		"port":            float64(7199),
		"portName":        "jmx",
		"protocol":        "TCP",
		"service":         "cassandra",
		"serviceDns":      "cassandra-temporal-dev.cassandra-temporal-dev",
		"serviceEndpoint": "cassandra-temporal-dev",
		"storeID":         storeID.Hex(),
		"team":            "workflow-engine",
		"cluster":         "test-cluster",
		"runID":           testID.String(),
	}

	gdb := graphdb.NewProvider(t)
	gw := graphdb.NewAsyncVertexWriter(t)
	gw.EXPECT().Queue(ctx, vtx1).Return(nil).Once()
	gw.EXPECT().Queue(ctx, vtx2).Return(nil).Once()
	gw.EXPECT().Flush(ctx).Return(nil)
	gw.EXPECT().Close(ctx).Return(nil)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Endpoint"), c, mock.AnythingOfType("graphdb.WriterOption")).Return(gw, nil)

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
				RunID: testID,
				Cluster: config.DynamicClusterInfo{
					Name: "test-cluster",
				},
			},
		},
	}

	// Initialize
	err = ei.Initialize(ctx, deps)
	assert.NoError(t, err)

	// Run
	err = ei.Run(ctx)
	assert.NoError(t, err)

	// Close
	err = ei.Close(ctx)
	assert.NoError(t, err)
}
