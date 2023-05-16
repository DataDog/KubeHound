package converter

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
	corev1 "k8s.io/api/core/v1"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
)

var (
	ErrUnsupportedVolume   = errors.New("provided volume is not currently supported")
	ErrDanglingRoleBinding = errors.New("role binding found with no matching role")
	ErrNoCacheInitialized  = errors.New("cache reader required for conversion")
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
		Id:     store.StoreObjectID(),
		PodId:  parent.Id,
		NodeId: parent.NodeId,
		Inherited: store.ContainerInherited{
			PodName:        parent.K8.Name,
			NodeName:       parent.K8.Spec.NodeName,
			HostPID:        parent.K8.Spec.HostPID,
			HostIPC:        parent.K8.Spec.HostIPC,
			HostNetwork:    parent.K8.Spec.HostNetwork,
			ServiceAccount: parent.K8.Spec.ServiceAccountName,
		},
		K8: corev1.Container(*input),
	}, nil
}

// Node returns the store representation of a K8s node from an input K8s node object.
func (c *StoreConverter) Node(_ context.Context, input types.NodeType) (*store.Node, error) {
	output := &store.Node{
		Id: store.StoreObjectID(),
		K8: corev1.Node(*input),
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

	nid, err := c.cache.Get(ctx, cache.NodeKey(input.Spec.NodeName))
	if err != nil {
		return nil, err
	}

	onid, err := primitive.ObjectIDFromHex(nid)
	if err != nil {
		return nil, err
	}

	output := &store.Pod{
		Id:     store.StoreObjectID(),
		NodeId: onid,
		K8:     corev1.Pod(*input),
	}

	if len(input.Namespace) != 0 {
		output.IsNamespaced = true
	}

	return output, nil
}

// Volume returns the store representation of a K8s mounted volume from an input K8s volume object.
// NOTE: requires cache access (ContainerKey).
func (c *StoreConverter) Volume(ctx context.Context, input types.VolumeType, parent *store.Pod) (*store.Volume, error) {
	if c.cache == nil {
		return nil, ErrNoCacheInitialized
	}

	// Only a subset of volumes are currently supported
	switch {
	case input.HostPath != nil:
		break
	case input.Projected != nil:
		break
	default:
		return nil, ErrUnsupportedVolume
	}

	output := &store.Volume{
		Id:     store.StoreObjectID(),
		PodId:  parent.Id,
		NodeId: parent.NodeId,
		Name:   input.Name,
		Source: corev1.Volume(*input),
		Mounts: make([]store.VolumeMount, 0),
	}

	// TODO this is not quite right, we need to have a volume that is unique per node and append mounts to the same entry
	// For now this is ok but make data volumes larger than needed
	for _, container := range parent.K8.Spec.Containers {
		for _, mount := range container.VolumeMounts {
			if mount.Name == output.Source.Name {
				cid, err := c.cache.Get(ctx, cache.ContainerKey(parent.K8.Name, container.Name))
				if err != nil {
					return nil, err
				}

				ocid, err := primitive.ObjectIDFromHex(cid)
				if err != nil {
					return nil, err
				}

				output.Mounts = append(output.Mounts, store.VolumeMount{
					ContainerId: ocid,
					K8:          mount,
				})
			}
		}
	}

	return output, nil
}

// Role returns the store representation of a K8s role from an input K8s Role object.
func (c *StoreConverter) Role(_ context.Context, input types.RoleType) (*store.Role, error) {
	return &store.Role{
		Id:           store.StoreObjectID(),
		Name:         input.Name,
		IsNamespaced: true,
		Namespace:    input.Namespace,
		Rules:        input.Rules,
	}, nil
}

// ClusterRole returns the store representation of a K8s cluster role from an input K8s ClusterRole object.
func (c *StoreConverter) ClusterRole(_ context.Context, input types.ClusterRoleType) (*store.Role, error) {
	return &store.Role{
		Id:           store.StoreObjectID(),
		Name:         input.Name,
		IsNamespaced: false,
		Namespace:    "",
		Rules:        input.Rules,
	}, nil
}

// RoleBinding returns the store representation of a K8s role binding from an input K8s RoleBinding object.
// NOTE: requires cache access (RoleKey).
func (c *StoreConverter) RoleBinding(ctx context.Context, input types.RoleBindingType) (*store.RoleBinding, error) {
	if c.cache == nil {
		return nil, ErrNoCacheInitialized
	}

	var output *store.RoleBinding

	rid, err := c.cache.Get(ctx, cache.RoleKey(input.RoleRef.Name))
	if err != nil {
		// We can get cache misses here if bindings remain with no corresponding role.
		return nil, ErrDanglingRoleBinding
	}

	orid, err := primitive.ObjectIDFromHex(rid)
	if err != nil {
		return nil, err
	}

	subj := input.Subjects
	output = &store.RoleBinding{
		Id:           store.StoreObjectID(),
		RoleId:       orid,
		Name:         input.Name,
		IsNamespaced: true,
		Namespace:    input.Namespace,
		Subjects:     make([]store.BindSubject, 0, len(subj)),
	}

	for _, s := range subj {
		output.Subjects = append(output.Subjects, store.BindSubject{
			IdentityId: store.StoreObjectID(),
			Subject:    s,
		})
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

	rid, err := c.cache.Get(ctx, cache.RoleKey(input.RoleRef.Name))
	if err != nil {
		// We can get cache misses here if bindings remain with no corresponding role.
		return nil, ErrDanglingRoleBinding
	}

	orid, err := primitive.ObjectIDFromHex(rid)
	if err != nil {
		return nil, err
	}

	subj := input.Subjects
	output = &store.RoleBinding{
		Id:           store.StoreObjectID(),
		RoleId:       orid,
		Name:         input.Name,
		IsNamespaced: false,
		Namespace:    "",
		Subjects:     make([]store.BindSubject, 0, len(subj)),
	}

	for _, s := range subj {
		output.Subjects = append(output.Subjects, store.BindSubject{
			IdentityId: store.StoreObjectID(),
			Subject:    s,
		})
	}

	return output, nil
}

// Identity returns the store representation of a K8s identity role binding from an input store BindSubject (subfield of RoleBinding) object.
// NOTE: store.Identity does not map directly to a K8s API object and instead derives from the subject of a role binding.
func (c *StoreConverter) Identity(_ context.Context, input store.BindSubject) (*store.Identity, error) {
	output := &store.Identity{
		Id:        input.IdentityId,
		Name:      input.Subject.Name,
		Namespace: "",
		Type:      input.Subject.Kind,
	}

	if len(input.Subject.Namespace) != 0 {
		output.IsNamespaced = true
		output.Namespace = input.Subject.Namespace
	}

	return output, nil
}
