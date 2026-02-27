package converter

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/libkube"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
)

const (
	EmptyNamespace = ""
)

var (
	ErrUnsupportedVolume     = errors.New("provided volume is not currently supported")
	ErrNoDBInitialized       = errors.New("database required for conversion")
	ErrDanglingRoleBinding   = errors.New("role binding found with no matching role")
	ErrProjectedDefaultToken = errors.New("projected volume grant no access (default serviceaccount)")
	ErrEndpointTarget        = errors.New("target reference for an endpoint could not be resolved")
	ErrRoleCacheMiss         = errors.New("missing role in cache")
	ErrRoleBindProperties    = errors.New("incorrect combination of (cluster) role and (cluster) role binding properties")
)

// StoreConverter enables converting between an input K8s model to its equivalent store model.
type StoreConverter struct {
	db      *sql.DB
	runtime *config.DynamicConfig
}

// NewStore returns a new store converter instance.
func NewStore(cfg *config.KubehoundConfig) *StoreConverter {
	return &StoreConverter{
		runtime: &cfg.Dynamic,
	}
}

// NewStoreWithDB returns a new store converter instance with read access to the SQLite database.
func NewStoreWithDB(cfg *config.KubehoundConfig, db *sql.DB) *StoreConverter {
	return &StoreConverter{
		db:      db,
		runtime: &cfg.Dynamic,
	}
}

// queryNodeID looks up a node ID by name from the SQLite database.
func (c *StoreConverter) queryNodeID(ctx context.Context, name string) (int64, error) {
	var id int64
	err := c.db.QueryRowContext(ctx,
		"SELECT id FROM nodes WHERE name = ? AND run_id = ? AND cluster_name = ? LIMIT 1",
		name, c.runtime.RunID.String(), c.runtime.Cluster.Name).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("node lookup (name=%s): %w", name, err)
	}
	return id, nil
}

// queryRole looks up a role by name and namespace from the SQLite database.
func (c *StoreConverter) queryRole(ctx context.Context, name, namespace string) (*store.Role, error) {
	var role store.Role
	var rulesJSON string
	err := c.db.QueryRowContext(ctx,
		"SELECT id, name, is_namespaced, namespace, rules FROM roles WHERE name = ? AND namespace = ? AND run_id = ? AND cluster_name = ? LIMIT 1",
		name, namespace, c.runtime.RunID.String(), c.runtime.Cluster.Name).Scan(
		&role.Id, &role.Name, &role.IsNamespaced, &role.Namespace, &rulesJSON)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(rulesJSON), &role.Rules); err != nil {
		return nil, fmt.Errorf("unmarshal role rules: %w", err)
	}
	return &role, nil
}

// queryIdentityID looks up an identity ID by name and namespace from the SQLite database.
func (c *StoreConverter) queryIdentityID(ctx context.Context, name, namespace string) (int64, error) {
	var id int64
	err := c.db.QueryRowContext(ctx,
		"SELECT id FROM identities WHERE name = ? AND namespace = ? AND run_id = ? AND cluster_name = ? LIMIT 1",
		name, namespace, c.runtime.RunID.String(), c.runtime.Cluster.Name).Scan(&id)
	return id, err
}

