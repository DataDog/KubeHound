package dump

import (
	"context"
	"fmt"
	"path"
	"reflect"
	"testing"
	"time"

	"github.com/DataDog/KubeHound/pkg/collector"
	"github.com/DataDog/KubeHound/pkg/dump/writer"
	mockwriter "github.com/DataDog/KubeHound/pkg/dump/writer/mockwriter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func newFakeDumpIngestor(ctx context.Context, t *testing.T, clientset *fake.Clientset) *DumpIngestor {
	t.Helper()
	directoryOutput := "/tmp"
	dumpWriter := mockwriter.NewDumperWriter(t)
	collectorClient := collector.NewTestK8sAPICollector(ctx, clientset)

	// Generate path for the dump
	clusterName := "test-cluster"
	// ./<clusterName>/kubehound_<clusterName>_<date>
	resName := path.Join(clusterName, fmt.Sprintf("%s%s_%s", OfflineDumpPrefix, clusterName, time.Now().Format(OfflineDumpDateFormat)))

	return &DumpIngestor{
		directoryOutput: directoryOutput,
		collector:       collectorClient,
		ClusterName:     clusterName,
		ResName:         resName,
		writer:          dumpWriter,
	}
}

func TestDumper_IngestEndpoint(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Helper()
	clientset := fake.NewSimpleClientset()
	dumperIngestor := newFakeDumpIngestor(ctx, t, clientset)

	// no ingestion
	noIngest := func(t *testing.T, _ []*discoveryv1.EndpointSlice) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)

		return m
	}

	// ingesting n entries
	nIngests := func(t *testing.T, endpoints []*discoveryv1.EndpointSlice) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)

		for _, endpoint := range endpoints {
			m.EXPECT().Write(mock.Anything, mock.Anything, ingestEndpointPath(endpoint)).Return(nil).Once()
		}

		return m
	}

	type args struct {
		endpoints []*discoveryv1.EndpointSlice
	}
	tests := []struct {
		name           string
		dumperIngestor *DumpIngestor
		testfct        func(t *testing.T, endpoints []*discoveryv1.EndpointSlice) *mockwriter.DumperWriter
		args           args
		wantErr        bool
	}{
		{
			name:           "no entry",
			dumperIngestor: dumperIngestor,
			testfct:        noIngest,
			args: args{
				endpoints: []*discoveryv1.EndpointSlice{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:           "entries found",
			dumperIngestor: dumperIngestor,
			testfct:        nIngests,
			args: args{
				endpoints: []*discoveryv1.EndpointSlice{
					collector.FakeEndpoint("name1", "namespace1", []int32{int32(80)}),
					collector.FakeEndpoint("name2", "namespace2", []int32{int32(443)}),
				},
			},
			wantErr: false,
		},
		{
			name:           "endpoint with no port",
			dumperIngestor: dumperIngestor,
			testfct:        noIngest,
			args: args{
				endpoints: []*discoveryv1.EndpointSlice{
					collector.FakeEndpoint("name1", "namespace1", []int32{}),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.dumperIngestor.writer = tt.testfct(t, tt.args.endpoints)
			for _, endpoint := range tt.args.endpoints {
				if err := tt.dumperIngestor.IngestEndpoint(ctx, endpoint); (err != nil) != tt.wantErr {
					t.Errorf("Dumper.IngestEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestDumpIngestor_IngestNode(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Helper()
	clientset := fake.NewSimpleClientset()
	dumperIngestor := newFakeDumpIngestor(ctx, t, clientset)

	// no ingestion
	noIngest := func(t *testing.T, _ []*corev1.Node) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)

		return m
	}

	// ingesting n entries
	nIngest := func(t *testing.T, nodes []*corev1.Node) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)
		for range nodes {
			m.EXPECT().Write(mock.Anything, mock.Anything, collector.NodePath).Return(nil).Once()
		}

		return m
	}

	type args struct {
		nodes []*corev1.Node
	}
	tests := []struct {
		name           string
		dumperIngestor *DumpIngestor
		testfct        func(t *testing.T, nodes []*corev1.Node) *mockwriter.DumperWriter
		args           args
		wantErr        bool
	}{
		{
			name:           "no entry",
			dumperIngestor: dumperIngestor,
			testfct:        noIngest,
			args: args{
				nodes: []*corev1.Node{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:           "entries found",
			dumperIngestor: dumperIngestor,
			testfct:        nIngest,
			args: args{
				nodes: []*corev1.Node{
					collector.FakeNode("name1", "provider1"),
					collector.FakeNode("name2", "provider2"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.dumperIngestor.writer = tt.testfct(t, tt.args.nodes)
			for _, node := range tt.args.nodes {
				if err := tt.dumperIngestor.IngestNode(ctx, node); (err != nil) != tt.wantErr {
					t.Errorf("Dumper.IngestNode() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestDumpIngestor_IngestPod(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Helper()
	clientset := fake.NewSimpleClientset()
	dumperIngestor := newFakeDumpIngestor(ctx, t, clientset)

	// no ingestion
	noIngest := func(t *testing.T, _ []*corev1.Pod) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)

		return m
	}

	// ingesting n entries
	nIngests := func(t *testing.T, pods []*corev1.Pod) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)

		for _, pod := range pods {
			m.EXPECT().Write(mock.Anything, mock.Anything, ingestPodPath(pod)).Return(nil).Once()
		}

		return m
	}

	type args struct {
		pods []*corev1.Pod
	}
	tests := []struct {
		name           string
		dumperIngestor *DumpIngestor
		testfct        func(t *testing.T, pods []*corev1.Pod) *mockwriter.DumperWriter
		args           args
		wantErr        bool
	}{
		{
			name:           "no entry",
			dumperIngestor: dumperIngestor,
			testfct:        noIngest,
			args: args{
				pods: []*corev1.Pod{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:           "entries found",
			dumperIngestor: dumperIngestor,
			testfct:        nIngests,
			args: args{
				pods: []*corev1.Pod{
					collector.FakePod("name1", "image1", "Running"),
					collector.FakePod("name2", "image2", "Running"),
				},
			},
			wantErr: false,
		},
		{
			name:           "pods not running",
			dumperIngestor: dumperIngestor,
			testfct:        noIngest,
			args: args{
				pods: []*corev1.Pod{
					collector.FakePod("name1", "image1", "Failed"),
					collector.FakePod("name2", "image1", "Pending"),
					collector.FakePod("name3", "image1", "Succeeded"),
					collector.FakePod("name4", "image1", "Unknown"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.dumperIngestor.writer = tt.testfct(t, tt.args.pods)
			for _, pod := range tt.args.pods {
				if err := tt.dumperIngestor.IngestPod(ctx, pod); (err != nil) != tt.wantErr {
					t.Errorf("Dumper.IngestEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestDumpIngestor_IngestRoleBinding(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Helper()
	clientset := fake.NewSimpleClientset()
	dumperIngestor := newFakeDumpIngestor(ctx, t, clientset)

	// no ingestion
	noIngest := func(t *testing.T, _ []*rbacv1.RoleBinding) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)

		return m
	}

	// ingesting n entries
	nIngests := func(t *testing.T, roleBindings []*rbacv1.RoleBinding) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)

		for _, roleDinding := range roleBindings {
			m.EXPECT().Write(mock.Anything, mock.Anything, ingestRoleBindingPath(roleDinding)).Return(nil).Once()
		}

		return m
	}

	type args struct {
		roleBindings []*rbacv1.RoleBinding
	}
	tests := []struct {
		name           string
		dumperIngestor *DumpIngestor
		testfct        func(t *testing.T, roleBindings []*rbacv1.RoleBinding) *mockwriter.DumperWriter
		args           args
		wantErr        bool
	}{
		{
			name:           "no entry",
			dumperIngestor: dumperIngestor,
			testfct:        noIngest,
			args: args{
				roleBindings: []*rbacv1.RoleBinding{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:           "entries found",
			dumperIngestor: dumperIngestor,
			testfct:        nIngests,
			args: args{
				roleBindings: []*rbacv1.RoleBinding{
					collector.FakeRoleBinding("namespace1", "name1"),
					collector.FakeRoleBinding("namespace2", "name2"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.dumperIngestor.writer = tt.testfct(t, tt.args.roleBindings)
			for _, roleBinding := range tt.args.roleBindings {
				if err := tt.dumperIngestor.IngestRoleBinding(ctx, roleBinding); (err != nil) != tt.wantErr {
					t.Errorf("Dumper.IngestEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestDumpIngestor_IngestRole(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Helper()
	clientset := fake.NewSimpleClientset()
	dumperIngestor := newFakeDumpIngestor(ctx, t, clientset)

	// no ingestion
	noIngest := func(t *testing.T, _ []*rbacv1.Role) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)

		return m
	}

	// ingesting n entries
	nIngests := func(t *testing.T, roles []*rbacv1.Role) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)

		for _, role := range roles {
			m.EXPECT().Write(mock.Anything, mock.Anything, ingestRolePath(role)).Return(nil).Once()
		}

		return m
	}

	type args struct {
		roles []*rbacv1.Role
	}
	tests := []struct {
		name           string
		dumperIngestor *DumpIngestor
		testfct        func(t *testing.T, pods []*rbacv1.Role) *mockwriter.DumperWriter
		args           args
		wantErr        bool
	}{
		{
			name:           "no entry",
			dumperIngestor: dumperIngestor,
			testfct:        noIngest,
			args: args{
				roles: []*rbacv1.Role{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:           "entries found",
			dumperIngestor: dumperIngestor,
			testfct:        nIngests,
			args: args{
				roles: []*rbacv1.Role{
					collector.FakeRole("namespace1", "name1"),
					collector.FakeRole("namespace2", "name2"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.dumperIngestor.writer = tt.testfct(t, tt.args.roles)
			for _, role := range tt.args.roles {
				if err := tt.dumperIngestor.IngestRole(ctx, role); (err != nil) != tt.wantErr {
					t.Errorf("Dumper.IngestEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestDumpIngestor_IngestClusterRole(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Helper()
	clientset := fake.NewSimpleClientset()
	dumperIngestor := newFakeDumpIngestor(ctx, t, clientset)

	// no ingestion
	noIngest := func(t *testing.T, _ []*rbacv1.ClusterRole) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)

		return m
	}

	// ingesting n entries
	nIngest := func(t *testing.T, clusterRole []*rbacv1.ClusterRole) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)
		for range clusterRole {
			m.EXPECT().Write(mock.Anything, mock.Anything, collector.ClusterRolesPath).Return(nil).Once()
		}

		return m
	}

	type args struct {
		clusterRole []*rbacv1.ClusterRole
	}
	tests := []struct {
		name           string
		dumperIngestor *DumpIngestor
		testfct        func(t *testing.T, clusterRole []*rbacv1.ClusterRole) *mockwriter.DumperWriter
		args           args
		wantErr        bool
	}{
		{
			name:           "no entry",
			dumperIngestor: dumperIngestor,
			testfct:        noIngest,
			args: args{
				clusterRole: []*rbacv1.ClusterRole{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:           "entries found",
			dumperIngestor: dumperIngestor,
			testfct:        nIngest,
			args: args{
				clusterRole: []*rbacv1.ClusterRole{
					collector.FakeClusterRole("name1"),
					collector.FakeClusterRole("name2"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.dumperIngestor.writer = tt.testfct(t, tt.args.clusterRole)
			for _, clusterRole := range tt.args.clusterRole {
				if err := tt.dumperIngestor.IngestClusterRole(ctx, clusterRole); (err != nil) != tt.wantErr {
					t.Errorf("Dumper.IngestNode() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestDumpIngestor_IngestClusterRoleBinding(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	t.Helper()
	clientset := fake.NewSimpleClientset()
	dumperIngestor := newFakeDumpIngestor(ctx, t, clientset)

	// no ingestion
	noIngest := func(t *testing.T, _ []*rbacv1.ClusterRoleBinding) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)

		return m
	}

	// ingesting n entries
	nIngest := func(t *testing.T, clusterRoleBidings []*rbacv1.ClusterRoleBinding) *mockwriter.DumperWriter {
		t.Helper()
		m := mockwriter.NewDumperWriter(t)
		for range clusterRoleBidings {
			m.EXPECT().Write(mock.Anything, mock.Anything, collector.ClusterRoleBindingsPath).Return(nil).Once()
		}

		return m
	}

	type args struct {
		clusterRoleBidings []*rbacv1.ClusterRoleBinding
	}
	tests := []struct {
		name           string
		dumperIngestor *DumpIngestor
		testfct        func(t *testing.T, clusterRoleBidings []*rbacv1.ClusterRoleBinding) *mockwriter.DumperWriter
		args           args
		wantErr        bool
	}{
		{
			name:           "no entry",
			dumperIngestor: dumperIngestor,
			testfct:        noIngest,
			args: args{
				clusterRoleBidings: []*rbacv1.ClusterRoleBinding{
					nil,
				},
			},
			wantErr: true,
		},
		{
			name:           "entries found",
			dumperIngestor: dumperIngestor,
			testfct:        nIngest,
			args: args{
				clusterRoleBidings: []*rbacv1.ClusterRoleBinding{
					collector.FakeClusterRoleBinding("name1"),
					collector.FakeClusterRoleBinding("name2"),
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.dumperIngestor.writer = tt.testfct(t, tt.args.clusterRoleBidings)
			for _, clusterRoleBiding := range tt.args.clusterRoleBidings {
				if err := tt.dumperIngestor.IngestClusterRoleBinding(ctx, clusterRoleBiding); (err != nil) != tt.wantErr {
					t.Errorf("Dumper.IngestNode() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestNewDumpIngestor(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	clientset := fake.NewSimpleClientset()
	collectorClient := collector.NewTestK8sAPICollector(ctx, clientset)
	directoryOutput := "/tmp"

	type args struct {
		collectorClient collector.CollectorClient
		compression     bool
		directoryOutput string
	}
	tests := []struct {
		name    string
		args    args
		want    *DumpIngestor
		wantErr bool
	}{

		{
			name: "no compression",
			args: args{
				collectorClient: collectorClient,
				compression:     false,
				directoryOutput: directoryOutput,
			},
			want: &DumpIngestor{
				directoryOutput: directoryOutput,

				writer: &writer.FileWriter{},
			},
			wantErr: false,
		},
		{
			name: "compression activated",
			args: args{
				collectorClient: collectorClient,
				compression:     true,
				directoryOutput: directoryOutput,
			},
			want: &DumpIngestor{
				directoryOutput: directoryOutput,
				writer:          &writer.TarWriter{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := NewDumpIngestor(ctx, tt.args.collectorClient, tt.args.compression, tt.args.directoryOutput)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDumpIngestorsss() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if !assert.Equal(t, reflect.TypeOf(got.writer), reflect.TypeOf(tt.want.writer)) {
				t.Errorf("NewDumpIngestor() = %v, want %v", reflect.TypeOf(got.writer), reflect.TypeOf(tt.want.writer))
			}

			if !assert.Equal(t, got.directoryOutput, tt.want.directoryOutput) {
				t.Errorf("NewDumpIngestor() = %v, want %v", got.directoryOutput, tt.want.directoryOutput)
			}
		})
	}
}
