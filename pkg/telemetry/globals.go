package telemetry

var (
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

var (
	MetricCollectorNodesCount               = "kubehound.collector.nodes.count"
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

var (
	TagTypeJanusGraph = "type:janusgraph"
	TagTypeMongodb    = "type:mongodb"
	// BaseTags represents the minimal tags sent by the application
	// Each sub-component of the app will add to their local usage their own tags depending on their needs.
	BaseTags = []string{}
)
