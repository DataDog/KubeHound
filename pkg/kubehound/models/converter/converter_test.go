package converter

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	rbacv1 "k8s.io/api/rbac/v1"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/cachekey"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

	ctx := context.Background()
	input, err := loadTestObject[types.NodeType]("testdata/node.json")
	assert.NoError(t, err, "node load error")

	c := mocks.NewCacheReader(t)
	id := store.ObjectID().Hex()
	c.EXPECT().Get(ctx, cachekey.Identity("system:node:node-1", "")).Return(&cache.CacheResult{
		Value: nil,
		Err:   cache.ErrNoEntry,
	}).Once()
	c.EXPECT().Get(ctx, cachekey.Identity("system:nodes", "")).Return(&cache.CacheResult{
		Value: id,
		Err:   nil,
	}).Once()

	// Collector input -> store model
	storeNode, err := NewStoreWithCache(c).Node(ctx, input)
	assert.NoError(t, err, "store node convert error")

	assert.Equal(t, storeNode.K8.Name, input.Name)

	// Store model -> graph model
	graphNode, err := NewGraph().Node(storeNode)
	assert.NoError(t, err, "graph node convert error")

	assert.Equal(t, storeNode.Id.Hex(), graphNode.StoreID)
	assert.Equal(t, graphNode.Team, "test-team")
	assert.Equal(t, storeNode.K8.Name, graphNode.Name)
	assert.False(t, storeNode.IsNamespaced)
	assert.Equal(t, storeNode.K8.Namespace, graphNode.Namespace)
	assert.False(t, graphNode.Critical)
	assert.Equal(t, shared.CompromiseNone, graphNode.Compromised)
}

func TestConverter_RolePipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.RoleType]("testdata/role.json")
	assert.NoError(t, err, "role load error")

	// Collector input -> store model
	storeRole, err := NewStore().Role(context.TODO(), input)
	assert.NoError(t, err, "store role convert error")

	assert.Equal(t, storeRole.Name, input.Name)
	assert.True(t, storeRole.IsNamespaced)
	assert.Equal(t, storeRole.Namespace, input.Namespace)
	assert.Equal(t, storeRole.Rules, input.Rules)
}

func TestConverter_ClusterRolePipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.ClusterRoleType]("testdata/clusterrole.json")
	assert.NoError(t, err, "cluster role load error")

	// Collector input -> store model
	storeRole, err := NewStore().ClusterRole(context.TODO(), input)
	assert.NoError(t, err, "store role convert error")

	assert.Equal(t, storeRole.Name, input.Name)
	assert.False(t, storeRole.IsNamespaced)
	assert.Empty(t, storeRole.Namespace)
	assert.Equal(t, storeRole.Rules, input.Rules)
}

