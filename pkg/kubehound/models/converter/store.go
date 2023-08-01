package converter

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kube"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
)

const (
	EmptyNamespace = ""
)

var (
	ErrUnsupportedVolume     = errors.New("provided volume is not currently supported")
	ErrNoCacheInitialized    = errors.New("cache reader required for conversion")
	ErrDanglingRoleBinding   = errors.New("role binding found with no matching role")
	ErrProjectedDefaultToken = errors.New("projected volume grant no access (default serviceaccount)")
)

// StoreConverter enables converting between an input K8s model to its equivalent store model.
type StoreConverter struct {
	cache cache.CacheReader
}

// NewStore returns a new store converter instance.
func NewStore() *StoreConverter {
	return &StoreConverter{}
}

// NewStoreWithCache returns a new store converter instance with read access to the cache.
func NewStoreWithCache(cache cache.CacheReader) *StoreConverter {
	return &StoreConverter{
		cache: cache,
	}
}

// Container returns the store representation of a K8s container from an input K8s container object.
func (c *StoreConverter) Container(_ context.Context, input types.ContainerType, parent *store.Pod) (*store.Container, error) {
	return &store.Container{
		Id:     store.ObjectID(),
		PodId:  parent.Id,
		NodeId: parent.NodeId,
		Inherited: store.ContainerInherited{
			PodName:        parent.K8.Name,
			NodeName:       parent.K8.Spec.NodeName,
			Namespace:      parent.K8.Namespace,
			HostPID:        parent.K8.Spec.HostPID,
			HostIPC:        parent.K8.Spec.HostIPC,
			HostNetwork:    parent.K8.Spec.HostNetwork,
			ServiceAccount: parent.K8.Spec.ServiceAccountName,
		},
		K8:        corev1.Container(*input),
		Ownership: store.ExtractOwnership(parent.K8.Labels),
	}, nil
}

// Node returns the store representation of a K8s node from an input K8s node object.
func (c *StoreConverter) Node(_ context.Context, input types.NodeType) (*store.Node, error) {
	output := &store.Node{
		Id:        store.ObjectID(),
		K8:        corev1.Node(*input),
		Ownership: store.ExtractOwnership(input.ObjectMeta.Labels),
	}

	if len(input.Namespace) != 0 {
		output.IsNamespaced = true
	}

	return output, nil
}

// Pod returns the store representation of a K8s pod from an input K8s pod object.
// NOTE: requires cache access (NodeKey).
func (c *StoreConverter) Pod(ctx context.Context, input types.PodType) (*store.Pod, error) {
	if c.cache == nil {
		return nil, ErrNoCacheInitialized
	}

	nid, err := c.cache.Get(ctx, cachekey.Node(input.Spec.NodeName)).ObjectID()
	if err != nil {
		return nil, err
	}

	output := &store.Pod{
		Id:        store.ObjectID(),
		NodeId:    nid,
		K8:        corev1.Pod(*input),
		Ownership: store.ExtractOwnership(input.ObjectMeta.Labels),
	}

	if len(input.Namespace) != 0 {
		output.IsNamespaced = true
	}

	return output, nil
}

// handleProjectedToken returns the identity store ID and source path corresponding to a projected token volume mount.
func (c *StoreConverter) handleProjectedToken(ctx context.Context, input types.VolumeMountType,
	volume *corev1.Volume, pod *store.Pod) (primitive.ObjectID, string, error) {

	// Retrieve the associated identity store ID from the cache
	said, err := c.cache.Get(ctx, cachekey.Identity(pod.K8.Spec.ServiceAccountName, pod.K8.Namespace)).ObjectID()
	switch err {
	case nil:
		// We have a matching identity object in the store, continue to create a volume
	case cache.ErrNoEntry:
		// This is completely fine. Most pods will run under a default account with no permissions which we ignore.
		return primitive.NilObjectID, "", ErrProjectedDefaultToken
	default:
		return primitive.NilObjectID, "", err
	}

	// Loop through looking for the service account token projection
	var sourcePath string
	for _, proj := range volume.Projected.Sources {
		if proj.ServiceAccountToken != nil {
			sourcePath = kube.ServiceAccountTokenPath(string(pod.K8.ObjectMeta.UID), input.Name)
			break
		}
	}

	return said, sourcePath, nil
}

