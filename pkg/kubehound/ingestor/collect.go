package ingestor

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/ingestor/preflight"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/converter"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	corev1 "k8s.io/api/core/v1"
)

// Collect runs the collection phase: stream K8s objects from the collector,
// convert to store models, and write to SQLite. No graph writes happen here.
func Collect(ctx context.Context, cfg *config.KubehoundConfig, collect collector.CollectorClient, storeDB storedb.Provider) error {
	db, ok := storeDB.Reader().(*sql.DB)
	if !ok {
		return fmt.Errorf("store provider reader is not *sql.DB")
	}
	conv := converter.NewStoreWithDB(cfg, db)

	// Dependency-respecting order:
	// 1. Nodes (no deps, but need identities for user_id lookup)
	// 2. Roles + ClusterRoles (no deps)
	// 3. RoleBindings + ClusterRoleBindings (need roles -> create identities, permissionsets)
	// 4. Nodes again? No - NodeIdentity does a lookup that may miss. Process identities before nodes.
	//    Actually, the original pipeline ran roles/bindings in parallel with nodes.
	//    Node.UserId lookup uses identities, which come from bindings.
	//    For correct lookups: roles -> bindings (creates identities) -> nodes -> endpoints -> pods
	steps := []struct {
		name string
		fn   func() error
	}{
		{"roles", func() error { return collect.StreamRoles(ctx, &roleCollector{conv, storeDB}) }},
		{"clusterroles", func() error { return collect.StreamClusterRoles(ctx, &clusterRoleCollector{conv, storeDB}) }},
		{"rolebindings", func() error { return collect.StreamRoleBindings(ctx, &roleBindingCollector{conv, storeDB, db}) }},
		{"clusterrolebindings", func() error {
			return collect.StreamClusterRoleBindings(ctx, &clusterRoleBindingCollector{conv, storeDB, db})
		}},
		{"nodes", func() error { return collect.StreamNodes(ctx, &nodeCollector{conv, storeDB}) }},
		{"endpoints", func() error { return collect.StreamEndpoints(ctx, &endpointCollector{conv, storeDB}) }},
		{"pods", func() error { return collect.StreamPods(ctx, &podCollector{conv, storeDB, db}) }},
	}
	for _, s := range steps {
		log.Trace(ctx).Infof("Collecting %s", s.name)
		if err := s.fn(); err != nil {
			return fmt.Errorf("collecting %s: %w", s.name, err)
		}
	}
	return nil
}

// --- Node collector ---

type nodeCollector struct {
	conv  *converter.StoreConverter
	store storedb.Provider
}

func (c *nodeCollector) IngestNode(ctx context.Context, node types.NodeType) error {
	if ok, err := preflight.CheckNode(node); !ok {
		return err
	}
	o, err := c.conv.Node(ctx, node)
	if err != nil {
		return err
	}
	return c.store.Write(ctx, o)
}

func (c *nodeCollector) Complete(_ context.Context) error { return nil }

// --- Role collector ---

type roleCollector struct {
	conv  *converter.StoreConverter
	store storedb.Provider
}

func (c *roleCollector) IngestRole(ctx context.Context, role types.RoleType) error {
	if ok, err := preflight.CheckRole(role); !ok {
		return err
	}
	o, err := c.conv.Role(ctx, role)
	if err != nil {
		return err
	}
	return c.store.Write(ctx, o)
}

func (c *roleCollector) Complete(_ context.Context) error { return nil }

// --- ClusterRole collector ---

type clusterRoleCollector struct {
	conv  *converter.StoreConverter
	store storedb.Provider
}

func (c *clusterRoleCollector) IngestClusterRole(ctx context.Context, role types.ClusterRoleType) error {
	if ok, err := preflight.CheckClusterRole(role); !ok {
		return err
	}
	o, err := c.conv.ClusterRole(ctx, role)
	if err != nil {
		return err
	}
	return c.store.Write(ctx, o)
}

func (c *clusterRoleCollector) Complete(_ context.Context) error { return nil }

// --- RoleBinding collector ---

type roleBindingCollector struct {
	conv  *converter.StoreConverter
	store storedb.Provider
	db    *sql.DB
}

