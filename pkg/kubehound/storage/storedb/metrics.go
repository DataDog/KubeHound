package storedb

var (
	baseTags = []string{}
)

var (
	MetricBackgroundWriterCall = "kubehound.storage.storedb.background"
	MetricBatchWrite           = "kubehound.storage.storedb.batchwrite.size"
	MetricQueueSize            = "kubehound.storage.storedb.queue.size"
)