func TestConverter_RoleBindingPipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.RoleBindingType]("testdata/rolebinding.json")
	assert.NoError(t, err, "role binding load error")

	rawRole, err := loadTestObject[types.RoleType]("testdata/role.json")
	assert.NoError(t, err, "role load error")

	linkedRole, err := NewStore().Role(context.TODO(), rawRole)
	assert.NoError(t, err, "role convert error")

	c := mocks.NewCacheReader(t)
	c.EXPECT().Get(mock.Anything, cachekey.Identity("app-monitors", "test-app")).Return(&cache.CacheResult{
		Value: nil,
		Err:   cache.ErrNoEntry,
	})
	c.EXPECT().Get(mock.Anything, cachekey.Role("test-reader", "test-app")).Return(&cache.CacheResult{
		Value: *linkedRole,
		Err:   nil,
	})

	// Collector input -> store rolebinding
	storeBinding, err := NewStoreWithCache(c).RoleBinding(context.TODO(), input)
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
	storeIdentity, err := NewStore().Identity(context.TODO(), &subject, storeBinding)
	assert.NoError(t, err, "store identity convert error")

	assert.Equal(t, subject.Subject.Name, storeIdentity.Name)
	assert.Equal(t, subject.Subject.Namespace, storeIdentity.Namespace)
	assert.Equal(t, subject.Subject.Kind, storeIdentity.Type)

	// Collector input -> store permissions
	storePermissionSet, err := NewStoreWithCache(c).PermissionSet(context.TODO(), storeBinding)
	assert.NoError(t, err, "store permission set convert error")

	assert.True(t, storePermissionSet.IsNamespaced)
	assert.Equal(t, storeBinding.Namespace, storePermissionSet.Namespace)
	assert.Equal(t, linkedRole.Name, storePermissionSet.RoleName)
	assert.Equal(t, storeBinding.Name, storePermissionSet.RoleBindingName)
	assert.Equal(t, linkedRole.Id, storePermissionSet.RoleId)
	assert.Equal(t, storeBinding.Id, storePermissionSet.RoleBindingId)

	// Store identity -> graph identity
	graphIdentity, err := NewGraph().Identity(storeIdentity)
	assert.NoError(t, err, "graph role binding convert error")

	assert.Equal(t, storeIdentity.Id.Hex(), graphIdentity.StoreID)
	assert.Equal(t, graphIdentity.App, "test-app")
	assert.Equal(t, graphIdentity.Service, "test-service")
	assert.Equal(t, graphIdentity.Team, "test-team")
	assert.Equal(t, storeIdentity.Name, graphIdentity.Name)
	assert.Equal(t, storeIdentity.Namespace, graphIdentity.Namespace)
	assert.Equal(t, storeIdentity.Type, graphIdentity.Type)

	// Store model -> graph model
	graphPermissions, err := NewGraph().PermissionSet(storePermissionSet)
	assert.NoError(t, err, "graph role convert error")

	assert.Equal(t, storePermissionSet.Id.Hex(), graphPermissions.StoreID)
	assert.Equal(t, graphPermissions.App, "test-app")
	assert.Equal(t, graphPermissions.Service, "test-service")
	assert.Equal(t, graphPermissions.Team, "test-team")
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

	linkedRole, err := NewStore().ClusterRole(context.TODO(), rawRole)
	assert.NoError(t, err, "role convert error")

	c := mocks.NewCacheReader(t)
	c.EXPECT().Get(mock.Anything, cachekey.Role("test-reader", "")).Return(&cache.CacheResult{
		Value: *linkedRole,
		Err:   nil,
	})
	c.EXPECT().Get(mock.Anything, cachekey.Identity("app-monitors-cluster", "test-app")).Return(&cache.CacheResult{
		Value: nil,
		Err:   cache.ErrNoEntry,
	})

	// Collector input -> store rolebinding
	storeBinding, err := NewStoreWithCache(c).ClusterRoleBinding(context.TODO(), input)
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
	storePermissionSet, err := NewStoreWithCache(c).PermissionSetCluster(context.TODO(), storeBinding)
	assert.NoError(t, err, "store permission set convert error")

	assert.False(t, storePermissionSet.IsNamespaced)
	assert.Equal(t, storeBinding.Namespace, storePermissionSet.Namespace)
	assert.Equal(t, linkedRole.Name, storePermissionSet.RoleName)
	assert.Equal(t, storeBinding.Name, storePermissionSet.RoleBindingName)
	assert.Equal(t, linkedRole.Id, storePermissionSet.RoleId)
	assert.Equal(t, storeBinding.Id, storePermissionSet.RoleBindingId)

	// Collector input -> store identity
	storeIdentity, err := NewStore().Identity(context.TODO(), &subject, storeBinding)
	assert.NoError(t, err, "store identity convert error")

	assert.Equal(t, subject.Subject.Name, storeIdentity.Name)
	assert.Equal(t, subject.Subject.Namespace, storeIdentity.Namespace)
	assert.Equal(t, subject.Subject.Kind, storeIdentity.Type)

	// Store identity -> graph identity
	graphIdentity, err := NewGraph().Identity(storeIdentity)
	assert.NoError(t, err, "graph role binding convert error")

	assert.Equal(t, storeIdentity.Id.Hex(), graphIdentity.StoreID)
	assert.Equal(t, graphIdentity.App, "test-app")
	assert.Equal(t, graphIdentity.Service, "test-service")
	assert.Equal(t, graphIdentity.Team, "test-team")
	assert.Equal(t, storeIdentity.Name, graphIdentity.Name)
	assert.Equal(t, storeIdentity.Namespace, graphIdentity.Namespace)
	assert.Equal(t, storeIdentity.Type, graphIdentity.Type)

	// Store model -> graph model
	graphPermissions, err := NewGraph().PermissionSet(storePermissionSet)
	assert.NoError(t, err, "graph role convert error")

	assert.Equal(t, storePermissionSet.Id.Hex(), graphPermissions.StoreID)
	assert.Equal(t, graphPermissions.App, "test-app")
	assert.Equal(t, graphPermissions.Service, "test-service")
	assert.Equal(t, graphPermissions.Team, "test-team")
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
		K8: rbacv1.RoleRef{
			Kind: "ClusterRole",
			Name: "test-cluster-role",
		},
	}

	c := mocks.NewCacheReader(t)
	c.EXPECT().Get(mock.Anything, cachekey.Role("test-cluster-role", "")).Return(&cache.CacheResult{
		Value: cr,
		Err:   nil,
	})

	ps, err := NewStoreWithCache(c).PermissionSet(context.TODO(), &rb)
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
		K8: rbacv1.RoleRef{
			Kind: "Role",
			Name: "test-ns1-role",
		},
	}

	c := mocks.NewCacheReader(t)
	c.EXPECT().Get(mock.Anything, cachekey.Role("test-ns1-role", "test-ns1")).Return(&cache.CacheResult{
		Value: r,
		Err:   nil,
	})

	ps, err := NewStoreWithCache(c).PermissionSet(context.TODO(), &rb)
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
		K8: rbacv1.RoleRef{
			Kind: "Role",
			Name: "test-ns1-role",
		},
	}

	c := mocks.NewCacheReader(t)
	c.EXPECT().Get(mock.Anything, cachekey.Role("test-ns1-role", "test-ns2")).Return(&cache.CacheResult{
		Value: r,
		Err:   nil,
	})

	_, err := NewStoreWithCache(c).PermissionSet(context.TODO(), &rb)
	assert.ErrorContains(t, err, "incorrect combination ")
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
		K8: rbacv1.RoleRef{
			Kind: "Role",
			Name: "test-ns1-role",
		},
	}

	c := mocks.NewCacheReader(t)
	c.EXPECT().Get(mock.Anything, cachekey.Role("test-ns1-role", "test-ns1")).Return(&cache.CacheResult{
		Value: r,
		Err:   nil,
	})

	_, err := NewStoreWithCache(c).PermissionSet(context.TODO(), &rb)
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
		K8: rbacv1.RoleRef{
			Kind: "Role",
			Name: "test-ns1-role",
		},
	}

	c := mocks.NewCacheReader(t)
	c.EXPECT().Get(mock.Anything, cachekey.Role("test-ns1-role", "")).Return(&cache.CacheResult{
		Value: r,
		Err:   nil,
	})

	_, err := NewStoreWithCache(c).PermissionSetCluster(context.TODO(), &crb)
	assert.ErrorContains(t, err, "incorrect combination ")
}

