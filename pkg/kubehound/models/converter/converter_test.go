package converter

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/DataDog/KubeHound/pkg/globals/types"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/store"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/cache"
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

	input, err := loadTestObject[types.NodeType]("testdata/node.json")
	assert.NoError(t, err, "node load error")

	// Collector input -> store model
	storeNode, err := NewStore().Node(context.TODO(), input)
	assert.NoError(t, err, "store node convert error")

	assert.Equal(t, storeNode.K8.Name, input.Name)

	// Store model -> graph model
	graphNode, err := NewGraph().Node(storeNode)
	assert.NoError(t, err, "graph node convert error")

	assert.Equal(t, storeNode.Id.Hex(), graphNode.StoreId)
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

	// Store model -> graph model
	graphRole, err := NewGraph().Role(storeRole)
	assert.NoError(t, err, "graph role convert error")

	assert.Equal(t, storeRole.Id.Hex(), graphRole.StoreId)
	assert.Equal(t, storeRole.Name, graphRole.Name)
	assert.Equal(t, storeRole.Namespace, graphRole.Namespace)

	rules := []string{
		"API()::R(pods)::N()::V(get,list)",
		"API()::R(configmaps)::N()::V(get)",
		"API(apps)::R(statefulsets)::N()::V(get,list)",
	}

	assert.Equal(t, rules, graphRole.Rules)
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

	// Store model -> graph model
	graphRole, err := NewGraph().Role(storeRole)
	assert.NoError(t, err, "graph role convert error")

	assert.Equal(t, storeRole.Id.Hex(), graphRole.StoreId)
	assert.Equal(t, storeRole.Name, graphRole.Name)
	assert.Equal(t, storeRole.Namespace, graphRole.Namespace)

	rules := []string{
		"API()::R(pods)::N()::V(get,list)",
		"API()::R(configmaps)::N()::V(get)",
		"API(apps)::R(statefulsets)::N()::V(get,list)",
	}

	assert.Equal(t, rules, graphRole.Rules)
}

func TestConverter_RoleBindingPipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.RoleBindingType]("testdata/rolebinding.json")
	assert.NoError(t, err, "role binding load error")

	c := mocks.NewCacheReader(t)
	k := cache.RoleKey("test-reader")
	id := store.ObjectID().Hex()
	c.EXPECT().Get(mock.Anything, k).Return(id, nil)

	// Collector input -> store rolebinding
	storeBinding, err := NewStoreWithCache(c).RoleBinding(context.TODO(), input)
	assert.NoError(t, err, "store role binding convert error")

	assert.Equal(t, storeBinding.Name, input.Name)
	assert.Equal(t, storeBinding.RoleId.Hex(), id)
	assert.True(t, storeBinding.IsNamespaced)
	assert.Equal(t, storeBinding.Namespace, input.Namespace)

	assert.Equal(t, 1, len(storeBinding.Subjects))
	subject := storeBinding.Subjects[0]

	assert.NotEmpty(t, subject.IdentityId)
	assert.Equal(t, subject.Subject, input.Subjects[0])

	// Collector input -> store identity
	storeIdentity, err := NewStore().Identity(context.TODO(), &subject)
	assert.NoError(t, err, "store identity convert error")

	assert.Equal(t, subject.Subject.Name, storeIdentity.Name)
	assert.Equal(t, subject.Subject.Namespace, storeIdentity.Namespace)
	assert.Equal(t, subject.Subject.Kind, storeIdentity.Type)

	// Store identity -> graph identity
	graphIdentity, err := NewGraph().Identity(storeIdentity)
	assert.NoError(t, err, "graph role binding convert error")

	assert.Equal(t, storeIdentity.Id.Hex(), graphIdentity.StoreId)
	assert.Equal(t, storeIdentity.Name, graphIdentity.Name)
	assert.Equal(t, storeIdentity.Namespace, graphIdentity.Namespace)
	assert.Equal(t, storeIdentity.Type, graphIdentity.Type)
}

func TestConverter_ClusterRoleBindingPipeline(t *testing.T) {
	t.Parallel()

	input, err := loadTestObject[types.ClusterRoleBindingType]("testdata/clusterrolebinding.json")
	assert.NoError(t, err, "cluster role binding load error")

	c := mocks.NewCacheReader(t)
	k := cache.RoleKey("test-reader")
	id := store.ObjectID().Hex()
	c.EXPECT().Get(mock.Anything, k).Return(id, nil)

	// Collector input -> store rolebinding
	storeBinding, err := NewStoreWithCache(c).ClusterRoleBinding(context.TODO(), input)
	assert.NoError(t, err, "store cluster role binding convert error")

	assert.Equal(t, storeBinding.Name, input.Name)
	assert.Equal(t, storeBinding.RoleId.Hex(), id)
	assert.False(t, storeBinding.IsNamespaced)
	assert.Empty(t, storeBinding.Namespace)

	assert.Equal(t, 1, len(storeBinding.Subjects))
	subject := storeBinding.Subjects[0]

	assert.NotEmpty(t, subject.IdentityId)
	assert.Equal(t, subject.Subject, input.Subjects[0])

	// Collector input -> store identity
	storeIdentity, err := NewStore().Identity(context.TODO(), &subject)
	assert.NoError(t, err, "store identity convert error")

	assert.Equal(t, subject.Subject.Name, storeIdentity.Name)
	assert.Equal(t, subject.Subject.Namespace, storeIdentity.Namespace)
	assert.Equal(t, subject.Subject.Kind, storeIdentity.Type)

	// Store identity -> graph identity
	graphIdentity, err := NewGraph().Identity(storeIdentity)
	assert.NoError(t, err, "graph role binding convert error")

	assert.Equal(t, storeIdentity.Id.Hex(), graphIdentity.StoreId)
	assert.Equal(t, storeIdentity.Name, graphIdentity.Name)
	assert.Equal(t, storeIdentity.Namespace, graphIdentity.Namespace)
	assert.Equal(t, storeIdentity.Type, graphIdentity.Type)
}

