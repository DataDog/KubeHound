package metric

// Collector metrics
var (
	CollectorCount = "kubehound.collector.count"
	CollectorWait  = "kubehound.collector.wait"
	CollectorSkip  = "kubehound.collector.skip"
)

// Pipeline storage metrics
var (
	ObjectWrite          = "kubehound.storage.batchwrite.object"
	VertexWrite          = "kubehound.storage.batchwrite.vertex"
	EdgeWrite            = "kubehound.storage.batchwrite.edge"
	QueueSize            = "kubehound.storage.queue.size"
	BackgroundWriterCall = "kubehound.storage.writer.background"
	FlushWriterCall      = "kubehound.storage.writer.flush"
)

// Cache metrics
var (
	CacheMiss      = "kubehound.cache.miss"
	CacheHit       = "kubehound.cache.hit"
	CacheWrite     = "kubehound.cache.write"
	CacheDuplicate = "kubehound.cache.duplicate"
)
