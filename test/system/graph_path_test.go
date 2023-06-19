package system

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

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

	// rawCount, err := g.V().
	// 	Has("class", vertex.VolumeLabel).
	// 	Repeat(__.Out().SimplePath()).
	// 	Until(__.Has("class", vertex.IdentityLabel)).
	// 	Path().
	// 	Count().
	// 	Next()
	// ids := []interface{}{
	// 	"648c8af0c798a6244dea7644", "648c8af1c798a6244dea7656", "648c8af1c798a6244dea766e",
	// }
	// var P = gremlingo.P
	// rawCount, err := g.V().
	// 	Has("class", vertex.IdentityLabel).
	// 	Has("storeID", P.Within(ids...)).
	// 	Count().
	// 	Next()

	verticesToInsert := []map[string]interface{}{
		{"name": "PodTest1", "namespace": "nonexistant"},
		{"name": "PodTest2", "namespace": "nonexistant"},
		{"name": "PodTest3", "namespace": "nonexistant"},
	}

	// Build the traversal query
	res := g.Inject(verticesToInsert).
		Unfold().As("properties").
		AddV("Pod").
		Property("class", "Pod").
		Property("name", __.Select("properties").Select("name")).
		Property("namespace", __.Select("properties").Select("namespace")).
		Iterate()
	err := <-res
	assert.NoError(suite.T(), err)
	// pathCount, err := rawCount.GetInt()
	// assert.NoError(suite.T(), err)
	// assert.NotEqual(suite.T(), pathCount, 0)

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

	// assert.Equal(suite.T(), expectedTokenCount, pathCount)
}

func TestPathTestSuite(t *testing.T) {
	suite.Run(t, new(PathTestSuite))
}

func (suite *PathTestSuite) TearDownTest() {
	suite.gdb.Close(context.Background())
}
