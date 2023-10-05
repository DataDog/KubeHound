package system

import (
	"context"
	"testing"

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

func (suite *EdgeTestSuite) TestEdge_IDENTITY_ASSUME_Node() {
	results, err := suite.g.V().
		HasLabel("Node").
		OutE().HasLabel("IDENTITY_ASSUME").
		InV().HasLabel("Identity").
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
		HasLabel("PermissionSet").
		Has("namespace", "default").
		OutE().HasLabel("POD_PATCH").
		InV().HasLabel("Pod").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[impersonate-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[modload-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[pod-create-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[rolebind-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[tokenlist-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[netadmin-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[priv-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[tokenget-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[nsenter-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[varlog-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[sharedps-pod1]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[sharedps-pod2]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[umh-core-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[pod-patch-pod]",
		"path[map[name:[patch-pods::pod-patch-pods]], map[], map[name:[pod-exec-pod]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_POD_CREATE() {
	// We have one bespoke container running with pod/create permissions which should reach all nodes
	// since they are not namespaced
	results, err := suite.g.V().
		HasLabel("PermissionSet").
		Has("namespace", "default").
		OutE().HasLabel("POD_CREATE").
		InV().HasLabel("Node").
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
		HasLabel("PermissionSet").
		Has("namespace", "default").
		OutE().HasLabel("POD_EXEC").
		InV().HasLabel("Pod").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[impersonate-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[modload-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[pod-create-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[rolebind-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[tokenlist-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[netadmin-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[priv-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[tokenget-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[nsenter-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[varlog-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[sharedps-pod1]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[sharedps-pod2]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[umh-core-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[pod-patch-pod]",
		"path[map[name:[exec-pods::pod-exec-pods]], map[], map[name:[pod-exec-pod]",
	}
	suite.Subset(paths, expected)
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
		HasLabel("Identity").
		Has("namespace", "default").
		OutE().HasLabel("PERMISSION_DISCOVER").
		InV().HasLabel("PermissionSet").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 6)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[rolebind-sa]], map[], map[name:[rolebind::pod-bind-role]",
		"path[map[name:[pod-patch-sa]], map[], map[name:[patch-pods::pod-patch-pods]",
		"path[map[name:[pod-create-sa]], map[], map[name:[create-pods::pod-create-pods]",
		"path[map[name:[tokenget-sa]], map[], map[name:[read-secrets::pod-get-secrets]",
		"path[map[name:[tokenlist-sa]], map[], map[name:[list-secrets::pod-list-secrets]",
		"path[map[name:[pod-exec-sa]], map[], map[name:[exec-pods::pod-exec-pods]",
		"path[map[name:[impersonate-sa]], map[], map[name:[impersonate::pod-impersonate]",
		"path[map[name:[varlog-sa]], map[], map[name:[read-logs::pod-read-logs]",
	}

	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_VOLUME_ACCESS() {
	// Every volume should have a VOLUME_ACCESS incoming from a node
	rawCount, err := suite.g.V().
		HasLabel("Volume").
		Count().Next()

	suite.NoError(err)
	volumeCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.NotEqual(volumeCount, 0)

	rawCount, err = suite.g.V().
		HasLabel("Node").
		OutE().HasLabel("VOLUME_ACCESS").
		InV().HasLabel("Volume").
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
		HasLabel("Volume").
		Count().Next()

	suite.NoError(err)
	volumeCount, err := rawCount.GetInt()
	suite.NoError(err)
	suite.NotEqual(volumeCount, 0)

	rawCount, err = suite.g.V().
		HasLabel("Container").
		OutE().HasLabel("VOLUME_DISCOVER").
		InV().HasLabel("Volume").
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
		HasLabel("PermissionSet").
		Has("namespace", "default").
		OutE().HasLabel("TOKEN_BRUTEFORCE").
		InV().HasLabel("Identity").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 8)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[pod-patch-sa]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[impersonate-sa]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[tokenlist-sa]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[pod-exec-sa]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[tokenget-sa]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[rolebind-sa]",
		"path[map[name:[read-secrets::pod-get-secrets]], map[], map[name:[pod-create-sa]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_TOKEN_LIST() {
	results, err := suite.g.V().
		HasLabel("PermissionSet").
		Has("namespace", "default").
		OutE().HasLabel("TOKEN_LIST").
		InV().HasLabel("Identity").
		Path().
		By(__.ValueMap("name")).
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 8)

	paths := suite.pathsToStringArray(results)
	expected := []string{
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[pod-patch-sa]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[impersonate-sa]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[tokenlist-sa]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[pod-exec-sa]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[tokenget-sa]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[rolebind-sa]",
		"path[map[name:[list-secrets::pod-list-secrets]], map[], map[name:[pod-create-sa]",
	}
	suite.Subset(paths, expected)
}

func (suite *EdgeTestSuite) TestEdge_TOKEN_STEAL() {
	// Every pod in our test cluster should have projected volume holding a token. BUT we only
	// save those with a non-default service account token as shown below.
	results, err := suite.g.V().
		HasLabel("Volume").
		OutE().
		HasLabel("TOKEN_STEAL").
		InV().
		HasLabel("Identity").
		Has("namespace", "default").
		Values("name").
		ToList()

	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 7)

	identities := suite.resultsToStringArray(results)
	expected := []string{
		"tokenget-sa", "impersonate-sa", "pod-create-sa", "pod-patch-sa", "pod-exec-sa", "tokenlist-sa", "rolebind-sa",
	}
	suite.Subset(identities, expected)
}

func (suite *EdgeTestSuite) TestEdge_EXPLOIT_HOST_READ() {
	results, err := suite.g.V().
		HasLabel("Container").
		OutE().HasLabel("VOLUME_DISCOVER").
		InV().HasLabel("Volume").
		Where(__.OutE().HasLabel("EXPLOIT_HOST_READ").
			InV().HasLabel("Node")).
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
		HasLabel("Container").
		OutE().HasLabel("VOLUME_DISCOVER").
		InV().HasLabel("Volume").
		Where(__.OutE().HasLabel("EXPLOIT_HOST_WRITE").
			InV().HasLabel("Node")).
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
			HasLabel("Container").
			Has("name", c).
			Values("node").As("n").
			V().HasLabel("Container").
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
			HasLabel("Container").
			Has("name", c).
			OutE().HasLabel("VOLUME_DISCOVER").
			InV().HasLabel("Volume").
			OutE().HasLabel("EXPLOIT_HOST_TRAVERSE").
			InV().HasLabel("Volume").
			OutE().HasLabel("TOKEN_STEAL").
			InV().HasLabel("Identity").
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
		HasLabel("Endpoint").
		Where(
			__.Has("exposure", P.Eq(int(shared.EndpointExposureClusterIP))).
				OutE("ENDPOINT_EXPLOIT").
				InV().
				HasLabel("Container")).
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
		HasLabel("Endpoint").
		Where(
			__.Has("exposure", P.Eq(int(shared.EndpointExposureNodeIP))).
				OutE("ENDPOINT_EXPLOIT").
				InV().
				HasLabel("Container")).
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
		HasLabel("Endpoint").
		Where(
			__.Has("exposure", P.Eq(int(shared.EndpointExposureExternal))).
				OutE("ENDPOINT_EXPLOIT").
				InV().
				HasLabel("Container")).
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
		HasLabel("Container").
		OutE().HasLabel("SHARE_PS_NAMESPACE").
		InV().HasLabel("Container").
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

func (suite *EdgeTestSuite) Test_NoEdgeCase() {
	// The control pod has no interesting properties and therefore should have NO outgoing edges
	results, err := suite.g.V().
		HasLabel("Container").
		Has("name", "control-pod").
		Out().
		ToList()

	suite.NoError(err)
	suite.Equal(len(results), 0)
}

func (suite *EdgeTestSuite) Test_CE_VAR_LOG_SYMLINK() {
	results, err := suite.g.V().
		HasLabel("Container").
		Has("name", "varlog-pod").
		Out().
		ToList()

	suite.NoError(err)
	suite.Equal(len(results), 1)
}

func TestEdgeTestSuite(t *testing.T) {
	suite.Run(t, new(EdgeTestSuite))
}

func (suite *EdgeTestSuite) TearDownTest() {
	suite.gdb.Close(context.Background())
}
