//nolint:containedctx
package collector

import (
	"context"
	"testing"

	mocks "github.com/DataDog/KubeHound/pkg/collector/mockingest"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestNewK8sAPICollectorConfig(t *testing.T) {
	ctx := context.Background()
	t.Parallel()

	type args struct {
		ctx    context.Context
		path   string
		values config.K8SAPICollectorConfig
	}
	tests := []struct {
		name    string
		args    args
		want    CollectorClient
		wantErr bool
	}{
		{
			name: "default",
			args: args{
				ctx:  ctx,
				path: "testdata/kubehound-test-live-default.yaml",
				values: config.K8SAPICollectorConfig{
					PageSize:           config.DefaultK8sAPIPageSize,
					PageBufferSize:     config.DefaultK8sAPIPageBufferSize,
					RateLimitPerSecond: config.DefaultK8sAPIRateLimitPerSecond,
				},
			},
			wantErr: false,
		},
		{
			name: "full settings",
			args: args{
				ctx:  ctx,
				path: "testdata/kubehound-test-live-full-spec.yaml",
				values: config.K8SAPICollectorConfig{
					PageSize:           int64(123),
					PageBufferSize:     int32(456),
					RateLimitPerSecond: int(789),
				},
			},
			wantErr: false,
		},
		{
			name: "mixed settings",
			args: args{
				ctx:  ctx,
				path: "testdata/kubehound-test-live-mixed-spec.yaml",
				values: config.K8SAPICollectorConfig{
					PageSize:           int64(123),
					PageBufferSize:     int32(456),
					RateLimitPerSecond: config.DefaultK8sAPIRateLimitPerSecond,
				},
			},
			wantErr: false,
		},
		{
			name: "wrong file settings",
			args: args{
				ctx:  ctx,
				path: "testdata/kubehound-test.yaml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			v := viper.New()
			cfg, err := config.NewConfig(context.TODO(), v, tt.args.path)
			assert.NoError(t, err)
			err = checkK8sAPICollectorConfig(cfg.Collector.Type)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewK8sAPICollectorConfig() error = %v, wantErr %v", err, tt.wantErr)

				return
			}

			if tt.wantErr && err != nil {
				return
			}

			assert.Equal(t, *cfg.Collector.Live, tt.args.values)
		})
	}
}

