package system

import (
	"context"
	"testing"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/kubehound/storage/graphdb"
	gremlingo "github.com/apache/tinkerpop/gremlin-go/v3/driver"
	"github.com/stretchr/testify/suite"
)

type DslTestSuite struct {
	suite.Suite
	gdb    graphdb.Provider
	client *gremlingo.DriverRemoteConnection
}

func (suite *DslTestSuite) SetupTest() {
	gdb, err := graphdb.Factory(context.Background(), config.MustLoadConfig(KubeHoundConfigPath))
	suite.Require().NoError(err)
	suite.gdb = gdb
	suite.client = gdb.Raw().(*gremlingo.DriverRemoteConnection)
}

func (suite *DslTestSuite) testScriptArray(script string) []string {
	raw, err := suite.client.Submit(script)
	suite.NoError(err)

	results, err := raw.All()
	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	vals := make([]string, 0, len(results))
	for _, r := range results {
		val := r.GetString()
		vals = append(vals, val)
	}

	return vals
}

func (suite *DslTestSuite) testScriptPath(script string) []string {
	raw, err := suite.client.Submit(script)
	suite.NoError(err)

	results, err := raw.All()
	suite.NoError(err)
	suite.GreaterOrEqual(len(results), 1)

	paths := make([]string, 0, len(results))
	for _, r := range results {
		path, err := r.GetPath()
		suite.NoError(err)
		paths = append(paths, path.String())
	}

	return paths
}

func (suite *DslTestSuite) TestTraversalSource_containers() {
	containers := suite.testScriptArray("kh.containers().has('namespace', 'default').values('name')")
	expected := make([]string, 0)

	for k, _ := range expectedContainers {
		expected = append(expected, k)
	}

	suite.ElementsMatch(containers, expected)
}

func (suite *DslTestSuite) TestTraversalSource_pods() {
	pods := suite.testScriptArray("kh.pods().has('namespace', 'default').values('name')")
	expected := make([]string, 0)

	for k, _ := range expectedPods {
		expected = append(expected, k)
	}

	suite.ElementsMatch(pods, expected)
}

func (suite *DslTestSuite) TestTraversalSource_nodes() {
	nodes := suite.testScriptArray("kh.nodes().values('name')")
	expected := make([]string, 0)

	for k, _ := range expectedNodes {
		expected = append(expected, k)
	}

	suite.ElementsMatch(nodes, expected)
}

func (suite *DslTestSuite) TestTraversalSource_escapes() {
	escapes := suite.testScriptPath("kh.escapes().by('name').by(label).by(label)")
	expected := []string{
		"path[kube-proxy, CE_MODULE_LOAD, Node]",
		"path[kube-proxy, CE_PRIV_MOUNT, Node]",
		"path[sys-ptrace-pod, CE_SYS_PTRACE, Node]",
		"path[priv-pod, CE_MODULE_LOAD, Node]",
		"path[priv-pod, CE_PRIV_MOUNT, Node]",
		"path[nsenter-pod, CE_NSENTER, Node]",
		"path[nsenter-pod, CE_MODULE_LOAD, Node]",
		"path[nsenter-pod, CE_PRIV_MOUNT, Node]",
		"path[kube-proxy, CE_MODULE_LOAD, Node]",
		"path[kube-proxy, CE_PRIV_MOUNT, Node]",
		"path[endpoints-pod, CE_NSENTER, Node]",
		"path[endpoints-pod, CE_MODULE_LOAD, Node]",
		"path[endpoints-pod, CE_PRIV_MOUNT, Node]",
		"path[modload-pod, CE_MODULE_LOAD, Node]",
		"path[kube-proxy, CE_MODULE_LOAD, Node]",
		"path[kube-proxy, CE_PRIV_MOUNT, Node]",
	}

	suite.ElementsMatch(escapes, expected)
}

func (suite *DslTestSuite) TestTraversalSource_endpoints() {
	eps := suite.testScriptArray("kh.endpoints().has('namespace', 'default').values('portName')")
	expected := []string{
		"jmx", "host-port-svc", "webproxy-service-port",
	}

	suite.ElementsMatch(eps, expected)

	eps = suite.testScriptArray("kh.endpoints(EndpointExposure.NodeIP).has('namespace', 'default').values('portName')")
	expected = []string{
		"host-port-svc", "webproxy-service-port",
	}

	suite.ElementsMatch(eps, expected)
}

func (suite *DslTestSuite) TestTraversalSource_services() {
	eps := suite.testScriptArray("kh.services().has('namespace', 'default').values('portName')")
	expected := []string{
		"webproxy-service-port",
	}

	suite.ElementsMatch(eps, expected)
}

func (suite *DslTestSuite) TestTraversalSource_volumes() {
	volumes := suite.testScriptArray("kh.volumes().has('namespace', 'default').values('name')")
	expected := make([]string, 0)

	for k, _ := range expectedVolumes {
		expected = append(expected, k)
	}

	suite.Greater(len(volumes), len(expected)) // Also have projected volumes
	suite.Subset(volumes, expected)
}

func (suite *DslTestSuite) TestTraversalSource_hostMounts() {
	mounts := suite.testScriptArray("kh.hostMounts().has('namespace', 'default').values('name')")
	expected := make([]string, 0)

	for k, _ := range expectedVolumes {
		expected = append(expected, k)
	}

	suite.ElementsMatch(mounts, expected)
}