func (c *roleBindingCollector) IngestRoleBinding(ctx context.Context, rb types.RoleBindingType) error {
	if ok, err := preflight.CheckRoleBinding(rb); !ok {
		return err
	}

	o, err := c.conv.RoleBinding(ctx, rb)
	if err != nil {
		if errors.Is(err, converter.ErrDanglingRoleBinding) {
			log.Trace(ctx).Debugf("Role binding dropped (%s::%s): %s", rb.Namespace, rb.Name, err.Error())
			return nil
		}
		return err
	}

	if err := c.store.Write(ctx, o); err != nil {
		return err
	}

	// Process subjects as identity objects
	for _, subj := range o.Subjects {
		s := subj
		if err := c.processSubject(ctx, &s, o); err != nil {
			return err
		}
	}

	// Create permission set from role binding
	err = c.createPermissionSet(ctx, o)
	switch {
	case err == nil:
	case errors.Is(err, converter.ErrRoleCacheMiss):
		fallthrough
	case errors.Is(err, converter.ErrRoleBindProperties):
		log.Trace(ctx).Debugf("Permission set dropped (%s::%s): %v", rb.Namespace, rb.Name, err)
	default:
		return err
	}

	return nil
}

func (c *roleBindingCollector) processSubject(ctx context.Context, subj *store.BindSubject, parent *store.RoleBinding) error {
	sid, err := c.conv.Identity(ctx, subj, parent)
	if err != nil {
		return err
	}

	if identityExists(c.db, ctx, sid.Name, sid.Namespace) {
		log.Trace(ctx).Debugf("identity %s/%s already exists, skipping insert", sid.Namespace, sid.Name)
		return nil
	}

	return c.store.Write(ctx, sid)
}

func (c *roleBindingCollector) createPermissionSet(ctx context.Context, rb *store.RoleBinding) error {
	o, err := c.conv.PermissionSet(ctx, rb)
	if err != nil {
		return err
	}
	return c.store.Write(ctx, o)
}

func (c *roleBindingCollector) Complete(_ context.Context) error { return nil }

// --- ClusterRoleBinding collector ---

type clusterRoleBindingCollector struct {
	conv  *converter.StoreConverter
	store storedb.Provider
	db    *sql.DB
}

func (c *clusterRoleBindingCollector) IngestClusterRoleBinding(ctx context.Context, crb types.ClusterRoleBindingType) error {
	if ok, err := preflight.CheckClusterRoleBinding(crb); !ok {
		return err
	}

	o, err := c.conv.ClusterRoleBinding(ctx, crb)
	if err != nil {
		if errors.Is(err, converter.ErrDanglingRoleBinding) {
			log.Trace(ctx).Debugf("Cluster role binding dropped: %s: %s", err.Error(), crb.Name)
			return nil
		}
		return err
	}

	if err := c.store.Write(ctx, o); err != nil {
		return err
	}

	// Process subjects as identity objects
	for _, subj := range o.Subjects {
		s := subj
		if err := c.processSubject(ctx, &s, o); err != nil {
			return err
		}
	}

	// Create permission set from cluster role binding
	err = c.createPermissionSet(ctx, o)
	switch {
	case err == nil:
	case errors.Is(err, converter.ErrRoleCacheMiss):
		fallthrough
	case errors.Is(err, converter.ErrRoleBindProperties):
		log.Trace(ctx).Debugf("Permission set dropped (%s::%s): %v", crb.Namespace, crb.Name, err)
	default:
		return err
	}

	return nil
}

func (c *clusterRoleBindingCollector) processSubject(ctx context.Context, subj *store.BindSubject, parent *store.RoleBinding) error {
	sid, err := c.conv.Identity(ctx, subj, parent)
	if err != nil {
		return err
	}

	if identityExists(c.db, ctx, sid.Name, sid.Namespace) {
		log.Trace(ctx).Debugf("identity %s/%s already exists, skipping insert", sid.Namespace, sid.Name)
		return nil
	}

	return c.store.Write(ctx, sid)
}

func (c *clusterRoleBindingCollector) createPermissionSet(ctx context.Context, crb *store.RoleBinding) error {
	o, err := c.conv.PermissionSetCluster(ctx, crb)
	if err != nil {
		return err
	}
	return c.store.Write(ctx, o)
}

func (c *clusterRoleBindingCollector) Complete(_ context.Context) error { return nil }

// --- Endpoint collector ---

type endpointCollector struct {
	conv  *converter.StoreConverter
	store storedb.Provider
}