func TestConverter_RoleCacheFailure(t *testing.T) {
	t.Parallel()

	c := mocks.NewCacheReader(t)
	c.EXPECT().Get(mock.Anything, mock.Anything).Return(&cache.CacheResult{
		Value: "",
		Err:   errors.New("not found"),
	}).Times(4)

	rb, err := loadTestObject[types.RoleBindingType]("testdata/rolebinding.json")
	assert.NoError(t, err, "role binding load error")

	_, err = NewStoreWithCache(c).RoleBinding(context.TODO(), rb)
	assert.ErrorContains(t, err, "role binding found with no matching role")

	crb, err := loadTestObject[types.ClusterRoleBindingType]("testdata/clusterrolebinding.json")
	assert.NoError(t, err, "cluster role binding load error")

	_, err = NewStoreWithCache(c).ClusterRoleBinding(context.TODO(), crb)
	assert.ErrorContains(t, err, "role binding found with no matching role")
}

func TestConverter_PodPipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.PodType]("testdata/pod.json")
	assert.NoError(t, err, "pod load error")

	c := mocks.NewCacheReader(t)
	k := cachekey.Node("test-node.ec2.internal")
	id := store.ObjectID().Hex()
	c.EXPECT().Get(mock.Anything, k).Return(&cache.CacheResult{
		Value: id,
		Err:   nil,
	})

	// Collector input -> store pod
	storePod, err := NewStoreWithCache(c).Pod(context.TODO(), input)
	assert.NoError(t, err, "store pod convert error")

	assert.Equal(t, storePod.NodeId.Hex(), id)
	assert.Equal(t, storePod.K8.Name, input.Name)
	assert.True(t, storePod.IsNamespaced)
	assert.Equal(t, storePod.K8.Namespace, input.Namespace)

	// Store pod -> graph pod
	graphPod, err := NewGraph().Pod(storePod)
	assert.NoError(t, err, "graph pod convert error")

	assert.Equal(t, storePod.Id.Hex(), graphPod.StoreID)
	assert.Equal(t, graphPod.App, "test-app")
	assert.Equal(t, graphPod.Service, "test-service")
	assert.Equal(t, graphPod.Team, "test-team")
	assert.Equal(t, storePod.K8.Name, graphPod.Name)
	assert.Equal(t, storePod.K8.Namespace, graphPod.Namespace)
	assert.False(t, graphPod.ShareProcessNamespace)
	assert.Equal(t, storePod.K8.Spec.ServiceAccountName, graphPod.ServiceAccount)
	assert.Equal(t, storePod.K8.Spec.NodeName, graphPod.Node)
	assert.False(t, graphPod.Critical)
	assert.Equal(t, shared.CompromiseNone, graphPod.Compromised)
}

