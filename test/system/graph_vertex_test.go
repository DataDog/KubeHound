//nolint:all
package system

import (
	"context"
	"strings"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/slices"
)

//go:generate go run ./generator ../setup/test-cluster ./vertex.gen.go

var containerToSkip = []string{
	"kube-apiserver",
	"kindnet-cni",
	"kube-controller-manager",
	"kube-apiserver",
	"kube-proxy",
	"kube-scheduler",
	"coredns",
	"etcd",
	"local-path-provisioner",
}

var podToSkip = []string{
	"kube-apiserver",
	"kindnet",
	"kube-controller-manager",
	"kube-apiserver",
	"kube-proxy",
	"kube-scheduler",
	"coredns",
	"etcd",
	"local-path-provisioner",
}

func prefixInSlice(str string, list []string) bool {
	for _, l := range list {
		if strings.HasPrefix(str, l) {
			return true
		}
	}
	return false
}

// numberOfKindDefaultContainer represent the current base count of containers created by Kind
// 13 is composed of:
// - 3 "kindnet" (1 per node)
// - 3 "kubeproxy" (1 per node)
// - 2 coredns
// - 1 etcd
// - 1 kube-scheduler
// - 1 kube-apiserver
// - 1 kube-controller
// - 1 local-path-provisioner
const numberOfKindDefaultContainer = 13

// numberOfKindDefaultPod represent the current base count of containers created by Kind
// 13 is composed of:
// - 3 "kindnet" (1 per node)
// - 3 "kubeproxy" (1 per node)
// - 2 coredns
// - 1 etcd
// - 1 kube-scheduler
// - 1 kube-apiserver
// - 1 kube-controller
// - 1 local-path-provisioner
const numberOfKindDefaultPod = 13

type VertexTestSuite struct {
	suite.Suite
	gdb    graphdb.Provider
	client *gremlingo.DriverRemoteConnection
	g      *gremlingo.GraphTraversalSource
}

func (suite *VertexTestSuite) SetupSuite() {
	require := suite.Require()
	ctx := context.Background()
	cfg := config.MustLoadConfig(ctx, "./kubehound.yaml")

	// JanusGraph
	gdb, err := graphdb.Factory(ctx, cfg)
	require.NoError(err, "error deleting the graphdb")
	suite.gdb = gdb
	suite.client = gdb.Raw().(*gremlingo.DriverRemoteConnection)

	suite.g = gremlingo.Traversal_().WithRemote(suite.client)
}

func (suite *VertexTestSuite) resultsToStringArray(results []*gremlingo.Result) []string {
	vals := make([]string, 0, len(results))
	for _, r := range results {
		val := r.GetString()
		vals = append(vals, val)
	}

	return vals
}

func (suite *VertexTestSuite) TestVertexContainer() {
	results, err := suite.g.V().Has("class", vertex.ContainerLabel).ElementMap().ToList()
	suite.NoError(err)

	suite.Equal(len(expectedContainers), len(results)-numberOfKindDefaultContainer)
	resultsMap := map[string]graph.Container{}
	for _, res := range results {
		res := res.GetInterface()
		converted := res.(map[any]any)

		containerName, ok := converted["name"].(string)
		suite.True(ok, "failed to convert container name to string")

		// This is most likely going to be flaky if we try to match these
		// because the node may change between run if they have the same specs
		// so we ignore the return value and just check that it exists and is a readable as a string
		_, ok = converted["node"].(string)
		suite.True(ok, "failed to convert node name to string")

		podName, ok := converted["pod"].(string)
		suite.True(ok, "failed to convert pod name to string")

		imageName, ok := converted["image"].(string)
		suite.True(ok, "failed to convert image name to string")

		compromised, ok := converted["compromised"].(int32)
		suite.True(ok, "failed to convert compromised field to CompromiseType")

		privileged, ok := converted["privileged"].(bool)
		suite.True(ok, "failed to convert privileged field to bool")

		namespace, ok := converted["namespace"].(string)
		suite.True(ok, "failed to convert privileged field to string")

		hostPID, ok := converted["hostPid"].(bool)
		suite.True(ok, "failed to convert privileged field to bool")

		hostNetwork, ok := converted["hostNetwork"].(bool)
		suite.True(ok, "failed to convert privileged field to bool")

		privEsc, ok := converted["privesc"].(bool)
		suite.True(ok, "failed to convert privileged field to bool")

		hostIPC, ok := converted["hostIpc"].(bool)
		suite.True(ok, "failed to convert privileged field to bool")

		runAsUser, ok := converted["runAsUser"].(int64)
		suite.True(ok, "failed to convert compromised field to CompromiseType")

		// We skip these because they are built by Kind itself
		if slices.Contains(containerToSkip, containerName) {
			continue
		}

		resultsMap[containerName] = graph.Container{
			StoreID:      "",
			Name:         containerName,
			Image:        imageName,
			Command:      []string{},
			Args:         []string{},
			Capabilities: []string{},
			Privileged:   privileged,
			PrivEsc:      privEsc,
			HostPID:      hostPID,
			HostIPC:      hostIPC,
			HostNetwork:  hostNetwork,
			RunAsUser:    runAsUser,
			Ports:        []string{},
			Pod:          podName,
			Namespace:    namespace,
			// Node:         nodeName, // see comments for converted["node"].(string)
			Compromised: shared.CompromiseType(compromised),
		}
	}
	suite.Equal(expectedContainers, resultsMap)
}

