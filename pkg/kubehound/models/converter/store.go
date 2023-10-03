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
	"github.com/DataDog/KubeHound/pkg/kubehound/libkube"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

const (
	EmptyNamespace = ""
)

var (
	ErrUnsupportedVolume     = errors.New("provided volume is not currently supported")
	ErrNoCacheInitialized    = errors.New("cache reader required for conversion")
	ErrDanglingRoleBinding   = errors.New("role binding found with no matching role")
	ErrProjectedDefaultToken = errors.New("projected volume grant no access (default serviceaccount)")
	ErrEndpointTarget        = errors.New("target reference for an endpoint could not be resolved")
	ErrRoleCacheMiss         = errors.New("missing role in cache")
	ErrRoleBindProperties    = errors.New("incorrect combination of (cluster) role and (cluster) role binding properties")
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
	output := &store.Container{
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
		K8:        *input,
		Ownership: store.ExtractOwnership(parent.K8.Labels),
	}

	// Certain fields are set by the PodSecurityContext and overridden by the container's SecurityContext.
	// Currently we only consider the RunAsUser field.
	if input.SecurityContext != nil && input.SecurityContext.RunAsUser != nil {
		output.Inherited.RunAsUser = *input.SecurityContext.RunAsUser
	} else if parent.K8.Spec.SecurityContext != nil && parent.K8.Spec.SecurityContext.RunAsUser != nil {
		output.Inherited.RunAsUser = *parent.K8.Spec.SecurityContext.RunAsUser
	}

	return output, nil
}