func TestConverter_PodChildPipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.PodType]("testdata/pod.json")
	assert.NoError(t, err, "pod load error")

	c := mocks.NewCacheReader(t)
	nk := cachekey.Node("test-node.ec2.internal")
	nid := store.ObjectID().Hex()
	c.EXPECT().Get(mock.Anything, nk).Return(&cache.CacheResult{
		Value: nid,
		Err:   nil,
	})

	ik := cachekey.Identity("app-monitors", "test-app")
	iid := store.ObjectID().Hex()
	c.EXPECT().Get(mock.Anything, ik).Return(&cache.CacheResult{
		Value: iid,
		Err:   nil,
	})

	// Collector input -> store pod
	storePod, err := NewStoreWithCache(c).Pod(context.TODO(), input)
	assert.NoError(t, err, "store pod convert error")

	// Collector container -> store container
	assert.Equal(t, 1, len(input.Spec.Containers))
	inContainer := input.Spec.Containers[0]
	storeContainer, err := NewStoreWithCache(c).Container(context.TODO(), &inContainer, storePod)
	assert.NoError(t, err, "store container convert error")

	assert.Equal(t, storeContainer.NodeId.Hex(), nid)
	assert.Equal(t, storeContainer.PodId, storePod.Id)
	assert.Equal(t, storeContainer.Inherited.PodName, storePod.K8.Name)
	assert.Equal(t, storeContainer.Inherited.NodeName, storePod.K8.Spec.NodeName)
	assert.Equal(t, storeContainer.Inherited.ServiceAccount, storePod.K8.Spec.ServiceAccountName)

	// Store container -> graph container
	graphContainer, err := NewGraph().Container(storeContainer, storePod)
	assert.NoError(t, err, "graph container convert error")

	assert.Equal(t, storeContainer.Id.Hex(), graphContainer.StoreID)
	assert.Equal(t, graphContainer.App, "test-app")
	assert.Equal(t, graphContainer.Service, "test-service")
	assert.Equal(t, graphContainer.Team, "test-team")
	assert.Equal(t, storeContainer.K8.Name, graphContainer.Name)
	assert.Equal(t, storeContainer.K8.Image, graphContainer.Image)
	assert.Equal(t, storeContainer.K8.Command, graphContainer.Command)
	assert.Equal(t, storeContainer.K8.Args, graphContainer.Args)
	assert.Equal(t, storeContainer.Inherited.PodName, graphContainer.Pod)
	assert.Equal(t, storeContainer.Inherited.NodeName, graphContainer.Node)
	assert.Equal(t, []string{"9200", "9300"}, graphContainer.Ports)
	assert.Equal(t, shared.CompromiseNone, graphContainer.Compromised)

	// Collector volume -> store volume
	assert.Equal(t, 2, len(input.Spec.Volumes))
	inVolume0 := storeContainer.K8.VolumeMounts[0]
	storeVolume0, err := NewStoreWithCache(c).Volume(context.TODO(), &inVolume0, storePod, storeContainer)
	assert.NoError(t, err, "store volume convert error")

	assert.Equal(t, storeVolume0.NodeId.Hex(), nid)
	assert.Equal(t, storeVolume0.PodId.Hex(), storePod.Id.Hex())
	assert.Equal(t, storeVolume0.Name, inVolume0.Name)
	assert.Equal(t, storeVolume0.Type, shared.VolumeTypeHost)
	assert.Equal(t, storeVolume0.MountPath, inVolume0.MountPath)
	assert.Equal(t, storeVolume0.SourcePath, "/var/run/datadog-agent")
	assert.False(t, storeVolume0.ReadOnly)
	assert.Empty(t, storeVolume0.ProjectedId)

	inVolume1 := storeContainer.K8.VolumeMounts[1]
	storeVolume1, err := NewStoreWithCache(c).Volume(context.TODO(), &inVolume1, storePod, storeContainer)
	assert.NoError(t, err, "store volume convert error")

	assert.Equal(t, storeVolume1.NodeId.Hex(), nid)
	assert.Equal(t, storeVolume1.PodId.Hex(), storePod.Id.Hex())
	assert.Equal(t, storeVolume1.Name, inVolume1.Name)
	assert.Equal(t, storeVolume1.Type, shared.VolumeTypeProjected)
	assert.Equal(t, storeVolume1.MountPath, inVolume1.MountPath)
	assert.Equal(t, storeVolume1.SourcePath, "/var/lib/kubelet/pods/5a9fc508-8410-444a-bf63-9f11e5979bee/volumes/kubernetes.io~projected/kube-api-access-4x9fz/token")
	assert.True(t, storeVolume1.ReadOnly)
	assert.Equal(t, storeVolume1.ProjectedId.Hex(), iid)

	// Store container -> graph container
	graphVolume, err := NewGraph().Volume(storeVolume0, storePod)
	assert.NoError(t, err, "graph volume convert error")

	assert.Equal(t, storeVolume0.Id.Hex(), graphVolume.StoreID)
	assert.Equal(t, graphVolume.App, "test-app")
	assert.Equal(t, graphVolume.Service, "test-service")
	assert.Equal(t, graphVolume.Team, "test-team")
	assert.Equal(t, storeVolume0.Name, graphVolume.Name)
	assert.Equal(t, shared.VolumeTypeHost, graphVolume.Type)
	assert.Equal(t, "/var/run/datadog-agent", graphVolume.SourcePath)
	assert.Equal(t, "/var/run/datadog-agent", graphVolume.MountPath)
	assert.False(t, graphVolume.Readonly)

	graphVolume, err = NewGraph().Volume(storeVolume1, storePod)
	assert.NoError(t, err, "graph volume convert error")

	assert.Equal(t, storeVolume1.Id.Hex(), graphVolume.StoreID)
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

	c := mocks.NewCacheReader(t)
	c.EXPECT().Get(mock.Anything, mock.Anything).Return(&cache.CacheResult{
		Value: "",
		Err:   errors.New("not found"),
	})

	input, err := loadTestObject[types.PodType]("testdata/pod.json")
	assert.NoError(t, err, "pod load error")

	_, err = NewStoreWithCache(c).Pod(context.TODO(), input)
	assert.ErrorContains(t, err, "not found")
}