func (suite *VertexTestSuite) TestVertexNode() {
	results, err := suite.g.V().Has("class", vertex.NodeLabel).ElementMap().ToList()
	suite.NoError(err)

	suite.Equal(len(expectedNodes), len(results))
	resultsMap := map[string]graph.Node{}
	for _, res := range results {
		res := res.GetInterface()
		converted := res.(map[any]any)

		nodeName, ok := converted["name"].(string)
		suite.True(ok, "failed to convert node name to string")

		compromised, ok := converted["compromised"].(int32)
		suite.True(ok, "failed to convert compromised field to CompromiseType")

		isNamespaced, ok := converted["isNamespaced"].(bool)
		suite.True(ok, "failed to convert isNamespaced field to bool")

		namespace, ok := converted["namespace"].(string)
		suite.True(ok, "failed to convert namespace field to string")

		critical, ok := converted["critical"].(bool)
		suite.True(ok, "failed to convert critical field to bool")

		resultsMap[nodeName] = graph.Node{
			Name:         nodeName,
			Compromised:  shared.CompromiseType(compromised),
			IsNamespaced: isNamespaced,
			Namespace:    namespace,
			Critical:     critical,
		}
	}
	suite.Equal(expectedNodes, resultsMap)
}

func (suite *VertexTestSuite) TestVertexPod() {
	results, err := suite.g.V().Has("class", vertex.PodLabel).ElementMap().ToList()
	suite.NoError(err)

	suite.Equal(len(expectedPods), len(results)-numberOfKindDefaultPod)
	resultsMap := map[string]graph.Pod{}
	for _, res := range results {
		res := res.GetInterface()
		converted := res.(map[any]any)

		podName, ok := converted["name"].(string)
		suite.True(ok, "failed to convert pod name to string")

		compromised, ok := converted["compromised"].(int32)
		suite.True(ok, "failed to convert compromised field to CompromiseType")

		isNamespaced, ok := converted["isNamespaced"].(bool)
		suite.True(ok, "failed to convert isNamespaced field to bool")

		namespace, ok := converted["namespace"].(string)
		suite.True(ok, "failed to convert namespace field to string")

		critical, ok := converted["critical"].(bool)
		suite.True(ok, "failed to convert critical field to bool")

		shareProcessNamespace, ok := converted["shareProcessNamespace"].(bool)
		suite.True(ok, "failed to convert shareProcessNamespace field to bool")

		serviceAccount, ok := converted["serviceAccount"].(string)
		suite.True(ok, "failed to convert serviceAccount field to bool")

		// We skip pods created by kind automatically
		if prefixInSlice(podName, podToSkip) {
			continue
		}

		resultsMap[podName] = graph.Pod{
			Name:                  podName,
			ServiceAccount:        serviceAccount,
			Compromised:           shared.CompromiseType(compromised),
			ShareProcessNamespace: shareProcessNamespace,
			IsNamespaced:          isNamespaced,
			Namespace:             namespace,
			Critical:              critical,
		}
	}
	suite.Equal(expectedPods, resultsMap)
}

func (suite *VertexTestSuite) TestVertexPermissionSet() {
	results, err := suite.g.V().
		Has("class", vertex.PermissionSetLabel).
		Has("namespace", "default").
		Values("name").
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	present := suite.resultsToStringArray(results)
	expected := []string{}
	for _, perm := range expectedPermissionSets {
		if perm.Namespace == "default" {
			expected = append(expected, perm.Name)
		}
	}

	suite.ElementsMatch(present, expected)

	//suite.Subset(present, expected)
}

func (suite *VertexTestSuite) TestVertexCritical() {
	results, err := suite.g.V().
		Has("class", vertex.PermissionSetLabel).
		Has("critical", true).
		Values("role").
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	present := suite.resultsToStringArray(results)
	expected := []string{
		"cluster-admin",
		"system:node-bootstrapper",
		"system:kube-scheduler",
	}

	suite.Subset(present, expected)
}

func (suite *VertexTestSuite) TestVertexVolume() {
	results, err := suite.g.V().Has("class", vertex.VolumeLabel).ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(61, len(results))

	results, err = suite.g.V().Has("class", vertex.VolumeLabel).Has("sourcePath", "/proc/sys/kernel").Has("name", "nodeproc").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(1, len(results))

	results, err = suite.g.V().Has("class", vertex.VolumeLabel).Has("sourcePath", "/lib/modules").Has("name", "lib-modules").ElementMap().ToList()
	suite.NoError(err)
	suite.Greater(len(results), 1) // Not sure why it has "6"

	results, err = suite.g.V().Has("class", vertex.VolumeLabel).Has("sourcePath", "/var/log").Has("name", "nodelog").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)
}

func (suite *VertexTestSuite) TestVertexIdentity() {
	results, err := suite.g.V().Has("class", vertex.IdentityLabel).ElementMap().ToList()
	suite.NoError(err)
	suite.Greater(len(results), 50)

	results, err = suite.g.V().Has("class", vertex.IdentityLabel).Has("name", "tokenget-sa").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)

	results, err = suite.g.V().Has("class", vertex.IdentityLabel).Has("name", "impersonate-sa").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)

	results, err = suite.g.V().Has("class", vertex.IdentityLabel).Has("name", "tokenlist-sa").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)

	results, err = suite.g.V().Has("class", vertex.IdentityLabel).Has("name", "pod-patch-sa").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)

	results, err = suite.g.V().Has("class", vertex.IdentityLabel).Has("name", "pod-create-sa").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)
}

func (suite *VertexTestSuite) TestVertexClusterProperty() {
	// All vertices should have the cluster property set
	results, err := suite.g.V().
		Values("cluster").
		Dedup().
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	present := suite.resultsToStringArray(results)
	expected := []string{
		"kind-kubehound.test.local",
	}

	suite.Subset(present, expected)
}

func (suite *VertexTestSuite) TearDownSuite() {
	suite.gdb.Close(context.Background())
}
