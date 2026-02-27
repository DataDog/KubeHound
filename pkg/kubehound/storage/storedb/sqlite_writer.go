package storedb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
)

var _ AsyncWriter = (*SQLiteWriter)(nil)

// SQLiteWriter implements AsyncWriter with immediate synchronous INSERT OR IGNORE.
type SQLiteWriter struct {
	db         *sql.DB
	collection collections.Collection
}

// NewSQLiteWriter creates a new synchronous SQLite writer.
func NewSQLiteWriter(db *sql.DB, collection collections.Collection) *SQLiteWriter {
	return &SQLiteWriter{
		db:         db,
		collection: collection,
	}
}

// Queue performs an immediate INSERT OR IGNORE. No batching.
func (w *SQLiteWriter) Queue(ctx context.Context, model any) error {
	switch m := model.(type) {
	case *store.Node:
		return w.insertNode(ctx, m)
	case *store.Pod:
		return w.insertPod(ctx, m)
	case *store.Container:
		return w.insertContainer(ctx, m)
	case *store.Volume:
		return w.insertVolume(ctx, m)
	case *store.Role:
		return w.insertRole(ctx, m)
	case *store.RoleBinding:
		return w.insertRoleBinding(ctx, m)
	case *store.Identity:
		return w.insertIdentity(ctx, m)
	case *store.PermissionSet:
		return w.insertPermissionSet(ctx, m)
	case *store.Endpoint:
		return w.insertEndpoint(ctx, m)
	default:
		return fmt.Errorf("sqlite writer: unsupported model type %T", model)
	}
}

// Flush is a no-op for synchronous writes.
func (w *SQLiteWriter) Flush(_ context.Context) error {
	return nil
}

// Close is a no-op — the database connection is owned by the provider.
func (w *SQLiteWriter) Close(_ context.Context) error {
	return nil
}

func (w *SQLiteWriter) insertNode(ctx context.Context, m *store.Node) error {
	_, err := w.db.ExecContext(ctx, `INSERT OR IGNORE INTO nodes
		(id, user_id, is_namespaced, name, namespace, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.UserId, boolToInt(m.IsNamespaced), m.Name, m.Namespace,
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (w *SQLiteWriter) insertPod(ctx context.Context, m *store.Pod) error {
	_, err := w.db.ExecContext(ctx, `INSERT OR IGNORE INTO pods
		(id, node_id, is_namespaced, name, namespace, node_name, service_account, host_pid, host_ipc, host_network, share_process_namespace, pod_ip, uid, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.NodeId, boolToInt(m.IsNamespaced), m.Name, m.Namespace,
		m.NodeName, m.ServiceAccount,
		boolToInt(m.HostPID), boolToInt(m.HostIPC), boolToInt(m.HostNetwork),
		boolToInt(m.ShareProcessNamespace), m.PodIP, m.UID,
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (w *SQLiteWriter) insertContainer(ctx context.Context, m *store.Container) error {
	capsJSON, _ := json.Marshal(m.Capabilities)
	cmdJSON, _ := json.Marshal(m.Command)
	argsJSON, _ := json.Marshal(m.Args)
	portsJSON, _ := json.Marshal(m.Ports)

	_, err := w.db.ExecContext(ctx, `INSERT OR IGNORE INTO containers
		(id, pod_id, node_id, name, image, command, args, capabilities_add, privileged, priv_esc, host_pid, host_ipc, host_network, run_as_user, pod_name, node_name, namespace, service_account, ports, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.PodId, m.NodeId,
		m.Name, m.Image, string(cmdJSON), string(argsJSON),
		string(capsJSON), boolToInt(m.Privileged), boolToInt(m.PrivEsc),
		boolToInt(m.Inherited.HostPID), boolToInt(m.Inherited.HostIPC), boolToInt(m.Inherited.HostNetwork),
		m.Inherited.RunAsUser,
		m.Inherited.PodName, m.Inherited.NodeName, m.Inherited.Namespace, m.Inherited.ServiceAccount,
		string(portsJSON),
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (w *SQLiteWriter) insertVolume(ctx context.Context, m *store.Volume) error {
	_, err := w.db.ExecContext(ctx, `INSERT OR IGNORE INTO volumes
		(id, pod_id, node_id, container_id, projected_id, name, type, source, mount, target_name, target_namespace, readonly, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.PodId, m.NodeId, m.ContainerId, m.ProjectedId,
		m.Name, m.Type, m.SourcePath, m.MountPath, m.TargetName, m.TargetNamespace,
		boolToInt(m.ReadOnly),
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (w *SQLiteWriter) insertRole(ctx context.Context, m *store.Role) error {
	rulesJSON, _ := json.Marshal(m.Rules)
	_, err := w.db.ExecContext(ctx, `INSERT OR IGNORE INTO roles
		(id, name, is_namespaced, namespace, rules, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.Name, boolToInt(m.IsNamespaced), m.Namespace, string(rulesJSON),
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (w *SQLiteWriter) insertRoleBinding(ctx context.Context, m *store.RoleBinding) error {
	subjJSON, _ := json.Marshal(m.Subjects)
	_, err := w.db.ExecContext(ctx, `INSERT OR IGNORE INTO rolebindings
		(id, name, role_id, is_namespaced, namespace, subjects, roleref_kind, roleref_name, roleref_api_group, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.Name, m.RoleId, boolToInt(m.IsNamespaced), m.Namespace,
		string(subjJSON), m.RoleRef.Kind, m.RoleRef.Name, m.RoleRef.APIGroup,
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (w *SQLiteWriter) insertIdentity(ctx context.Context, m *store.Identity) error {
	_, err := w.db.ExecContext(ctx, `INSERT OR IGNORE INTO identities
		(id, name, is_namespaced, namespace, type, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.Name, boolToInt(m.IsNamespaced), m.Namespace, m.Type,
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (w *SQLiteWriter) insertPermissionSet(ctx context.Context, m *store.PermissionSet) error {
	rulesJSON, _ := json.Marshal(m.Rules)
	_, err := w.db.ExecContext(ctx, `INSERT OR IGNORE INTO permissionsets
		(id, role_id, role_name, role_binding_id, role_binding_name, name, is_namespaced, namespace, rules, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.RoleId, m.RoleName, m.RoleBindingId, m.RoleBindingName,
		m.Name, boolToInt(m.IsNamespaced), m.Namespace, string(rulesJSON),
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (w *SQLiteWriter) insertEndpoint(ctx context.Context, m *store.Endpoint) error {
	addrJSON, _ := json.Marshal(m.Addresses)
	_, err := w.db.ExecContext(ctx, `INSERT OR IGNORE INTO endpoints
		(id, container_id, pod_name, pod_namespace, node_name, is_namespaced, namespace, name, has_slice, service_name, service_dns, address_type, addresses, port, port_name, protocol, exposure, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.ContainerId,
		m.PodName, m.PodNamespace, m.NodeName,
		boolToInt(m.IsNamespaced), m.Namespace, m.Name,
		boolToInt(m.HasSlice), m.ServiceName, m.ServiceDns,
		string(m.AddressType), string(addrJSON),
		m.Port, m.PortName, m.Protocol, int(m.Exposure),
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
