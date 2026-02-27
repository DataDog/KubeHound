package storedb

import "database/sql"

// SQLite schema DDL for the KubeHound store database.
// Decomposed columns matching rkh's design — no JSON blob for K8s objects.

type Table string

const (
	TableNodes          Table = "nodes"
	TablePods           Table = "pods"
	TableContainers     Table = "containers"
	TableVolumes        Table = "volumes"
	TableRoles          Table = "roles"
	TableRoleBindings   Table = "rolebindings"
	TableIdentities     Table = "identities"
	TablePermissionSets Table = "permissionsets"
	TableEndpoints      Table = "endpoints"
)

var Tables = []Table{
	TableNodes, TablePods, TableContainers, TableVolumes,
	TableRoles, TableRoleBindings, TableIdentities,
	TablePermissionSets, TableEndpoints,
}

// InitSchema initializes the schema and indices on the given database.
func InitSchema(db *sql.DB) error {
	if _, err := db.Exec(schemaDDL); err != nil {
		return err
	}
	if _, err := db.Exec(schemaIndices); err != nil {
		return err
	}
	return nil
}

const schemaDDL = `
CREATE TABLE IF NOT EXISTS nodes (
	id INTEGER PRIMARY KEY,
	user_id INTEGER NOT NULL DEFAULT 0,
	is_namespaced INTEGER NOT NULL DEFAULT 0,
	name TEXT NOT NULL DEFAULT '',
	namespace TEXT NOT NULL DEFAULT '',
	app TEXT NOT NULL DEFAULT '',
	team TEXT NOT NULL DEFAULT '',
	service TEXT NOT NULL DEFAULT '',
	run_id TEXT NOT NULL DEFAULT '',
	cluster_name TEXT NOT NULL DEFAULT '',
	cluster_version_major TEXT NOT NULL DEFAULT '',
	cluster_version_minor TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS pods (
	id INTEGER PRIMARY KEY,
	node_id INTEGER NOT NULL DEFAULT 0,
	is_namespaced INTEGER NOT NULL DEFAULT 0,
	name TEXT NOT NULL DEFAULT '',
	namespace TEXT NOT NULL DEFAULT '',
	node_name TEXT NOT NULL DEFAULT '',
	service_account TEXT NOT NULL DEFAULT '',
	host_pid INTEGER NOT NULL DEFAULT 0,
	host_ipc INTEGER NOT NULL DEFAULT 0,
	host_network INTEGER NOT NULL DEFAULT 0,
	share_process_namespace INTEGER NOT NULL DEFAULT 0,
	pod_ip TEXT NOT NULL DEFAULT '',
	uid TEXT NOT NULL DEFAULT '',
	app TEXT NOT NULL DEFAULT '',
	team TEXT NOT NULL DEFAULT '',
	service TEXT NOT NULL DEFAULT '',
	run_id TEXT NOT NULL DEFAULT '',
	cluster_name TEXT NOT NULL DEFAULT '',
	cluster_version_major TEXT NOT NULL DEFAULT '',
	cluster_version_minor TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS containers (
	id INTEGER PRIMARY KEY,
	pod_id INTEGER NOT NULL DEFAULT 0,
	node_id INTEGER NOT NULL DEFAULT 0,
	name TEXT NOT NULL DEFAULT '',
	image TEXT NOT NULL DEFAULT '',
	command TEXT NOT NULL DEFAULT '[]',
	args TEXT NOT NULL DEFAULT '[]',
	capabilities_add TEXT NOT NULL DEFAULT '[]',
	privileged INTEGER NOT NULL DEFAULT 0,
	priv_esc INTEGER NOT NULL DEFAULT 0,
	host_pid INTEGER NOT NULL DEFAULT 0,
	host_ipc INTEGER NOT NULL DEFAULT 0,
	host_network INTEGER NOT NULL DEFAULT 0,
	run_as_user INTEGER NOT NULL DEFAULT 0,
	pod_name TEXT NOT NULL DEFAULT '',
	node_name TEXT NOT NULL DEFAULT '',
	namespace TEXT NOT NULL DEFAULT '',
	service_account TEXT NOT NULL DEFAULT '',
	ports TEXT NOT NULL DEFAULT '[]',
	app TEXT NOT NULL DEFAULT '',
	team TEXT NOT NULL DEFAULT '',
	service TEXT NOT NULL DEFAULT '',
	run_id TEXT NOT NULL DEFAULT '',
	cluster_name TEXT NOT NULL DEFAULT '',
	cluster_version_major TEXT NOT NULL DEFAULT '',
	cluster_version_minor TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS volumes (
	id INTEGER PRIMARY KEY,
	pod_id INTEGER NOT NULL DEFAULT 0,
	node_id INTEGER NOT NULL DEFAULT 0,
	container_id INTEGER NOT NULL DEFAULT 0,
	projected_id INTEGER NOT NULL DEFAULT 0,
	name TEXT NOT NULL DEFAULT '',
	type TEXT NOT NULL DEFAULT '',
	source TEXT NOT NULL DEFAULT '',
	mount TEXT NOT NULL DEFAULT '',
	target_name TEXT NOT NULL DEFAULT '',
	target_namespace TEXT NOT NULL DEFAULT '',
	readonly INTEGER NOT NULL DEFAULT 0,
	app TEXT NOT NULL DEFAULT '',
	team TEXT NOT NULL DEFAULT '',
	service TEXT NOT NULL DEFAULT '',
	run_id TEXT NOT NULL DEFAULT '',
	cluster_name TEXT NOT NULL DEFAULT '',
	cluster_version_major TEXT NOT NULL DEFAULT '',
	cluster_version_minor TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS roles (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL DEFAULT '',
	is_namespaced INTEGER NOT NULL DEFAULT 0,
	namespace TEXT NOT NULL DEFAULT '',
	rules TEXT NOT NULL DEFAULT '[]',
	app TEXT NOT NULL DEFAULT '',
	team TEXT NOT NULL DEFAULT '',
	service TEXT NOT NULL DEFAULT '',
	run_id TEXT NOT NULL DEFAULT '',
	cluster_name TEXT NOT NULL DEFAULT '',
	cluster_version_major TEXT NOT NULL DEFAULT '',
	cluster_version_minor TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS rolebindings (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL DEFAULT '',
	role_id INTEGER NOT NULL DEFAULT 0,
	is_namespaced INTEGER NOT NULL DEFAULT 0,
	namespace TEXT NOT NULL DEFAULT '',
	subjects TEXT NOT NULL DEFAULT '[]',
	roleref_kind TEXT NOT NULL DEFAULT '',
	roleref_name TEXT NOT NULL DEFAULT '',
	roleref_api_group TEXT NOT NULL DEFAULT '',
	app TEXT NOT NULL DEFAULT '',
	team TEXT NOT NULL DEFAULT '',
	service TEXT NOT NULL DEFAULT '',
	run_id TEXT NOT NULL DEFAULT '',
	cluster_name TEXT NOT NULL DEFAULT '',
	cluster_version_major TEXT NOT NULL DEFAULT '',
	cluster_version_minor TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS identities (
	id INTEGER PRIMARY KEY,
	name TEXT NOT NULL DEFAULT '',
	is_namespaced INTEGER NOT NULL DEFAULT 0,
	namespace TEXT NOT NULL DEFAULT '',
	type TEXT NOT NULL DEFAULT '',
	app TEXT NOT NULL DEFAULT '',
	team TEXT NOT NULL DEFAULT '',
	service TEXT NOT NULL DEFAULT '',
	run_id TEXT NOT NULL DEFAULT '',
	cluster_name TEXT NOT NULL DEFAULT '',
	cluster_version_major TEXT NOT NULL DEFAULT '',
	cluster_version_minor TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS permissionsets (
	id INTEGER PRIMARY KEY,
	role_id INTEGER NOT NULL DEFAULT 0,
	role_name TEXT NOT NULL DEFAULT '',
	role_binding_id INTEGER NOT NULL DEFAULT 0,
	role_binding_name TEXT NOT NULL DEFAULT '',
	name TEXT NOT NULL DEFAULT '',
	is_namespaced INTEGER NOT NULL DEFAULT 0,
	namespace TEXT NOT NULL DEFAULT '',
	rules TEXT NOT NULL DEFAULT '[]',
	app TEXT NOT NULL DEFAULT '',
	team TEXT NOT NULL DEFAULT '',
	service TEXT NOT NULL DEFAULT '',
	run_id TEXT NOT NULL DEFAULT '',
	cluster_name TEXT NOT NULL DEFAULT '',
	cluster_version_major TEXT NOT NULL DEFAULT '',
	cluster_version_minor TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS endpoints (
	id INTEGER PRIMARY KEY,
	container_id INTEGER NOT NULL DEFAULT 0,
	pod_name TEXT NOT NULL DEFAULT '',
	pod_namespace TEXT NOT NULL DEFAULT '',
	node_name TEXT NOT NULL DEFAULT '',
	is_namespaced INTEGER NOT NULL DEFAULT 0,
	namespace TEXT NOT NULL DEFAULT '',
	name TEXT NOT NULL DEFAULT '',
	has_slice INTEGER NOT NULL DEFAULT 0,
	service_name TEXT NOT NULL DEFAULT '',
	service_dns TEXT NOT NULL DEFAULT '',
	address_type TEXT NOT NULL DEFAULT '',
	addresses TEXT NOT NULL DEFAULT '[]',
	port INTEGER NOT NULL DEFAULT 0,
	port_name TEXT NOT NULL DEFAULT '',
	protocol TEXT NOT NULL DEFAULT 'TCP',
	exposure INTEGER NOT NULL DEFAULT 0,
	app TEXT NOT NULL DEFAULT '',
	team TEXT NOT NULL DEFAULT '',
	service TEXT NOT NULL DEFAULT '',
	run_id TEXT NOT NULL DEFAULT '',
	cluster_name TEXT NOT NULL DEFAULT '',
	cluster_version_major TEXT NOT NULL DEFAULT '',
	cluster_version_minor TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS store_graph_id_map (
	store_id TEXT NOT NULL,
	vertex_id INTEGER NOT NULL,
	PRIMARY KEY (store_id)
);
`

