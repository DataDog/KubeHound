package collector

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/DataDog/KubeHound/pkg/cmd"
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
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/pager"
	ctrl "sigs.k8s.io/controller-runtime"
)

// FileCollector implements a collector based on local K8s API json files generated outside the KubeHound application via e.g kubectl.
type k8sAPICollector struct {
	clientset           kubernetes.Interface
	rl                  ratelimit.Limiter
	cfg                 *config.K8SAPICollectorConfig
	tags                collectorTags
	waitTime            map[string]time.Duration
	startTime           time.Time
	mu                  *sync.Mutex
	isStreaming         bool
	clusterName         string
	clusterVersionMajor string
	clusterVersionMinor string
	runID               string
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

// NewK8sAPICollector creates a new instance of the k8s live API collector from the provided application config.
func NewK8sAPICollector(ctx context.Context, cfg *config.KubehoundConfig) (CollectorClient, error) {
	ctx = context.WithValue(ctx, log.ContextFieldComponent, K8sAPICollectorName)
	l := log.Trace(ctx)

	clusterName, err := config.GetClusterName(ctx)
	if err != nil {
		return nil, err
	}
	if clusterName == "" {
		return nil, errors.New("cluster name is empty. Did you forget to set `KUBECONFIG` or use `kubectx` to select a cluster?")
	}

	if !cfg.Collector.NonInteractive {
		l.Warn(fmt.Sprintf("About to dump k8s cluster %s - Do you want to continue ? [Yes/No]", clusterName), log.String(log.FieldClusterKey, clusterName))
		proceed, err := cmd.AskForConfirmation(ctx)
		if err != nil {
			return nil, err
		}

		if !proceed {
			return nil, errors.New("user did not confirm")
		}
	} else {
		l.Warn(fmt.Sprintf("Non-interactive mode enabled, proceeding with k8s cluster dump: %s", clusterName), log.String(log.FieldClusterKey, clusterName))
	}

	err = checkK8sAPICollectorConfig(cfg.Collector.Type)
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

	discoveryClient, err := discovery.NewDiscoveryClientForConfig(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("getting discovery client: %w", err)
	}

	serverVersion, err := discoveryClient.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("getting server version: %w", err)
	}

	return &k8sAPICollector{
		cfg:                 cfg.Collector.Live,
		clientset:           clientset,
		rl:                  ratelimit.New(cfg.Collector.Live.RateLimitPerSecond), // per second
		tags:                newCollectorTags(),
		waitTime:            map[string]time.Duration{},
		startTime:           time.Now(),
		mu:                  &sync.Mutex{},
		clusterName:         clusterName,
		clusterVersionMajor: serverVersion.Major,
		clusterVersionMinor: serverVersion.Minor,
		runID:               cfg.Dynamic.RunID.String(),
	}, nil
}

func (c *k8sAPICollector) ComputeMetadata(ctx context.Context, ingestor MetadataIngestor) error {
	metrics, err := c.computeMetrics(ctx)
	if err != nil {
		return fmt.Errorf("error computing metrics: %w", err)
	}

	metadata := Metadata{
		Cluster: ClusterInfo{
			Name:         c.clusterName,
			VersionMajor: c.clusterVersionMajor,
			VersionMinor: c.clusterVersionMinor,
		},
		RunID:   c.runID,
		Metrics: metrics,
	}

	err = ingestor.DumpMetadata(ctx, metadata)
	if err != nil {
		return fmt.Errorf("ingesting metadata: %w", err)
	}

	return nil
}

func (c *k8sAPICollector) wait(ctx context.Context, resourceType string, tags []string) {
	l := log.Logger(ctx)
	c.mu.Lock()
	prev := time.Now()
	now := c.rl.Take()

	waitTime := now.Sub(prev)
	defer c.mu.Unlock()
	c.waitTime[resourceType] += waitTime

	// Display a message to tell the user the streaming has started (only once after the approval has been made)
	if !c.isStreaming {
		l.Info("Streaming data from the K8s API")
		c.isStreaming = true
	}

	// entity := tag.Entity(resourceType)
	err := statsd.Gauge(ctx, metric.CollectorWait, float64(c.waitTime[resourceType]), tags, 1)
	if err != nil {
		l.Error("could not send gauge", log.ErrorField(err))
	}
}

