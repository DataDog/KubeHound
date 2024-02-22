package collector

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/DataDog/KubeHound/pkg/config"
	"github.com/DataDog/KubeHound/pkg/telemetry/log"
	"github.com/DataDog/KubeHound/pkg/telemetry/metric"
	"github.com/DataDog/KubeHound/pkg/telemetry/span"
	"github.com/DataDog/KubeHound/pkg/telemetry/statsd"
	"github.com/DataDog/KubeHound/pkg/telemetry/tag"

	"go.uber.org/ratelimit"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/pager"
	ctrl "sigs.k8s.io/controller-runtime"
)

// FileCollector implements a collector based on local K8s API json files generated outside the KubeHound application via e.g kubectl.
type k8sAPICollector struct {
	clientset kubernetes.Interface
	log       *log.KubehoundLogger
	rl        ratelimit.Limiter
	cfg       *config.K8SAPICollectorConfig
	tags      []string
	waitTime  map[string]time.Duration
	startTime time.Time
	mu        *sync.Mutex
}

const (
	K8sAPICollectorName = "k8s-api-collector"
)

var (
	CollectorUserAgent = fmt.Sprintf("KubeHound-Collector-v%s", config.BuildVersion)
)

// checkK8sAPICollectorConfig made for unit testing to avoid using NewK8sAPICollector that is bind to a kubernetes config file
func checkK8sAPICollectorConfig(collectorType string) error {
	if collectorType != config.CollectorTypeK8sAPI {
		return fmt.Errorf("invalid collector type in config: %s", collectorType)
	}

	return nil
}

func tunedListOptions() metav1.ListOptions {
	// Optimized for speed. See: https://blog.palark.com/kubernetes-api-list-case-troubleshooting/
	return metav1.ListOptions{
		ResourceVersion:      "0",
		ResourceVersionMatch: metav1.ResourceVersionMatchNotOlderThan,
	}
}

func newClusterInfo(_ context.Context) (*ClusterInfo, error) {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	raw, err := kubeConfig.RawConfig()
	if err != nil {
		return nil, fmt.Errorf("raw config get: %w", err)
	}

	return &ClusterInfo{
		Name: raw.CurrentContext,
	}, nil
}

func GetClusterName(ctx context.Context) (string, error) {
	cluster, err := newClusterInfo(ctx)
	if err != nil {
		return "", fmt.Errorf("collector cluster info: %w", err)
	}

	return cluster.Name, nil
}

// NewK8sAPICollector creates a new instance of the k8s live API collector from the provided application config.
func NewK8sAPICollector(ctx context.Context, cfg *config.KubehoundConfig) (CollectorClient, error) {
	tags := tag.BaseTags
	tags = append(tags, tag.Collector(K8sAPICollectorName))
	l := log.Trace(ctx, log.WithComponent(K8sAPICollectorName))

	err := checkK8sAPICollectorConfig(cfg.Collector.Type)
	if err != nil {
		return nil, err
	}

	kubeConfig, err := ctrl.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("building kubernetes config: %w", err)
	}

	kubeConfig.UserAgent = CollectorUserAgent

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("getting kubernetes config: %w", err)
	}

	return &k8sAPICollector{
		cfg:       cfg.Collector.Live,
		clientset: clientset,
		log:       l,
		rl:        ratelimit.New(cfg.Collector.Live.RateLimitPerSecond), // per second
		tags:      tags,
		waitTime:  map[string]time.Duration{},
		startTime: time.Now(),
		mu:        &sync.Mutex{},
	}, nil
}

func (c *k8sAPICollector) wait(_ context.Context, resourceType string) {
	prev := time.Now()
	now := c.rl.Take()

	waitTime := now.Sub(prev)
	c.mu.Lock()
	defer c.mu.Unlock()
	c.waitTime[resourceType] += waitTime

	entity := tag.Entity(resourceType)
	err := statsd.Gauge(metric.CollectorWait, float64(c.waitTime[resourceType]), append(c.tags, entity), 1)
	if err != nil {
		log.I.Error(err)
	}
}

func (c *k8sAPICollector) waitTimeByResource(resourceType string, span ddtrace.Span) {
	c.mu.Lock()
	defer c.mu.Unlock()

	waitTime := c.waitTime[resourceType]
	span.SetTag(tag.WaitTag, waitTime)
	log.I.Debugf("Wait time for %s: %s", resourceType, waitTime)
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

func (c *k8sAPICollector) Tags(ctx context.Context) []string {
	return c.tags
}

func (c *k8sAPICollector) ClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	return newClusterInfo(ctx)
}

