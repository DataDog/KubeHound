package collector

import (
	"context"
	"testing"

	mocks "github.com/DataDog/KubeHound/pkg/collector/mockingest"
	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/ratelimit"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func NewTestK8sAPICollector(ctx context.Context, clientset *fake.Clientset) CollectorClient {
	cfg := &config.K8SAPICollectorConfig{
		PageSize:           config.DefaultK8sAPIPageSize,
		PageBufferSize:     config.DefaultK8sAPIPageBufferSize,
		RateLimitPerSecond: config.DefaultK8sAPIRateLimitPerSecond,
	}

	return &k8sAPICollector{
		cfg:       cfg,
		clientset: clientset,
		log:       log.Trace(ctx, log.WithComponent(K8sAPICollectorName)),
		rl:        ratelimit.New(config.DefaultK8sAPIRateLimitPerSecond), // per second
	}
}

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
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg, err := config.NewConfig(tt.args.path)
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

func fakePod(namespace string, image string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  image,
					Image: "nginx:latest",
				},
			},
			RestartPolicy: corev1.RestartPolicyAlways,
		},
	}
}

func Test_k8sAPICollector_streamPodsNamespace(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 pod found
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.PodIngestor) {
		clientset := fake.NewSimpleClientset()
		m := mocks.NewPodIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()
		return clientset, m
	}

	// Listing pods from all namespaces
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.PodIngestor) {
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				fakePod("namespace1", "iamge1"),
				fakePod("namespace2", "image2"),
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
		tt := tt
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

func fakeRole(namespace string, name string) *rbacv1.Role {
	return &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Verbs:     []string{"get", "list", "watch"},
				Resources: []string{"pods", "configmaps"},
			},
		},
	}
}

func Test_k8sAPICollector_StreamRoles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 roles
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.RoleIngestor) {
		clientset := fake.NewSimpleClientset()
		m := mocks.NewRoleIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()
		return clientset, m
	}

	// Listing all the roles from all namespaces
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.RoleIngestor) {
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				fakeRole("namespace1", "name1"),
				fakeRole("namespace2", "name2"),
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
		tt := tt
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

func fakeRoleBinding(namespace, name string) *rbacv1.RoleBinding {
	return &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "Role",
			Name:     "my-role",
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "User",
				Name:      "my-user",
				Namespace: namespace,
			},
		},
	}
}

func Test_k8sAPICollector_StreamRoleBindings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 role bindings found
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.RoleBindingIngestor) {
		clientset := fake.NewSimpleClientset()
		m := mocks.NewRoleBindingIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()
		return clientset, m
	}

	// Listing all the roles bindings from all namespaces
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.RoleBindingIngestor) {
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				fakeRoleBinding("namespace1", "name1"),
				fakeRoleBinding("namespace2", "name2"),
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
		tt := tt
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

func fakeNode(name string, providerID string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.NodeSpec{
			ProviderID: providerID,
		},
	}
}

func Test_k8sAPICollector_StreamNodes(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 nodes found
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.NodeIngestor) {
		clientset := fake.NewSimpleClientset()
		m := mocks.NewNodeIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()
		return clientset, m
	}

	// Listing all the nodes in the cluster
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.NodeIngestor) {
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				fakeNode("name1", "uid1"),
				fakeNode("name2", "uid2"),
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
		tt := tt
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

func fakeClusterRole(name string) *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Verbs:     []string{"get", "list", "watch"},
				Resources: []string{"pods", "configmaps"},
			},
		},
	}
}

func Test_k8sAPICollector_StreamClusterRoles(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 cluster roles found
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.ClusterRoleIngestor) {
		clientset := fake.NewSimpleClientset()
		m := mocks.NewClusterRoleIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()
		return clientset, m
	}

	// Listing all the cluster roles
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.ClusterRoleIngestor) {
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				fakeClusterRole("name1"),
				fakeClusterRole("name2"),
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
		tt := tt
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

func fakeClusterRoleBinding(name string) *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		RoleRef: rbacv1.RoleRef{
			Kind:     "ClusterRole",
			Name:     name,
			APIGroup: "rbac.authorization.k8s.io",
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "User",
				Name: "user1",
			},
		},
	}
}

func Test_k8sAPICollector_StreamClusterRoleBindings(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// 0 cluster role bindings found
	test1 := func(t *testing.T) (*fake.Clientset, *mocks.ClusterRoleBindingIngestor) {
		clientset := fake.NewSimpleClientset()
		m := mocks.NewClusterRoleBindingIngestor(t)
		m.EXPECT().Complete(mock.Anything).Return(nil).Once()
		return clientset, m
	}

	// Listing all the cluster roles bindings
	test2 := func(t *testing.T) (*fake.Clientset, *mocks.ClusterRoleBindingIngestor) {
		clienset := fake.NewSimpleClientset(
			[]runtime.Object{
				fakeClusterRoleBinding("name1"),
				fakeClusterRoleBinding("name2"),
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
		tt := tt
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
