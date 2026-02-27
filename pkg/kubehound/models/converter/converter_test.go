package converter

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"testing"

	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/libkube"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
)

var testConfig = &config.KubehoundConfig{
	Dynamic: config.DynamicConfig{
		RunID: config.NewRunID(),
		Cluster: config.DynamicClusterInfo{
			Name: "test-cluster",
		},
	},
}

func testDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	if err := storedb.InitSchema(db); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func insertIdentity(t *testing.T, db *sql.DB, id int64, name, namespace string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO identities (id, name, namespace, run_id, cluster_name) VALUES (?, ?, ?, ?, ?)",
		id, name, namespace, testConfig.Dynamic.RunID.String(), testConfig.Dynamic.Cluster.Name)
	if err != nil {
		t.Fatalf("insert identity: %v", err)
	}
}

func insertNode(t *testing.T, db *sql.DB, id int64, name, namespace string) {
	t.Helper()
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO nodes (id, name, namespace, run_id, cluster_name) VALUES (?, ?, ?, ?, ?)",
		id, name, namespace, testConfig.Dynamic.RunID.String(), testConfig.Dynamic.Cluster.Name)
	if err != nil {
		t.Fatalf("insert node: %v", err)
	}
}

func insertRole(t *testing.T, db *sql.DB, role *store.Role) {
	t.Helper()
	rulesJSON, err := json.Marshal(role.Rules)
	if err != nil {
		t.Fatalf("marshal role rules: %v", err)
	}
	_, err = db.ExecContext(context.Background(),
		"INSERT INTO roles (id, name, is_namespaced, namespace, rules, run_id, cluster_name) VALUES (?, ?, ?, ?, ?, ?, ?)",
		role.Id, role.Name, boolToInt(role.IsNamespaced), role.Namespace, string(rulesJSON),
		testConfig.Dynamic.RunID.String(), testConfig.Dynamic.Cluster.Name)
	if err != nil {
		t.Fatalf("insert role: %v", err)
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func loadTestObject[T types.InputType](filename string) (T, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	var output T
	err = decoder.Decode(&output)
	if err != nil {
		return nil, err
	}

	return output, nil
}

func TestConverter_NodePipeline(t *testing.T) {
	t.Parallel()

	// Reset the sync.Once used by libkube.DefaultNodeIdentity so each test gets a fresh lookup.
	libkube.ResetOnce()

	ctx := t.Context()
	input, err := loadTestObject[types.NodeType]("testdata/node.json")
	assert.NoError(t, err, "node load error")

	db := testDB(t)

	// Insert identity for "system:nodes" (the default node group identity).
	identityID := store.ObjectID()
	insertIdentity(t, db, identityID, "system:nodes", "")

	// Collector input -> store model
	storeNode, err := NewStoreWithDB(testConfig, db).Node(ctx, input)
	assert.NoError(t, err, "store node convert error")

	assert.Equal(t, storeNode.Name, input.Name)

	// Store model -> graph model
	graphNode, err := NewGraph(testConfig).Node(storeNode)
	assert.NoError(t, err, "graph node convert error")

	assert.Equal(t, store.Hex(storeNode.Id), graphNode.StoreID)
	assert.Equal(t, graphNode.Team, "test-team")
	assert.Equal(t, storeNode.Name, graphNode.Name)
	assert.False(t, storeNode.IsNamespaced)
	assert.Equal(t, storeNode.Namespace, graphNode.Namespace)
	assert.False(t, graphNode.Critical)
	assert.Equal(t, shared.CompromiseNone, graphNode.Compromised)
}

func TestConverter_RolePipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.RoleType]("testdata/role.json")
	assert.NoError(t, err, "role load error")

	// Collector input -> store model
	storeRole, err := NewStore(testConfig).Role(t.Context(), input)
	assert.NoError(t, err, "store role convert error")

	assert.Equal(t, storeRole.Name, input.Name)
	assert.True(t, storeRole.IsNamespaced)
	assert.Equal(t, storeRole.Namespace, input.Namespace)
	assert.Equal(t, storeRole.Rules, input.Rules)
	assert.Equal(t, storeRole.Runtime.Cluster.Name, testConfig.Dynamic.Cluster.Name)
	assert.Equal(t, storeRole.Runtime.RunID, testConfig.Dynamic.RunID.String())
}