func (c *k8sAPICollector) waitTimeByResource(ctx context.Context, resourceType string, span ddtrace.Span) {
	l := log.Logger(ctx)
	c.mu.Lock()
	defer c.mu.Unlock()

	waitTime := c.waitTime[resourceType]
	span.SetTag(tag.WaitTag, waitTime)
	l.Debugf("Wait time for %s: %s", resourceType, waitTime)
}

func (c *k8sAPICollector) Name() string {
	return K8sAPICollectorName
}

func (c *k8sAPICollector) HealthCheck(ctx context.Context) (bool, error) {
	l := log.Logger(ctx)
	l.Debug("Requesting /healthz endpoint")

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

func (c *k8sAPICollector) ClusterInfo(ctx context.Context) (*ClusterInfo, error) {
	// The config cluster info does not contain the version info as it is populated at collection time
	cfgClusterInfo, err := config.NewClusterInfo(ctx)
	if err != nil {
		return nil, err
	}

	return &ClusterInfo{
		Name:         cfgClusterInfo.Name,
		VersionMajor: c.clusterVersionMajor,
		VersionMinor: c.clusterVersionMinor,
	}, nil
}

// Generate metrics for k8sAPI collector
func (c *k8sAPICollector) computeMetrics(ctx context.Context) (Metrics, error) {
	l := log.Logger(ctx)
	var errMetric error
	var runTotalWaitTime time.Duration
	for _, wait := range c.waitTime {
		runTotalWaitTime += wait
	}

	runDuration := time.Since(c.startTime)
	err := statsd.Gauge(ctx, metric.CollectorRunWait, float64(runTotalWaitTime), c.tags.baseTags, 1)
	if err != nil {
		errMetric = errors.Join(errMetric, err)
		l.Error("could not send gauge", log.ErrorField(err))
	}
	err = statsd.Gauge(ctx, metric.CollectorRunDuration, float64(runDuration), c.tags.baseTags, 1)
	if err != nil {
		errMetric = errors.Join(errMetric, err)
		l.Error("could not send gauge", log.ErrorField(err))
	}

	runThrottlingPercentage := 1 - (float64(runDuration-runTotalWaitTime) / float64(runDuration))
	err = statsd.Gauge(ctx, metric.CollectorRunThrottling, runThrottlingPercentage, c.tags.baseTags, 1)
	if err != nil {
		errMetric = errors.Join(errMetric, err)
		l.Error("could not send gauge", log.ErrorField(err))
	}
	l.Info("Stats for the run time duration", log.Dur("run", runDuration), log.Dur("wait", runTotalWaitTime), log.Percent("throttling_percent", 100*runThrottlingPercentage, 100))

	// SaveMetadata
	metadata := Metrics{
		DumpTime:             time.Now(),
		RunDuration:          runDuration,
		TotalWaitTime:        runTotalWaitTime,
		ThrottlingPercentage: runThrottlingPercentage,
	}

	return metadata, errMetric
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
		_ = statsd.Incr(ctx, metric.CollectorCount, c.tags.pod, 1)
		c.wait(ctx, entity, c.tags.pod)
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
	span, ctx := span.SpanRunFromContext(ctx, span.CollectorStream)
	span.SetTag(tag.EntityTag, entity)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	// passing an empty namespace will collect all namespaces
	err = c.streamPodsNamespace(ctx, "", ingestor)
	if err != nil {
		return err
	}

	c.waitTimeByResource(ctx, entity, span)

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
		_ = statsd.Incr(ctx, metric.CollectorCount, c.tags.role, 1)
		c.wait(ctx, entity, c.tags.role)
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
	span, ctx := span.SpanRunFromContext(ctx, span.CollectorStream)
	span.SetTag(tag.EntityTag, entity)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	// passing an empty namespace will collect all namespaces
	err = c.streamRolesNamespace(ctx, "", ingestor)
	if err != nil {
		return err
	}

	c.waitTimeByResource(ctx, entity, span)

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
		_ = statsd.Incr(ctx, metric.CollectorCount, c.tags.rolebinding, 1)
		c.wait(ctx, entity, c.tags.rolebinding)
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
	span, ctx := span.SpanRunFromContext(ctx, span.CollectorStream)
	span.SetTag(tag.EntityTag, entity)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	// passing an empty namespace will collect all namespaces
	err = c.streamRoleBindingsNamespace(ctx, "", ingestor)
	if err != nil {
		return err
	}

	c.waitTimeByResource(ctx, entity, span)

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
		_ = statsd.Incr(ctx, metric.CollectorCount, c.tags.endpoint, 1)
		c.wait(ctx, entity, c.tags.endpoint)
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
	span, ctx := span.SpanRunFromContext(ctx, span.CollectorStream)
	span.SetTag(tag.EntityTag, tag.EntityEndpoints)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	// passing an empty namespace will collect all namespaces
	err = c.streamEndpointsNamespace(ctx, "", ingestor)
	if err != nil {
		return err
	}

	c.waitTimeByResource(ctx, entity, span)

	return ingestor.Complete(ctx)
}

func (c *k8sAPICollector) StreamNodes(ctx context.Context, ingestor NodeIngestor) error {
	entity := tag.EntityNodes
	span, ctx := span.SpanRunFromContext(ctx, span.CollectorStream)
	span.SetTag(tag.EntityTag, entity)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	opts := tunedListOptions()
	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		entries, err := c.clientset.CoreV1().Nodes().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s nodes: %w", err)
		}

		return entries, err
	}))

	c.setPagerConfig(pager)

	err = pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(ctx, metric.CollectorCount, c.tags.node, 1)
		c.wait(ctx, entity, c.tags.node)
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

	c.waitTimeByResource(ctx, entity, span)

	return ingestor.Complete(ctx)
}

