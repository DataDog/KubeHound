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
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	mockcache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPodIngest_Pipeline(t *testing.T) {
	t.Parallel()

	pi := &PodIngest{}
	ctx := context.Background()
	fakePod, err := loadTestObject[types.PodType]("testdata/pod.json")
	assert.NoError(t, err)

	client := mockcollect.NewCollectorClient(t)
	client.EXPECT().StreamPods(ctx, pi).
		RunAndReturn(func(ctx context.Context, i collector.PodIngestor) error {
			// Fake the stream of a single pod from the collector client
			err := i.IngestPod(ctx, fakePod)
			if err != nil {
				return err
			}

			return i.Complete(ctx)
		})

	// Cache setup
	c := mockcache.NewCacheProvider(t)
	cw := mockcache.NewAsyncWriter(t)
	cw.EXPECT().Queue(ctx, mock.AnythingOfType("*cachekey.containerCacheKey"), mock.AnythingOfType("string")).Return(nil).Once()

	cw.EXPECT().Flush(ctx).Return(nil)
	cw.EXPECT().Close(ctx).Return(nil)
	c.EXPECT().BulkWriter(ctx, mock.AnythingOfType("cache.WriterOption")).Return(cw, nil)
	c.EXPECT().Get(ctx, mock.AnythingOfType("*cachekey.nodeCacheKey")).Return(&cache.CacheResult{
		Value: store.ObjectID().Hex(),
		Err:   nil,
	})
	c.EXPECT().Get(ctx, mock.AnythingOfType("*cachekey.identityCacheKey")).Return(&cache.CacheResult{
		Value: store.ObjectID().Hex(),
		Err:   nil,
	})
	c.EXPECT().Get(ctx, mock.AnythingOfType("*cachekey.endpointCacheKey")).Return(&cache.CacheResult{
		Value: nil,
		Err:   cache.ErrNoEntry,
	})

	// Store setup - pods
	sdb := storedb.NewProvider(t)
	psw := storedb.NewAsyncWriter(t)
	pods := collections.Pod{}
	pid := store.ObjectID()
	psw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.Pod")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.Pod).Id = pid

			return nil
		}).Once()
	psw.EXPECT().Flush(ctx).Return(nil)
	psw.EXPECT().Close(ctx).Return(nil)

	// Store setup - containers
	csw := storedb.NewAsyncWriter(t)
	containers := collections.Container{}
	cid := store.ObjectID()
	csw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.Container")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.Container).Id = cid

			return nil
		}).Once()
	csw.EXPECT().Flush(ctx).Return(nil)
	csw.EXPECT().Close(ctx).Return(nil)

	// Store setup - volumes
	vsw := storedb.NewAsyncWriter(t)
	volumes := collections.Volume{}
	vid := store.ObjectID()
	vsw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.Volume")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.Volume).Id = vid

			return nil
		}).Once()

	vsw.EXPECT().Flush(ctx).Return(nil)
	vsw.EXPECT().Close(ctx).Return(nil)

	// Store setup - endpoint
	esw := storedb.NewAsyncWriter(t)
	endpoints := collections.Endpoint{}
	eid := store.ObjectID()
	esw.EXPECT().Queue(ctx, mock.AnythingOfType("*store.Endpoint")).
		RunAndReturn(func(ctx context.Context, i any) error {
			i.(*store.Endpoint).Id = eid

			return nil
		}).Once()

	esw.EXPECT().Flush(ctx).Return(nil)
	esw.EXPECT().Close(ctx).Return(nil)

	sdb.EXPECT().BulkWriter(ctx, pods, mock.Anything).Return(psw, nil)
	sdb.EXPECT().BulkWriter(ctx, containers, mock.Anything).Return(csw, nil)
	sdb.EXPECT().BulkWriter(ctx, volumes, mock.Anything).Return(vsw, nil)
	sdb.EXPECT().BulkWriter(ctx, endpoints, mock.Anything).Return(esw, nil)

	// Graph setup - pods
	pv := map[string]any{
		"compromised":           float64(0),
		"critical":              false,
		"isNamespaced":          true,
		"name":                  "app-monitors-client-78cb6d7899-j2rjp",
		"namespace":             "test-app",
		"node":                  "test-node.ec2.internal",
		"serviceAccount":        "app-monitors",
		"shareProcessNamespace": false,
		"storeID":               pid.Hex(),
		"team":                  "test-team",
		"app":                   "test-app",
		"service":               "test-service",
	}

	gdb := graphdb.NewProvider(t)
	pgw := graphdb.NewAsyncVertexWriter(t)
	pgw.EXPECT().Queue(ctx, pv).Return(nil).Once()
	pgw.EXPECT().Flush(ctx).Return(nil)
	pgw.EXPECT().Close(ctx).Return(nil)

	// Graph setup - containers
	cv := map[string]any{
		"args":         any(nil),
		"isNamespaced": true,
		"namespace":    "test-app",
		"capabilities": []any{},
		"command":      any(nil),
		"compromised":  float64(0),
		"hostIpc":      false,
		"hostNetwork":  false,
		"hostPid":      false,
		"image":        "dockerhub.com/elasticsearch:latest",
		"name":         "elasticsearch",
		"node":         "test-node.ec2.internal",
		"pod":          "app-monitors-client-78cb6d7899-j2rjp",
		"ports":        []any{"9200"},
		"privesc":      false,
		"privileged":   false,
		"runAsUser":    float64(1000),
		"storeID":      cid.Hex(),
		"team":         "test-team",
		"app":          "test-app",
		"service":      "test-service",
	}

	cgw := graphdb.NewAsyncVertexWriter(t)
	cgw.EXPECT().Queue(ctx, cv).Return(nil).Once()
	cgw.EXPECT().Flush(ctx).Return(nil)
	cgw.EXPECT().Close(ctx).Return(nil)

	// Graph setup - volumes

	vv := map[string]any{
		"name":         "kube-api-access-4x9fz",
		"isNamespaced": true,
		"namespace":    "test-app",
		"sourcePath":   "/var/lib/kubelet/pods//volumes/kubernetes.io~projected/kube-api-access-4x9fz/token",
		"mountPath":    "/var/run/secrets/kubernetes.io/serviceaccount",
		"storeID":      vid.Hex(),
		"team":         "test-team",
		"app":          "test-app",
		"service":      "test-service",
		"type":         "Projected",
		"readonly":     true,
	}
	vgw := graphdb.NewAsyncVertexWriter(t)
	vgw.EXPECT().Queue(ctx, vv).Return(nil).Once()
	vgw.EXPECT().Flush(ctx).Return(nil)
	vgw.EXPECT().Close(ctx).Return(nil)

	// Graph setup - endpoints

	ev := map[string]interface{}{
		"addressType":     "IPv4",
		"addresses":       []interface{}{"10.1.1.2"},
		"app":             "test-app",
		"compromised":     float64(0),
		"exposure":        float64(shared.EndpointExposureNodeIP),
		"isNamespaced":    true,
		"name":            "test-app::app-monitors-client-78cb6d7899-j2rjp::TCP::9200",
		"namespace":       "test-app",
		"port":            float64(9200),
		"portName":        "http",
		"protocol":        "TCP",
		"service":         "test-service",
		"serviceDns":      "",
		"serviceEndpoint": "http",
		"storeID":         eid.Hex(),
		"team":            "test-team",
	}
	egw := graphdb.NewAsyncVertexWriter(t)
	egw.EXPECT().Queue(ctx, ev).Return(nil).Once()
	egw.EXPECT().Flush(ctx).Return(nil)
	egw.EXPECT().Close(ctx).Return(nil)

	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Pod"), c, mock.AnythingOfType("graphdb.WriterOption")).Return(pgw, nil)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Container"), c, mock.AnythingOfType("graphdb.WriterOption")).Return(cgw, nil)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Volume"), c, mock.AnythingOfType("graphdb.WriterOption")).Return(vgw, nil)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("*vertex.Endpoint"), c, mock.AnythingOfType("graphdb.WriterOption")).Return(egw, nil)

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

	// Initialize
	err = pi.Initialize(ctx, deps)
	assert.NoError(t, err)

	// Run
	err = pi.Run(ctx)
	assert.NoError(t, err)

	// Close
	err = pi.Close(ctx)
	assert.NoError(t, err)
}