func TestConverter_ClusterRolePipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.ClusterRoleType]("testdata/clusterrole.json")
	assert.NoError(t, err, "cluster role load error")

	// Collector input -> store model
	storeRole, err := NewStore(testConfig).ClusterRole(t.Context(), input)
	assert.NoError(t, err, "store role convert error")

	assert.Equal(t, storeRole.Name, input.Name)
	assert.False(t, storeRole.IsNamespaced)
	assert.Empty(t, storeRole.Namespace)
	assert.Equal(t, storeRole.Rules, input.Rules)
	assert.Equal(t, storeRole.Runtime.Cluster.Name, testConfig.Dynamic.Cluster.Name)
	assert.Equal(t, storeRole.Runtime.RunID, testConfig.Dynamic.RunID.String())
}

func TestConverter_RoleBindingPipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.RoleBindingType]("testdata/rolebinding.json")
	assert.NoError(t, err, "role binding load error")

	rawRole, err := loadTestObject[types.RoleType]("testdata/role.json")
	assert.NoError(t, err, "role load error")

	linkedRole, err := NewStore(testConfig).Role(t.Context(), rawRole)
	assert.NoError(t, err, "role convert error")

	db := testDB(t)

	// Insert the role into SQLite so the converter can look it up.
	insertRole(t, db, linkedRole)

	converter := NewStoreWithDB(testConfig, db)

	// Collector input -> store rolebinding
	storeBinding, err := converter.RoleBinding(t.Context(), input)
	assert.NoError(t, err, "store role binding convert error")

	assert.Equal(t, storeBinding.Name, input.Name)
	assert.Equal(t, storeBinding.RoleId, linkedRole.Id)
	assert.True(t, storeBinding.IsNamespaced)
	assert.Equal(t, storeBinding.Namespace, input.Namespace)

	assert.Equal(t, 1, len(storeBinding.Subjects))
	subject := storeBinding.Subjects[0]

	assert.NotEmpty(t, subject.IdentityId)
	assert.Equal(t, subject.Subject, input.Subjects[0])

	// Collector input -> store identity
	storeIdentity, err := NewStore(testConfig).Identity(t.Context(), &subject, storeBinding)
	assert.NoError(t, err, "store identity convert error")

	assert.Equal(t, subject.Subject.Name, storeIdentity.Name)
	assert.Equal(t, subject.Subject.Namespace, storeIdentity.Namespace)
	assert.Equal(t, subject.Subject.Kind, storeIdentity.Type)

	// Collector input -> store permissions
	storePermissionSet, err := converter.PermissionSet(t.Context(), storeBinding)
	assert.NoError(t, err, "store permission set convert error")

	assert.True(t, storePermissionSet.IsNamespaced)
	assert.Equal(t, storeBinding.Namespace, storePermissionSet.Namespace)
	assert.Equal(t, linkedRole.Name, storePermissionSet.RoleName)
	assert.Equal(t, storeBinding.Name, storePermissionSet.RoleBindingName)
	assert.Equal(t, linkedRole.Id, storePermissionSet.RoleId)
	assert.Equal(t, storeBinding.Id, storePermissionSet.RoleBindingId)

	// Store identity -> graph identity
	graphIdentity, err := NewGraph(testConfig).Identity(storeIdentity)
	assert.NoError(t, err, "graph role binding convert error")

	assert.Equal(t, store.Hex(storeIdentity.Id), graphIdentity.StoreID)
	assert.Equal(t, graphIdentity.App, "test-app")
	assert.Equal(t, graphIdentity.Service, "test-service")
	assert.Equal(t, graphIdentity.Team, "test-team")
	assert.Equal(t, storeIdentity.Name, graphIdentity.Name)
	assert.Equal(t, storeIdentity.Namespace, graphIdentity.Namespace)
	assert.Equal(t, storeIdentity.Type, graphIdentity.Type)

	// Store model -> graph model
	graphPermissions, err := NewGraph(testConfig).PermissionSet(storePermissionSet)
	assert.NoError(t, err, "graph role convert error")

	assert.Equal(t, store.Hex(storePermissionSet.Id), graphPermissions.StoreID)
	assert.Equal(t, storePermissionSet.Ownership.Application, graphPermissions.App)
	assert.Equal(t, storePermissionSet.Ownership.Service, graphPermissions.Service)
	assert.Equal(t, storePermissionSet.Ownership.Team, graphPermissions.Team)
	assert.Equal(t, storePermissionSet.Name, graphPermissions.Name)
	assert.Equal(t, storePermissionSet.Namespace, graphPermissions.Namespace)
	assert.Equal(t, storePermissionSet.RoleName, graphPermissions.Role)
	assert.Equal(t, storePermissionSet.RoleBindingName, graphPermissions.RoleBinding)

	rules := []string{
		"API()::R(pods)::N()::V(get,list)",
		"API()::R(configmaps)::N()::V(get)",
		"API(apps)::R(statefulsets)::N()::V(get,list)",
	}

	assert.Equal(t, rules, graphPermissions.Rules)
}