func (c *k8sAPICollector) StreamClusterRoles(ctx context.Context, ingestor ClusterRoleIngestor) error {
	entity := tag.EntityClusterRoles
	span, ctx := span.SpanRunFromContext(ctx, span.CollectorStream)
	span.SetTag(tag.EntityTag, entity)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	opts := tunedListOptions()
	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		entries, err := c.clientset.RbacV1().ClusterRoles().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s cluster roles: %w", err)
		}

		return entries, err
	}))

	c.setPagerConfig(pager)

	err = pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(ctx, metric.CollectorCount, c.tags.clusterrole, 1)
		c.wait(ctx, entity, c.tags.clusterrole)
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

	c.waitTimeByResource(ctx, entity, span)

	return ingestor.Complete(ctx)
}

func (c *k8sAPICollector) StreamClusterRoleBindings(ctx context.Context, ingestor ClusterRoleBindingIngestor) error {
	entity := tag.EntityClusterRolebindings
	span, ctx := span.SpanRunFromContext(ctx, span.CollectorStream)
	span.SetTag(tag.EntityTag, entity)
	var err error
	defer func() { span.Finish(tracer.WithError(err)) }()

	opts := tunedListOptions()
	pager := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		entries, err := c.clientset.RbacV1().ClusterRoleBindings().List(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("getting K8s cluster roles: %w", err)
		}

		return entries, err
	}))

	c.setPagerConfig(pager)

	err = pager.EachListItem(ctx, opts, func(obj runtime.Object) error {
		_ = statsd.Incr(ctx, metric.CollectorCount, c.tags.clusterrolebinding, 1)
		c.wait(ctx, entity, c.tags.clusterrolebinding)
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

	c.waitTimeByResource(ctx, entity, span)

	return ingestor.Complete(ctx)
}