func TestConverter_RoleCacheFailure(t *testing.T) {
	t.Parallel()

	c := mocks.NewCacheReader(t)
	c.EXPECT().Get(mock.Anything, mock.Anything).Return("", errors.New("not found")).Twice()

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
	k := cache.NodeKey("test-node.ec2.internal")
	id := store.ObjectID().Hex()
	c.EXPECT().Get(mock.Anything, k).Return(id, nil)

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

	assert.Equal(t, storePod.Id.Hex(), graphPod.StoreId)
	assert.Equal(t, storePod.K8.Name, graphPod.Name)
	assert.Equal(t, storePod.K8.Namespace, graphPod.Namespace)
	assert.False(t, graphPod.SharedProcessNamespace)
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
	nk := cache.NodeKey("test-node.ec2.internal")
	nid := store.ObjectID().Hex()
	c.EXPECT().Get(mock.Anything, nk).Return(nid, nil)

	ck := cache.ContainerKey("app-monitors-client-78cb6d7899-j2rjp", "elasticsearch")
	cid := store.ObjectID().Hex()
	c.EXPECT().Get(mock.Anything, ck).Return(cid, nil)

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
	graphContainer, err := NewGraph().Container(storeContainer)
	assert.NoError(t, err, "graph container convert error")

	assert.Equal(t, storeContainer.Id.Hex(), graphContainer.StoreId)
	assert.Equal(t, storeContainer.K8.Name, graphContainer.Name)
	assert.Equal(t, storeContainer.K8.Image, graphContainer.Image)
	assert.Equal(t, storeContainer.K8.Command, graphContainer.Command)
	assert.Equal(t, storeContainer.K8.Args, graphContainer.Args)
	assert.Equal(t, storeContainer.Inherited.PodName, graphContainer.Pod)
	assert.Equal(t, storeContainer.Inherited.NodeName, graphContainer.Node)
	assert.Equal(t, []int{9200, 9300}, graphContainer.Ports)
	assert.Equal(t, shared.CompromiseNone, graphContainer.Compromised)
	assert.False(t, graphContainer.Critical)

	// Collector volume -> store volume
	assert.Equal(t, 2, len(input.Spec.Volumes))
	inVolume0 := input.Spec.Volumes[0]
	storeVolume0, err := NewStoreWithCache(c).Volume(context.TODO(), &inVolume0, storePod)
	assert.NoError(t, err, "store volume convert error")

	assert.Equal(t, storeVolume0.NodeId.Hex(), nid)
	assert.Equal(t, storeVolume0.PodId.Hex(), storePod.Id.Hex())
	assert.Equal(t, storeVolume0.Name, inVolume0.Name)
	assert.Equal(t, storeVolume0.Source, inVolume0)

	inVolume1 := input.Spec.Volumes[1]
	storeVolume1, err := NewStoreWithCache(c).Volume(context.TODO(), &inVolume1, storePod)
	assert.NoError(t, err, "store volume convert error")

	assert.Equal(t, storeVolume1.NodeId.Hex(), nid)
	assert.Equal(t, storeVolume1.PodId.Hex(), storePod.Id.Hex())
	assert.Equal(t, storeVolume1.Name, inVolume1.Name)
	assert.Equal(t, storeVolume1.Source, inVolume1)

	// Store container -> graph container
	graphVolume, err := NewGraph().Volume(storeVolume0)
	assert.NoError(t, err, "graph volume convert error")

	assert.Equal(t, storeVolume0.Id.Hex(), graphVolume.StoreId)
	assert.Equal(t, storeVolume0.Name, graphVolume.Name)
	assert.Equal(t, shared.VolumeTypeProjected, graphVolume.Type)
	assert.Equal(t, "/var/lib/kubelet/pods/5a9fc508-8410-444a-bf63-9f11e5979bee/volumes/kubernetes.io~projected/kube-api-access-4x9fz/token", graphVolume.NodePath)

	graphVolume, err = NewGraph().Volume(storeVolume1)
	assert.NoError(t, err, "graph volume convert error")

	assert.Equal(t, storeVolume1.Id.Hex(), graphVolume.StoreId)
	assert.Equal(t, storeVolume1.Name, graphVolume.Name)
	assert.Equal(t, shared.VolumeTypeHost, graphVolume.Type)
	assert.Equal(t, "/var/run/datadog-agent", graphVolume.NodePath)
}

func TestConverter_PodCacheFailure(t *testing.T) {
	t.Parallel()

	c := mocks.NewCacheReader(t)
	c.EXPECT().Get(mock.Anything, mock.Anything).Return("", errors.New("not found"))

	input, err := loadTestObject[types.PodType]("testdata/pod.json")
	assert.NoError(t, err, "pod load error")

	_, err = NewStoreWithCache(c).Pod(context.TODO(), input)
	assert.ErrorContains(t, err, "not found")
}
