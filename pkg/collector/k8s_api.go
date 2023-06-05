package collector

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/globals"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/pager"
	ctrl "sigs.k8s.io/controller-runtime"

	"go.uber.org/ratelimit"
)

// FileCollector implements a collector based on local K8s API json files generated outside the KubeHound application via e.g kubectl.
type k8sAPICollector struct {
	clientset kubernetes.Interface
	log       *log.KubehoundLogger
	rl        ratelimit.Limiter
	cfg       *config.K8SAPICollectorConfig
}

const (
	K8sAPICollectorName               = "k8s-api-collector"
	K8sAPIDefaultPageSize       int64 = 5000
	K8sAPIDefaultPageBufferSize int32 = 10
	K8sAPIRateLimitPerSecond    int   = 100
)

// NewK8sAPICollector creates a new instance of the k8s live API collector from the provided application config.
func NewK8sAPICollector(ctx context.Context, cfg *config.KubehoundConfig) (CollectorClient, error) {
	l := log.Trace(ctx, log.WithComponent(K8sAPICollectorName))

	if cfg.Collector.Type != config.CollectorTypeK8sAPI {
		return nil, fmt.Errorf("invalid collector type in config: %s", cfg.Collector.Type)
	}

	if cfg.Collector.Live == nil {
		cfg.Collector.Live = &config.K8SAPICollectorConfig{}
	}

	if cfg.Collector.Live.PageSize == nil {
		cfg.Collector.Live.PageSize = globals.Ptr(K8sAPIDefaultPageSize)
		l.Warnf("setting PageSize with default value: %d", K8sAPIDefaultPageSize)
	}

	if cfg.Collector.Live.PageBufferSize == nil {
		cfg.Collector.Live.PageBufferSize = globals.Ptr(K8sAPIDefaultPageBufferSize)
		l.Warnf("setting PageBufferSize with default value: %d", K8sAPIDefaultPageBufferSize)
	}

	if cfg.Collector.Live.RateLimitPerSecond == nil {
		cfg.Collector.Live.RateLimitPerSecond = globals.Ptr(K8sAPIRateLimitPerSecond)
		l.Warnf("setting RateLimitPerSecond with default value: %d", K8sAPIRateLimitPerSecond)
	}

	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("building kubernetes config: %v", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("getting kubernetes config: %v", err)
	}

	return &k8sAPICollector{
		cfg:       cfg.Collector.Live,
		clientset: clientset,
		log:       l,
		rl:        ratelimit.New(*cfg.Collector.Live.RateLimitPerSecond), // per second
	}, nil
}

func (c *k8sAPICollector) Name() string {
	return K8sAPICollectorName
}

func (c *k8sAPICollector) HealthCheck(ctx context.Context) (bool, error) {
	c.log.Debugf("Requestin /healthz endpoint")

	rawRes, err := c.clientset.Discovery().RESTClient().Get().AbsPath("/healthz").DoRaw(ctx)
	if err != nil {
		return false, fmt.Errorf("/healthz bad request: %s", err.Error())
	}

	res := string(rawRes)
	if res != "ok" {
		return false, fmt.Errorf("/healthz request not ok: response: %s", res)
	}

	return true, nil
}

func (c *k8sAPICollector) Close(ctx context.Context) error {
	return nil
}

// checkNamespaceExists checks if a namespace exists
func (c *k8sAPICollector) checkNamespaceExists(ctx context.Context, namespace string) error {
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})

	if err == nil || namespace == "" {
		return nil
	}

	if errors.IsNotFound(err) {
		return fmt.Errorf("namespace %s does not exist", namespace)
	}

	return fmt.Errorf("checking namespace %s: %v", namespace, err)
}

func (c *k8sAPICollector) setPagerConfig(pager *pager.ListPager) {
	pager.PageSize = *c.cfg.PageSize
	pager.PageBufferSize = *c.cfg.PageBufferSize
}

// streamPodsNamespace streams the pod objects corresponding to a cluster namespace.
func (c *k8sAPICollector) streamPodsNamespace(ctx context.Context, namespace string, ingestor PodIngestor) error {
	err := c.checkNamespaceExists(ctx, namespace)
	if err != nil {
		return err
	}

	opts := metav1.ListOptions{}

	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		pods, err := c.clientset.CoreV1().Pods(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s pods for namespace %s: %v", namespace, err)
		}
		return pods, err
	}))

	c.setPagerConfig(pager)

	return pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		c.rl.Take()
		item := obj.(*corev1.Pod)
		err := ingestor.IngestPod(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s pod %s for namespace %s: %w", item.Name, namespace, err)
		}
		return nil
	})
}

