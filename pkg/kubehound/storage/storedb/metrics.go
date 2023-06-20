package storedb

var (
	baseTags = []string{
		"database:mongodb",
	}
)

var (
	MetricBackgroundWriterCall = "kubehound.storage.storedb.background"
	MetricBatchWrite           = "kubehound.storage.storedb.batchwrite.size"
	MetricQueueSize            = "kubehound.storage.storedb.queue.size"
)

var (
	SpanOperationFlush      = "kubehound.mongo.flush"
	SpanOperationBatchWrite = "kubehound.mongo.batchwrite"
)