const schemaPragmas = `
PRAGMA journal_mode=WAL;
PRAGMA synchronous=NORMAL;
PRAGMA foreign_keys=ON;
PRAGMA busy_timeout=5000;
`

const schemaIndices = `
CREATE INDEX IF NOT EXISTS idx_nodes_name ON nodes(name, run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_nodes_runtime ON nodes(run_id, cluster_name);

CREATE INDEX IF NOT EXISTS idx_pods_node_id ON pods(node_id);
CREATE INDEX IF NOT EXISTS idx_pods_runtime ON pods(run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_pods_namespace ON pods(namespace, run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_pods_share_process ON pods(share_process_namespace, run_id, cluster_name);

CREATE INDEX IF NOT EXISTS idx_containers_pod_id ON containers(pod_id);
CREATE INDEX IF NOT EXISTS idx_containers_node_id ON containers(node_id);
CREATE INDEX IF NOT EXISTS idx_containers_runtime ON containers(run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_containers_namespace ON containers(namespace, run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_containers_sa ON containers(service_account, namespace, run_id, cluster_name);

CREATE INDEX IF NOT EXISTS idx_volumes_pod_id ON volumes(pod_id);
CREATE INDEX IF NOT EXISTS idx_volumes_node_id ON volumes(node_id);
CREATE INDEX IF NOT EXISTS idx_volumes_container_id ON volumes(container_id);
CREATE INDEX IF NOT EXISTS idx_volumes_type ON volumes(type, run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_volumes_runtime ON volumes(run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_volumes_source_type ON volumes(source, type, run_id, cluster_name);

CREATE INDEX IF NOT EXISTS idx_roles_name_ns ON roles(name, namespace, run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_roles_runtime ON roles(run_id, cluster_name);

CREATE INDEX IF NOT EXISTS idx_rolebindings_role_id ON rolebindings(role_id);
CREATE INDEX IF NOT EXISTS idx_rolebindings_runtime ON rolebindings(run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_rolebindings_namespace ON rolebindings(namespace, run_id, cluster_name);

CREATE INDEX IF NOT EXISTS idx_identities_name_ns ON identities(name, namespace, run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_identities_runtime ON identities(run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_identities_type ON identities(type, run_id, cluster_name);

CREATE INDEX IF NOT EXISTS idx_permissionsets_runtime ON permissionsets(run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_permissionsets_namespace ON permissionsets(namespace, run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_permissionsets_role_id ON permissionsets(role_id);

CREATE INDEX IF NOT EXISTS idx_endpoints_runtime ON endpoints(run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_endpoints_has_slice ON endpoints(has_slice, run_id, cluster_name);
CREATE INDEX IF NOT EXISTS idx_endpoints_ns_pod ON endpoints(namespace, pod_name, protocol, port);
`
