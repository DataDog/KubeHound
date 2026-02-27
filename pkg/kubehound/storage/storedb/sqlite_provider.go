package storedb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/store/collections"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	_ "modernc.org/sqlite"
)

var _ Provider = (*SQLiteProvider)(nil)

// SQLiteProvider implements the storedb.Provider interface using SQLite.
type SQLiteProvider struct {
	db *sql.DB
}

// NewSQLiteProvider creates a new SQLite provider instance.
func NewSQLiteProvider(ctx context.Context, cfg *config.KubehoundConfig) (*SQLiteProvider, error) {
	path := cfg.SQLite.Path
	if path == "" {
		path = config.DefaultSQLitePath
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening sqlite database %s: %w", path, err)
	}

	// Set connection pool to 1 for SQLite (serialized writes)
	db.SetMaxOpenConns(1)

	// Run PRAGMAs
	for _, line := range strings.Split(schemaPragmas, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if _, err := db.ExecContext(ctx, line); err != nil {
			db.Close()
			return nil, fmt.Errorf("sqlite pragma %q: %w", line, err)
		}
	}

	return &SQLiteProvider{db: db}, nil
}

// NewSQLiteProviderFromDB creates a provider from an existing *sql.DB handle (e.g. for testing).
func NewSQLiteProviderFromDB(db *sql.DB) *SQLiteProvider {
	return &SQLiteProvider{db: db}
}

func (p *SQLiteProvider) Name() string {
	return "sqlite"
}

func (p *SQLiteProvider) HealthCheck(ctx context.Context) (bool, error) {
	if err := p.db.PingContext(ctx); err != nil {
		return false, err
	}
	return true, nil
}

// Prepare creates all tables and indices.
func (p *SQLiteProvider) Prepare(ctx context.Context) error {
	l := log.Logger(ctx)

	l.Info("Creating SQLite schema")
	if _, err := p.db.ExecContext(ctx, schemaDDL); err != nil {
		return fmt.Errorf("sqlite schema creation: %w", err)
	}

	l.Info("Creating SQLite indices")
	if _, err := p.db.ExecContext(ctx, schemaIndices); err != nil {
		return fmt.Errorf("sqlite index creation: %w", err)
	}

	return nil
}

// Clean deletes data for a specific run/cluster combination.
func (p *SQLiteProvider) Clean(ctx context.Context, runID string, clusterName string) error {
	tables := collections.GetCollections()
	for _, table := range tables {
		query := fmt.Sprintf("DELETE FROM %s WHERE run_id = ? AND cluster_name = ?", table)
		if _, err := p.db.ExecContext(ctx, query, runID, clusterName); err != nil {
			return fmt.Errorf("sqlite clean table %s: %w", table, err)
		}
	}

	return nil
}

// Reader returns the underlying *sql.DB handle for queries.
func (p *SQLiteProvider) Reader() any {
	return p.db
}

// Write performs a synchronous INSERT OR IGNORE for the given store model.
func (p *SQLiteProvider) Write(ctx context.Context, model any) error {
	switch m := model.(type) {
	case *store.Node:
		return p.insertNode(ctx, m)
	case *store.Pod:
		return p.insertPod(ctx, m)
	case *store.Container:
		return p.insertContainer(ctx, m)
	case *store.Volume:
		return p.insertVolume(ctx, m)
	case *store.Role:
		return p.insertRole(ctx, m)
	case *store.RoleBinding:
		return p.insertRoleBinding(ctx, m)
	case *store.Identity:
		return p.insertIdentity(ctx, m)
	case *store.PermissionSet:
		return p.insertPermissionSet(ctx, m)
	case *store.Endpoint:
		return p.insertEndpoint(ctx, m)
	default:
		return fmt.Errorf("sqlite write: unsupported model type %T", model)
	}
}

// Close closes the database connection.
func (p *SQLiteProvider) Close(ctx context.Context) error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

