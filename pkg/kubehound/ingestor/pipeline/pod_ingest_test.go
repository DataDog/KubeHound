package pipeline

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockcollect "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	cache "github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	graphdb "github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb/mocks"
	storedb "github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb/mocks"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPodIngest_Pipeline(t *testing.T) {
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
	c := cache.NewCacheProvider(t)
	cw := cache.NewAsyncWriter(t)
	cw.EXPECT().Queue(ctx, mock.AnythingOfType("*cachekey.containerCacheKey"), mock.AnythingOfType("string")).Return(nil).Once()

	cw.EXPECT().Flush(ctx).Return(nil)
	cw.EXPECT().Close(ctx).Return(nil)
	c.EXPECT().BulkWriter(ctx).Return(cw, nil)
	c.EXPECT().Get(ctx, mock.AnythingOfType("*cachekey.nodeCacheKey")).Return(store.ObjectID().Hex(), nil)
	c.EXPECT().Get(ctx, mock.AnythingOfType("*cachekey.containerCacheKey")).Return(store.ObjectID().Hex(), nil)

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

	sdb.EXPECT().BulkWriter(ctx, pods, mock.Anything).Return(psw, nil)
	sdb.EXPECT().BulkWriter(ctx, containers, mock.Anything).Return(csw, nil)
	sdb.EXPECT().BulkWriter(ctx, volumes, mock.Anything).Return(vsw, nil)

	// Graph setup - pods
	pv := map[string]any{
		"compromised":            float64(0),
		"critical":               false,
		"isNamespaced":           true,
		"name":                   "app-monitors-client-78cb6d7899-j2rjp",
		"namespace":              "test-app",
		"node":                   "test-node.ec2.internal",
		"serviceAccount":         "app-monitors",
		"sharedProcessNamespace": false,
		"storeID":                pid.Hex(),
	}

	gdb := graphdb.NewProvider(t)
	pgw := graphdb.NewAsyncVertexWriter(t)
	pgw.EXPECT().Queue(ctx, pv).Return(nil).Once()
	pgw.EXPECT().Flush(ctx).Return(nil)
	pgw.EXPECT().Close(ctx).Return(nil)

	// Graph setup - containers
	cv := map[string]any{
		"args":         any(nil),
		"capabilities": []any{},
		"command":      any(nil),
		"compromised":  float64(0),
		"hostIpc":      false,
		"hostNetwork":  false,
		"hostPath":     false,
		"hostPid":      false,
		"image":        "dockerhub.com/elasticsearch:latest",
		"name":         "elasticsearch",
		"node":         "test-node.ec2.internal",
		"pod":          "app-monitors-client-78cb6d7899-j2rjp",
		"ports":        []any{"9200", "9300"},
		"privesc":      false,
		"privileged":   false,
		"runAsUser":    float64(0),
		"storeID":      cid.Hex()}

	cgw := graphdb.NewAsyncVertexWriter(t)
	cgw.EXPECT().Queue(ctx, cv).Return(nil).Once()
	cgw.EXPECT().Flush(ctx).Return(nil)
	cgw.EXPECT().Close(ctx).Return(nil)

	// Graph setup - volumes

	vv := map[string]any{
		"name":    "kube-api-access-4x9fz",
		"path":    "/var/lib/kubelet/pods//volumes/kubernetes.io~projected/kube-api-access-4x9fz/token",
		"storeID": vid.Hex(),
		"type":    "Projected",
	}
	vgw := graphdb.NewAsyncVertexWriter(t)
	vgw.EXPECT().Queue(ctx, vv).Return(nil).Once()
	vgw.EXPECT().Flush(ctx).Return(nil)
	vgw.EXPECT().Close(ctx).Return(nil)

	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("vertex.Pod"), mock.AnythingOfType("graphdb.WriterOption")).Return(pgw, nil)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("vertex.Container"), mock.AnythingOfType("graphdb.WriterOption")).Return(cgw, nil)
	gdb.EXPECT().VertexWriter(ctx, mock.AnythingOfType("vertex.Volume"), mock.AnythingOfType("graphdb.WriterOption")).Return(vgw, nil)

	deps := &Dependencies{
		Collector: client,
		Cache:     c,
		GraphDB:   gdb,
		StoreDB:   sdb,
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