func TestConverter_ClusterRoleBindingPipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.ClusterRoleBindingType]("testdata/clusterrolebinding.json")
	assert.NoError(t, err, "cluster role binding load error")

	rawRole, err := loadTestObject[types.ClusterRoleType]("testdata/clusterrole.json")
	assert.NoError(t, err, "role load error")

	linkedRole, err := NewStore(testConfig).ClusterRole(t.Context(), rawRole)
	assert.NoError(t, err, "role convert error")

	db := testDB(t)

	// Insert the role into SQLite so the converter can look it up.
	insertRole(t, db, linkedRole)

	converter := NewStoreWithDB(testConfig, db)

	// Collector input -> store rolebinding
	storeBinding, err := converter.ClusterRoleBinding(t.Context(), input)
	assert.NoError(t, err, "store cluster role binding convert error")

	assert.Equal(t, storeBinding.Name, input.Name)
	assert.Equal(t, storeBinding.RoleId, linkedRole.Id)
	assert.False(t, storeBinding.IsNamespaced)
	assert.Empty(t, storeBinding.Namespace)

	assert.Equal(t, 1, len(storeBinding.Subjects))
	subject := storeBinding.Subjects[0]

	assert.NotEmpty(t, subject.IdentityId)
	assert.Equal(t, subject.Subject, input.Subjects[0])

	// Collector input -> store permissions
	storePermissionSet, err := converter.PermissionSetCluster(t.Context(), storeBinding)
	assert.NoError(t, err, "store permission set convert error")

	assert.False(t, storePermissionSet.IsNamespaced)
	assert.Equal(t, storeBinding.Namespace, storePermissionSet.Namespace)
	assert.Equal(t, linkedRole.Name, storePermissionSet.RoleName)
	assert.Equal(t, storeBinding.Name, storePermissionSet.RoleBindingName)
	assert.Equal(t, linkedRole.Id, storePermissionSet.RoleId)
	assert.Equal(t, storeBinding.Id, storePermissionSet.RoleBindingId)

	// Collector input -> store identity
	storeIdentity, err := NewStore(testConfig).Identity(t.Context(), &subject, storeBinding)
	assert.NoError(t, err, "store identity convert error")

	assert.Equal(t, subject.Subject.Name, storeIdentity.Name)
	assert.Equal(t, subject.Subject.Namespace, storeIdentity.Namespace)
	assert.Equal(t, subject.Subject.Kind, storeIdentity.Type)

	// Store identity -> graph identity
	graphIdentity, err := NewGraph(testConfig).Identity(storeIdentity)
	assert.NoError(t, err, "graph role binding convert error")

	assert.Equal(t, store.Hex(storeIdentity.Id), graphIdentity.StoreID)
	assert.Equal(t, graphIdentity.App, "test-app")
	assert.Equal(t, graphIdentity.Service, "test-service")
	assert.Equal(t, graphIdentity.Team, "test-team")
	assert.Equal(t, storeIdentity.Name, graphIdentity.Name)
	assert.Equal(t, storeIdentity.Namespace, graphIdentity.Namespace)
	assert.Equal(t, storeIdentity.Type, graphIdentity.Type)

	// Store model -> graph model
	graphPermissions, err := NewGraph(testConfig).PermissionSet(storePermissionSet)
	assert.NoError(t, err, "graph role convert error")

	assert.Equal(t, store.Hex(storePermissionSet.Id), graphPermissions.StoreID)
	assert.Equal(t, storePermissionSet.Ownership.Application, graphPermissions.App)
	assert.Equal(t, storePermissionSet.Ownership.Service, graphPermissions.Service)
	assert.Equal(t, storePermissionSet.Ownership.Team, graphPermissions.Team)
	assert.Equal(t, storePermissionSet.Name, graphPermissions.Name)
	assert.Equal(t, storePermissionSet.Namespace, graphPermissions.Namespace)
	assert.Equal(t, storePermissionSet.RoleName, graphPermissions.Role)
	assert.Equal(t, storePermissionSet.RoleBindingName, graphPermissions.RoleBinding)

	rules := []string{
		"API()::R(pods)::N()::V(get,list)",
		"API()::R(configmaps)::N()::V(get)",
		"API(apps)::R(statefulsets)::N()::V(get,list)",
	}

	assert.Equal(t, rules, graphPermissions.Rules)
}