func (p *SQLiteProvider) insertNode(ctx context.Context, m *store.Node) error {
	_, err := p.db.ExecContext(ctx, `INSERT OR IGNORE INTO nodes
		(id, user_id, is_namespaced, name, namespace, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.UserId, boolToInt(m.IsNamespaced), m.Name, m.Namespace,
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (p *SQLiteProvider) insertPod(ctx context.Context, m *store.Pod) error {
	_, err := p.db.ExecContext(ctx, `INSERT OR IGNORE INTO pods
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

func (p *SQLiteProvider) insertContainer(ctx context.Context, m *store.Container) error {
	capsJSON, _ := json.Marshal(m.Capabilities)
	cmdJSON, _ := json.Marshal(m.Command)
	argsJSON, _ := json.Marshal(m.Args)
	portsJSON, _ := json.Marshal(m.Ports)

	_, err := p.db.ExecContext(ctx, `INSERT OR IGNORE INTO containers
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

func (p *SQLiteProvider) insertVolume(ctx context.Context, m *store.Volume) error {
	_, err := p.db.ExecContext(ctx, `INSERT OR IGNORE INTO volumes
		(id, pod_id, node_id, container_id, projected_id, name, type, source, mount, target_name, target_namespace, readonly, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.PodId, m.NodeId, m.ContainerId, m.ProjectedId,
		m.Name, m.Type, m.SourcePath, m.MountPath, m.TargetName, m.TargetNamespace,
		boolToInt(m.ReadOnly),
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (p *SQLiteProvider) insertRole(ctx context.Context, m *store.Role) error {
	rulesJSON, _ := json.Marshal(m.Rules)
	_, err := p.db.ExecContext(ctx, `INSERT OR IGNORE INTO roles
		(id, name, is_namespaced, namespace, rules, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.Name, boolToInt(m.IsNamespaced), m.Namespace, string(rulesJSON),
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (p *SQLiteProvider) insertRoleBinding(ctx context.Context, m *store.RoleBinding) error {
	subjJSON, _ := json.Marshal(m.Subjects)
	_, err := p.db.ExecContext(ctx, `INSERT OR IGNORE INTO rolebindings
		(id, name, role_id, is_namespaced, namespace, subjects, roleref_kind, roleref_name, roleref_api_group, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.Name, m.RoleId, boolToInt(m.IsNamespaced), m.Namespace,
		string(subjJSON), m.RoleRef.Kind, m.RoleRef.Name, m.RoleRef.APIGroup,
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (p *SQLiteProvider) insertIdentity(ctx context.Context, m *store.Identity) error {
	_, err := p.db.ExecContext(ctx, `INSERT OR IGNORE INTO identities
		(id, name, is_namespaced, namespace, type, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.Name, boolToInt(m.IsNamespaced), m.Namespace, m.Type,
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (p *SQLiteProvider) insertPermissionSet(ctx context.Context, m *store.PermissionSet) error {
	rulesJSON, _ := json.Marshal(m.Rules)
	_, err := p.db.ExecContext(ctx, `INSERT OR IGNORE INTO permissionsets
		(id, role_id, role_name, role_binding_id, role_binding_name, name, is_namespaced, namespace, rules, app, team, service, run_id, cluster_name, cluster_version_major, cluster_version_minor)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		m.Id, m.RoleId, m.RoleName, m.RoleBindingId, m.RoleBindingName,
		m.Name, boolToInt(m.IsNamespaced), m.Namespace, string(rulesJSON),
		m.Ownership.Application, m.Ownership.Team, m.Ownership.Service,
		m.Runtime.RunID, m.Runtime.Cluster.Name, m.Runtime.Cluster.VersionMajor, m.Runtime.Cluster.VersionMinor)
	return err
}

func (p *SQLiteProvider) insertEndpoint(ctx context.Context, m *store.Endpoint) error {
	addrJSON, _ := json.Marshal(m.Addresses)
	_, err := p.db.ExecContext(ctx, `INSERT OR IGNORE INTO endpoints
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