func Test_k8sAPICollector_streamPodsNamespace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 pod found
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.PodIngestor) {
		t.Helper()
		clientset := fake.NewSimpleClientset()
		m := mocks.NewPodIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clientset, m
	}

	// Listing pods from all namespaces
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.PodIngestor) {
		t.Helper()
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				FakePod("namespace1", "iamge1", "Running"),
				FakePod("namespace2", "image2", "Running"),
			}...,
		)
		m := mocks.NewPodIngestor(t)
		m.EXPECT().IngestPod(mock.Anything, mock.AnythingOfType("types.PodType")).Return(nil).Twice()
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clienset, m
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		testfct func(t *testing.T) (*fake.Clientset, *mocks.PodIngestor)
		args    args
		wantErr bool
	}{
		{
			name:    "no entry",
			testfct: test1,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name:    "all namespace",
			testfct: test2,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			clientset, mock := tt.testfct(t)
			c := NewTestK8sAPICollector(tt.args.ctx, clientset)
			if err := c.StreamPods(tt.args.ctx, mock); (err != nil) != tt.wantErr {
				t.Errorf("k8sAPICollector.StreamPods() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_k8sAPICollector_StreamRoles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 roles
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.RoleIngestor) {
		t.Helper()
		clientset := fake.NewSimpleClientset()
		m := mocks.NewRoleIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clientset, m
	}

	// Listing all the roles from all namespaces
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.RoleIngestor) {
		t.Helper()
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				FakeRole("namespace1", "name1"),
				FakeRole("namespace2", "name2"),
			}...,
		)
		m := mocks.NewRoleIngestor(t)
		m.EXPECT().IngestRole(mock.Anything, mock.AnythingOfType("types.RoleType")).Return(nil).Twice()
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clienset, m
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		testfct func(t *testing.T) (*fake.Clientset, *mocks.RoleIngestor)
		args    args
		wantErr bool
	}{
		{
			name:    "no entry",
			testfct: test1,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name:    "all namespace",
			testfct: test2,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			clientset, mock := tt.testfct(t)
			c := NewTestK8sAPICollector(tt.args.ctx, clientset)
			if err := c.StreamRoles(tt.args.ctx, mock); (err != nil) != tt.wantErr {
				t.Errorf("k8sAPICollector.StreamRoles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_k8sAPICollector_StreamRoleBindings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 role bindings found
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.RoleBindingIngestor) {
		t.Helper()
		clientset := fake.NewSimpleClientset()
		m := mocks.NewRoleBindingIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clientset, m
	}

	// Listing all the roles bindings from all namespaces
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.RoleBindingIngestor) {
		t.Helper()
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				FakeRoleBinding("namespace1", "name1"),
				FakeRoleBinding("namespace2", "name2"),
			}...,
		)
		m := mocks.NewRoleBindingIngestor(t)
		m.EXPECT().IngestRoleBinding(mock.Anything, mock.AnythingOfType("types.RoleBindingType")).Return(nil).Twice()
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clienset, m
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		testfct func(t *testing.T) (*fake.Clientset, *mocks.RoleBindingIngestor)
		args    args
		wantErr bool
	}{
		{
			name:    "no entry",
			testfct: test1,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name:    "all namespace",
			testfct: test2,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			clientset, mock := tt.testfct(t)
			c := NewTestK8sAPICollector(tt.args.ctx, clientset)
			if err := c.StreamRoleBindings(tt.args.ctx, mock); (err != nil) != tt.wantErr {
				t.Errorf("k8sAPICollector.StreamRoleBindings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_k8sAPICollector_StreamNodes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 nodes found
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.NodeIngestor) {
		t.Helper()
		clientset := fake.NewSimpleClientset()
		m := mocks.NewNodeIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clientset, m
	}

	// Listing all the nodes in the cluster
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.NodeIngestor) {
		t.Helper()
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				FakeNode("name1", "uid1"),
				FakeNode("name2", "uid2"),
			}...,
		)
		m := mocks.NewNodeIngestor(t)
		m.EXPECT().IngestNode(mock.Anything, mock.AnythingOfType("types.NodeType")).Return(nil).Twice()
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clienset, m
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		testfct func(t *testing.T) (*fake.Clientset, *mocks.NodeIngestor)
		args    args
		wantErr bool
	}{
		{
			name:    "no entry",
			testfct: test1,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name:    "all namespace",
			testfct: test2,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			clientset, mock := tt.testfct(t)
			c := NewTestK8sAPICollector(tt.args.ctx, clientset)
			if err := c.StreamNodes(tt.args.ctx, mock); (err != nil) != tt.wantErr {
				t.Errorf("k8sAPICollector.StreamNodes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_k8sAPICollector_StreamClusterRoles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 cluster roles found
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.ClusterRoleIngestor) {
		t.Helper()
		clientset := fake.NewSimpleClientset()
		m := mocks.NewClusterRoleIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clientset, m
	}

	// Listing all the cluster roles
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.ClusterRoleIngestor) {
		t.Helper()
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				FakeClusterRole("name1"),
				FakeClusterRole("name2"),
			}...,
		)
		m := mocks.NewClusterRoleIngestor(t)
		m.EXPECT().IngestClusterRole(mock.Anything, mock.AnythingOfType("types.ClusterRoleType")).Return(nil).Twice()
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clienset, m
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		testfct func(t *testing.T) (*fake.Clientset, *mocks.ClusterRoleIngestor)
		args    args
		wantErr bool
	}{
		{
			name:    "no entry",
			testfct: test1,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name:    "all entries",
			testfct: test2,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			clientset, mock := tt.testfct(t)
			c := NewTestK8sAPICollector(tt.args.ctx, clientset)
			if err := c.StreamClusterRoles(tt.args.ctx, mock); (err != nil) != tt.wantErr {
				t.Errorf("k8sAPICollector.StreamClusterRoles() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_k8sAPICollector_StreamClusterRoleBindings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 cluster role bindings found
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.ClusterRoleBindingIngestor) {
		t.Helper()
		clientset := fake.NewSimpleClientset()
		m := mocks.NewClusterRoleBindingIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clientset, m
	}

	// Listing all the cluster roles bindings
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.ClusterRoleBindingIngestor) {
		t.Helper()
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				FakeClusterRoleBinding("name1"),
				FakeClusterRoleBinding("name2"),
			}...,
		)
		m := mocks.NewClusterRoleBindingIngestor(t)
		m.EXPECT().IngestClusterRoleBinding(mock.Anything, mock.AnythingOfType("types.ClusterRoleBindingType")).Return(nil).Twice()
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clienset, m
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		testfct func(t *testing.T) (*fake.Clientset, *mocks.ClusterRoleBindingIngestor)
		args    args
		wantErr bool
	}{
		{
			name:    "no entry",
			testfct: test1,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name:    "all entries",
			testfct: test2,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			clientset, mock := tt.testfct(t)
			c := NewTestK8sAPICollector(tt.args.ctx, clientset)
			if err := c.StreamClusterRoleBindings(tt.args.ctx, mock); (err != nil) != tt.wantErr {
				t.Errorf("k8sAPICollector.StreamClusterRoleBindings() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_k8sAPICollector_StreamEndpoints(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 endpoints found
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.EndpointIngestor) {
		t.Helper()
		clientset := fake.NewSimpleClientset()
		m := mocks.NewEndpointIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clientset, m
	}

	// Listing all the endpoints bindings from all namespaces
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.EndpointIngestor) {
		t.Helper()
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				FakeEndpoint("namespace1", "name1", []int32{80, 443}),
				FakeEndpoint("namespace2", "name2", []int32{80, 443}),
			}...,
		)
		m := mocks.NewEndpointIngestor(t)
		m.EXPECT().IngestEndpoint(mock.Anything, mock.AnythingOfType("types.EndpointType")).Return(nil).Twice()
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()

		return clienset, m
	}

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		testfct func(t *testing.T) (*fake.Clientset, *mocks.EndpointIngestor)
		args    args
		wantErr bool
	}{
		{
			name:    "no entry",
			testfct: test1,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
		{
			name:    "all namespace",
			testfct: test2,
			args: args{
				ctx: ctx,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			clientset, mock := tt.testfct(t)
			c := NewTestK8sAPICollector(tt.args.ctx, clientset)
			if err := c.StreamEndpoints(tt.args.ctx, mock); (err != nil) != tt.wantErr {
				t.Errorf("k8sAPICollector.StreamEndpoints() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
