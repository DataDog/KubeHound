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
	g      *gremlingo.GraphTraversalSource
}

func (suite *EdgeTestSuite) SetupTest() {
	gdb, err := graphdb.Factory(context.Background(), config.MustLoadConfig(KubeHoundConfigPath))
	suite.Require().NoError(err)
	suite.gdb = gdb
	suite.client = gdb.Raw().(*gremlingo.DriverRemoteConnection)
	suite.g = gremlingo.Traversal_().WithRemote(suite.client)

}

var DefaultContainerEscapeNodes = map[string]bool{
	"kubehound.test.local-worker":  true,
	"kubehound.test.local-worker2": true,
}

func (suite *EdgeTestSuite) pathsToStringArray(results []*gremlingo.Result) []string {
	paths := make([]string, 0, len(results))
	for _, r := range results {
		path, err := r.GetPath()
		suite.NoError(err)
		suite.Len(path.Objects, 3)
		paths = append(paths, path.String())
	}

	return paths
}

func (suite *EdgeTestSuite) _testContainerEscape(edgeLabel string, nodes map[string]bool, containers map[string]bool) {
	results, err := suite.g.V().
		Has("class", "Container").
		OutE(edgeLabel).
		InV().
		Dedup().
		Values("name").
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1) // multiple matches as overlaps other container escapes

	matched := false
	for _, r := range results {
		if nodes[r.GetString()] {
			matched = true
			break
		}
	}
	suite.True(matched)

	results, err = suite.g.V().
		Has("class", "Container").
		OutE(edgeLabel).
		OutV().Dedup().
		Values("name").
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1) // multiple matches as overlaps other container escapes

	matched = false
	for _, r := range results {
		if containers[r.GetString()] {
			matched = true
			break
		}
	}

	suite.True(matched)
}

func (suite *EdgeTestSuite) TestEdge_CE_MODULE_LOAD() {
	containers := map[string]bool{
		"modload-pod": true,
	}

	suite._testContainerEscape("CE_MODULE_LOAD", DefaultContainerEscapeNodes, containers)
}

func (suite *EdgeTestSuite) TestEdge_CE_NSENTER() {
	containers := map[string]bool{
		"nsenter-pod": true,
	}

	suite._testContainerEscape("CE_NSENTER", DefaultContainerEscapeNodes, containers)
}

func (suite *EdgeTestSuite) TestEdge_CE_PRIV_MOUNT() {
	containers := map[string]bool{
		"priv-pod": true,
	}

	suite._testContainerEscape("CE_PRIV_MOUNT", DefaultContainerEscapeNodes, containers)
}

func (suite *EdgeTestSuite) TestEdge_CONTAINER_ATTACH() {
	suite.True(false)
}

func (suite *EdgeTestSuite) TestEdge_IDENTITY_ASSUME() {
	// We currently have 6 custom accounts configured (excluding the default)
	// 	➜  KubeHound ✗ k get sa
	// NAME             SECRETS   AGE
	// default          0         7h39m
	// impersonate-sa   0         7h39m
	// pod-create-sa    0         7h39m
	// pod-patch-sa     0         7h39m
	// rolebind-sa      0         7h39m
	// tokenget-sa      0         7h39m
	// tokenlist-sa     0         7h39m

	results, err := suite.g.V().
		HasLabel("Container").
		OutE().HasLabel("IDENTITY_ASSUME").
		InV().HasLabel("Identity").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 6)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[rolebind-pod]], map[], map[name:[rolebind-sa]",
		"path[map[name:[tokenget-pod]], map[], map[name:[tokenget-sa]",
		"path[map[name:[pod-patch-pod]], map[], map[name:[pod-patch-sa]",
		"path[map[name:[tokenlist-pod]], map[], map[name:[tokenlist-sa]",
		"path[map[name:[pod-create-pod]], map[], map[name:[pod-create-sa]",
		"path[map[name:[impersonate-pod]], map[], map[name:[impersonate-sa]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_POD_ATTACH() {

}

func (suite *EdgeTestSuite) TestEdge_POD_PATCH() {
	// We have one bespoke container running with pod/patch permissions
	// 	➜  KubeHound ✗ k get pods
	// NAME              READY   STATUS    RESTARTS   AGE
	// impersonate-pod   1/1     Running   0          8h
	// modload-pod       1/1     Running   0          8h
	// netadmin-pod      1/1     Running   0          8h
	// nsenter-pod       1/1     Running   0          8h
	// pod-create-pod    1/1     Running   0          8h
	// pod-patch-pod     1/1     Running   0          8h
	// priv-pod          1/1     Running   0          8h
	// rolebind-pod      1/1     Running   0          8h
	// sharedps-pod      1/1     Running   0          8h
	// tokenget-pod      1/1     Running   0          8h
	// tokenlist-pod     1/1     Running   0          8h
	// umh-core-pod      1/1     Running   0          8h
	// varlog-pod        1/1     Running   0          8h
	results, err := suite.g.V().
		HasLabel("Role").
		OutE().HasLabel("POD_PATCH").
		InV().HasLabel("Pod").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[patch-pods]], map[], map[name:[impersonate-pod]",
		"path[map[name:[patch-pods]], map[], map[name:[modload-pod]",
		"path[map[name:[patch-pods]], map[], map[name:[pod-create-pod]",
		"path[map[name:[patch-pods]], map[], map[name:[rolebind-pod]",
		"path[map[name:[patch-pods]], map[], map[name:[tokenlist-pod]",
		"path[map[name:[patch-pods]], map[], map[name:[netadmin-pod]",
		"path[map[name:[patch-pods]], map[], map[name:[priv-pod]",
		"path[map[name:[patch-pods]], map[], map[name:[tokenget-pod]",
		"path[map[name:[patch-pods]], map[], map[name:[nsenter-pod]",
		"path[map[name:[patch-pods]], map[], map[name:[varlog-pod]",
		"path[map[name:[patch-pods]], map[], map[name:[sharedps-pod]",
		"path[map[name:[patch-pods]], map[], map[name:[umh-core-pod]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_ROLE_GRANT() {

	// We currently have 6 custom accounts configured (excluding the default)
	// 	➜  KubeHound ✗ k get sa
	// NAME             SECRETS   AGE
	// default          0         7h39m
	// impersonate-sa   0         7h39m
	// pod-create-sa    0         7h39m
	// pod-patch-sa     0         7h39m
	// rolebind-sa      0         7h39m
	// tokenget-sa      0         7h39m
	// tokenlist-sa     0         7h39m
	results, err := suite.g.V().
		HasLabel("Identity").
		OutE().HasLabel("ROLE_GRANT").
		InV().HasLabel("Role").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 6)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[pod-create-sa]], map[], map[name:[create-pods]",
		"path[map[name:[rolebind-sa]], map[], map[name:[rolebind]",
		"path[map[name:[tokenlist-sa]], map[], map[name:[list-secrets]",
		"path[map[name:[impersonate-sa]], map[], map[name:[impersonate]",
		"path[map[name:[tokenget-sa]], map[], map[name:[read-secrets]",
		"path[map[name:[pod-patch-sa]], map[], map[name:[patch-pods]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_VOLUME_MOUNT() {
	suite.True(false)
}

func TestEdgeTestSuite(t *testing.T) {
	suite.Run(t, new(EdgeTestSuite))
}

func (suite *EdgeTestSuite) TearDownTest() {
	suite.gdb.Close(context.Background())
}
