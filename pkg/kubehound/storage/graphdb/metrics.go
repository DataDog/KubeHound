package graphdb

var (
	baseTags = []string{}
)

var (
	MetricBackgroundWriterCall = "kubehound.storage.graphdb.background"
	MetricBatchWrite           = "kubehound.storage.graphdb.batchwrite.size"
	MetricQueueSize            = "kubehound.storage.graphdb.queue.size"
)

var (
	SpanOperationFlush      = "kubehound.janusgraph.flush"
	SpanOperationBatchWrite = "kubehound.janusgraph.batchwrite"
)