// Volume returns the store representation of a K8s mounted volume from an input K8s volume object.
// NOTE: requires cache access (IdentityKey).
func (c *StoreConverter) Volume(ctx context.Context, input types.VolumeMountType, pod *store.Pod,
	container *store.Container) (*store.Volume, error) {

	if c.cache == nil {
		return nil, ErrNoCacheInitialized
	}

	output := &store.Volume{
		Id:          store.ObjectID(),
		PodId:       pod.Id,
		NodeId:      pod.NodeId,
		ContainerId: container.Id,
		Name:        input.Name,
		MountPath:   input.MountPath,
		ReadOnly:    input.ReadOnly,
		Ownership:   store.ExtractOwnership(pod.K8.Labels),
	}

	// Resolve the volume to the underlying name
	found := false

	// Expect a small size array so iterating through this is quicker than building up a map for lookup
	for _, volume := range pod.K8.Spec.Volumes {
		if volume.Name == input.Name {
			found = true

			// Only a subset of volumes are currently supported
			switch {
			case volume.HostPath != nil:
				output.Type = shared.VolumeTypeHost
				output.SourcePath = volume.HostPath.Path
			case volume.Projected != nil:
				said, source, err := c.handleProjectedToken(ctx, input, &volume, pod)
				if err != nil {
					return nil, fmt.Errorf("projected token volume (%s) processing: %w", volume.Name, err)
				}

				output.Type = shared.VolumeTypeProjected
				output.SourcePath = source
				output.ProjectedId = said
			default:
				return nil, ErrUnsupportedVolume
			}

			output.K8 = volume
		}
	}

	if !found {
		return nil, fmt.Errorf("mount has no corresponding volume: %s", input.Name)
	}

	return output, nil
}

// Role returns the store representation of a K8s role from an input K8s Role object.
func (c *StoreConverter) Role(_ context.Context, input types.RoleType) (*store.Role, error) {
	return &store.Role{
		Id:           store.ObjectID(),
		Name:         input.Name,
		IsNamespaced: true,
		Namespace:    input.Namespace,
		Rules:        input.Rules,
		Ownership:    store.ExtractOwnership(input.ObjectMeta.Labels),
	}, nil
}

// ClusterRole returns the store representation of a K8s cluster role from an input K8s ClusterRole object.
func (c *StoreConverter) ClusterRole(_ context.Context, input types.ClusterRoleType) (*store.Role, error) {
	return &store.Role{
		Id:           store.ObjectID(),
		Name:         input.Name,
		IsNamespaced: false,
		Namespace:    "",
		Rules:        input.Rules,
		Ownership:    store.ExtractOwnership(input.ObjectMeta.Labels),
	}, nil
}

func (c *StoreConverter) convertSubject(ctx context.Context, subj rbacv1.Subject) (store.BindSubject, error) {
	// Check if identity already exists and use that ID, otherwise generate a new one
	sid, err := c.cache.Get(ctx, cachekey.Identity(subj.Name, subj.Namespace)).ObjectID()
	switch err {
	case nil:
		// Entry already exists, use the cached id value
	case cache.ErrNoEntry:
		// Entry does not exist, create a new id value
		sid = store.ObjectID()
	default:
		return store.BindSubject{}, err
	}

	return store.BindSubject{
		IdentityId: sid,
		Subject:    subj,
	}, nil
}

// RoleBinding returns the store representation of a K8s role binding from an input K8s RoleBinding object.
// NOTE: requires cache access (RoleKey).
func (c *StoreConverter) RoleBinding(ctx context.Context, input types.RoleBindingType) (*store.RoleBinding, error) {
	if c.cache == nil {
		return nil, ErrNoCacheInitialized
	}

	var output *store.RoleBinding

	rid, err := c.cache.Get(ctx, cachekey.Role(input.RoleRef.Name, input.Namespace)).ObjectID()
	if err != nil {
		// We can get cache misses here if binding corresponds to a cluster role
		rid, err = c.cache.Get(ctx, cachekey.Role(input.RoleRef.Name, EmptyNamespace)).ObjectID()
		if err != nil {
			return nil, ErrDanglingRoleBinding
		}
	}

	subj := input.Subjects
	output = &store.RoleBinding{
		Id:           store.ObjectID(),
		RoleId:       rid,
		Name:         input.Name,
		IsNamespaced: true,
		Namespace:    input.Namespace,
		Subjects:     make([]store.BindSubject, 0, len(subj)),
		Ownership:    store.ExtractOwnership(input.ObjectMeta.Labels),
	}

	for _, s := range subj {
		s, err := c.convertSubject(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("role binding subject convert: %w", err)
		}

		output.Subjects = append(output.Subjects, s)
	}

	return output, nil
}