// Container returns the store representation of a K8s container from an input K8s container object.
func (c *StoreConverter) Container(_ context.Context, input types.ContainerType, parent *store.Pod) (*store.Container, error) {
	output := &store.Container{
		Id:     store.ObjectID(),
		PodId:  parent.Id,
		NodeId: parent.NodeId,
		Name:   input.Name,
		Image:  input.Image,
		Inherited: store.ContainerInherited{
			PodName:        parent.Name,
			NodeName:       parent.NodeName,
			Namespace:      parent.Namespace,
			HostPID:        parent.HostPID,
			HostIPC:        parent.HostIPC,
			HostNetwork:    parent.HostNetwork,
			ServiceAccount: parent.ServiceAccount,
		},
		Ownership: parent.Ownership,
		Runtime:   store.Runtime(c.runtime),
	}

	// Copy command and args
	output.Command = input.Command
	output.Args = input.Args

	// Certain fields are set by the PodSecurityContext and overridden by the container's SecurityContext.
	if input.SecurityContext != nil && input.SecurityContext.RunAsUser != nil {
		output.Inherited.RunAsUser = *input.SecurityContext.RunAsUser
	}

	// Privileged
	if input.SecurityContext != nil && input.SecurityContext.Privileged != nil {
		output.Privileged = *input.SecurityContext.Privileged
	}

	// Privilege escalation
	if input.SecurityContext != nil && input.SecurityContext.AllowPrivilegeEscalation != nil {
		output.PrivEsc = *input.SecurityContext.AllowPrivilegeEscalation
	}

	// Capabilities
	output.Capabilities = make([]string, 0)
	if input.SecurityContext != nil && input.SecurityContext.Capabilities != nil {
		for _, cap := range input.SecurityContext.Capabilities.Add {
			output.Capabilities = append(output.Capabilities, string(cap))
		}
	}

	// Ports
	output.Ports = make([]int32, 0, len(input.Ports))
	for _, p := range input.Ports {
		output.Ports = append(output.Ports, p.ContainerPort)
	}

	return output, nil
}

// Node returns the store representation of a K8s node from an input K8s node object.
func (c *StoreConverter) Node(ctx context.Context, input types.NodeType) (*store.Node, error) {
	if c.db == nil {
		return nil, ErrNoDBInitialized
	}

	output := &store.Node{
		Id:        store.ObjectID(),
		Name:      input.Name,
		Namespace: input.Namespace,
		Ownership: store.ExtractOwnership(input.Labels),
		Runtime:   store.Runtime(c.runtime),
	}

	if len(input.Namespace) != 0 {
		output.IsNamespaced = true
	}

	// Retrieve the associated identity store ID from the database
	uid, err := libkube.NodeIdentity(ctx, c.db, input.Name, c.runtime.RunID.String(), c.runtime.Cluster.Name)
	switch {
	case err == nil:
		output.UserId = uid
	case errors.Is(err, libkube.ErrMissingNodeUser):
		// Most nodes run under a default account with no permissions which we ignore.
	default:
		return nil, err
	}

	return output, nil
}

// Pod returns the store representation of a K8s pod from an input K8s pod object.
// NOTE: requires database access (node lookup).
func (c *StoreConverter) Pod(ctx context.Context, input types.PodType) (*store.Pod, error) {
	if c.db == nil {
		return nil, ErrNoDBInitialized
	}

	nid, err := c.queryNodeID(ctx, input.Spec.NodeName)
	if err != nil {
		return nil, err
	}

	output := &store.Pod{
		Id:             store.ObjectID(),
		NodeId:         nid,
		Name:           input.Name,
		Namespace:      input.Namespace,
		NodeName:       input.Spec.NodeName,
		ServiceAccount: input.Spec.ServiceAccountName,
		HostPID:        input.Spec.HostPID,
		HostIPC:        input.Spec.HostIPC,
		HostNetwork:    input.Spec.HostNetwork,
		PodIP:          input.Status.PodIP,
		UID:            string(input.UID),
		Ownership:      store.ExtractOwnership(input.Labels),
		Runtime:        store.Runtime(c.runtime),
	}

	if input.Spec.ShareProcessNamespace != nil {
		output.ShareProcessNamespace = *input.Spec.ShareProcessNamespace
	}

	if len(input.Namespace) != 0 {
		output.IsNamespaced = true
	}

	return output, nil
}

// handleProjectedToken returns the identity store ID and source path corresponding to a projected token volume mount.
func (c *StoreConverter) handleProjectedToken(ctx context.Context, input types.VolumeMountType,
	volume *corev1.Volume, pod *store.Pod) (int64, string, error) {

	// Retrieve the associated identity store ID from the database
	said, err := c.queryIdentityID(ctx, pod.ServiceAccount, pod.Namespace)
	switch {
	case err == nil:
		// We have a matching identity object in the store, continue to create a volume
	case errors.Is(err, sql.ErrNoRows):
		// Most pods run under a default account with no permissions which we ignore.
		return 0, "", ErrProjectedDefaultToken
	default:
		return 0, "", err
	}

	// Loop through looking for the service account token projection
	var sourcePath string
	for _, proj := range volume.Projected.Sources {
		if proj.ServiceAccountToken != nil {
			sourcePath = libkube.ServiceAccountTokenPath(pod.UID, input.Name)
			break
		}
	}

	return said, sourcePath, nil
}

