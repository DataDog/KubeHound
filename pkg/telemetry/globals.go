package telemetry

const (
	SpanJanusGraphOperationFlush      = "kubehound.janusgraph.flush"
	SpanJanusGraphOperationBatchWrite = "kubehound.janusgraph.batchwrite"

	SpanMongodbOperationFlush      = "kubehound.mongo.flush"
	SpanMongodbOperationBatchWrite = "kubehound.mongo.batchwrite"

	SpanOperationIngestData = "kubehound.ingestData"
	SpanOperationBuildGraph = "kubehound.buildGraph"
	SpanOperationLaunch     = "kubehound.launch"

	SpanOperationStream   = "kubehound.collector.stream"
	SpanOperationReadFile = "kubehound.collector.readFile"

	SpanOperationRun = "kubehound.graph.builder.run"
)

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

const (
	TagTypeJanusGraph      = "type:janusgraph"
	TagTypeMongodb         = "type:mongodb"
	TagCollectorTypeK8sApi = "collector:k8s-api"
	TagCollectorTypeFile   = "collector:file"

	TagKeyResource = "resource"
	TagKeyLabel    = "label"
	TagKeyRunId    = "run_id"

	TagResourcePods                = "pods"
	TagResourceRoles               = "roles"
	TagResourceRolebindings        = "rolebindings"
	TagResourceNodes               = "nodes"
	TagResourceEndpoints           = "endpoints"
	TagResourceClusterRoles        = "clusterroles"
	TagResourceClusterRolebindings = "clusterrolebindings"
	// BaseTags represents the minimal tags sent by the application
	// Each sub-component of the app will add to their local usage their own tags depending on their needs.
)

var (
	BaseTags = []string{}
)