// ClusterRoleBinding returns the store representation of a K8s cluster role binding from an input K8s ClusterRoleBinding object.
// NOTE: requires cache access (RoleKey).
func (c *StoreConverter) ClusterRoleBinding(ctx context.Context, input types.ClusterRoleBindingType) (*store.RoleBinding, error) {
	if c.cache == nil {
		return nil, ErrNoCacheInitialized
	}

	var output *store.RoleBinding

	rid, err := c.cache.Get(ctx, cachekey.Role(input.RoleRef.Name, input.Namespace)).ObjectID()
	if err != nil {
		// We can get cache misses here if binding corresponds to a cluster role
		rid, err = c.cache.Get(ctx, cachekey.Role(input.RoleRef.Name, EmptyNamespace)).ObjectID()
		if err != nil {
			return nil, ErrDanglingRoleBinding
		}
	}

	subj := input.Subjects
	output = &store.RoleBinding{
		Id:           store.ObjectID(),
		RoleId:       rid,
		Name:         input.Name,
		IsNamespaced: false,
		Namespace:    "",
		Subjects:     make([]store.BindSubject, 0, len(subj)),
		Ownership:    store.ExtractOwnership(input.ObjectMeta.Labels),
	}

	for _, s := range subj {
		s, err := c.convertSubject(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("cluster role binding subject convert: %w", err)
		}

		output.Subjects = append(output.Subjects, s)
	}

	return output, nil
}

// Identity returns the store representation of a K8s identity role binding from an input store BindSubject (subfield of RoleBinding) object.
// NOTE: store.Identity does not map directly to a K8s API object and instead derives from the subject of a role binding.
func (c *StoreConverter) Identity(_ context.Context, input *store.BindSubject, parent *store.RoleBinding) (*store.Identity, error) {
	output := &store.Identity{
		Id:        input.IdentityId,
		Name:      input.Subject.Name,
		Namespace: "",
		Type:      input.Subject.Kind,
		Ownership: parent.Ownership,
	}

	if len(input.Subject.Namespace) != 0 {
		output.IsNamespaced = true
		output.Namespace = input.Subject.Namespace
	}

	return output, nil
}

func (c *StoreConverter) Endpoint(_ context.Context, addr discoveryv1.Endpoint,
	port discoveryv1.EndpointPort, parent types.EndpointType) (*store.Endpoint, error) {

	// Ensure our assumption that the target is always a pod holds
	if addr.TargetRef.Kind != "Pod" {
		return nil, fmt.Errorf("unexpected endpoint target %#v", addr.TargetRef)
	}

	output := &store.Endpoint{
		Id:           store.ObjectID(),
		PodName:      addr.TargetRef.Name,
		PodNamespace: addr.TargetRef.Namespace,
		Name:         fmt.Sprintf("%s::%s::%s", parent.Name, addr.TargetRef.Name, *port.Name),
		HasSlice:     true,
		ServiceName:  parent.Labels["kubernetes.io/service-name"], // TODO tidy
		AddressType:  parent.AddressType,
		Backend:      addr,
		Port:         port,
		Ownership:    store.ExtractOwnership(parent.ObjectMeta.Labels),
		K8:           parent.ObjectMeta,
		Access:       shared.EndpointAccessExternal, // If created via the ingestion pipeline the endpoint corresponds to a k8s endpoint slice
	}

	if addr.NodeName != nil {
		output.NodeName = *addr.NodeName
	}

	if len(parent.Namespace) != 0 {
		output.IsNamespaced = true
		output.Namespace = parent.Namespace
	}

	return output, nil
}

func (c *StoreConverter) EndpointPrivate(_ context.Context, port *corev1.ContainerPort,
	pod *store.Pod, container *store.Container) (*store.Endpoint, error) {

	output := &store.Endpoint{
		Id:           store.ObjectID(),
		ContainerId:  container.Id,
		PodName:      pod.K8.Name,
		PodNamespace: pod.K8.Namespace,
		ServiceName:  fmt.Sprintf("%s::%s::%s", port.Name, pod.K8.Namespace, pod.K8.Name),
		Name:         fmt.Sprintf("%s::%s::%d", container.K8.Name, pod.K8.Status.HostIP, port.ContainerPort),
		NodeName:     pod.K8.Spec.NodeName,
		AddressType:  "IPv4", // TODO figure out address tyoe from inpout
		Backend: discoveryv1.Endpoint{
			// TODO other fields
			Addresses: []string{pod.K8.Status.PodIP},
			TargetRef: &corev1.ObjectReference{
				Kind:            pod.K8.Kind,
				APIVersion:      pod.K8.APIVersion,
				Name:            pod.K8.Name,
				Namespace:       pod.K8.Namespace,
				UID:             pod.K8.UID,
				ResourceVersion: pod.K8.ResourceVersion,
			},
			NodeName: &pod.K8.Spec.NodeName,
		},
		Port: discoveryv1.EndpointPort{
			Name:     &port.Name,
			Protocol: &port.Protocol,
			Port:     &port.ContainerPort, // TOOD host port

		},
		Ownership: container.Ownership,

		// If created via the pod ingestion pipeline the endpoint does not correspond to a k8s
		// endpoint slice and thus is only accessible from within the cluster.
		Access: shared.EndpointAccessInternal,
	}

	if len(pod.K8.Namespace) != 0 {
		output.IsNamespaced = true
		output.Namespace = pod.K8.Namespace
	}

	return output, nil
}