// Generate metrics for k8sAPI collector
func (c *k8sAPICollector) computeMetrics(_ context.Context) error {
	var errMetric error
	var runTotalWaitTime time.Duration
	for _, wait := range c.waitTime {
		runTotalWaitTime += wait
	}

	runDuration := time.Since(c.startTime)
	err := statsd.Gauge(metric.CollectorRunWait, float64(runTotalWaitTime), c.tags, 1)
	if err != nil {
		errMetric = errors.Join(errMetric, err)
		log.I.Error(err)
	}
	err = statsd.Gauge(metric.CollectorRunDuration, float64(runDuration), c.tags, 1)
	if err != nil {
		errMetric = errors.Join(errMetric, err)
		log.I.Error(err)
	}

	runThrottlingPercentage := 1 - (float64(runDuration-runTotalWaitTime) / float64(runDuration))
	err = statsd.Gauge(metric.CollectorRunThrottling, runThrottlingPercentage, c.tags, 1)
	if err != nil {
		errMetric = errors.Join(errMetric, err)
		log.I.Error(err)
	}
	log.I.Infof("Stats for the run time duration: %s / wait: %s / throttling: %f%%", runDuration, runTotalWaitTime, 100*runThrottlingPercentage) //nolint:gomnd

	return errMetric
}

func (c *k8sAPICollector) Close(ctx context.Context) error {
	err := c.computeMetrics(ctx)
	if err != nil {
		// We don't want to return an error here as it is just metrics and won't affect the collection of data
		log.I.Errorf("Error computing metrics: %s", err)
	}

	return nil
}

// checkNamespaceExists checks if a namespace exists
func (c *k8sAPICollector) checkNamespaceExists(ctx context.Context, namespace string) error {
	_, err := c.clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})

	if err == nil || namespace == "" {
		return nil
	}

	if kerrors.IsNotFound(err) {
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
	entity := tag.EntityPods
	err := c.checkNamespaceExists(ctx, namespace)
	if err != nil {
		return err
	}

	opts := tunedListOptions()
	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		entries, err := c.clientset.CoreV1().Pods(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s pods for namespace %s: %w", namespace, err)
		}

		return entries, err
	}))

	c.setPagerConfig(pager)

	return pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(metric.CollectorCount, append(c.tags, tag.Entity(entity)), 1)
		c.wait(ctx, entity)
		item, ok := obj.(*corev1.Pod)
		if !ok {
			return fmt.Errorf("pod stream type conversion error: %T", obj)
		}

		err := ingestor.IngestPod(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s pod %s for namespace %s: %w", item.Name, namespace, err)
		}

		return nil
	})
}

func (c *k8sAPICollector) StreamPods(ctx context.Context, ingestor PodIngestor) error {
	entity := tag.EntityPods
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, entity)
	defer span.Finish()

	// passing an empty namespace will collect all namespaces
	err := c.streamPodsNamespace(ctx, "", ingestor)
	if err != nil {
		return err
	}

	c.waitTimeByResource(entity, span)

	return ingestor.Complete(ctx)
}

// streamRolesNamespace streams the role objects corresponding to a cluster namespace.
func (c *k8sAPICollector) streamRolesNamespace(ctx context.Context, namespace string, ingestor RoleIngestor) error {
	entity := tag.EntityRoles
	err := c.checkNamespaceExists(ctx, namespace)
	if err != nil {
		return err
	}

	opts := tunedListOptions()
	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		entries, err := c.clientset.RbacV1().Roles(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s roles for namespace %s: %w", namespace, err)
		}

		return entries, err
	}))

	c.setPagerConfig(pager)

	return pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(metric.CollectorCount, append(c.tags, tag.Entity(entity)), 1)
		c.wait(ctx, entity)
		item, ok := obj.(*rbacv1.Role)
		if !ok {
			return fmt.Errorf("role stream type conversion error: %T", obj)
		}

		err := ingestor.IngestRole(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s roles %s for namespace %s: %w", item.Name, namespace, err)
		}

		return nil
	})
}

func (c *k8sAPICollector) StreamRoles(ctx context.Context, ingestor RoleIngestor) error {
	entity := tag.EntityRoles
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, entity)
	defer span.Finish()

	// passing an empty namespace will collect all namespaces
	err := c.streamRolesNamespace(ctx, "", ingestor)
	if err != nil {
		return err
	}

	c.waitTimeByResource(entity, span)

	return ingestor.Complete(ctx)
}

// streamRoleBindingsNamespace streams the role bindings objects corresponding to a cluster namespace.
func (c *k8sAPICollector) streamRoleBindingsNamespace(ctx context.Context, namespace string, ingestor RoleBindingIngestor) error {
	entity := tag.EntityRolebindings
	err := c.checkNamespaceExists(ctx, namespace)
	if err != nil {
		return err
	}

	opts := tunedListOptions()
	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		entries, err := c.clientset.RbacV1().RoleBindings(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s rolebinding for namespace %s: %w", namespace, err)
		}

		return entries, err
	}))

	c.setPagerConfig(pager)

	return pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(metric.CollectorCount, append(c.tags, tag.Entity(entity)), 1)
		c.wait(ctx, entity)
		item, ok := obj.(*rbacv1.RoleBinding)
		if !ok {
			return fmt.Errorf("role binding stream type conversion error: %T", obj)
		}

		err := ingestor.IngestRoleBinding(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s rolebinding %s for namespace %s: %w", item.Name, namespace, err)
		}

		return nil
	})
}