func TestConverter_EndpointPipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.EndpointType]("testdata/endpointslice.json")
	assert.NoError(t, err, "endpoint slice load error")

	// Collector input -> store model
	storeEp, err := NewStore().Endpoint(context.TODO(), input.Endpoints[0], input.Ports[0], input)
	assert.NoError(t, err, "endpoint convert error")

	assert.Equal(t, storeEp.Name, "cassandra-temporal-dev-kmwfp::TCP::cql")
	assert.True(t, storeEp.IsNamespaced)
	assert.Equal(t, storeEp.Namespace, input.Namespace)
	assert.Equal(t, storeEp.ServiceName, "cassandra-temporal-dev")
	assert.Equal(t, storeEp.ServiceDns, "cassandra-temporal-dev.cassandra-temporal-dev")
	assert.Equal(t, storeEp.AddressType, discoveryv1.AddressType("IPv4"))
	assert.Equal(t, storeEp.Backend.Addresses, []string{"10.1.1.1"})
	assert.Equal(t, *storeEp.Backend.NodeName, "node.ec2.internal")
	assert.Equal(t, *storeEp.Port.Port, int32(9042))
	assert.Equal(t, *storeEp.Port.Protocol, v1.Protocol("TCP"))
	assert.Equal(t, *storeEp.Port.Name, "cql")

	// Store model -> graph model
	graphEp, err := NewGraph().Endpoint(storeEp)
	assert.NoError(t, err, "graph endpoint convert error")

	assert.Equal(t, storeEp.Id.Hex(), graphEp.StoreID)
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

	ctx := context.Background()
	input, err := loadTestObject[types.PodType]("testdata/pod.json")
	assert.NoError(t, err, "endpoint slice load error")

	c := mocks.NewCacheReader(t)
	c.EXPECT().Get(mock.Anything, mock.AnythingOfType("*cachekey.nodeCacheKey")).Return(&cache.CacheResult{
		Value: store.ObjectID().Hex(),
		Err:   nil,
	})
	converter := NewStoreWithCache(c)

	// Collector input -> store model
	pod, err := converter.Pod(ctx, input)
	assert.NoError(t, err)
	container, err := converter.Container(ctx, &pod.K8.Spec.Containers[0], pod)
	assert.NoError(t, err)
	containerPort := container.K8.Ports[0]

	storeEp, err := converter.EndpointPrivate(ctx, &containerPort, pod, container)
	assert.NoError(t, err, "endpoint convert error")

	assert.Equal(t, storeEp.Name, "test-app::app-monitors-client-78cb6d7899-j2rjp::TCP::9200")
	assert.True(t, storeEp.IsNamespaced)
	assert.Equal(t, storeEp.Namespace, pod.K8.Namespace)
	assert.Equal(t, storeEp.ServiceName, "http")
	assert.Equal(t, storeEp.ServiceDns, "")
	assert.Equal(t, storeEp.AddressType, discoveryv1.AddressType("IPv4"))
	assert.Equal(t, storeEp.Backend.Addresses, []string{"10.1.1.2"})
	assert.Equal(t, *storeEp.Backend.NodeName, "test-node.ec2.internal")
	assert.Equal(t, *storeEp.Port.Port, int32(9200))
	assert.Equal(t, *storeEp.Port.Protocol, v1.Protocol("TCP"))
	assert.Equal(t, *storeEp.Port.Name, "http")

	// Store model -> graph model
	graphEp, err := NewGraph().Endpoint(storeEp)
	assert.NoError(t, err, "graph endpoint convert error")

	assert.Equal(t, storeEp.Id.Hex(), graphEp.StoreID)
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