func (suite *DslTestSuite) TestTraversalSource_identities() {
	ids := suite.testScriptArray("kh.identities().has('namespace', 'default').values('name')")
	expected := []string{
		"varrlog-sa", "rolebind-sa", "tokenget-sa", "pod-create-sa",
		"impersonate-sa", "pod-exec-sa", "tokenlist-sa", "pod-patch-sa",
	}

	suite.ElementsMatch(ids, expected)
}

func (suite *DslTestSuite) TestTraversalSource_sas() {
	ids := suite.testScriptArray("kh.sas().has('namespace', 'default').values('name')")
	expected := []string{
		"varrlog-sa", "rolebind-sa", "tokenget-sa", "pod-create-sa",
		"impersonate-sa", "pod-exec-sa", "tokenlist-sa", "pod-patch-sa",
	}

	suite.ElementsMatch(ids, expected)
}

func (suite *DslTestSuite) TestTraversalSource_users() {
	ids := suite.testScriptArray("kh.users().values('name')")
	expected := []string{
		"system:anonymous", "system:kube-proxy", "system:kube-controller-manager", "system:kube-scheduler",
	}

	suite.ElementsMatch(ids, expected)
}

func (suite *DslTestSuite) TestTraversalSource_groups() {
	ids := suite.testScriptArray("kh.groups().values('name')")
	expected := []string{
		"system:monitoring", "system:unauthenticated", "system:serviceaccounts", "system:nodes",
		"system:bootstrappers:kubeadm:default-node-token", "system:authenticated", "system:masters",
	}

	suite.ElementsMatch(ids, expected)
}

func (suite *DslTestSuite) TestTraversalSource_permissions() {
	ps := suite.testScriptArray("kh.permissions().has('namespace', 'default').values('name')")
	expected := []string{
		"create-pods::pod-create-pods", "impersonate::pod-impersonate", "rolebind::pod-bind-role",
		"read-secrets::pod-get-secrets", "list-secrets::pod-list-secrets", "exec-pods::pod-exec-pods",
		"patch-pods::pod-patch-pods", "read-logs::pod-read-logs",
	}

	suite.ElementsMatch(ps, expected)
}

func (suite *DslTestSuite) TestTraversal_attacks() {
	attacks := suite.testScriptPath("kh.containers('nsenter-pod').attacks().by('name').by(label).by(label)")
	expected := []string{
		"path[nsenter-pod, CE_NSENTER, Node]",
		"path[nsenter-pod, CE_MODULE_LOAD, Node]",
		"path[nsenter-pod, CE_PRIV_MOUNT, Node]",
	}

	suite.ElementsMatch(attacks, expected)
}

func (suite *DslTestSuite) TestTraversal_criticalPaths() {
	attacks := suite.testScriptPath("kh.services().criticalPaths().by(label).by(label).dedup()")

	// There are A LOT of paths in the test cluster. Just sample a few
	expected := []string{
		"path[Endpoint, ENDPOINT_EXPLOIT, Container, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet]",
		"path[Endpoint, ENDPOINT_EXPLOIT, Container, VOLUME_DISCOVER, Volume, TOKEN_STEAL, Identity, PERMISSION_DISCOVER, PermissionSet]",
		"path[Endpoint, ENDPOINT_EXPLOIT, Container, CE_NSENTER, Node, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet]",
		"path[Endpoint, ENDPOINT_EXPLOIT, Container, CE_MODULE_LOAD, Node, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet]",
		"path[Endpoint, ENDPOINT_EXPLOIT, Container, CE_PRIV_MOUNT, Node, IDENTITY_ASSUME, Identity, PERMISSION_DISCOVER, PermissionSet]",
	}

	suite.Subset(attacks, expected)
}

func (suite *DslTestSuite) TestTraversal_hasCriticalPath() {
	attacks := suite.testScriptArray("kh.containers('modload-pod').hasCriticalPath().values('name')")
	suite.ElementsMatch(attacks, []string{"modload-pod"})
}

func (suite *DslTestSuite) TestTraversal_criticalPathsFilter() {
	// There are critical paths from this container
	attacks := suite.testScriptPath("kh.containers('modload-pod').criticalPaths().by(label).by(label).dedup()")
	suite.GreaterOrEqual(len(attacks), 1)

	// But NOT if we exclude CE_MODULE_LOAD
	raw, err := suite.client.Submit("kh.containers('modload-pod').criticalPathsFilter(10, 'CE_MODULE_LOAD').by(label).by(label).dedup()")
	suite.NoError(err)

	results, err := raw.All()
	suite.NoError(err)
	suite.Empty(results)
}

func (suite *DslTestSuite) TestTraversal_critical() {
	raw, err := suite.client.Submit("kh.containers('control-pod').critical().hasNext()")
	suite.NoError(err)

	critical, ok, err := raw.One()
	suite.NoError(err)
	suite.True(ok)
	suite.False(critical.GetBool())

	raw, err = suite.client.Submit("kh.permissions('cluster-admin').critical().hasNext()")
	suite.NoError(err)

	critical, ok, err = raw.One()
	suite.NoError(err)
	suite.True(ok)
	suite.True(critical.GetBool())
}

func TestDslTestSuite(t *testing.T) {
	suite.Run(t, new(DslTestSuite))
}

func (suite *DslTestSuite) TearDownTest() {
	suite.gdb.Close(context.Background())
}
