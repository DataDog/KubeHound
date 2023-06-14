package system

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/driver"
	"github.com/stretchr/testify/suite"
)

// Optional syntactic sugar.
var __ = gremlingo.T__

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
	err = runKubeHound()
	suite.NoError(err)
}

func (suite *VertexTestSuite) TestVertexContainer() {
	g := gremlingo.Traversal_().WithRemote(suite.client)
	results, err := g.V().HasLabel(vertex.ContainerLabel).ElementMap().ToList()
	suite.NoError(err)
	suite.T().Errorf("results: %s", results)
	for _, res := range results {
		suite.T().Errorf("res: %s", res.String())
	}
}

func (suite *VertexTestSuite) TestVertexIdentity() {
	g := gremlingo.Traversal_().WithRemote(suite.client)
	results, err := g.V().HasLabel(vertex.IdentityLabel).ElementMap().ToList()
	suite.NoError(err)
	suite.T().Errorf("results: %s", results)
	for _, res := range results {
		suite.T().Errorf("res: %s", res.String())
	}
}

func (suite *VertexTestSuite) TestVertexNode() {
	g := gremlingo.Traversal_().WithRemote(suite.client)
	results, err := g.V().HasLabel(vertex.NodeLabel).ElementMap().ToList()
	suite.NoError(err)
	suite.T().Errorf("results: %s", results)

	suite.Equal(3, len(results))
	expectedNodeNames := map[string]bool{
		"kubehound.test.local-control-plane": true,
		"kubehound.test.local-worker":        true,
		"kubehound.test.local-worker2":       true,
	}

	for _, res := range results {
		suite.T().Errorf("res: %s", res)
		// cast me
		// resConverted := ??
		suite.Contains(expectedNodeNames, resConverted.Name)
	}
}

func (suite *VertexTestSuite) TestVertexPod() {
	g := gremlingo.Traversal_().WithRemote(suite.client)
	results, err := g.V().HasLabel(vertex.PodLabel).ElementMap().ToList()
	suite.NoError(err)
	suite.T().Errorf("results: %s", results)
	for _, res := range results {
		suite.T().Errorf("res: %s", res.String())
	}
}

func (suite *VertexTestSuite) TestVertexRole() {
	g := gremlingo.Traversal_().WithRemote(suite.client)
	results, err := g.V().HasLabel(vertex.RoleLabel).ElementMap().ToList()
	suite.NoError(err)
	suite.T().Errorf("results: %s", results)
	for _, res := range results {
		suite.T().Errorf("res: %s", res.String())
	}
}

func (suite *VertexTestSuite) TestVertexVolume() {
	g := gremlingo.Traversal_().WithRemote(suite.client)
	results, err := g.V().HasLabel(vertex.VolumeLabel).ElementMap().ToList()
	suite.NoError(err)
	suite.T().Errorf("results: %s", results)
	for _, res := range results {
		suite.T().Errorf("res: %s", res.String())
	}
}

func TestVertexTestSuite(t *testing.T) {
	suite.Run(t, new(VertexTestSuite))
}

func (suite *VertexTestSuite) TearDownTest() {
	suite.gdb.Close(context.Background())
}
