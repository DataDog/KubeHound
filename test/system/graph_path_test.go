package system

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Optional syntactic sugar.
var __ = gremlingo.T__

type PathTestSuite struct {
	suite.Suite
	gdb    graphdb.Provider
	client *gremlingo.DriverRemoteConnection
}

func (suite *PathTestSuite) SetupTest() {
	gdb, err := graphdb.Factory(context.Background(), config.MustLoadConfig(KubeHoundConfigPath))
	suite.NoError(err)
	suite.gdb = gdb
	suite.client = gdb.Raw().(*gremlingo.DriverRemoteConnection)

	suite.NoError(err)
}

func (suite *PathTestSuite) TestPath_TOKEN_STEAL() {
	g := gremlingo.Traversal_().WithRemote(suite.client)

	rawCount, err := g.V().
		HasLabel(vertex.VolumeLabel).
		Repeat(__.Out().SimplePath()).
		Until(__.HasLabel(vertex.IdentityLabel)).
		Path().
		Count().
		Next()

	assert.NoError(suite.T(), err)
	pathCount, err := rawCount.GetInt()
	assert.NoError(suite.T(), err)
	assert.NotEqual(suite.T(), pathCount, 0)

	// Every pod in our test cluster should have projected volume holding a token. BUT we only
	// save those with a non-default service account token as shown below.
	//
	// $ kubectl get sa
	// NAME             SECRETS   AGE
	// default          0         28h
	// impersonate-sa   0         28h
	// pod-create-sa    0         28h
	// pod-patch-sa     0         28h
	// rolebind-sa      0         28h
	// tokenget-sa      0         28h
	// tokenlist-sa     0         28h
	const expectedTokenCount = 6

	assert.Equal(suite.T(), expectedTokenCount, pathCount)
}

func TestPathTestSuite(t *testing.T) {
	suite.Run(t, new(PathTestSuite))
}

func (suite *PathTestSuite) TearDownTest() {
	suite.gdb.Close(context.Background())
}
