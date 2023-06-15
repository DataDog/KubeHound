package system

import (
	"context"
	"testing"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/graph"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/storedb"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
	"github.com/stretchr/testify/suite"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/exp/slices"
)

//go:generate go run ./generator ../setup/test-cluster ./vertex.gen.go

var containerToVerify = []string{
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

type VertexTestSuite struct {
	suite.Suite
	gdb    graphdb.Provider
	client *gremlingo.DriverRemoteConnection
	g      *gremlingo.GraphTraversalSource
}

func (suite *VertexTestSuite) SetupSuite() {
	ctx := context.Background()
	gdb, err := graphdb.Factory(ctx, config.MustLoadConfig("./kubehound.yaml"))
	suite.NoError(err)
	suite.gdb = gdb
	suite.client = gdb.Raw().(*gremlingo.DriverRemoteConnection)
	suite.g = gremlingo.Traversal_().WithRemote(suite.client)
	errChan := suite.g.V().Drop().Iterate()
	err = <-errChan
	if err != nil {
		suite.Errorf(err, "error deleting the graphdb:\n")
	}

	provider, err := storedb.NewMongoProvider(ctx, storedb.MongoLocalDatabaseURL, 1*time.Second)
	mongoclient := provider.Raw().(*mongo.Client)
	db := mongoclient.Database("kubehound")
	err = db.Drop(ctx)
	if err != nil {
		suite.Errorf(err, "error deleting the mongodb:\n")
	}
	time.Sleep(2 * time.Second)
	err = runKubeHound()
	suite.NoError(err)
}

func (suite *VertexTestSuite) TestVertexContainer() {
	results, err := suite.g.V().HasLabel(vertex.ContainerLabel).ElementMap().ToList()
	suite.NoError(err)

	suite.Equal(len(expectedContainers), len(results))
	resultsMap := map[string]graph.Container{}
	for _, res := range results {
		res := res.GetInterface()
		converted := res.(map[any]any)

		containerName, ok := converted["name"].(string)
		suite.True(ok, "failed to convert container name to string")

		// This is most likely going to be flaky if we try to match these
		// nodeName, ok := converted["node"].(string)
		// suite.True(ok, "failed to convert node name to string")

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
			Privileged:   privileged,
			PrivEsc:      false,
			HostPID:      false,
			HostPath:     false,
			HostIPC:      false,
			HostNetwork:  false,
			RunAsUser:    0,
			Ports:        []int{},
			Pod:          podName,
			// Node:         nodeName,
			// Compromised:  shared.CompromiseType(compromised),
			Critical: critical,
		}
	}
	suite.Equal(expectedContainers, resultsMap)
}

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
