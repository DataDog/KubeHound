package system

import (
	"context"
	"strings"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
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
	cfg := config.MustLoadConfig("./kubehound.yaml")

	// JanusGraph
	gdb, err := graphdb.Factory(ctx, cfg)
	require.NoError(err, "error deleting the graphdb")
	suite.gdb = gdb
	suite.client = gdb.Raw().(*gremlingo.DriverRemoteConnection)

	suite.g = gremlingo.Traversal_().WithRemote(suite.client)
}

func (suite *VertexTestSuite) TestVertexContainer() {
	results, err := suite.g.V().HasLabel(vertex.ContainerLabel).ElementMap().ToList()
	suite.NoError(err)

	suite.Equal(len(expectedContainers)+numberOfKindDefaultContainer, len(results))
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

		// compromised, ok := converted["compromised"].(int)
		// suite.True(ok, "failed to convert compromised field to CompromiseType")

		critical, ok := converted["critical"].(bool)
		suite.True(ok, "failed to convert critical field to bool")

		privileged, ok := converted["privileged"].(bool)
		suite.True(ok, "failed to convert privileged field to bool")

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
			PrivEsc:      false,
			HostPID:      false,
			HostPath:     false,
			HostIPC:      false,
			HostNetwork:  false,
			RunAsUser:    0,
			Ports:        []int{},
			Pod:          podName,
			// Node:         nodeName, // see comments for converted["node"].(string)
			// Compromised:  shared.CompromiseType(compromised),
			Critical: critical,
		}
	}
	suite.Equal(expectedContainers, resultsMap)
}

func (suite *VertexTestSuite) TestVertexNode() {
	results, err := suite.g.V().HasLabel(vertex.NodeLabel).ElementMap().ToList()
	suite.NoError(err)

	suite.Equal(len(expectedNodes), len(results))
	resultsMap := map[string]graph.Node{}
	for _, res := range results {
		res := res.GetInterface()
		converted := res.(map[any]any)

		nodeName, ok := converted["name"].(string)
		suite.True(ok, "failed to convert node name to string")

		// compromised, ok := converted["compromised"].(int)
		// suite.True(ok, "failed to convert compromised field to CompromiseType")

		isNamespaced, ok := converted["isNamespaced"].(bool)
		suite.True(ok, "failed to convert isNamespaced field to bool")

		namespace, ok := converted["namespace"].(string)
		suite.True(ok, "failed to convert namespace field to string")

		critical, ok := converted["critical"].(bool)
		suite.True(ok, "failed to convert critical field to bool")

		resultsMap[nodeName] = graph.Node{
			Name: nodeName,
			// Compromised:  shared.CompromiseType(compromised),
			IsNamespaced: isNamespaced,
			Namespace:    namespace,
			Critical:     critical,
		}
	}
	suite.Equal(expectedNodes, resultsMap)
}

func (suite *VertexTestSuite) TestVertexPod() {
	results, err := suite.g.V().HasLabel(vertex.PodLabel).ElementMap().ToList()
	suite.NoError(err)

	suite.Equal(len(expectedPods)+numberOfKindDefaultPod, len(results))
	resultsMap := map[string]graph.Pod{}
	for _, res := range results {
		res := res.GetInterface()
		converted := res.(map[any]any)

		podName, ok := converted["name"].(string)
		suite.True(ok, "failed to convert pod name to string")

		// compromised, ok := converted["compromised"].(int)
		// suite.True(ok, "failed to convert compromised field to CompromiseType")

		isNamespaced, ok := converted["isNamespaced"].(bool)
		suite.True(ok, "failed to convert isNamespaced field to bool")

		namespace, ok := converted["namespace"].(string)
		suite.True(ok, "failed to convert namespace field to string")

		critical, ok := converted["critical"].(bool)
		suite.True(ok, "failed to convert critical field to bool")

		sharedProcessNamespace, ok := converted["sharedProcessNamespace"].(bool)
		suite.True(ok, "failed to convert sharedProcessNamespace field to bool")

		serviceAccount, ok := converted["serviceAccount"].(string)
		suite.True(ok, "failed to convert serviceAccount field to bool")

		// We skip pods created by kind automatically
		if prefixInSlice(podName, podToSkip) {
			continue
		}

		resultsMap[podName] = graph.Pod{
			Name:           podName,
			ServiceAccount: serviceAccount,
			// Compromised:            shared.CompromiseType(compromised),
			SharedProcessNamespace: sharedProcessNamespace,
			IsNamespaced:           isNamespaced,
			Namespace:              namespace,
			Critical:               critical,
		}
	}
	suite.Equal(expectedPods, resultsMap)
}

func (suite *VertexTestSuite) TestVertexRole() {
	results, err := suite.g.V().HasLabel(vertex.RoleLabel).Has("name", "impersonate").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(1, len(results))

	results, err = suite.g.V().HasLabel(vertex.RoleLabel).Has("name", "read-secrets").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(1, len(results))

	results, err = suite.g.V().HasLabel(vertex.RoleLabel).Has("name", "create-pods").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(1, len(results))

	results, err = suite.g.V().HasLabel(vertex.RoleLabel).Has("name", "patch-pods").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(1, len(results))

	results, err = suite.g.V().HasLabel(vertex.RoleLabel).Has("name", "rolebind").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(1, len(results))
}

func (suite *VertexTestSuite) TestVertexToken() {
	results, err := suite.g.V().HasLabel(vertex.TokenLabel).Has("type", "ServiceAccount").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(6, len(results))

	results, err = suite.g.V().HasLabel(vertex.TokenLabel).Has("identity", "pod-patch-sa").Has("namespace", "default").Has("critical", false).ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(1, len(results))
}

func (suite *VertexTestSuite) TestVertexVolume() {
	results, err := suite.g.V().HasLabel(vertex.VolumeLabel).ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(54, len(results))

	results, err = suite.g.V().HasLabel(vertex.VolumeLabel).Has("path", "/proc/sys/kernel").Has("name", "nodeproc").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(1, len(results))

	results, err = suite.g.V().HasLabel(vertex.VolumeLabel).Has("path", "/lib/modules").Has("name", "lib-modules").ElementMap().ToList()
	suite.NoError(err)
	suite.Greater(len(results), 1) // Not sure why it has "6"

	results, err = suite.g.V().HasLabel(vertex.VolumeLabel).Has("path", "/var/log").Has("name", "nodelog").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)
}

func (suite *VertexTestSuite) TestVertexIdentity() {
	results, err := suite.g.V().HasLabel(vertex.IdentityLabel).ElementMap().ToList()
	suite.NoError(err)
	suite.Greater(len(results), 50)

	results, err = suite.g.V().HasLabel(vertex.IdentityLabel).Has("name", "tokenget-sa").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)

	results, err = suite.g.V().HasLabel(vertex.IdentityLabel).Has("name", "impersonate-sa").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)

	results, err = suite.g.V().HasLabel(vertex.IdentityLabel).Has("name", "tokenlist-sa").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)

	results, err = suite.g.V().HasLabel(vertex.IdentityLabel).Has("name", "pod-patch-sa").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)

	results, err = suite.g.V().HasLabel(vertex.IdentityLabel).Has("name", "rolebind-sa").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)

	results, err = suite.g.V().HasLabel(vertex.IdentityLabel).Has("name", "pod-create-sa").ElementMap().ToList()
	suite.NoError(err)
	suite.Equal(len(results), 1)
}

func TestVertexTestSuite(t *testing.T) {
	suite.Run(t, new(VertexTestSuite))
}

func (suite *VertexTestSuite) TearDownSuite() {
	suite.gdb.Close(context.Background())
}