func (c *k8sAPICollector) StreamPods(ctx context.Context, ingestor PodIngestor) error {
	// passing an empty namespace will collect all namespaces
	err := c.streamPodsNamespace(ctx, "", ingestor)
	if err != nil {
		return err
	}
	return ingestor.Complete(ctx)
}

// streamRolesNamespace streams the role objects corresponding to a cluster namespace.
func (c *k8sAPICollector) streamRolesNamespace(ctx context.Context, namespace string, ingestor RoleIngestor) error {
	err := c.checkNamespaceExists(ctx, namespace)
	if err != nil {
		return err
	}

	opts := metav1.ListOptions{}

	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		pods, err := c.clientset.RbacV1().Roles(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s roles for namespace %s: %v", namespace, err)
		}
		return pods, err
	}))

	c.setPagerConfig(pager)

	return pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		c.rl.Take()
		item := obj.(*rbacv1.Role)
		err := ingestor.IngestRole(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s pod %s for namespace %s: %w", item.Name, namespace, err)
		}
		return nil
	})
}

func (c *k8sAPICollector) StreamRoles(ctx context.Context, ingestor RoleIngestor) error {
	// passing an empty namespace will collect all namespaces
	err := c.streamRolesNamespace(ctx, "", ingestor)
	if err != nil {
		return err
	}
	return ingestor.Complete(ctx)
}

// streamRoleBindingsNamespace streams the role bindings objects corresponding to a cluster namespace.
func (c *k8sAPICollector) streamRoleBindingsNamespace(ctx context.Context, namespace string, ingestor RoleBindingIngestor) error {
	err := c.checkNamespaceExists(ctx, namespace)
	if err != nil {
		return err
	}

	opts := metav1.ListOptions{}

	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		pods, err := c.clientset.RbacV1().RoleBindings(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s rolebinding for namespace %s: %v", namespace, err)
		}
		return pods, err
	}))

	c.setPagerConfig(pager)

	return pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		c.rl.Take()
		item := obj.(*rbacv1.RoleBinding)
		err := ingestor.IngestRoleBinding(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s role binding %s for namespace %s: %w", item.Name, namespace, err)
		}
		return nil
	})
}

func (c *k8sAPICollector) StreamRoleBindings(ctx context.Context, ingestor RoleBindingIngestor) error {
	// passing an empty namespace will collect all namespaces
	err := c.streamRoleBindingsNamespace(ctx, "", ingestor)
	if err != nil {
		return err
	}
	return ingestor.Complete(ctx)
}

func (c *k8sAPICollector) StreamNodes(ctx context.Context, ingestor NodeIngestor) error {
	opts := metav1.ListOptions{}

	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		pods, err := c.clientset.CoreV1().Nodes().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s nodes: %v", err)
		}
		return pods, err
	}))

	c.setPagerConfig(pager)

	err := pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		c.rl.Take()
		item := obj.(*corev1.Node)
		err := ingestor.IngestNode(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s node %s: %w", item.Name, err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return ingestor.Complete(ctx)
}

func (c *k8sAPICollector) StreamClusterRoles(ctx context.Context, ingestor ClusterRoleIngestor) error {
	opts := metav1.ListOptions{}

	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		pods, err := c.clientset.RbacV1().ClusterRoles().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s cluster roles: %v", err)
		}
		return pods, err
	}))

	c.setPagerConfig(pager)

	err := pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		c.rl.Take()
		item := obj.(*rbacv1.ClusterRole)
		err := ingestor.IngestClusterRole(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s cluster role %s: %w", item.Name, err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return ingestor.Complete(ctx)
}

func (c *k8sAPICollector) StreamClusterRoleBindings(ctx context.Context, ingestor ClusterRoleBindingIngestor) error {
	opts := metav1.ListOptions{}

	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		pods, err := c.clientset.RbacV1().ClusterRoleBindings().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s cluster roles: %v", err)
		}
		return pods, err
	}))

	c.setPagerConfig(pager)

	err := pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		c.rl.Take()
		item := obj.(*rbacv1.ClusterRoleBinding)
		err := ingestor.IngestClusterRoleBinding(ctx, obj.(*rbacv1.ClusterRoleBinding))
		if err != nil {
			return fmt.Errorf("processing K8s cluster role binding %s: %w", item.Name, err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return ingestor.Complete(ctx)
}