func (c *k8sAPICollector) StreamRoleBindings(ctx context.Context, ingestor RoleBindingIngestor) error {
	entity := tag.EntityRolebindings
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, entity)
	defer span.Finish()

	// passing an empty namespace will collect all namespaces
	err := c.streamRoleBindingsNamespace(ctx, "", ingestor)
	if err != nil {
		return err
	}

	c.waitTimeByResource(entity, span)

	return ingestor.Complete(ctx)
}

// streamEndpointsNamespace streams the endpoint slice objects corresponding to a cluster namespace.
func (c *k8sAPICollector) streamEndpointsNamespace(ctx context.Context, namespace string, ingestor EndpointIngestor) error {
	entity := tag.EntityEndpoints
	err := c.checkNamespaceExists(ctx, namespace)
	if err != nil {
		return err
	}

	opts := tunedListOptions()
	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		entries, err := c.clientset.DiscoveryV1().EndpointSlices(namespace).List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s endpoint slices for namespace %s: %w", namespace, err)
		}

		return entries, err
	}))

	c.setPagerConfig(pager)

	return pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(metric.CollectorCount, append(c.tags, tag.Entity(entity)), 1)
		c.wait(ctx, entity)
		item, ok := obj.(*discoveryv1.EndpointSlice)
		if !ok {
			return fmt.Errorf("endpoint stream type conversion error: %T", obj)
		}

		err := ingestor.IngestEndpoint(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s endpoint slice %s for namespace %s: %w", item.Name, namespace, err)
		}

		return nil
	})
}

func (c *k8sAPICollector) StreamEndpoints(ctx context.Context, ingestor EndpointIngestor) error {
	entity := tag.EntityEndpoints
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, tag.EntityEndpoints)
	defer span.Finish()

	// passing an empty namespace will collect all namespaces
	err := c.streamEndpointsNamespace(ctx, "", ingestor)
	if err != nil {
		return err
	}

	c.waitTimeByResource(entity, span)

	return ingestor.Complete(ctx)
}

func (c *k8sAPICollector) StreamNodes(ctx context.Context, ingestor NodeIngestor) error {
	entity := tag.EntityNodes
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, entity)
	defer span.Finish()

	opts := tunedListOptions()
	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		entries, err := c.clientset.CoreV1().Nodes().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s nodes: %w", err)
		}

		return entries, err
	}))

	c.setPagerConfig(pager)

	err := pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(metric.CollectorCount, append(c.tags, tag.Entity(entity)), 1)
		c.wait(ctx, entity)
		item, ok := obj.(*corev1.Node)
		if !ok {
			return fmt.Errorf("node stream type conversion error: %T", obj)
		}

		err := ingestor.IngestNode(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s node %s: %w", item.Name, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	c.waitTimeByResource(entity, span)

	return ingestor.Complete(ctx)
}

func (c *k8sAPICollector) StreamClusterRoles(ctx context.Context, ingestor ClusterRoleIngestor) error {
	entity := tag.EntityClusterRoles
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, entity)
	defer span.Finish()

	opts := tunedListOptions()
	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		entries, err := c.clientset.RbacV1().ClusterRoles().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s cluster roles: %w", err)
		}

		return entries, err
	}))

	c.setPagerConfig(pager)

	err := pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(metric.CollectorCount, append(c.tags, tag.Entity(entity)), 1)
		c.wait(ctx, entity)
		item, ok := obj.(*rbacv1.ClusterRole)
		if !ok {
			return fmt.Errorf("cluster role stream type conversion error: %T", obj)
		}

		err := ingestor.IngestClusterRole(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s cluster role %s: %w", item.Name, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	c.waitTimeByResource(entity, span)

	return ingestor.Complete(ctx)
}

func (c *k8sAPICollector) StreamClusterRoleBindings(ctx context.Context, ingestor ClusterRoleBindingIngestor) error {
	entity := tag.EntityClusterRolebindings
	span, ctx := tracer.StartSpanFromContext(ctx, span.CollectorStream, tracer.Measured())
	span.SetTag(tag.EntityTag, entity)
	defer span.Finish()

	opts := tunedListOptions()
	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		entries, err := c.clientset.RbacV1().ClusterRoleBindings().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s cluster roles: %w", err)
		}

		return entries, err
	}))

	c.setPagerConfig(pager)

	err := pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(metric.CollectorCount, append(c.tags, tag.Entity(entity)), 1)
		c.wait(ctx, entity)
		item, ok := obj.(*rbacv1.ClusterRoleBinding)
		if !ok {
			return fmt.Errorf("cluster role binding stream type conversion error: %T", obj)
		}

		err := ingestor.IngestClusterRoleBinding(ctx, item)
		if err != nil {
			return fmt.Errorf("processing K8s cluster role binding %s: %w", item.Name, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	c.waitTimeByResource(entity, span)

	return ingestor.Complete(ctx)
}
