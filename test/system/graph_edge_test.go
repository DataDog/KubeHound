package system

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/stretchr/testify/suite"
)

type EdgeTestSuite struct {
	suite.Suite
	gdb    graphdb.Provider
	client *gremlingo.DriverRemoteConnection
}

func (suite *EdgeTestSuite) SetupTest() {
	gdb, err := graphdb.Factory(context.Background(), config.MustLoadConfig(KubeHoundConfigPath))
	suite.NoError(err)
	suite.gdb = gdb
	suite.client = gdb.Raw().(*gremlingo.DriverRemoteConnection)

	suite.NoError(err)
}

// // find the vertices connected by the CE_MODULE_LOAD edge and get the name of the pod and node
// func (suite *EdgeTestSuite) TestEdge_CE_MODULE_LOAD() {
// 	g := gremlingo.Traversal_().WithRemote(suite.client)

// 	// query the name of the node
// 	g.V().Has("class", vertex.PodLabel).OutE("CE_MODULE_LOAD").InV().Values("name").Next()

// 	// query the name of the outE container
// 	g.V().Has("class", vertex.PodLabel).OutE("CE_MODULE_LOAD").Values("name").Next()

// 	// Count of containers
// 	g.V().has("class", "Container").outE("CE_MODULE_LOAD").outV().dedup().values("name")
// 	[priv-pod, nsenter-pod, kube-proxy, kube-proxy, modload-pod, kube-proxy]

// 	g.V().has("class", "Container").outE("CE_MODULE_LOAD").inV().dedup().values("name")
// 	[kubehound.test.local-worker2, kubehound.test.local-worker, kubehound.test.local-control-plane]

// 	g.V().has("class", "Container").outE("CE_MODULE_LOAD").outV().values("name").Count()
// }

// // Find the vertices connected by the ESCAPE_MODULE_LOAD edge
// func (suite *EdgeTestSuite) TestEdge_ESCAPE_MODULE_LOAD() {
// 	g := gremlingo.Traversal_().WithRemote(suite.client)

// 	// We will probably have more than just our example due to the OR condition. So just look for
// 	// our specific example pod.

// 	//

// 	rawCount, err := g.V().
// 		Has("class", vertex.VolumeLabel).
// 		Repeat(__.Out().SimplePath()).
// 		Until(__.Has("class", vertex.IdentityLabel)).
// 		Path().
// 		Count().
// 		Next()

// 	assert.NoError(suite.T(), err)
// 	_, err = rawCount.GetInt()
// 	assert.NoError(suite.T(), err)
// 	// assert.NotEqual(suite.T(), pathCount, 0)

// 	const expectedTokenCount = 6

// 	assert.Equal(suite.T(), expectedTokenCount, pathCount)
// }

func TestEdgeTestSuite(t *testing.T) {
	suite.Run(t, new(EdgeTestSuite))
}

func (suite *EdgeTestSuite) TearDownTest() {
	suite.gdb.Close(context.Background())
}
