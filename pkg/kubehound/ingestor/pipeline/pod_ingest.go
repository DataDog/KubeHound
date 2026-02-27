package pipeline

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	corev1 "k8s.io/api/core/v1"
)

const (
	PodIngestName = "k8s-pod-ingest"
)

type PodIngest struct {
	v []vertex.Builder
	r *IngestResources
}

var _ ObjectIngest = (*PodIngest)(nil)

func (i *PodIngest) Name() string {
	return PodIngestName
}

func (i *PodIngest) Initialize(ctx context.Context, deps *Dependencies) error {
	var err error

	i.v = []vertex.Builder{
		&vertex.Pod{},
		&vertex.Container{},
		&vertex.Volume{},
		&vertex.Endpoint{},
	}

	opts := make([]IngestResourceOption, 0)
	opts = append(opts, WithConverterDB())
	for _, vtx := range i.v {
		opts = append(opts, WithGraphWriter(vtx))
	}

	i.r, err = CreateResources(ctx, deps, opts...)
	if err != nil {
		return err
	}

	return nil
}

// processEndpoints will handle the ingestion pipeline for endpoints belonging to a processed K8s pod input.
func (i *PodIngest) processEndpoints(ctx context.Context, port *corev1.ContainerPort, pod *store.Pod, container *store.Container) error {
	// Normalize endpoint to temporary store object format
	tmp, err := i.r.storeConvert.EndpointPrivate(ctx, port, pod, container)
	if err != nil {
		return err
	}

	// Check whether this exposed container endpoint has an associated endpoint slice.
	// If it does, nothing further to do. If it does NOT, write the container port as a private endpoint.
	if i.r.endpointSliceExists(ctx, tmp.Namespace, tmp.PodName, tmp.SafeProtocol(), tmp.SafePort()) {
		// Validate our assumptions
		if port.HostPort != 0 && port.ContainerPort != port.HostPort {
			log.Trace(ctx).Warnf("assumption failure: host port set on container with associated endpoint slice (%s::%s::%s::%d)",
				tmp.Namespace, tmp.PodName, tmp.SafeProtocol(), tmp.SafePort())
		}
		return nil
	}

	// Promote the temporary object to an object that will be written to our databases.
	se := tmp

	// Write to store
	if err := i.r.writeStore(ctx, se); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.Endpoint(se)
	if err != nil {
		return err
	}

	// Write to graph
	if err := i.r.writeVertex(ctx, i.v[3], insert); err != nil {
		return err
	}

	return nil
}

// processContainer will handle the ingestion pipeline for a container belonging to a processed K8s pod input.
func (i *PodIngest) processContainer(ctx context.Context, parent *store.Pod, k8sPod types.PodType, container types.ContainerType) error {
	if ok, err := preflight.CheckContainer(container); !ok {
		return err
	}

	// Normalize container to store object format
	sc, err := i.r.storeConvert.Container(ctx, container, parent)
	if err != nil {
		return err
	}

	// Write to store
	if err := i.r.writeStore(ctx, sc); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.Container(sc, parent)
	if err != nil {
		return err
	}

	// Write to graph
	if err := i.r.writeVertex(ctx, i.v[1], insert); err != nil {
		return err
	}

	// Handle volume mounts
	for _, volumeMount := range container.VolumeMounts {
		vm := volumeMount
		err := i.processVolumeMount(ctx, &vm, k8sPod.Spec.Volumes, parent, sc)
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
func (i *PodIngest) processVolumeMount(ctx context.Context, volumeMount types.VolumeMountType, k8sVolumes []corev1.Volume, pod *store.Pod, container *store.Container) error {
	if ok, err := preflight.CheckVolume(volumeMount); !ok {
		return err
	}

	// Normalize volume to store object format
	sv, err := i.r.storeConvert.VolumeFromK8s(ctx, volumeMount, k8sVolumes, pod, container)
	if err != nil {
		log.Trace(ctx).Debugf("process volume type: %v (continuing)", err)

		return nil
	}

	// Write to store
	if err := i.r.writeStore(ctx, sv); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.Volume(sv, pod)
	if err != nil {
		return err
	}

	// Write to graph
	return i.r.writeVertex(ctx, i.v[2], insert)
}

// streamCallback is invoked by the collector for each pod collected.
func (i *PodIngest) IngestPod(ctx context.Context, pod types.PodType) error {
	if ok, err := preflight.CheckPod(ctx, pod); !ok {
		return err
	}

	// Normalize pod to store object format
	sp, err := i.r.storeConvert.Pod(ctx, pod)
	if err != nil {
		log.Trace(ctx).Warnf("process pod %s error (continuing): %v", pod.Name, err)

		return nil
	}

	// Write to store
	if err := i.r.writeStore(ctx, sp); err != nil {
		return err
	}

	// Transform store model to vertex input
	insert, err := i.r.graphConvert.Pod(sp) //nolint: contextcheck
	if err != nil {
		return err
	}

	// Write to graph
	if err := i.r.writeVertex(ctx, i.v[0], insert); err != nil {
		return err
	}

	// Handle containers
	for _, container := range pod.Spec.Containers {
		c := container
		err := i.processContainer(ctx, sp, pod, &c)
		if err != nil {
			return err
		}
	}

	return nil
}

// completeCallback is invoked by the collector when all pods have been streamed.
func (i *PodIngest) Complete(ctx context.Context) error {
	return i.r.flushWriters(ctx)
}

func (i *PodIngest) Run(ctx context.Context) error {
	return i.r.collect.StreamPods(ctx, i)
}

func (i *PodIngest) Close(ctx context.Context) error {
	return i.r.cleanupAll(ctx)
}
