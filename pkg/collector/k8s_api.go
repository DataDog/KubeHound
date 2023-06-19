package collector

import (
	"context"
	"fmt"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
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
	K8sAPICollectorName = "k8s-api-collector"
)

// checkK8sAPICollectorConfig made for unit testing to avoid using NewK8sAPICollector that is bind to a kubernetes config file
func checkK8sAPICollectorConfig(collectorType string) error {
	if collectorType != config.CollectorTypeK8sAPI {
		return fmt.Errorf("invalid collector type in config: %s", collectorType)
	}
	return nil
}

// NewK8sAPICollector creates a new instance of the k8s live API collector from the provided application config.
func NewK8sAPICollector(ctx context.Context, cfg *config.KubehoundConfig) (CollectorClient, error) {
	baseTags = append(baseTags, "collector:k8s-api")
	l := log.Trace(ctx, log.WithComponent(K8sAPICollectorName))

	err := checkK8sAPICollectorConfig(cfg.Collector.Type)
	if err != nil {
		return nil, err
	}

	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("building kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("getting kubernetes config: %w", err)
	}

	return &k8sAPICollector{
		cfg:       cfg.Collector.Live,
		clientset: clientset,
		log:       l,
		rl:        ratelimit.New(cfg.Collector.Live.RateLimitPerSecond), // per second
	}, nil
}

func (c *k8sAPICollector) Name() string {
	return K8sAPICollectorName
}

func (c *k8sAPICollector) HealthCheck(ctx context.Context) (bool, error) {
	c.log.Debugf("Requesting /healthz endpoint")

	rawRes, err := c.clientset.Discovery().RESTClient().Get().AbsPath("/healthz").DoRaw(ctx)
	if err != nil {
		return false, fmt.Errorf("/healthz bad request: %w", err)
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

	return fmt.Errorf("checking namespace %s: %w", namespace, err)
}

func (c *k8sAPICollector) setPagerConfig(pager *pager.ListPager) {
	pager.PageSize = c.cfg.PageSize
	pager.PageBufferSize = c.cfg.PageBufferSize
}

// streamPodsNamespace streams the pod objects corresponding to a cluster namespace.
func (c *k8sAPICollector) streamPodsNamespace(ctx context.Context, namespace string, ingestor PodIngestor) error {
	err := c.checkNamespaceExists(ctx, namespace)
	if err != nil {
		return err
	}

	opts := metav1.ListOptions{}

	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		entries, err := c.clientset.CoreV1().Pods(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s pods for namespace %s: %w", namespace, err)
		}
		return entries, err
	}))

	c.setPagerConfig(pager)

	return pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(MetricCollectorPodsCount, baseTags, 1)
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
		entries, err := c.clientset.RbacV1().Roles(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s roles for namespace %s: %w", namespace, err)
		}
		return entries, err
	}))

	c.setPagerConfig(pager)

	return pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(MetricCollectorRolesCount, baseTags, 1)
		c.rl.Take()
		item := obj.(*rbacv1.Role)
		err := ingestor.IngestRole(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s roles %s for namespace %s: %w", item.Name, namespace, err)
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
		entries, err := c.clientset.RbacV1().RoleBindings(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s rolebinding for namespace %s: %w", namespace, err)
		}
		return entries, err
	}))

	c.setPagerConfig(pager)

	return pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(MetricCollectorRoleBindingsCount, baseTags, 1)
		c.rl.Take()
		item := obj.(*rbacv1.RoleBinding)
		err := ingestor.IngestRoleBinding(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s rolebinding %s for namespace %s: %w", item.Name, namespace, err)
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
		entries, err := c.clientset.CoreV1().Nodes().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s nodes: %w", err)
		}
		return entries, err
	}))

	c.setPagerConfig(pager)

	err := pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(MetricCollectorNodesCount, baseTags, 1)
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
		entries, err := c.clientset.RbacV1().ClusterRoles().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s cluster roles: %w", err)
		}
		return entries, err
	}))

	c.setPagerConfig(pager)

	err := pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(MetricCollectorClusterRolesCount, baseTags, 1)
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
		entries, err := c.clientset.RbacV1().ClusterRoleBindings().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s cluster roles: %w", err)
		}
		return entries, err
	}))

	c.setPagerConfig(pager)

	err := pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(MetricCollectorClusterRoleBindingsCount, baseTags, 1)
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