// Volume returns the store representation of a K8s mounted volume from an input K8s volume object.
// NOTE: requires database access (IdentityKey).
func (c *StoreConverter) Volume(ctx context.Context, input types.VolumeMountType, pod *store.Pod,
	container *store.Container) (*store.Volume, error) {

	if c.db == nil {
		return nil, ErrNoDBInitialized
	}

	output := &store.Volume{
		Id:          store.ObjectID(),
		PodId:       pod.Id,
		NodeId:      pod.NodeId,
		ContainerId: container.Id,
		Name:        input.Name,
		MountPath:   input.MountPath,
		ReadOnly:    input.ReadOnly,
		Ownership:   pod.Ownership,
		Runtime:     store.Runtime(c.runtime),
	}

	return output, nil
}

// VolumeFromK8s resolves a volume mount using the K8s pod spec volumes list.
func (c *StoreConverter) VolumeFromK8s(ctx context.Context, input types.VolumeMountType,
	k8sVolumes []corev1.Volume, pod *store.Pod, container *store.Container) (*store.Volume, error) {

	if c.db == nil {
		return nil, ErrNoDBInitialized
	}

	output := &store.Volume{
		Id:          store.ObjectID(),
		PodId:       pod.Id,
		NodeId:      pod.NodeId,
		ContainerId: container.Id,
		Name:        input.Name,
		MountPath:   input.MountPath,
		ReadOnly:    input.ReadOnly,
		Ownership:   pod.Ownership,
		Runtime:     store.Runtime(c.runtime),
	}

	found := false
	for _, volume := range k8sVolumes {
		v := volume
		if v.Name == input.Name {
			found = true
			switch {
			case v.Secret != nil:
				output.Type = shared.VolumeTypeSecret
				output.TargetName = v.Secret.SecretName
				output.TargetNamespace = pod.Namespace
			case v.HostPath != nil:
				output.Type = shared.VolumeTypeHost
				output.SourcePath = v.HostPath.Path
			case v.Projected != nil:
				said, source, err := c.handleProjectedToken(ctx, input, &v, pod)
				if err != nil {
					return nil, fmt.Errorf("projected token volume (%s) processing: %w", v.Name, err)
				}
				output.Type = shared.VolumeTypeProjected
				output.SourcePath = source
				output.ProjectedId = said
			default:
				return nil, ErrUnsupportedVolume
			}
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
		Ownership:    store.ExtractOwnership(input.Labels),
		Runtime:      store.Runtime(c.runtime),
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
		Ownership:    store.ExtractOwnership(input.Labels),
		Runtime:      store.Runtime(c.runtime),
	}, nil
}

func (c *StoreConverter) convertSubject(ctx context.Context, subj rbacv1.Subject) (store.BindSubject, error) {
	// Check if identity already exists in SQLite and use that ID, otherwise generate a new one
	sid, err := c.queryIdentityID(ctx, subj.Name, subj.Namespace)
	switch {
	case err == nil:
		// Entry already exists, use the existing id value
	case errors.Is(err, sql.ErrNoRows):
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
// NOTE: requires database access (RoleKey).
func (c *StoreConverter) RoleBinding(ctx context.Context, input types.RoleBindingType) (*store.RoleBinding, error) {
	if c.db == nil {
		return nil, ErrNoDBInitialized
	}

	role, err := c.queryRole(ctx, input.RoleRef.Name, input.Namespace)
	if err != nil {
		role, err = c.queryRole(ctx, input.RoleRef.Name, EmptyNamespace)
		if err != nil {
			return nil, ErrDanglingRoleBinding
		}
	}

	subj := input.Subjects
	output := &store.RoleBinding{
		Id:           store.ObjectID(),
		RoleId:       role.Id,
		Name:         input.Name,
		IsNamespaced: true,
		Namespace:    input.Namespace,
		Subjects:     make([]store.BindSubject, 0, len(subj)),
		Ownership:    store.ExtractOwnership(input.Labels),
		Runtime:      store.Runtime(c.runtime),
		RoleRef:      input.RoleRef,
	}

	for _, s := range subj {
		if s.Namespace == "" && s.Kind == rbacv1.ServiceAccountKind {
			s.Namespace = input.Namespace
		}
		s, err := c.convertSubject(ctx, s)
		if err != nil {
			return nil, fmt.Errorf("role binding subject convert: %w", err)
		}
		output.Subjects = append(output.Subjects, s)
	}

	return output, nil
}

// ClusterRoleBinding returns the store representation of a K8s cluster role binding from an input K8s ClusterRoleBinding object.
// NOTE: requires database access (RoleKey).
func (c *StoreConverter) ClusterRoleBinding(ctx context.Context, input types.ClusterRoleBindingType) (*store.RoleBinding, error) {
	if c.db == nil {
		return nil, ErrNoDBInitialized
	}

	role, err := c.queryRole(ctx, input.RoleRef.Name, input.Namespace)
	if err != nil {
		role, err = c.queryRole(ctx, input.RoleRef.Name, EmptyNamespace)
		if err != nil {
			return nil, ErrDanglingRoleBinding
		}
	}

	subj := input.Subjects
	output := &store.RoleBinding{
		Id:           store.ObjectID(),
		RoleId:       role.Id,
		Name:         input.Name,
		IsNamespaced: false,
		Namespace:    "",
		Subjects:     make([]store.BindSubject, 0, len(subj)),
		Ownership:    store.ExtractOwnership(input.Labels),
		Runtime:      store.Runtime(c.runtime),
		RoleRef:      input.RoleRef,
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

// Identity returns the store representation of a K8s identity from an input store BindSubject.
func (c *StoreConverter) Identity(ctx context.Context, input *store.BindSubject, parent *store.RoleBinding) (*store.Identity, error) {
	output := &store.Identity{
		Id:        input.IdentityId,
		Name:      input.Subject.Name,
		Namespace: "",
		Type:      input.Subject.Kind,
		Ownership: parent.Ownership,
		Runtime:   store.Runtime(c.runtime),
	}

	if input.Subject.Kind == "ServiceAccount" && len(input.Subject.Namespace) == 0 {
		if len(parent.Namespace) == 0 {
			log.Trace(ctx).Errorf("Namespace not found for service account (%s), using input(rolebinding) namespace (%s) for PermissionSet (%d)\n", input.Subject.Name, parent.Namespace, input.IdentityId)
		} else {
			output.Namespace = parent.Namespace
			output.IsNamespaced = true
		}
		return output, nil
	}

	if len(input.Subject.Namespace) != 0 {
		output.IsNamespaced = true
		output.Namespace = input.Subject.Namespace
	}

	return output, nil
}

// PermissionSet returns the store representation of a K8s role / rolebinding combination.
func (c *StoreConverter) PermissionSet(ctx context.Context, roleBinding *store.RoleBinding) (*store.PermissionSet, error) {
	if c.db == nil {
		return nil, ErrNoDBInitialized
	}

	if !roleBinding.IsNamespaced {
		return nil, fmt.Errorf("invalid input (%s), use converter.PermissionSetCluster for cluster role bindings", roleBinding.Name)
	}

	var roleName, roleNS string
	if roleBinding.RoleRef.Kind == "ClusterRole" {
		roleName = roleBinding.RoleRef.Name
		roleNS = EmptyNamespace
	} else {
		roleName = roleBinding.RoleRef.Name
		roleNS = roleBinding.Namespace
	}

	role, err := c.queryRole(ctx, roleName, roleNS)
	if err != nil {
		return nil, ErrRoleCacheMiss
	}

	if roleBinding.Namespace != role.Namespace && role.Namespace != EmptyNamespace {
		log.Trace(ctx).Debugf("The role namespace (%s) does not match the rolebinding namespace (%s)",
			role.Namespace, roleBinding.Namespace)
		return nil, ErrRoleBindProperties
	}

	isEffective := false
	for _, s := range roleBinding.Subjects {
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
		Runtime:         store.Runtime(c.runtime),
	}

	if !role.IsNamespaced {
		output.IsNamespaced = true
		output.Namespace = roleBinding.Namespace
	}

	return output, nil
}

// PermissionSetCluster returns the store representation of a K8s cluster role / cluster role binding combination.
func (c *StoreConverter) PermissionSetCluster(ctx context.Context, clusterRoleBinding *store.RoleBinding) (*store.PermissionSet, error) {
	if c.db == nil {
		return nil, ErrNoDBInitialized
	}

	if clusterRoleBinding.IsNamespaced {
		return nil, fmt.Errorf("invalid input (%s), use converter.PermissionSet for role bindings", clusterRoleBinding.Name)
	}

	role, err := c.queryRole(ctx, clusterRoleBinding.RoleRef.Name, clusterRoleBinding.Namespace)
	if err != nil {
		return nil, ErrRoleCacheMiss
	}

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
		Runtime:         store.Runtime(c.runtime),
	}

	return output, nil
}

// Endpoint returns the store representation of a K8s endpoint from an EndpointSlice.
func (c *StoreConverter) Endpoint(_ context.Context, addr discoveryv1.Endpoint,
	port discoveryv1.EndpointPort, parent types.EndpointType) (*store.Endpoint, error) {

	if addr.TargetRef == nil {
		return nil, ErrEndpointTarget
	}

	if addr.TargetRef.Kind != "Pod" {
		return nil, ErrEndpointTarget
	}

	protocol := store.DefaultEndpointProtocol
	if port.Protocol != nil {
		protocol = string(*port.Protocol)
	}

	portNum := 0
	if port.Port != nil {
		portNum = int(*port.Port)
	}

	portName := store.DefaultPortName
	if port.Name != nil {
		portName = *port.Name
	}

	output := &store.Endpoint{
		Id:           store.ObjectID(),
		PodName:      addr.TargetRef.Name,
		PodNamespace: addr.TargetRef.Namespace,
		Name:         fmt.Sprintf("%s::%s::%s", parent.Name, protocol, portName),
		HasSlice:     true,
		ServiceName:  libkube.ServiceName(parent),
		ServiceDns:   libkube.ServiceDns(parent),
		AddressType:  string(parent.AddressType),
		Addresses:    addr.Addresses,
		Port:         portNum,
		PortName:     portName,
		Protocol:     protocol,
		Ownership:    store.ExtractOwnership(parent.Labels),
		Runtime:      store.Runtime(c.runtime),
		Exposure:     shared.EndpointExposureExternal,
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

// EndpointPrivate returns the store representation of a private endpoint (no endpoint slice).
func (c *StoreConverter) EndpointPrivate(_ context.Context, port *corev1.ContainerPort,
	pod *store.Pod, container *store.Container) (*store.Endpoint, error) {

	addrType, err := libkube.AddressType(pod.PodIP)
	if err != nil {
		return nil, err
	}

	output := &store.Endpoint{
		Id:           store.ObjectID(),
		ContainerId:  container.Id,
		PodName:      pod.Name,
		PodNamespace: pod.Namespace,
		Name:         fmt.Sprintf("%s::%s::%s::%d", pod.Namespace, pod.Name, port.Protocol, port.ContainerPort),
		NodeName:     pod.NodeName,
		AddressType:  string(addrType),
		Addresses:    []string{pod.PodIP},
		Port:         int(port.ContainerPort),
		PortName:     port.Name,
		Protocol:     string(port.Protocol),
		Ownership:    container.Ownership,
		Runtime:      store.Runtime(c.runtime),
	}

	if len(pod.Namespace) != 0 {
		output.IsNamespaced = true
		output.Namespace = pod.Namespace
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
		output.Exposure = shared.EndpointExposureNodeIP
	} else {
		output.Exposure = shared.EndpointExposureClusterIP
	}

	return output, nil
}
