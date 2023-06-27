package system

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/graph/vertex"
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
	// Every container should have a CONTAINER_ATTACH incoming from a pod
	rawCount, err := suite.g.V().
		HasLabel("Container").
		Count().Next()

	suite.NoError(err)
	containerCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.NotEqual(containerCount, 0)

	rawCount, err = suite.g.V().
		HasLabel("Pod").
		OutE().HasLabel("CONTAINER_ATTACH").
		InV().HasLabel("Container").
		Dedup().
		Path().
		Count().Next()

	suite.NoError(err)
	pathCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.Equal(containerCount, pathCount)
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
	// Every pod should have a POD_ATTACH incoming from a node
	rawCount, err := suite.g.V().
		HasLabel("Pod").
		Count().Next()

	suite.NoError(err)
	podCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.NotEqual(podCount, 0)

	rawCount, err = suite.g.V().
		HasLabel("Node").
		OutE().HasLabel("POD_ATTACH").
		InV().HasLabel("Pod").
		Dedup().
		Path().
		Count().Next()

	suite.NoError(err)
	pathCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.Equal(podCount, pathCount)
}

func (suite *EdgeTestSuite) TestEdge_POD_PATCH() {
	// We have one bespoke container running with pod/patch permissions which should reach all nodes
	// since they are not namespaced
	results, err := suite.g.V().
		HasLabel("Role").
		OutE().HasLabel("POD_PATCH").
		InV().HasLabel("Node").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 3)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[patch-pods]], map[], map[name:[kubehound.test.local-control-plane]",
		"path[map[name:[patch-pods]], map[], map[name:[kubehound.test.local-worker]",
		"path[map[name:[patch-pods]], map[], map[name:[kubehound.test.local-worker2]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_POD_CREATE() {
	// We have one bespoke container running with pod/create permissions which should reach all nodes
	// since they are not namespaced
	results, err := suite.g.V().
		HasLabel("Role").
		OutE().HasLabel("POD_CREATE").
		InV().HasLabel("Node").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 3)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[create-pods]], map[], map[name:[kubehound.test.local-control-plane]",
		"path[map[name:[create-pods]], map[], map[name:[kubehound.test.local-worker]",
		"path[map[name:[create-pods]], map[], map[name:[kubehound.test.local-worker2]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_POD_EXEC() {
	// We have one bespoke container running with pod/exec permissions which should reach all pods in the namespace
	results, err := suite.g.V().
		HasLabel("Role").
		OutE().HasLabel("POD_EXEC").
		InV().HasLabel("Pod").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[exec-pods]], map[], map[name:[impersonate-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[modload-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[pod-create-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[rolebind-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[tokenlist-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[netadmin-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[priv-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[tokenget-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[nsenter-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[varlog-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[sharedps-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[umh-core-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[pod-patch-pod]",
		"path[map[name:[exec-pods]], map[], map[name:[pod-exec-pod]",
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
	}
	suite.Subset(paths, expected)
}
func (suite *EdgeTestSuite) TestEdge_SHARE_PS_NAMESPACE() {
	// WIP / FIX ME
	results, err := suite.g.V().
		HasLabel(vertex.PodLabel).
		OutE().HasLabel("SHARE_PS_NAMESPACE").
		InV().HasLabel("Pods").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[pod-create-sa]], map[], map[name:[create-pods]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_VOLUME_MOUNT() {
	// Every volume should have a VOLUME_MOUNT incoming from a node
	rawCount, err := suite.g.V().
		HasLabel("Volume").
		Count().Next()

	suite.NoError(err)
	volumeCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.NotEqual(volumeCount, 0)

	rawCount, err = suite.g.V().
		HasLabel("Node").
		OutE().HasLabel("VOLUME_MOUNT").
		InV().HasLabel("Volume").
		Dedup().
		Path().
		Count().Next()

	suite.NoError(err)
	pathCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.Equal(volumeCount, pathCount)

	// Every volume should have a VOLUME_MOUNT incoming from a container
	rawCount, err = suite.g.V().
		HasLabel("Container").
		OutE().HasLabel("VOLUME_MOUNT").
		InV().HasLabel("Volume").
		Dedup().
		Path().
		Count().Next()

	suite.NoError(err)
	pathCount, err = rawCount.GetInt()
	suite.NoError(err)
	suite.Equal(volumeCount, pathCount)
}

func (suite *EdgeTestSuite) TestEdge_TOKEN_BRUTEFOCE() {
	results, err := suite.g.V().
		HasLabel("Role").
		OutE().HasLabel("TOKEN_BRUTEFORCE").
		InV().HasLabel("Identity").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 7)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[read-secrets]], map[], map[name:[pod-patch-sa]",
		"path[map[name:[read-secrets]], map[], map[name:[impersonate-sa]",
		"path[map[name:[read-secrets]], map[], map[name:[tokenlist-sa]",
		"path[map[name:[read-secrets]], map[], map[name:[pod-exec-sa]",
		"path[map[name:[read-secrets]], map[], map[name:[tokenget-sa]",
		"path[map[name:[read-secrets]], map[], map[name:[rolebind-sa]",
		"path[map[name:[read-secrets]], map[], map[name:[pod-create-sa]",
		"path[map[name:[read-secrets]], map[], map[name:[system:kube-proxy]",
	}
	suite.Subset(paths, expected)
}

func TestEdgeTestSuite(t *testing.T) {
	suite.Run(t, new(EdgeTestSuite))
}

func (suite *EdgeTestSuite) TearDownTest() {
	suite.gdb.Close(context.Background())
}
