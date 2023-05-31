package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
)

const (
	PodIngestName = "k8s-pod-ingest"
)

type objectIndex int

const (
	podIndex objectIndex = iota
	containerIndex
	volumeIndex
	maxObjectIndex
)

type PodIngest struct {
	v []vertex.Vertex
	c []collections.Collection
	r *IngestResources
}

var _ ObjectIngest = (*PodIngest)(nil)

func (i *PodIngest) Name() string {
	return PodIngestName
}

func (i *PodIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	var err error

	//
	// Pods will create other objects such as volumes (from the pod volume mount list) and containers
	// from the (container/init container lists). As such we need to intialize a list of the writers we need.
	//

	i.v = []vertex.Vertex{
		vertex.Pod{},
		vertex.Container{},
		vertex.Volume{},
	}

	i.c = []collections.Collection{
		collections.Pod{},
		collections.Container{},
		collections.Volume{},
	}

	opts := make([]IngestResourceOption, 0)
	opts = append(opts, WithCacheWriter())
	opts = append(opts, WithConverterCache())
	for objIndex := podIndex; objIndex < maxObjectIndex; objIndex++ {
		opts = append(opts, WithStoreWriter(i.c[objIndex]))
		opts = append(opts, WithGraphWriter(i.v[objIndex]))
	}

	i.r, err = CreateResources(ctx, deps, opts...)
	if err != nil {
		return err
	}

	return nil
}

// processContainer will handle the ingestion pipeline for a container belonging to a processed K8s pod input.
func (i *PodIngest) processContainer(ctx context.Context, parent *store.Pod, container types.ContainerType) error {
	// Normalize container to store object format
	sc, err := i.r.storeConvert.Container(ctx, container, parent)
	if err != nil {
		return err
	}

	// Async write to store
	if err := i.r.storeWriter(i.c[containerIndex]).Queue(ctx, sc); err != nil {
		return err
	}

	// Async write to cache
	if err := i.r.cacheWriter.Queue(ctx, cache.ContainerKey(parent.K8.Name, sc.K8.Name), sc.Id.Hex()); err != nil {
		return err
	}

	// Transform store model to vertex input
	vc, err := i.r.graphConvert.Container(sc)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.r.graphWriter(i.v[containerIndex]).Queue(ctx, vc); err != nil {
		return err
	}

	return nil
}

// processVolume will handle the ingestion pipeline for a volume belonging to a processed K8s pod input.
func (i *PodIngest) processVolume(ctx context.Context, parent *store.Pod, volume types.VolumeType) error {
	// Normalize volume to store object format
	sv, err := i.r.storeConvert.Volume(ctx, volume, parent)
	if err != nil {
		return err
	}

	// Async write to store
	if err := i.r.storeWriter(i.c[volumeIndex]).Queue(ctx, sv); err != nil {
		return err
	}

	// Transform store model to vertex input
	vv, err := i.r.graphConvert.Volume(sv)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.r.graphWriter(i.v[volumeIndex]).Queue(ctx, vv); err != nil {
		return err
	}

	return nil
}

// streamCallback is invoked by the collector for each pod collected.
// The function ingests an input pod object into the cache/store/graph and then ingests
// all child objects (containers, volumes, etc) through their own ingestion pipeline.
func (i *PodIngest) IngestPod(ctx context.Context, pod types.PodType) error {
	// Normalize pod to store object format
	sp, err := i.r.storeConvert.Pod(ctx, pod)
	if err != nil {
		return err
	}

	// Async write to store
	if err := i.r.storeWriter(i.c[podIndex]).Queue(ctx, sp); err != nil {
		return err
	}

	// Transform store model to vertex input
	vp, err := i.r.graphConvert.Pod(sp)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.r.graphWriter(i.v[podIndex]).Queue(ctx, vp); err != nil {
		return err
	}

	// Handle containers
	// TODO: review handling of InitContainers
	for _, container := range pod.Spec.Containers {
		c := container
		err := i.processContainer(ctx, sp, &c)
		if err != nil {
			return err
		}
	}

	// Handle volumes
	for _, volume := range pod.Spec.Volumes {
		v := volume
		err := i.processVolume(ctx, sp, &v)
		if err != nil {
			return err
		}
	}

	return nil
}

// completeCallback is invoked by the collector when all pods have been streamed.
// The function flushes all writers and waits for completion.
func (i *PodIngest) Complete(ctx context.Context) error {
	return i.r.flushWriters(ctx)
}

func (i *PodIngest) Run(ctx context.Context) error {
	return i.r.collect.StreamPods(ctx, i)
}

func (i *PodIngest) Close(ctx context.Context) error {
	return i.r.cleanupAll(ctx)
}
