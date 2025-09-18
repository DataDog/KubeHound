package collector

import (
	"context"
	"sync"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"go.uber.org/ratelimit"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func FakePod(namespace string, name string, status string) *corev1.Pod {
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  name,
					Image: "nginx:latest",
				},
			},
			RestartPolicy: corev1.RestartPolicyAlways,
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodPhase(status),
		},
	}
}

func FakeRole(namespace string, name string) *rbacv1.Role {
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

func FakeRoleBinding(namespace, name string) *rbacv1.RoleBinding {
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

func FakeNode(name string, providerID string) *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Spec: corev1.NodeSpec{
			ProviderID: providerID,
		},
	}
}

func FakeClusterRole(name string) *rbacv1.ClusterRole {
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

func FakeClusterRoleBinding(name string) *rbacv1.ClusterRoleBinding {
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

func FakePort(name string, port int32) *discoveryv1.EndpointPort {
	return &discoveryv1.EndpointPort{
		Port: &port,
		Name: &name,
	}
}

func FakeEndpoint(name string, namespace string, ports []int32) *discoveryv1.EndpointSlice {
	endpointPorts := []discoveryv1.EndpointPort{}
	if len(ports) == 0 {
		endpointPorts = nil
	} else {
		for _, port := range ports {
			endpointPorts = append(endpointPorts, *FakePort(name, port))
		}
	}

	return &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Ports: endpointPorts,
	}
}

func NewTestK8sAPICollector(ctx context.Context, clientset *fake.Clientset) CollectorClient {
	cfg := &config.K8SAPICollectorConfig{
		PageSize:           config.DefaultK8sAPIPageSize,
		PageBufferSize:     config.DefaultK8sAPIPageBufferSize,
		RateLimitPerSecond: config.DefaultK8sAPIRateLimitPerSecond,
	}

	return &k8sAPICollector{
		cfg:       cfg,
		clientset: clientset,
		// log:       log.Trace(ctx, log.WithComponent(K8sAPICollectorName)),
		rl:       ratelimit.New(config.DefaultK8sAPIRateLimitPerSecond), // per second
		waitTime: map[string]time.Duration{},
		mu:       &sync.Mutex{},
		cluster: &ClusterInfo{
			Name:         "test-cluster",
			VersionMajor: "1",
			VersionMinor: "33",
		},
		runID: "test-run-id",
	}
}
