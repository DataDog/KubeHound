package system

import (
	"context"
	"fmt"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
	"github.com/stretchr/testify/suite"
	"golang.org/x/exp/slices"
)

//go:generate go run ./generator ../setup/test-cluster ./vertex.gen.go

const (
	nodePrefix = "kubehound.test.local-"
)

// Optional syntactic sugar.
var __ = gremlingo.T__

var containerToVerify = []string{
	"kube-apiserver",
	"kindnet-cni",
	"kube-controller-manager",
	"kube-apiserver",
}

type VertexTestSuite struct {
	suite.Suite
	gdb    graphdb.Provider
	client *gremlingo.DriverRemoteConnection
}

func (suite *VertexTestSuite) SetupTest() {
	gdb, err := graphdb.Factory(context.Background(), config.MustLoadConfig("./kubehound.yaml"))
	suite.NoError(err)
	suite.gdb = gdb
	suite.client = gdb.Raw().(*gremlingo.DriverRemoteConnection)
	// g := gremlingo.Traversal_().WithRemote(suite.client)
	// errChan := g.V().Drop().Iterate()
	// err = <-errChan
	// if err != nil {
	// 	suite.Errorf(err, "error deleting the graphdb:\n")
	// }
	// err = runKubeHound()
	// suite.NoError(err)
}

func (suite *VertexTestSuite) TestVertexContainer() {
	g := gremlingo.Traversal_().WithRemote(suite.client)
	results, err := g.V().HasLabel(vertex.ContainerLabel).ElementMap().ToList()
	suite.NoError(err)

	suite.Equal(len(expectedContainers), len(results))
	resultsMap := map[string]graph.Container{}
	for _, res := range results {
		res := res.GetInterface()
		converted := res.(map[any]any)

		containerName, ok := converted["name"].(string)
		suite.True(ok, "failed to convert container name to string")

		nodeName, ok := converted["node"].(string)
		suite.True(ok, "failed to convert node name to string")

		podName, ok := converted["pod"].(string)
		suite.True(ok, "failed to convert pod name to string")

		imageName, ok := converted["image"].(string)
		suite.True(ok, "failed to convert image name to string")

		compromised, ok := converted["compromised"].(int)
		suite.True(ok, "failed to convert compromised field to CompromiseType")

		critical, ok := converted["critical"].(bool)
		suite.True(ok, "failed to convert critical field to bool")

		// We skip these because they are built by Kind itself
		if slices.Contains(containerToVerify, containerName) {
			continue
		}

		resultsMap[containerName] = graph.Container{
			StoreID:      "",
			Name:         containerName,
			Image:        imageName,
			Command:      []string{},
			Args:         []string{},
			Capabilities: []string{},
			Privileged:   false,
			PrivEsc:      false,
			HostPID:      false,
			HostPath:     false,
			HostIPC:      false,
			HostNetwork:  false,
			RunAsUser:    0,
			Ports:        []int{},
			Pod:          podName,
			Node:         nodeName,
			Compromised:  shared.CompromiseType(compromised),
			Critical:     critical,
		}
	}
	suite.Equal(expectedContainers, resultsMap)
}

// func (suite *VertexTestSuite) TestVertexIdentity() {
// 	g := gremlingo.Traversal_().WithRemote(suite.client)
// 	results, err := g.V().HasLabel(vertex.IdentityLabel).ElementMap().ToList()
// 	suite.NoError(err)
// 	suite.T().Errorf("results: %s", results)
// 	for _, res := range results {
// 		suite.T().Errorf("res: %s", res.String())
// 	}
// }

func (suite *VertexTestSuite) TestVertexNode() {
	g := gremlingo.Traversal_().WithRemote(suite.client)
	results, err := g.V().HasLabel(vertex.NodeLabel).ElementMap().ToList()
	suite.NoError(err)

	suite.Equal(len(expectedNodes), len(results))
	resultsMap := map[string]graph.Node{}
	for _, res := range results {
		res := res.GetInterface()
		converted := res.(map[any]any)

		nodeName, ok := converted["name"].(string)
		suite.True(ok, "failed to convert node name to string")

		compromised, ok := converted["compromised"].(int)
		suite.True(ok, "failed to convert compromised field to CompromiseType")

		isNamespaced, ok := converted["isNamespaced"].(bool)
		suite.True(ok, "failed to convert isNamespaced field to bool")

		namespace, ok := converted["namespace"].(string)
		suite.True(ok, "failed to convert namespace field to string")

		critical, ok := converted["critical"].(bool)
		suite.True(ok, "failed to convert critical field to bool")

		// Prefix the node with the kind prefix
		// nodeName = nodePrefix + nodeName
		for {
			orig := nodeName
			count := 0
			_, exist := resultsMap[nodeName]
			if exist {
				nodeName = fmt.Sprintf("%s%d", orig, count)
				continue
			}
			break
		}
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

// func (suite *VertexTestSuite) TestVertexPod() {
// 	g := gremlingo.Traversal_().WithRemote(suite.client)
// 	results, err := g.V().HasLabel(vertex.PodLabel).ElementMap().ToList()
// 	suite.NoError(err)
// 	suite.T().Errorf("results: %s", results)
// 	for _, res := range results {
// 		suite.T().Errorf("res: %s", res.String())
// 	}
// }

// func (suite *VertexTestSuite) TestVertexRole() {
// 	g := gremlingo.Traversal_().WithRemote(suite.client)
// 	results, err := g.V().HasLabel(vertex.RoleLabel).ElementMap().ToList()
// 	suite.NoError(err)
// 	suite.T().Errorf("results: %s", results)
// 	for _, res := range results {
// 		suite.T().Errorf("res: %s", res.String())
// 	}
// }

// func (suite *VertexTestSuite) TestVertexVolume() {
// 	g := gremlingo.Traversal_().WithRemote(suite.client)
// 	results, err := g.V().HasLabel(vertex.VolumeLabel).ElementMap().ToList()
// 	suite.NoError(err)
// 	suite.T().Errorf("results: %s", results)
// 	for _, res := range results {
// 		suite.T().Errorf("res: %s", res.String())
// 	}
// }

func TestVertexTestSuite(t *testing.T) {
	suite.Run(t, new(VertexTestSuite))
}

func (suite *VertexTestSuite) TearDownTest() {
	suite.gdb.Close(context.Background())
}