func (c *endpointCollector) IngestEndpoint(ctx context.Context, eps types.EndpointType) error {
	if ok, err := preflight.CheckEndpoint(ctx, eps); !ok {
		return err
	}

	for _, port := range eps.Ports {
		for _, addr := range eps.Endpoints {
			o, err := c.conv.Endpoint(ctx, addr, port, eps)
			if err != nil {
				if errors.Is(err, converter.ErrEndpointTarget) {
					log.Trace(ctx).Debugf("Endpoint dropped: %s: %s", err.Error(), addr.TargetRef)
					return nil
				}
				return err
			}
			if err := c.store.Write(ctx, o); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *endpointCollector) Complete(_ context.Context) error { return nil }

// --- Pod collector ---

type podCollector struct {
	conv  *converter.StoreConverter
	store storedb.Provider
	db    *sql.DB
}

func (c *podCollector) IngestPod(ctx context.Context, pod types.PodType) error {
	if ok, err := preflight.CheckPod(ctx, pod); !ok {
		return err
	}

	sp, err := c.conv.Pod(ctx, pod)
	if err != nil {
		log.Trace(ctx).Warnf("process pod %s error (continuing): %v", pod.Name, err)
		return nil
	}

	if err := c.store.Write(ctx, sp); err != nil {
		return err
	}

	// Handle containers
	for _, container := range pod.Spec.Containers {
		cont := container
		if err := c.processContainer(ctx, sp, pod, &cont); err != nil {
			return err
		}
	}

	return nil
}

func (c *podCollector) processContainer(ctx context.Context, parent *store.Pod, k8sPod types.PodType, container types.ContainerType) error {
	if ok, err := preflight.CheckContainer(container); !ok {
		return err
	}

	sc, err := c.conv.Container(ctx, container, parent)
	if err != nil {
		return err
	}

	if err := c.store.Write(ctx, sc); err != nil {
		return err
	}

	// Handle volume mounts
	for _, volumeMount := range container.VolumeMounts {
		vm := volumeMount
		if err := c.processVolumeMount(ctx, &vm, k8sPod.Spec.Volumes, parent, sc); err != nil {
			return err
		}
	}

	// Handle endpoints (derived from container ports)
	for _, port := range container.Ports {
		p := port
		if err := c.processEndpoints(ctx, &p, parent, sc); err != nil {
			return err
		}
	}

	return nil
}

func (c *podCollector) processVolumeMount(ctx context.Context, volumeMount types.VolumeMountType, k8sVolumes []corev1.Volume, pod *store.Pod, container *store.Container) error {
	if ok, err := preflight.CheckVolume(volumeMount); !ok {
		return err
	}

	sv, err := c.conv.VolumeFromK8s(ctx, volumeMount, k8sVolumes, pod, container)
	if err != nil {
		log.Trace(ctx).Debugf("process volume type: %v (continuing)", err)
		return nil
	}

	return c.store.Write(ctx, sv)
}

func (c *podCollector) processEndpoints(ctx context.Context, port *corev1.ContainerPort, pod *store.Pod, container *store.Container) error {
	tmp, err := c.conv.EndpointPrivate(ctx, port, pod, container)
	if err != nil {
		return err
	}

	if endpointSliceExists(c.db, ctx, tmp.Namespace, tmp.PodName, tmp.SafeProtocol(), tmp.SafePort()) {
		if port.HostPort != 0 && port.ContainerPort != port.HostPort {
			log.Trace(ctx).Warnf("assumption failure: host port set on container with associated endpoint slice (%s::%s::%s::%d)",
				tmp.Namespace, tmp.PodName, tmp.SafeProtocol(), tmp.SafePort())
		}
		return nil
	}

	return c.store.Write(ctx, tmp)
}

func (c *podCollector) Complete(_ context.Context) error { return nil }

// --- Helpers ---

func identityExists(db *sql.DB, ctx context.Context, name, namespace string) bool {
	if db == nil {
		return false
	}
	var id int64
	err := db.QueryRowContext(ctx,
		"SELECT id FROM identities WHERE name = ? AND namespace = ? LIMIT 1",
		name, namespace).Scan(&id)
	return err == nil
}

func endpointSliceExists(db *sql.DB, ctx context.Context, namespace, podName, protocol string, port int) bool {
	if db == nil {
		return false
	}
	var id int64
	err := db.QueryRowContext(ctx,
		"SELECT id FROM endpoints WHERE namespace = ? AND pod_name = ? AND protocol = ? AND port = ? AND has_slice = 1 LIMIT 1",
		namespace, podName, protocol, port).Scan(&id)
	return err == nil
}
