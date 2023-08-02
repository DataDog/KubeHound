package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	corev1 "k8s.io/api/core/v1"
)

const (
	PodIngestName = "k8s-pod-ingest"
)

type objectIndex int

const (
	podIndex objectIndex = iota
	containerIndex
	volumeIndex
	endpointIndex
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
		&vertex.Pod{},
		&vertex.Container{},
		&vertex.Volume{},
		&vertex.Endpoint{},
	}

	i.c = []collections.Collection{
		collections.Pod{},
		collections.Container{},
		collections.Volume{},
		collections.Endpoint{},
	}

	opts := make([]IngestResourceOption, 0)
	opts = append(opts, WithCacheReader())
	opts = append(opts, WithCacheWriter(cache.WithTest()))
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

// processEndpoints will handle the ingestion pipeline for a endpoints belonging to a processed K8s pod input.
func (i *PodIngest) processEndpoints(ctx context.Context, port *corev1.ContainerPort,
	pod *store.Pod, container *store.Container) error {

	// Normalize endpoint to temporary store object format
	tmp, err := i.r.storeConvert.EndpointPrivate(ctx, port, pod, container)
	if err != nil {
		return err
	}

	// Check whether this exposed container endpoint has an associated endpoint slice. If so, we need do nothing
	// further. However if it does NOT we write the details of the container port as a private endpoint entry.
	ck := cachekey.Endpoint(tmp.Namespace, tmp.PodName, tmp.SafeProtocol(), tmp.SafePort())
	_, err = i.r.readCache(ctx, ck).Bool()
	switch err {
	case cache.ErrNoEntry:
		// No associated endpoint slice, create the endpoint from container parameters
	case nil:
		// Entry already has an associated store entry with the endpoint slice ingest pipeline
		// Nothing further to do...
		return nil
	default:
		return err
	}

	// Promote the temporary object to an object that will be written to our databases.
	se := tmp

	// Async write to store
	if err := i.r.writeStore(ctx, i.c[endpointIndex], se); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.Endpoint(se)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.r.writeVertex(ctx, i.v[endpointIndex], insert); err != nil {
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
	insert, err := i.r.graphConvert.Container(sc, parent)
	if err != nil {
		return err
	}

	// Aysnc write to graph
	if err := i.r.writeVertex(ctx, i.v[containerIndex], insert); err != nil {
		return err
	}

	// Handle volume mounts
	for _, volumeMount := range container.VolumeMounts {
		vm := volumeMount
		err := i.processVolumeMount(ctx, &vm, parent, sc)
		if err != nil {
			return err
		}
	}

	// Handle endpoints (derived from container ports)
	for _, port := range container.Ports {
		p := port
		err := i.processEndpoints(ctx, &p, parent, sc)
		if err != nil {
			return err
		}
	}

	return nil
}

// processVolumeMount will handle the ingestion pipeline for a volume belonging to a processed K8s pod input.
func (i *PodIngest) processVolumeMount(ctx context.Context, volumeMount types.VolumeMountType,
	pod *store.Pod, container *store.Container) error {

	// TODO can we skip known good e.g agent here to cuyt down the volume??
	if ok, err := preflight.CheckVolume(volumeMount); !ok {
		return err
	}

	// Normalize volume to store object format
	sv, err := i.r.storeConvert.Volume(ctx, volumeMount, pod, container)
	if err != nil {
		log.I.Debugf("process volume type: %v (continuing)", err)
		return nil
	}

	// Async write to store
	if err := i.r.writeStore(ctx, i.c[volumeIndex], sv); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.Volume(sv, pod)
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
		log.Trace(ctx).Warnf("process pod %s error (continuing): %v", pod.Name, err)
		return nil
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
	for _, container := range pod.Spec.Containers {
		c := container
		err := i.processContainer(ctx, sp, &c)
		if err != nil {
			return err
		}
	}

	// Currently do not process init containers. Any interesting identity they are running under will be the same as the container since
	// service accounts are defined at a pod level (although this may change in future K8s releases). Any interesting properties of the
	// container will be ephemeral, making any exploitation very complex. Thus we do not include init containers within our graph.

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
