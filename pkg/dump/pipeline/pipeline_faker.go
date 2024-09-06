package pipeline

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/DataDog/KubeHound/pkg/collector"
	mockcollector "github.com/DataDog/KubeHound/pkg/collector/mockcollector"
	mockwriter "github.com/DataDog/KubeHound/pkg/dump/writer/mockwriter"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	rbacv1 "k8s.io/api/rbac/v1"
)

func PipelineLiveTest(ctx context.Context, t *testing.T, workerNum int) (*mockwriter.DumperWriter, collector.CollectorClient) {
	t.Helper()
	mDumpWriter, collectorClient := NewFakeDumpIngestorPipeline(ctx, t, false)

	mDumpWriter.EXPECT().WorkerNumber().Return(workerNum)

	// Building the map of the cached k8s objects
	// For namespaced resources, the main key is the namespace
	// For non-namespaced resources, the only key is the k8s object type
	countK8sObjectsByFile := make(map[string]int)
	for _, rObj := range GenK8sObjects() {
		reflectType := reflect.TypeOf(rObj)
		switch reflectType {
		case reflect.TypeOf(&corev1.Node{}):
			countK8sObjectsByFile[collector.NodePath]++
		case reflect.TypeOf(&corev1.Pod{}):
			k8sObj, ok := rObj.(*corev1.Pod)
			if !ok {
				t.Fatalf("failed to cast object to PodType: %s", reflectType.String())
			}
			path := fmt.Sprintf("%s/%s", k8sObj.Namespace, collector.PodPath)
			countK8sObjectsByFile[path]++
		case reflect.TypeOf(&rbacv1.Role{}):
			k8sObj, ok := rObj.(*rbacv1.Role)
			if !ok {
				t.Fatalf("failed to cast object to RoleType: %s", reflectType.String())
			}
			path := fmt.Sprintf("%s/%s", k8sObj.Namespace, collector.RolesPath)
			countK8sObjectsByFile[path]++
		case reflect.TypeOf(&rbacv1.RoleBinding{}):
			k8sObj, ok := rObj.(*rbacv1.RoleBinding)
			if !ok {
				t.Fatalf("failed to cast object to RoleBindingType: %s", reflectType.String())
			}
			path := fmt.Sprintf("%s/%s", k8sObj.Namespace, collector.RoleBindingsPath)
			countK8sObjectsByFile[path]++
		case reflect.TypeOf(&rbacv1.ClusterRole{}):
			countK8sObjectsByFile[collector.ClusterRolesPath]++
		case reflect.TypeOf(&rbacv1.ClusterRoleBinding{}):
			countK8sObjectsByFile[collector.ClusterRoleBindingsPath]++
		case reflect.TypeOf(&discoveryv1.EndpointSlice{}):
			k8sObj, ok := rObj.(*discoveryv1.EndpointSlice)
			if !ok {
				t.Fatalf("failed to cast object to EndpointType: %s", reflectType.String())
			}
			path := fmt.Sprintf("%s/%s", k8sObj.Namespace, collector.EndpointPath)
			countK8sObjectsByFile[path]++
		default:
			t.Fatalf("unknown object type to cast: %s", reflectType.String())
		}
	}

	for range countK8sObjectsByFile {
		mDumpWriter.EXPECT().Write(mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	}

	mDumpWriter.EXPECT().Write(mock.Anything, mock.Anything, collector.MetadataPath).Return(nil).Once()

	return mDumpWriter, collectorClient
}

func NewFakeDumpIngestorPipeline(ctx context.Context, t *testing.T, mockCollector bool) (*mockwriter.DumperWriter, collector.CollectorClient) {
	t.Helper()

	mDumpWriter := mockwriter.NewDumperWriter(t)

	mCollectorClient := mockcollector.NewCollectorClient(t)
	clientset := fake.NewSimpleClientset(GenK8sObjects()...)
	collectorClient := collector.NewTestK8sAPICollector(ctx, clientset)

	if mockCollector {
		return mDumpWriter, mCollectorClient
	}

	return mDumpWriter, collectorClient

}

func GenK8sObjects() []runtime.Object {
	k8sOjb := []runtime.Object{
		collector.FakeNode("node1", "provider1"),
		collector.FakePod("namespace1", "name11", "Running"),
		collector.FakePod("namespace1", "name12", "Running"),
		collector.FakePod("namespace2", "name21", "Running"),
		collector.FakePod("namespace2", "name22", "Running"),
		collector.FakeRole("namespace1", "name11"),
		collector.FakeRole("namespace1", "name12"),
		collector.FakeRole("namespace2", "name21"),
		collector.FakeRole("namespace2", "name22"),
		collector.FakeClusterRole("name1"),
		collector.FakeRoleBinding("namespace1", "name11"),
		collector.FakeRoleBinding("namespace1", "name12"),
		collector.FakeRoleBinding("namespace2", "name21"),
		collector.FakeRoleBinding("namespace2", "name22"),
		collector.FakeClusterRoleBinding("name1"),
		collector.FakeEndpoint("namespace1", "name11", []int32{80}),
		collector.FakeEndpoint("namespace1", "name12", []int32{80}),
		collector.FakeEndpoint("namespace2", "name21", []int32{80}),
		collector.FakeEndpoint("namespace2", "name22", []int32{80}),
	}

	return k8sOjb
}
