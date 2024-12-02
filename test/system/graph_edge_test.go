//nolint:all
package system

import (
	"context"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/models/shared"
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
	ctx := context.Background()
	gdb, err := graphdb.Factory(ctx, config.MustLoadConfig(ctx, KubeHoundConfigPath))
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

func (suite *EdgeTestSuite) resultsToStringArray(results []*gremlingo.Result) []string {
	vals := make([]string, 0, len(results))
	for _, r := range results {
		val := r.GetString()
		vals = append(vals, val)
	}

	return vals
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

func (suite *EdgeTestSuite) TestEdge_CE_SYS_PTRACE() {
	containers := map[string]bool{
		"sys-ptrace-pod": true,
	}

	suite._testContainerEscape("CE_SYS_PTRACE", DefaultContainerEscapeNodes, containers)
}

func (suite *EdgeTestSuite) TestEdge_CE_UMH_CORE_PATTERN() {
	containers := map[string]bool{
		"umh-core-container": true,
	}

	suite._testContainerEscape("CE_UMH_CORE_PATTERN", DefaultContainerEscapeNodes, containers)
}

func (suite *EdgeTestSuite) TestEdge_CONTAINER_ATTACH() {
	// Every container should have a CONTAINER_ATTACH incoming from a pod
	rawCount, err := suite.g.V().
		Has("class", "Container").
		Count().Next()

	suite.NoError(err)
	containerCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.NotEqual(containerCount, 0)

	rawCount, err = suite.g.V().
		Has("class", "Pod").
		OutE().HasLabel("CONTAINER_ATTACH").
		InV().Has("class", "Container").
		Dedup().
		Path().
		Count().Next()

	suite.NoError(err)
	pathCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.Equal(containerCount, pathCount)
}

func (suite *EdgeTestSuite) TestEdge_IDENTITY_ASSUME_Container() {
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
		Has("class", "Container").
		OutE().HasLabel("IDENTITY_ASSUME").
		InV().Has("class", "Identity").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 6)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[impersonate-pod]], map[], map[name:[impersonate-sa]",
		"path[map[name:[pod-create-pod]], map[], map[name:[pod-create-sa]",
		"path[map[name:[pod-exec-pod]], map[], map[name:[pod-exec-sa]",
		"path[map[name:[pod-patch-pod]], map[], map[name:[pod-patch-sa]",
		"path[map[name:[rolebind-pod-crb-cr-crb-cr]], map[], map[name:[rolebind-sa-crb-cr-crb-cr]",
		"path[map[name:[rolebind-pod-crb-cr-crb-r-fail]], map[], map[name:[rolebind-sa-crb-cr-crb-r-fail]",
		"path[map[name:[rolebind-pod-crb-cr-rb-cr]], map[], map[name:[rolebind-sa-crb-cr-rb-cr]",
		"path[map[name:[rolebind-pod-crb-cr-rb-r]], map[], map[name:[rolebind-sa-crb-cr-rb-r]",
		"path[map[name:[rolebind-pod-rb-cr-crb-cr-fail]], map[], map[name:[rolebind-sa-rb-cr-crb-cr-fail]",
		"path[map[name:[rolebind-pod-rb-cr-rb-cr]], map[], map[name:[rolebind-sa-rb-cr-rb-cr]",
		"path[map[name:[rolebind-pod-rb-cr-rb-r]], map[], map[name:[rolebind-sa-rb-cr-rb-r]",
		"path[map[name:[rolebind-pod-rb-r-crb-cr-fail]], map[], map[name:[rolebind-sa-rb-r-crb-cr-fail]",
		"path[map[name:[rolebind-pod-rb-r-rb-crb]], map[], map[name:[rolebind-sa-rb-r-rb-crb]",
		"path[map[name:[rolebind-pod-rb-r-rb-r]], map[], map[name:[rolebind-sa-rb-r-rb-r]",
		"path[map[name:[tokenget-pod]], map[], map[name:[tokenget-sa]",
		"path[map[name:[tokenlist-pod]], map[], map[name:[tokenlist-sa]",
		"path[map[name:[varlog-container]], map[], map[name:[varlog-sa]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_IDENTITY_ASSUME_Node() {
	results, err := suite.g.V().
		Has("class", "Node").
		OutE().HasLabel("IDENTITY_ASSUME").
		InV().Has("class", "Identity").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 3)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[kubehound.test.local-control-plane]], map[], map[name:[system:nodes]",
		"path[map[name:[kubehound.test.local-worker]], map[], map[name:[system:nodes]",
		"path[map[name:[kubehound.test.local-worker2]], map[], map[name:[system:nodes]",
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
		Has("class", "Node").
		OutE().HasLabel("POD_ATTACH").
		InV().Has("class", "Pod").
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
		Has("class", "PermissionSet").
		Has("namespace", "default").
		OutE().HasLabel("POD_PATCH").
		InV().Has("class", "Pod").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[control-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[endpoints-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[host-read-exploit-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[host-write-exploit-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[impersonate-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[modload-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[netadmin-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[nsenter-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[pod-create-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[pod-exec-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[pod-patch-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[priv-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[rolebind-pod-crb-cr-crb-cr]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[rolebind-pod-crb-cr-crb-r-fail]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[rolebind-pod-crb-cr-rb-cr]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[rolebind-pod-crb-cr-rb-r]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[rolebind-pod-rb-cr-crb-cr-fail]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[rolebind-pod-rb-cr-rb-cr]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[rolebind-pod-rb-cr-rb-r]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[rolebind-pod-rb-r-crb-cr-fail]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[rolebind-pod-rb-r-rb-crb]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[rolebind-pod-rb-r-rb-r]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[sharedps-pod1]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[sharedps-pod2]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[sys-ptrace-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[tokenget-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[tokenlist-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[umh-core-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[varlog-pod]",
	}
	suite.ElementsMatch(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_POD_CREATE() {
	// We have one bespoke container running with pod/create permissions which should reach all nodes
	// since they are not namespaced
	results, err := suite.g.V().
		Has("class", "PermissionSet").
		Has("namespace", "default").
		OutE().HasLabel("POD_CREATE").
		InV().Has("class", "Node").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 3)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[create-pods::pod-create-pods]], map[], map[name:[kubehound.test.local-control-plane]",
		"path[map[name:[create-pods::pod-create-pods]], map[], map[name:[kubehound.test.local-worker]",
		"path[map[name:[create-pods::pod-create-pods]], map[], map[name:[kubehound.test.local-worker2]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_POD_EXEC() {
	// We have one bespoke container running with pod/exec permissions which should reach all pods in the namespace
	results, err := suite.g.V().
		Has("class", "PermissionSet").
		Has("namespace", "default").
		OutE().HasLabel("POD_EXEC").
		InV().Has("class", "Pod").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[control-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[endpoints-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[host-read-exploit-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[host-write-exploit-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[impersonate-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[modload-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[netadmin-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[nsenter-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[pod-create-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[pod-exec-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[pod-patch-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[priv-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[rolebind-pod-crb-cr-crb-cr]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[rolebind-pod-crb-cr-crb-r-fail]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[rolebind-pod-crb-cr-rb-cr]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[rolebind-pod-crb-cr-rb-r]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[rolebind-pod-rb-cr-crb-cr-fail]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[rolebind-pod-rb-cr-rb-cr]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[rolebind-pod-rb-cr-rb-r]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[rolebind-pod-rb-r-crb-cr-fail]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[rolebind-pod-rb-r-rb-crb]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[rolebind-pod-rb-r-rb-r]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[sharedps-pod1]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[sharedps-pod2]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[sys-ptrace-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[tokenget-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[tokenlist-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[umh-core-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[varlog-pod]",
	}
	suite.ElementsMatch(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_PERMISSION_DISCOVER() {

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
		Has("class", "Identity").
		Has("namespace", "default").
		OutE().HasLabel("PERMISSION_DISCOVER").
		InV().Has("class", "PermissionSet").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 6)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[impersonate-sa]], map[], map[name:[impersonate::pod-impersonate]",
		"path[map[name:[pod-create-sa]], map[], map[name:[create-pods::pod-create-pods]",
		"path[map[name:[pod-exec-sa]], map[], map[name:[exec-pods::pod-exec-pods]",
		"path[map[name:[pod-patch-sa]], map[], map[name:[patch-pods::pod-patch-pods]",
		"path[map[name:[rolebind-sa-crb-cr-crb-cr]], map[], map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]",
		"path[map[name:[rolebind-sa-crb-cr-crb-r-fail]], map[], map[name:[rolebind-crb-cr-crb-r-fail::pod-bind-role-crb-cr-crb-r-fail]",
		"path[map[name:[rolebind-sa-crb-cr-rb-cr]], map[], map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]",
		"path[map[name:[rolebind-sa-crb-cr-rb-r]], map[], map[name:[rolebind-crb-cr-rb-r::pod-bind-rola-crb-cr-rb-r]",
		"path[map[name:[rolebind-sa-rb-cr-crb-cr-fail]], map[], map[name:[rolebind-rb-cr-crb-cr-fail::pod-bind-role-rb-cr-crb-cr-fail]",
		"path[map[name:[rolebind-sa-rb-cr-rb-cr]], map[], map[name:[rolebind-rb-cr-rb-cr::pod-bind-role-rb-cr-rb-cr]",
		"path[map[name:[rolebind-sa-rb-cr-rb-r]], map[], map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]",
		"path[map[name:[rolebind-sa-rb-r-crb-cr-fail]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail]",
		"path[map[name:[rolebind-sa-rb-r-rb-crb]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb]",
		"path[map[name:[rolebind-sa-rb-r-rb-r]], map[], map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]",
		"path[map[name:[tokenget-sa]], map[], map[name:[read-secrets::pod-get-secrets]",
		"path[map[name:[tokenlist-sa]], map[], map[name:[list-secrets::pod-list-secrets]",
		"path[map[name:[varlog-sa]], map[], map[name:[read-logs::pod-read-logs]",
	}

	suite.ElementsMatch(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_VOLUME_ACCESS() {
	// Every volume should have a VOLUME_ACCESS incoming from a node
	rawCount, err := suite.g.V().
		Has("class", "Volume").
		Count().Next()

	suite.NoError(err)
	volumeCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.NotEqual(volumeCount, 0)

	rawCount, err = suite.g.V().
		Has("class", "Node").
		OutE().HasLabel("VOLUME_ACCESS").
		InV().Has("class", "Volume").
		Dedup().
		Path().
		Count().Next()

	suite.NoError(err)
	pathCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.Equal(volumeCount, pathCount)
}

func (suite *EdgeTestSuite) TestEdge_VOLUME_DISCOVER() {
	// Every volume should have a VOLUME_DISCOVER incoming from a container
	rawCount, err := suite.g.V().
		Has("class", "Volume").
		Count().Next()

	suite.NoError(err)
	volumeCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.NotEqual(volumeCount, 0)

	rawCount, err = suite.g.V().
		Has("class", "Container").
		OutE().HasLabel("VOLUME_DISCOVER").
		InV().Has("class", "Volume").
		Dedup().
		Path().
		Count().Next()

	suite.NoError(err)
	pathCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.Equal(volumeCount, pathCount)
}

func (suite *EdgeTestSuite) TestEdge_TOKEN_BRUTEFORCE() {
	results, err := suite.g.V().
		Has("class", "PermissionSet").
		Has("namespace", "default").
		OutE().HasLabel("TOKEN_BRUTEFORCE").
		InV().Has("class", "Identity").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 8)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[impersonate-sa]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[pod-create-sa]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[pod-exec-sa]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[pod-patch-sa]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[rolebind-sa-crb-cr-crb-cr]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[rolebind-sa-crb-cr-crb-r-fail]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[rolebind-sa-crb-cr-rb-cr]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[rolebind-sa-crb-cr-rb-r]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[rolebind-sa-rb-cr-crb-cr-fail]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[rolebind-sa-rb-cr-rb-cr]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[rolebind-sa-rb-cr-rb-r]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[rolebind-sa-rb-r-crb-cr-fail]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[rolebind-sa-rb-r-rb-crb]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[rolebind-sa-rb-r-rb-r]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[tokenget-sa]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[tokenlist-sa]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[varlog-sa]",
	}
	suite.ElementsMatch(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_TOKEN_LIST() {
	results, err := suite.g.V().
		Has("class", "PermissionSet").
		Has("namespace", "default").
		OutE().HasLabel("TOKEN_LIST").
		InV().Has("class", "Identity").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 8)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[impersonate-sa]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[pod-create-sa]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[pod-exec-sa]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[pod-patch-sa]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[rolebind-sa-crb-cr-crb-cr]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[rolebind-sa-crb-cr-crb-r-fail]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[rolebind-sa-crb-cr-rb-cr]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[rolebind-sa-crb-cr-rb-r]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[rolebind-sa-rb-cr-crb-cr-fail]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[rolebind-sa-rb-cr-rb-cr]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[rolebind-sa-rb-cr-rb-r]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[rolebind-sa-rb-r-crb-cr-fail]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[rolebind-sa-rb-r-rb-crb]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[rolebind-sa-rb-r-rb-r]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[tokenget-sa]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[tokenlist-sa]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[varlog-sa]",
	}
	suite.ElementsMatch(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_TOKEN_STEAL() {
	// Every pod in our test cluster should have projected volume holding a token. BUT we only
	// save those with a non-default service account token as shown below.
	results, err := suite.g.V().
		Has("class", "Volume").
		OutE().
		HasLabel("TOKEN_STEAL").
		InV().
		Has("class", "Identity").
		Has("namespace", "default").
		Values("name").
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 7)

	identities := suite.resultsToStringArray(results)
	expected := []string{
		"impersonate-sa",
		"pod-create-sa",
		"pod-exec-sa",
		"pod-patch-sa",
		"rolebind-sa-crb-cr-crb-cr",
		"rolebind-sa-crb-cr-crb-r-fail",
		"rolebind-sa-crb-cr-rb-cr",
		"rolebind-sa-crb-cr-rb-r",
		"rolebind-sa-rb-cr-crb-cr-fail",
		"rolebind-sa-rb-cr-rb-cr",
		"rolebind-sa-rb-cr-rb-r",
		"rolebind-sa-rb-r-crb-cr-fail",
		"rolebind-sa-rb-r-rb-crb",
		"rolebind-sa-rb-r-rb-r",
		"tokenget-sa",
		"tokenlist-sa",
		"varlog-sa",
	}
	suite.ElementsMatch(identities, expected)
}

func (suite *EdgeTestSuite) TestEdge_EXPLOIT_HOST_READ() {
	results, err := suite.g.V().
		Has("class", "Container").
		OutE().HasLabel("VOLUME_DISCOVER").
		InV().Has("class", "Volume").
		Where(__.OutE().HasLabel("EXPLOIT_HOST_READ").
			InV().Has("class", "Node")).
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[host-read-exploit-pod]], map[], map[name:[host-ssh]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_EXPLOIT_HOST_WRITE() {
	results, err := suite.g.V().
		Has("class", "Container").
		OutE().HasLabel("VOLUME_DISCOVER").
		InV().Has("class", "Volume").
		Where(__.OutE().HasLabel("EXPLOIT_HOST_WRITE").
			InV().Has("class", "Node")).
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[host-write-exploit-pod]], map[], map[name:[hostroot]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_EXPLOIT_HOST_TRAVERSE() {
	for _, c := range []string{"host-read-exploit-pod", "host-write-exploit-pod"} {
		// Find the containers on the same node as our vulnerable pod and map to their service accounts
		results, err := suite.g.V().
			Has("class", "Container").
			Has("name", c).
			Values("node").As("n").
			V().Has("class", "Container").
			Has("node", __.Where(P.Eq("n"))).
			OutE("IDENTITY_ASSUME").
			InV().
			Values("name").
			ToList()

		suite.NoError(err)
		suite.GreaterOrEqual(len(results), 1)
		expected := suite.resultsToStringArray(results)

		// Now find the identities our vulnerable pod can reach via doing a traverse to the projected token volume
		results, err = suite.g.V().
			Has("class", "Container").
			Has("name", c).
			OutE().HasLabel("VOLUME_DISCOVER").
			InV().Has("class", "Volume").
			OutE().HasLabel("EXPLOIT_HOST_TRAVERSE").
			InV().Has("class", "Volume").
			OutE().HasLabel("TOKEN_STEAL").
			InV().Has("class", "Identity").
			Values("name").
			ToList()

		suite.NoError(err)
		suite.GreaterOrEqual(len(results), 1)
		reachable := suite.resultsToStringArray(results)

		// Assert the 2 lists are the same
		suite.ElementsMatch(expected, reachable)
	}
}

func (suite *EdgeTestSuite) TestEdge_ENDPOINT_EXPLOIT_ContainerPort() {
	results, err := suite.g.V().
		Has("class", "Endpoint").
		Where(
			__.Has("exposure", P.Eq(int(shared.EndpointExposureClusterIP))).
				OutE("ENDPOINT_EXPLOIT").
				InV().
				Has("class", "Container")).
		Values("serviceEndpoint").
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.resultsToStringArray(results)
	expected := []string{
		"jmx",
	}

	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_ENDPOINT_EXPLOIT_NodePort() {
	results, err := suite.g.V().
		Has("class", "Endpoint").
		Where(
			__.Has("exposure", P.Eq(int(shared.EndpointExposureNodeIP))).
				OutE("ENDPOINT_EXPLOIT").
				InV().
				Has("class", "Container")).
		Values("serviceEndpoint").
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.resultsToStringArray(results)
	expected := []string{
		"host-port-svc",
	}

	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_ENDPOINT_EXPLOIT_External() {
	results, err := suite.g.V().
		Has("class", "Endpoint").
		Where(
			__.Has("exposure", P.Eq(int(shared.EndpointExposureExternal))).
				OutE("ENDPOINT_EXPLOIT").
				InV().
				Has("class", "Container")).
		Values("serviceEndpoint").
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.resultsToStringArray(results)
	expected := []string{
		"webproxy-service",
	}

	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_SHARE_PS_NAMESPACE() {
	results, err := suite.g.V().
		Has("class", "Container").
		OutE().HasLabel("SHARE_PS_NAMESPACE").
		InV().Has("class", "Container").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		// Pod1 a 3 containers (A,B,C) = 6 links
		"path[map[name:[sharedps-pod1-a]], map[], map[name:[sharedps-pod1-b]",
		"path[map[name:[sharedps-pod1-a]], map[], map[name:[sharedps-pod1-c]",
		"path[map[name:[sharedps-pod1-b]], map[], map[name:[sharedps-pod1-a]",
		"path[map[name:[sharedps-pod1-b]], map[], map[name:[sharedps-pod1-c]",
		"path[map[name:[sharedps-pod1-c]], map[], map[name:[sharedps-pod1-b]",
		"path[map[name:[sharedps-pod1-c]], map[], map[name:[sharedps-pod1-a]",

		// Pod1 a 2 containers (A,B) = 2 links
		"path[map[name:[sharedps-pod2-a]], map[], map[name:[sharedps-pod2-b]",
		"path[map[name:[sharedps-pod2-b]], map[], map[name:[sharedps-pod2-a]",
	}
	suite.ElementsMatch(paths, expected)
}

// Case 1 (cf docs)
func (suite *EdgeTestSuite) TestEdge_ROLE_BIND_CASE_1() {
	results, err := suite.g.V().
		Has("class", "PermissionSet").
		Has("isNamespaced", false).
		OutE().HasLabel("ROLE_BIND").
		InV().Has("class", "PermissionSet").
		Has("isNamespaced", false).
		// Scoping only to the roles related to the attacks to avoid dependency on the Kind Cluster default roles
		Has("name", gremlingo.TextP.StartingWith("rolebind")).
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]], map[], map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]",
		"path[map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]], map[], map[name:[rolebind-crb-cr-crb-r-fail::pod-bind-role-crb-cr-crb-r-fail]",
		"path[map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]], map[], map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]",
		"path[map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]], map[], map[name:[rolebind-crb-cr-rb-r::pod-bind-rola-crb-cr-rb-r]",
		"path[map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]], map[], map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]",
		"path[map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]], map[], map[name:[rolebind-crb-cr-crb-r-fail::pod-bind-role-crb-cr-crb-r-fail]",
		"path[map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]], map[], map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]",
		"path[map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]], map[], map[name:[rolebind-crb-cr-rb-r::pod-bind-rola-crb-cr-rb-r]",
	}
	suite.ElementsMatch(paths, expected)
}

// Case 2 (cf docs)
func (suite *EdgeTestSuite) TestEdge_ROLE_BIND_CASE_2() {
	results, err := suite.g.V().
		Has("class", "PermissionSet").
		Has("isNamespaced", false).
		OutE().HasLabel("ROLE_BIND").
		InV().Has("class", "PermissionSet").
		Has("isNamespaced", false).
		// Scoping only to the roles related to the attacks to avoid dependency on the Kind Cluster default roles
		Has("name", gremlingo.TextP.StartingWith("rolebind")).
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]], map[], map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]",
		"path[map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]], map[], map[name:[rolebind-crb-cr-crb-r-fail::pod-bind-role-crb-cr-crb-r-fail]",
		"path[map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]], map[], map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]",
		"path[map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]], map[], map[name:[rolebind-crb-cr-rb-r::pod-bind-rola-crb-cr-rb-r]",
		"path[map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]], map[], map[name:[rolebind-crb-cr-crb-cr::pod-bind-role-crb-cr-crb-cr]",
		"path[map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]], map[], map[name:[rolebind-crb-cr-crb-r-fail::pod-bind-role-crb-cr-crb-r-fail]",
		"path[map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]], map[], map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]",
		"path[map[name:[rolebind-crb-cr-rb-cr::pod-bind-role-crb-cr-rb-cr]], map[], map[name:[rolebind-crb-cr-rb-r::pod-bind-rola-crb-cr-rb-r]",
	}
	suite.ElementsMatch(paths, expected)
}

// Case 3 (cf docs)
func (suite *EdgeTestSuite) TestEdge_ROLE_BIND_CASE_3() {
	results, err := suite.g.V().
		Has("class", "PermissionSet").
		Has("isNamespaced", true).
		OutE().HasLabel("ROLE_BIND").
		InV().Has("class", "PermissionSet").
		Has("isNamespaced", true).
		// Scoping only to the roles related to the attacks to avoid dependency on the Kind Cluster default roles
		Has("name", gremlingo.TextP.StartingWith("rolebind-")).
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]], map[], map[name:[rolebind-rb-cr-crb-cr-fail::pod-bind-role-rb-cr-crb-cr-fail]",
		"path[map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail-g]",
		"path[map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail-u]",
		"path[map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail]",
		"path[map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb-g]",
		"path[map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb-u]",
		"path[map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb]",
		"path[map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]], map[], map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]",
		"path[map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]], map[], map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]",
		"path[map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]], map[], map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]], map[], map[name:[rolebind-rb-cr-crb-cr-fail::pod-bind-role-rb-cr-crb-cr-fail]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]], map[], map[name:[rolebind-rb-cr-rb-cr::pod-bind-role-rb-cr-rb-cr]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]], map[], map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail-g]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail-u]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb-g]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb-u]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]], map[], map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]], map[], map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]], map[], map[name:[rolebind-rb-cr-crb-cr-fail::pod-bind-role-rb-cr-crb-cr-fail]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]], map[], map[name:[rolebind-rb-cr-rb-cr::pod-bind-role-rb-cr-rb-cr]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]], map[], map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail-g]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail-u]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb-g]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb-u]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]], map[], map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]], map[], map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]], map[], map[name:[rolebind-rb-cr-crb-cr-fail::pod-bind-role-rb-cr-crb-cr-fail]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]], map[], map[name:[rolebind-rb-cr-rb-cr::pod-bind-role-rb-cr-rb-cr]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]], map[], map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail-g]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail-u]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]], map[], map[name:[rolebind-rb-r-crb-cr-fail::pod-bind-role-rb-r-crb-cr-fail]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb-g]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb-u]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]], map[], map[name:[rolebind-rb-r-rb-crb::pod-bind-role-rb-r-rb-crb]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]], map[], map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-g]",
		"path[map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r]], map[], map[name:[rolebind-rb-r-rb-r::pod-bind-role-rb-r-rb-r-u]",
		"path[map[name:[rolebind-rb-cr-rb-r::pod-bind-role-rb-cr-rb-r]], map[], map[name:[rolebind-rb-cr-rb-cr::pod-bind-role-rb-cr-rb-cr]",
	}
	suite.ElementsMatch(paths, expected)
}

// Case 4 (cf docs)
func (suite *EdgeTestSuite) TestEdge_ROLE_BIND_CASE_4() {
	results, err := suite.g.V().
		Has("class", "PermissionSet").
		Has("isNamespaced", true).
		OutE().HasLabel("ROLE_BIND").
		InV().Has("class", "PermissionSet").
		Has("isNamespaced", false).
		// Scoping only to the roles related to the attacks to avoid dependency on the Kind Cluster default roles
		Has("name", gremlingo.TextP.StartingWith("rolebind")).
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.Equal(0, 0)

	paths := suite.pathsToStringArray(results)
	// not implemented yet
	expected := []string{}
	suite.ElementsMatch(paths, expected)
}

func (suite *EdgeTestSuite) Test_NoEdgeCase() {
	// The control pod has no interesting properties and therefore should have NO outgoing edges
	results, err := suite.g.V().
		Has("class", "Container").
		Has("name", "control-pod").
		Out().
		ToList()

	suite.NoError(err)
	suite.Equal(len(results), 0)
}

func (suite *EdgeTestSuite) Test_CE_VAR_LOG_SYMLINK() {
	containers := map[string]bool{
		"varlog-container": true,
	}

	suite._testContainerEscape("CE_VAR_LOG_SYMLINK", DefaultContainerEscapeNodes, containers)
}

func (suite *EdgeTestSuite) TearDownTest() {
	suite.gdb.Close(context.Background())
}