func TestConverter_PermissionSet_ClusterRole_RoleBinding(t *testing.T) {
	t.Parallel()

	cr := store.Role{
		Id:           store.ObjectID(),
		Name:         "test-cluster-role",
		Namespace:    "",
		IsNamespaced: false,
	}

	rb := store.RoleBinding{
		Name:         "test-rolebinding",
		Namespace:    "test-ns",
		IsNamespaced: true,
		RoleId:       cr.Id,
		Subjects: []store.BindSubject{
			{
				Subject: rbacv1.Subject{
					Kind:      "ServiceAccount",
					Name:      "test-sa",
					Namespace: "",
				},
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "test-cluster-role",
		},
	}

	db := testDB(t)
	insertRole(t, db, &cr)

	ps, err := NewStoreWithDB(testConfig, db).PermissionSet(t.Context(), &rb)
	assert.NoError(t, err, "store permission set convert error")

	assert.True(t, ps.IsNamespaced)
	assert.Equal(t, rb.Namespace, ps.Namespace)
}

func TestConverter_PermissionSet_Role_RoleBinding_Namespace(t *testing.T) {
	t.Parallel()

	r := store.Role{
		Id:           store.ObjectID(),
		Name:         "test-ns1-role",
		Namespace:    "test-ns1",
		IsNamespaced: true,
	}

	rb := store.RoleBinding{
		Name:         "test-rolebinding",
		Namespace:    "test-ns1",
		IsNamespaced: true,
		RoleId:       r.Id,
		Subjects: []store.BindSubject{
			{
				Subject: rbacv1.Subject{
					Kind:      "ServiceAccount",
					Name:      "test-sa",
					Namespace: "test-ns2",
				},
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "Role",
			Name: "test-ns1-role",
		},
	}

	db := testDB(t)
	insertRole(t, db, &r)

	ps, err := NewStoreWithDB(testConfig, db).PermissionSet(t.Context(), &rb)
	assert.NoError(t, err, "store permission set convert error")

	assert.True(t, ps.IsNamespaced)
	assert.Equal(t, rb.Namespace, ps.Namespace)
}

func TestConverter_PermissionSet_InvalidCombination_Namespace(t *testing.T) {
	t.Parallel()

	r := store.Role{
		Id:           store.ObjectID(),
		Name:         "test-ns1-role",
		Namespace:    "test-ns1",
		IsNamespaced: true,
	}

	rb := store.RoleBinding{
		Name:         "test-rolebinding",
		Namespace:    "test-ns2",
		IsNamespaced: true,
		RoleId:       r.Id,
		Subjects: []store.BindSubject{
			{
				Subject: rbacv1.Subject{
					Kind:      "ServiceAccount",
					Name:      "test-sa",
					Namespace: "test-ns2",
				},
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "Role",
			Name: "test-ns1-role",
		},
	}

	db := testDB(t)
	// Insert the role in namespace "test-ns1". The PermissionSet method queries using
	// rb.Namespace ("test-ns2") for a Role kind, so the role won't be found.
	// With SQLite the cross-namespace mismatch is caught at query time.
	insertRole(t, db, &r)

	_, err := NewStoreWithDB(testConfig, db).PermissionSet(t.Context(), &rb)
	assert.ErrorContains(t, err, "missing role in cache")
}

func TestConverter_PermissionSet_InvalidCombination_Users(t *testing.T) {
	t.Parallel()

	r := store.Role{
		Id:           store.ObjectID(),
		Name:         "test-ns1-role",
		Namespace:    "test-ns1",
		IsNamespaced: true,
	}

	rb := store.RoleBinding{
		Name:         "test-rolebinding",
		Namespace:    "test-ns1",
		IsNamespaced: true,
		RoleId:       r.Id,
		Subjects: []store.BindSubject{
			{
				Subject: rbacv1.Subject{
					Kind:      "User",
					Name:      "test-user",
					Namespace: "test-ns2",
				},
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "Role",
			Name: "test-ns1-role",
		},
	}

	db := testDB(t)
	insertRole(t, db, &r)

	_, err := NewStoreWithDB(testConfig, db).PermissionSet(t.Context(), &rb)
	assert.ErrorContains(t, err, "incorrect combination ")
}

func TestConverter_PermissionSet_InvalidCombination_Types(t *testing.T) {
	t.Parallel()

	r := store.Role{
		Id:           store.ObjectID(),
		Name:         "test-ns1-role",
		Namespace:    "test-ns1",
		IsNamespaced: true,
	}

	crb := store.RoleBinding{
		Name:         "test-rolebinding",
		Namespace:    "",
		IsNamespaced: false,
		RoleId:       r.Id,
		Subjects: []store.BindSubject{
			{
				Subject: rbacv1.Subject{
					Kind:      "ServiceAccount",
					Name:      "test-sa",
					Namespace: "test-ns1",
				},
			},
		},
		RoleRef: rbacv1.RoleRef{
			Kind: "Role",
			Name: "test-ns1-role",
		},
	}

	db := testDB(t)
	// For PermissionSetCluster, queryRole uses crb.Namespace ("") so we insert the role
	// with namespace "" so it's found, but with IsNamespaced=true which triggers the mismatch.
	_, err := db.ExecContext(context.Background(),
		"INSERT INTO roles (id, name, is_namespaced, namespace, rules, run_id, cluster_name) VALUES (?, ?, ?, ?, ?, ?, ?)",
		r.Id, r.Name, boolToInt(r.IsNamespaced), "", "[]",
		testConfig.Dynamic.RunID.String(), testConfig.Dynamic.Cluster.Name)
	assert.NoError(t, err)

	_, err = NewStoreWithDB(testConfig, db).PermissionSetCluster(t.Context(), &crb)
	assert.ErrorContains(t, err, "incorrect combination ")
}

func TestConverter_RoleCacheFailure(t *testing.T) {
	t.Parallel()

	// Empty database - no roles inserted, so all role lookups will fail.
	db := testDB(t)

	rb, err := loadTestObject[types.RoleBindingType]("testdata/rolebinding.json")
	assert.NoError(t, err, "role binding load error")

	_, err = NewStoreWithDB(testConfig, db).RoleBinding(t.Context(), rb)
	assert.ErrorContains(t, err, "role binding found with no matching role")

	crb, err := loadTestObject[types.ClusterRoleBindingType]("testdata/clusterrolebinding.json")
	assert.NoError(t, err, "cluster role binding load error")

	_, err = NewStoreWithDB(testConfig, db).ClusterRoleBinding(t.Context(), crb)
	assert.ErrorContains(t, err, "role binding found with no matching role")
}

func TestConverter_PodPipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.PodType]("testdata/pod.json")
	assert.NoError(t, err, "pod load error")

	db := testDB(t)

	// Insert a node so the pod converter can look up the node ID.
	nodeID := store.ObjectID()
	insertNode(t, db, nodeID, "test-node.ec2.internal", "")

	// Collector input -> store pod
	storePod, err := NewStoreWithDB(testConfig, db).Pod(t.Context(), input)
	assert.NoError(t, err, "store pod convert error")

	assert.Equal(t, store.Hex(storePod.NodeId), store.Hex(nodeID))
	assert.Equal(t, storePod.Name, input.Name)
	assert.True(t, storePod.IsNamespaced)
	assert.Equal(t, storePod.Namespace, input.Namespace)

	// Store pod -> graph pod
	graphPod, err := NewGraph(testConfig).Pod(storePod)
	assert.NoError(t, err, "graph pod convert error")

	assert.Equal(t, store.Hex(storePod.Id), graphPod.StoreID)
	assert.Equal(t, graphPod.App, "test-app")
	assert.Equal(t, graphPod.Service, "test-service")
	assert.Equal(t, graphPod.Team, "test-team")
	assert.Equal(t, storePod.Name, graphPod.Name)
	assert.Equal(t, storePod.Namespace, graphPod.Namespace)
	assert.False(t, graphPod.ShareProcessNamespace)
	assert.Equal(t, storePod.ServiceAccount, graphPod.ServiceAccount)
	assert.Equal(t, storePod.NodeName, graphPod.Node)
	assert.False(t, graphPod.Critical)
	assert.Equal(t, shared.CompromiseNone, graphPod.Compromised)
}

func TestConverter_PodChildPipeline(t *testing.T) {
	t.Parallel()

	// Reset the sync.Once used by libkube.DefaultNodeIdentity so each test gets a fresh lookup.
	libkube.ResetOnce()

	input, err := loadTestObject[types.PodType]("testdata/pod.json")
	assert.NoError(t, err, "pod load error")

	db := testDB(t)

	// Insert a node so the pod converter can look up the node ID.
	nodeID := store.ObjectID()
	insertNode(t, db, nodeID, "test-node.ec2.internal", "")

	// Insert an identity for the service account "app-monitors" in namespace "test-app"
	// so the projected volume converter can find it.
	identityID := store.ObjectID()
	insertIdentity(t, db, identityID, "app-monitors", "test-app")

	converter := NewStoreWithDB(testConfig, db)

	// Collector input -> store pod
	storePod, err := converter.Pod(t.Context(), input)
	assert.NoError(t, err, "store pod convert error")

	// Collector container -> store container
	assert.Equal(t, 1, len(input.Spec.Containers))
	inContainer := input.Spec.Containers[0]
	storeContainer, err := converter.Container(t.Context(), &inContainer, storePod)
	assert.NoError(t, err, "store container convert error")

	assert.Equal(t, store.Hex(storeContainer.NodeId), store.Hex(nodeID))
	assert.Equal(t, storeContainer.PodId, storePod.Id)
	assert.Equal(t, storeContainer.Inherited.PodName, storePod.Name)
	assert.Equal(t, storeContainer.Inherited.NodeName, storePod.NodeName)
	assert.Equal(t, storeContainer.Inherited.ServiceAccount, storePod.ServiceAccount)

	// Store container -> graph container
	graphContainer, err := NewGraph(testConfig).Container(storeContainer, storePod)
	assert.NoError(t, err, "graph container convert error")

	assert.Equal(t, store.Hex(storeContainer.Id), graphContainer.StoreID)
	assert.Equal(t, graphContainer.App, "test-app")
	assert.Equal(t, graphContainer.Service, "test-service")
	assert.Equal(t, graphContainer.Team, "test-team")
	assert.Equal(t, storeContainer.Name, graphContainer.Name)
	assert.Equal(t, storeContainer.Image, graphContainer.Image)
	assert.Equal(t, storeContainer.Command, graphContainer.Command)
	assert.Equal(t, storeContainer.Args, graphContainer.Args)
	assert.Equal(t, storeContainer.Inherited.PodName, graphContainer.Pod)
	assert.Equal(t, storeContainer.Inherited.NodeName, graphContainer.Node)
	assert.Equal(t, []string{"9200", "9300"}, graphContainer.Ports)
	assert.Equal(t, shared.CompromiseNone, graphContainer.Compromised)

	// Collector volume -> store volume (using K8s volumes from the input pod)
	assert.Equal(t, 2, len(input.Spec.Volumes))

	// VolumeMounts come from the input K8s container (not the store container).
	inVolume0 := inContainer.VolumeMounts[0]
	storeVolume0, err := converter.VolumeFromK8s(t.Context(), &inVolume0, input.Spec.Volumes, storePod, storeContainer)
	assert.NoError(t, err, "store volume convert error")

	assert.Equal(t, store.Hex(storeVolume0.NodeId), store.Hex(nodeID))
	assert.Equal(t, store.Hex(storeVolume0.PodId), store.Hex(storePod.Id))
	assert.Equal(t, storeVolume0.Name, inVolume0.Name)
	assert.Equal(t, storeVolume0.Type, shared.VolumeTypeHost)
	assert.Equal(t, storeVolume0.MountPath, inVolume0.MountPath)
	assert.Equal(t, storeVolume0.SourcePath, "/var/run/datadog-agent")
	assert.False(t, storeVolume0.ReadOnly)
	assert.Empty(t, storeVolume0.ProjectedId)

	inVolume1 := inContainer.VolumeMounts[1]
	storeVolume1, err := converter.VolumeFromK8s(t.Context(), &inVolume1, input.Spec.Volumes, storePod, storeContainer)
	assert.NoError(t, err, "store volume convert error")

	assert.Equal(t, store.Hex(storeVolume1.NodeId), store.Hex(nodeID))
	assert.Equal(t, store.Hex(storeVolume1.PodId), store.Hex(storePod.Id))
	assert.Equal(t, storeVolume1.Name, inVolume1.Name)
	assert.Equal(t, storeVolume1.Type, shared.VolumeTypeProjected)
	assert.Equal(t, storeVolume1.MountPath, inVolume1.MountPath)
	assert.Equal(t, storeVolume1.SourcePath, "/var/lib/kubelet/pods/5a9fc508-8410-444a-bf63-9f11e5979bee/volumes/kubernetes.io~projected/kube-api-access-4x9fz/token")
	assert.True(t, storeVolume1.ReadOnly)
	assert.Equal(t, store.Hex(storeVolume1.ProjectedId), store.Hex(identityID))

	// Store container -> graph container
	graphVolume, err := NewGraph(testConfig).Volume(storeVolume0, storePod)
	assert.NoError(t, err, "graph volume convert error")

	assert.Equal(t, store.Hex(storeVolume0.Id), graphVolume.StoreID)
	assert.Equal(t, graphVolume.App, "test-app")
	assert.Equal(t, graphVolume.Service, "test-service")
	assert.Equal(t, graphVolume.Team, "test-team")
	assert.Equal(t, storeVolume0.Name, graphVolume.Name)
	assert.Equal(t, shared.VolumeTypeHost, graphVolume.Type)
	assert.Equal(t, "/var/run/datadog-agent", graphVolume.SourcePath)
	assert.Equal(t, "/var/run/datadog-agent", graphVolume.MountPath)
	assert.False(t, graphVolume.Readonly)

	graphVolume, err = NewGraph(testConfig).Volume(storeVolume1, storePod)
	assert.NoError(t, err, "graph volume convert error")

	assert.Equal(t, store.Hex(storeVolume1.Id), graphVolume.StoreID)
	assert.Equal(t, graphVolume.App, "test-app")
	assert.Equal(t, graphVolume.Service, "test-service")
	assert.Equal(t, graphVolume.Team, "test-team")
	assert.Equal(t, storeVolume1.Name, graphVolume.Name)
	assert.Equal(t, shared.VolumeTypeProjected, graphVolume.Type)
	assert.Equal(t, "/var/lib/kubelet/pods/5a9fc508-8410-444a-bf63-9f11e5979bee/volumes/kubernetes.io~projected/kube-api-access-4x9fz/token", graphVolume.SourcePath)
	assert.Equal(t, "/var/run/secrets/kubernetes.io/serviceaccount", graphVolume.MountPath)
	assert.True(t, graphVolume.Readonly)
}

func TestConverter_PodCacheFailure(t *testing.T) {
	t.Parallel()

	// Empty database - no nodes inserted, so node lookup will fail.
	db := testDB(t)

	input, err := loadTestObject[types.PodType]("testdata/pod.json")
	assert.NoError(t, err, "pod load error")

	_, err = NewStoreWithDB(testConfig, db).Pod(t.Context(), input)
	assert.ErrorContains(t, err, "node lookup")
}

func TestConverter_EndpointPipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.EndpointType]("testdata/endpointslice.json")
	assert.NoError(t, err, "endpoint slice load error")

	// Collector input -> store model
	storeEp, err := NewStore(testConfig).Endpoint(t.Context(), input.Endpoints[0], input.Ports[0], input)
	assert.NoError(t, err, "endpoint convert error")

	assert.Equal(t, storeEp.Name, "cassandra-temporal-dev-kmwfp::TCP::cql")
	assert.True(t, storeEp.IsNamespaced)
	assert.Equal(t, storeEp.Namespace, input.Namespace)
	assert.Equal(t, storeEp.ServiceName, "cassandra-temporal-dev")
	assert.Equal(t, storeEp.ServiceDns, "cassandra-temporal-dev.cassandra-temporal-dev")
	assert.Equal(t, storeEp.AddressType, "IPv4")
	assert.Equal(t, storeEp.Addresses, []string{"10.1.1.1"})
	assert.Equal(t, storeEp.NodeName, "node.ec2.internal")
	assert.Equal(t, storeEp.Port, 9042)
	assert.Equal(t, storeEp.Protocol, "TCP")
	assert.Equal(t, storeEp.PortName, "cql")

	// Store model -> graph model
	graphEp, err := NewGraph(testConfig).Endpoint(storeEp)
	assert.NoError(t, err, "graph endpoint convert error")

	assert.Equal(t, store.Hex(storeEp.Id), graphEp.StoreID)
	assert.Equal(t, graphEp.App, "test-app")
	assert.Equal(t, graphEp.Service, "test-service")
	assert.Equal(t, graphEp.Team, "test-team")
	assert.Equal(t, storeEp.Name, graphEp.Name)
	assert.Equal(t, storeEp.ServiceName, graphEp.ServiceEndpointName)
	assert.Equal(t, storeEp.ServiceDns, graphEp.ServiceDnsName)
	assert.Equal(t, "IPv4", graphEp.AddressType)
	assert.Equal(t, []string{"10.1.1.1"}, graphEp.Addresses)
	assert.Equal(t, 9042, graphEp.Port)
	assert.Equal(t, "cql", graphEp.PortName)
	assert.Equal(t, "TCP", graphEp.Protocol)
	assert.Equal(t, shared.EndpointExposureExternal, graphEp.Exposure)
}