// Node returns the store representation of a K8s node from an input K8s node object.
func (c *StoreConverter) Node(ctx context.Context, input types.NodeType) (*store.Node, error) {
	if c.cache == nil {
		return nil, ErrNoCacheInitialized
	}

	output := &store.Node{
		Id:        store.ObjectID(),
		K8:        *input,
		Ownership: store.ExtractOwnership(input.ObjectMeta.Labels),
	}

	if len(input.Namespace) != 0 {
		output.IsNamespaced = true
	}

	// Retrieve the associated identity store ID from the cache
	uid, err := libkube.NodeIdentity(ctx, c.cache, input.Name)
	switch err {
	case nil:
		// We have a matching node identity object in the store
		output.UserId = uid
	case libkube.ErrMissingNodeUser:
		// This is completely fine. Most nodes will run under a default account with no permissions which we ignore.
	default:
		return nil, err
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
		K8:        *input,
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
			sourcePath = libkube.ServiceAccountTokenPath(string(pod.K8.ObjectMeta.UID), input.Name)

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

	role, err := c.cache.Get(ctx, cachekey.Role(input.RoleRef.Name, input.Namespace)).Role()
	if err != nil {
		// We can get cache misses here if binding corresponds to a cluster role
		role, err = c.cache.Get(ctx, cachekey.Role(input.RoleRef.Name, EmptyNamespace)).Role()
		if err != nil {
			return nil, ErrDanglingRoleBinding
		}
	}

	subj := input.Subjects
	output = &store.RoleBinding{
		Id:           store.ObjectID(),
		RoleId:       role.Id,
		Name:         input.Name,
		IsNamespaced: true,
		Namespace:    input.Namespace,
		Subjects:     make([]store.BindSubject, 0, len(subj)),
		Ownership:    store.ExtractOwnership(input.ObjectMeta.Labels),
		K8:           input.RoleRef,
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

	role, err := c.cache.Get(ctx, cachekey.Role(input.RoleRef.Name, input.Namespace)).Role()
	if err != nil {
		// We can get cache misses here if binding corresponds to a cluster role
		role, err = c.cache.Get(ctx, cachekey.Role(input.RoleRef.Name, EmptyNamespace)).Role()
		if err != nil {
			return nil, ErrDanglingRoleBinding
		}
	}

	subj := input.Subjects
	output = &store.RoleBinding{
		Id:           store.ObjectID(),
		RoleId:       role.Id,
		Name:         input.Name,
		IsNamespaced: false,
		Namespace:    "",
		Subjects:     make([]store.BindSubject, 0, len(subj)),
		Ownership:    store.ExtractOwnership(input.ObjectMeta.Labels),
		K8:           input.RoleRef,
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

// PermissionSet returns the store representation of a K8s role / rolebinding combination from input K8s objects.
// RBAC rules and limitation:
//   - Roles and RoleBindings must exist in the same namespace.
//   - RoleBindings can exist in separate namespaces to Service Accounts.
//   - RoleBindings can link ClusterRoles, but they only grant access to the namespace of the RoleBinding.
func (c *StoreConverter) PermissionSet(ctx context.Context, roleBinding *store.RoleBinding) (*store.PermissionSet, error) {
	if c.cache == nil {
		return nil, ErrNoCacheInitialized
	}

	if !roleBinding.IsNamespaced {
		return nil, fmt.Errorf("invalid input (%s), use converter.PermissionSetCluster for cluster role bindings", roleBinding.Name)
	}

	// Get matching role from cache
	var ck cachekey.CacheKey
	if roleBinding.K8.Kind == "ClusterRole" {
		ck = cachekey.Role(roleBinding.K8.Name, EmptyNamespace)
	} else {
		ck = cachekey.Role(roleBinding.K8.Name, roleBinding.Namespace)
	}

	role, err := c.cache.Get(ctx, ck).Role()
	if err != nil {
		return nil, ErrRoleCacheMiss
	}

	// Roles and role bindings must exist in the same namespace or the role must be a ClusterRole
	if roleBinding.Namespace != role.Namespace && role.Namespace != EmptyNamespace {
		log.Trace(ctx).Debugf("The role namespace (%s) does not match the rolebinding namespace (%s)",
			role.Namespace, roleBinding.Namespace)
		return nil, ErrRoleBindProperties
	}

	// RoleBindings can exist in separate namespaces to Service Accounts.
	// Will be FULLY handled in the PERMISSION_DISCOVER edge, just checking if no match is being found
	isEffective := false
	for _, s := range roleBinding.Subjects {
		// Service Account
		// User or Group have to be on the same namespace
		if s.Subject.Kind == "ServiceAccount" || s.Subject.Namespace == roleBinding.Namespace || s.Subject.Namespace == EmptyNamespace {
			isEffective = true
		}
	}

	if !isEffective {
		log.Trace(ctx).Debugf("The rolebinding/subjects are ALL not in the same namespace: rb::%s/rb.sbj::%#v",
			roleBinding.Namespace, roleBinding.Subjects)

		return nil, ErrRoleBindProperties
	}

	output := &store.PermissionSet{
		Id:              store.ObjectID(),
		RoleId:          role.Id,
		RoleName:        role.Name,
		RoleBindingId:   roleBinding.Id,
		RoleBindingName: roleBinding.Name,
		Name:            fmt.Sprintf("%s::%s", role.Name, roleBinding.Name),
		IsNamespaced:    role.IsNamespaced,
		Namespace:       role.Namespace,
		Rules:           role.Rules,
		Ownership:       role.Ownership,
	}

	// RoleBindings can link ClusterRoles, but they only grant access to the namespace of the RoleBinding.
	if !role.IsNamespaced {
		output.IsNamespaced = true
		output.Namespace = roleBinding.Namespace
	}

	return output, nil
}

// PermissionSet returns the store representation of a K8s role / rolebinding combination from input K8s objects.
// RBAC rules and limitation:
//   - ClusterRoleBindings link accounts to ClusterRoles and grant access across all resources.
//   - ClusterRoleBindings can not reference Roles.
func (c *StoreConverter) PermissionSetCluster(ctx context.Context, clusterRoleBinding *store.RoleBinding) (*store.PermissionSet, error) {
	if c.cache == nil {
		return nil, ErrNoCacheInitialized
	}

	if clusterRoleBinding.IsNamespaced {
		return nil, fmt.Errorf("invalid input (%s), use converter.PermissionSet for role bindings", clusterRoleBinding.Name)
	}

	// Get matching role from cache
	role, err := c.cache.Get(ctx, cachekey.Role(clusterRoleBinding.K8.Name, clusterRoleBinding.Namespace)).Role()
	if err != nil {
		return nil, ErrRoleCacheMiss
	}

	// ClusterRoleBindings can not reference Roles.
	if role.IsNamespaced {
		log.Trace(ctx).Debugf("The clusterrolebinding bind a role and not a clusterrole, skipping the permissionset: r::%s/cr::%s",
			role.Namespace, clusterRoleBinding.Namespace)
		return nil, ErrRoleBindProperties
	}

	output := &store.PermissionSet{
		Id:              store.ObjectID(),
		RoleId:          role.Id,
		RoleName:        role.Name,
		RoleBindingId:   clusterRoleBinding.Id,
		RoleBindingName: clusterRoleBinding.Name,
		Name:            fmt.Sprintf("%s::%s", role.Name, clusterRoleBinding.Name),
		IsNamespaced:    role.IsNamespaced,
		Namespace:       role.Namespace,
		Rules:           role.Rules,
		Ownership:       role.Ownership,
	}

	return output, nil
}

// Endpoint returns the store representation of a K8s endpoint from an input Endpoint & EndpointPort objects (subfields of EndpointSlice).
// NOTE: store.Endpoint does not map directly to a K8s API object and instead derives from the elements of an EndpointSlice.
func (c *StoreConverter) Endpoint(_ context.Context, addr discoveryv1.Endpoint,
	port discoveryv1.EndpointPort, parent types.EndpointType) (*store.Endpoint, error) {

	// Ensure we have a target
	if addr.TargetRef == nil {
		return nil, ErrEndpointTarget
	}

	// Ensure our assumption that the target is always a pod holds
	if addr.TargetRef.Kind != "Pod" {
		return nil, ErrEndpointTarget
	}

	output := &store.Endpoint{
		Id:           store.ObjectID(),
		PodName:      addr.TargetRef.Name,
		PodNamespace: addr.TargetRef.Namespace,
		Name:         fmt.Sprintf("%s::%s::%s", parent.Name, *port.Protocol, *port.Name),
		HasSlice:     true,
		ServiceName:  libkube.ServiceName(parent),
		ServiceDns:   libkube.ServiceDns(parent),
		AddressType:  parent.AddressType,
		Backend:      addr,
		Port:         port,
		Ownership:    store.ExtractOwnership(parent.ObjectMeta.Labels),
		K8:           parent.ObjectMeta,

		// If created via the ingestion pipeline the endpoint corresponds to a k8s endpoint slice
		Exposure: shared.EndpointExposureExternal,
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

// EndpointPrivate returns the store representation of a K8s endpoint from an input port, container & pod.
// This variant handles the case when the provided container port does not match a known EndpointSlice. The generated endpoint will
// not be accessible from outside the cluster but can still provide value to an attacker with an presence inside the cluster.
func (c *StoreConverter) EndpointPrivate(_ context.Context, port *corev1.ContainerPort,
	pod *store.Pod, container *store.Container) (*store.Endpoint, error) {

	// Derive the address type from the pod IP
	podIP := pod.K8.Status.PodIP
	addrType, err := libkube.AddressType(podIP)
	if err != nil {
		return nil, err
	}

	output := &store.Endpoint{
		Id:           store.ObjectID(),
		ContainerId:  container.Id,
		PodName:      pod.K8.Name,
		PodNamespace: pod.K8.Namespace,
		Name:         fmt.Sprintf("%s::%s::%s::%d", pod.K8.Namespace, pod.K8.Name, port.Protocol, port.ContainerPort),
		NodeName:     pod.K8.Spec.NodeName,
		AddressType:  addrType,
		Backend: discoveryv1.Endpoint{
			Addresses: []string{podIP},
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
			Port:     &port.ContainerPort,
		},
		Ownership: container.Ownership,
	}

	if len(pod.K8.Namespace) != 0 {
		output.IsNamespaced = true
		output.Namespace = pod.K8.Namespace
	}

	switch {
	case len(port.Name) != 0:
		output.ServiceName = port.Name
	case port.HostPort != 0:
		output.ServiceName = fmt.Sprintf("%s::%d", port.Protocol, port.HostPort)
	default:
		output.ServiceName = fmt.Sprintf("%s::%d", port.Protocol, port.ContainerPort)
	}

	if port.HostPort != 0 {
		// With a host port field, endpoint is only accessible from the node IP
		output.Exposure = shared.EndpointExposureNodeIP

		// TODO future improvement - consider providing the node address as a backend here
	} else {
		// Without a host port field, endpoint is only accessible from within the cluster on the node IP
		output.Exposure = shared.EndpointExposureClusterIP
	}

	return output, nil
}
