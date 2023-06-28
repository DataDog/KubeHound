package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
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
	v []vertex.Builder
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

	i.v = []vertex.Builder{
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
	if ok, err := preflight.CheckContainer(container); !ok {
		return err
	}

	// Normalize container to store object format
	sc, err := i.r.storeConvert.Container(ctx, container, parent)
	if err != nil {
		return err
	}

	// Async write to store
	if err := i.r.writeStore(ctx, i.c[containerIndex], sc); err != nil {
		return err
	}

	// Async write to cache
	if err := i.r.writeCache(ctx, cachekey.Container(parent.K8.Name, sc.K8.Name, parent.K8.Namespace),
		sc.Id.Hex()); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.Container(sc)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.r.writeVertex(ctx, i.v[containerIndex], insert); err != nil {
		return err
	}

	return nil
}

// processVolume will handle the ingestion pipeline for a volume belonging to a processed K8s pod input.
func (i *PodIngest) processVolume(ctx context.Context, parent *store.Pod, volume types.VolumeType) error {
	if ok, err := preflight.CheckVolume(volume); !ok {
		return err
	}

	// Normalize volume to store object format
	sv, err := i.r.storeConvert.Volume(ctx, volume, parent)
	if err != nil {
		log.I.Debugf("process volume type: %v (continuing)", err)
		return nil
	}

	// Async write to store
	if err := i.r.writeStore(ctx, i.c[volumeIndex], sv); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.Volume(sv, parent)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.r.writeVertex(ctx, i.v[volumeIndex], insert); err != nil {
		return err
	}

	return nil
}

// streamCallback is invoked by the collector for each pod collected.
// The function ingests an input pod object into the cache/store/graph and then ingests
// all child objects (containers, volumes, etc) through their own ingestion pipeline.
func (i *PodIngest) IngestPod(ctx context.Context, pod types.PodType) error {
	if ok, err := preflight.CheckPod(pod); !ok {
		return err
	}

	// Normalize pod to store object format
	sp, err := i.r.storeConvert.Pod(ctx, pod)
	if err != nil {
		return err
	}

	// Async write to store
	if err := i.r.writeStore(ctx, i.c[podIndex], sp); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.Pod(sp)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.r.writeVertex(ctx, i.v[podIndex], insert); err != nil {
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
