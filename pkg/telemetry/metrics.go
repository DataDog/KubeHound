package telemetry

import (
	"fmt"
)

const (
	Prefix = "kubehound"
)

func metricFQN(name string) string {
	return fmt.Sprintf("%s.%s", Prefix, name)
}

// Collector metrics
var (
	CollectorCountMetric = metricFQN("collector.count")
)

// Pipeline storage metrics

const (
	MetricCollectorNodesCount               = "kubehound.collector.nodes.count"
	MetricCollectorEndpointCount            = "kubehound.collector.endpoints.count"
	MetricCollectorPodsCount                = "kubehound.collector.pods.count"
	MetricCollectorRolesCount               = "kubehound.collector.roles.count"
	MetricCollectorRoleBindingsCount        = "kubehound.collector.rolebindings.count"
	MetricCollectorClusterRolesCount        = "kubehound.collector.clusterroles.count"
	MetricCollectorClusterRoleBindingsCount = "kubehound.collector.clusterrolebindings.count"

	MetricStoredbBackgroundWriterCall = "kubehound.storage.storedb.background"
	MetricStoredbBatchWrite           = "kubehound.storage.storedb.batchwrite.size"
	MetricStoredbQueueSize            = "kubehound.storage.storedb.queue.size"

	MetricGraphdbBackgroundWriterCall = "kubehound.storage.graphdb.background"
	MetricGraphdbBatchWrite           = "kubehound.storage.graphdb.batchwrite.size"
	MetricGraphdbQueueSize            = "kubehound.storage.graphdb.queue.size"

	MetricCacheMiss           = "kubehound.cache.miss"
	MetricCacheHit            = "kubehound.cache.hit"
	MetricCacheWrite          = "kubehound.cache.write"
	MetricCacheDuplicateEntry = "kubehound.cache.duplicate"
)