func TestConverter_EndpointPrivatePipeline(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input, err := loadTestObject[types.PodType]("testdata/pod.json")
	assert.NoError(t, err, "endpoint slice load error")

	db := testDB(t)

	// Insert a node so the pod converter can look up the node ID.
	nodeID := store.ObjectID()
	insertNode(t, db, nodeID, "test-node.ec2.internal", "")

	converter := NewStoreWithDB(testConfig, db)

	// Collector input -> store model
	pod, err := converter.Pod(ctx, input)
	assert.NoError(t, err)
	container, err := converter.Container(ctx, &input.Spec.Containers[0], pod)
	assert.NoError(t, err)
	containerPort := input.Spec.Containers[0].Ports[0]

	storeEp, err := converter.EndpointPrivate(ctx, &containerPort, pod, container)
	assert.NoError(t, err, "endpoint convert error")

	assert.Equal(t, storeEp.Name, "test-app::app-monitors-client-78cb6d7899-j2rjp::TCP::9200")
	assert.True(t, storeEp.IsNamespaced)
	assert.Equal(t, storeEp.Namespace, pod.Namespace)
	assert.Equal(t, storeEp.ServiceName, "http")
	assert.Equal(t, storeEp.ServiceDns, "")
	assert.Equal(t, storeEp.AddressType, "IPv4")
	assert.Equal(t, storeEp.Addresses, []string{"10.1.1.2"})
	assert.Equal(t, storeEp.NodeName, "test-node.ec2.internal")
	assert.Equal(t, storeEp.Port, 9200)
	assert.Equal(t, storeEp.Protocol, "TCP")
	assert.Equal(t, storeEp.PortName, "http")

	// Store model -> graph model
	graphEp, err := NewGraph(testConfig).Endpoint(storeEp)
	assert.NoError(t, err, "graph endpoint convert error")

	assert.Equal(t, store.Hex(storeEp.Id), graphEp.StoreID)
	assert.Equal(t, graphEp.App, "test-app")
	assert.Equal(t, graphEp.Service, "test-service")
	assert.Equal(t, graphEp.Team, "test-team")
	assert.Equal(t, storeEp.Name, graphEp.Name)
	assert.Equal(t, storeEp.ServiceName, graphEp.ServiceEndpointName)
	assert.Equal(t, storeEp.ServiceDns, graphEp.ServiceDnsName)
	assert.Equal(t, "IPv4", graphEp.AddressType)
	assert.Equal(t, []string{"10.1.1.2"}, graphEp.Addresses)
	assert.Equal(t, 9200, graphEp.Port)
	assert.Equal(t, "http", graphEp.PortName)
	assert.Equal(t, "TCP", graphEp.Protocol)
	assert.Equal(t, shared.EndpointExposureNodeIP, graphEp.Exposure)
}
